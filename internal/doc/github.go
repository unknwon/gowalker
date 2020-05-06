// Copyright 2015 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package doc

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/unknwon/com"
	log "gopkg.in/clog.v1"

	"github.com/unknwon/gowalker/internal/base"
	"github.com/unknwon/gowalker/internal/db"
	"github.com/unknwon/gowalker/internal/httplib"
	"github.com/unknwon/gowalker/internal/setting"
)

var (
	githubRawHeader       = http.Header{"Accept": {"application/vnd.github-blob.raw"}}
	githubRevisionPattern = regexp.MustCompile(`value="[a-z0-9A-Z]+"`)
	githubPattern         = regexp.MustCompile(`^github\.com/(?P<owner>[a-z0-9A-Z_.\-]+)/(?P<repo>[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]*)?$`)
)

func getGithubRevision(importPath, tag string) (string, error) {
	data, err := com.HttpGetBytes(Client, fmt.Sprintf("https://%s/commits/"+tag, importPath), nil)
	if err != nil {
		return "", fmt.Errorf("fail to get revision(%s): %v", importPath, err)
	}

	i := bytes.Index(data, []byte(`commit-links-group BtnGroup`))
	if i == -1 {
		return "", fmt.Errorf("cannot find locater in page: %s", importPath)
	}
	data = data[i+1:]
	m := githubRevisionPattern.FindSubmatch(data)
	if m == nil {
		return "", fmt.Errorf("cannot find revision in page: %s", importPath)
	}
	return strings.TrimSuffix(strings.TrimPrefix(string(m[0]), `value="`), `"`), nil
}

type RepoInfo struct {
	DefaultBranch string `json:"default_branch"`
	Fork          bool   `json:"fork"`
	Parent        struct {
		FullName string `json:"full_name"`
	} `json:"parent"`
}

type RepoCommit struct {
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

func getGitHubDoc(match map[string]string, etag string) (*Package, error) {
	httpGet := func(url string, v interface{}) error {
		return httplib.Get(url).
			SetBasicAuth(setting.GitHub.ClientID, setting.GitHub.ClientSecret).
			ToJson(v)
	}
	repoInfo := new(RepoInfo)
	err := httpGet(com.Expand("https://api.github.com/repos/{owner}/{repo}", match), repoInfo)
	if err != nil {
		return nil, fmt.Errorf("get repo default branch: %v", err)
	}

	// Set default branch if not presented.
	if len(match["tag"]) == 0 {
		match["tag"] = repoInfo.DefaultBranch
	}

	// Check if last commit time is behind upstream for fork repository.
	if repoInfo.Fork {
		url := com.Expand("https://api.github.com/repos/{owner}/{repo}/commits?per_page=1&{cred}", match)
		forkCommits := make([]*RepoCommit, 0, 1)
		if err := httpGet(url, &forkCommits); err != nil {
			return nil, fmt.Errorf("get fork repository commits: %v", err)
		}
		if len(forkCommits) == 0 {
			return nil, fmt.Errorf("unexpected zero number of fork repository commits: %s", url)
		}

		match["parent"] = repoInfo.Parent.FullName
		url = com.Expand("https://api.github.com/repos/{parent}/commits?per_page=1&{cred}", match)
		parentCommits := make([]*RepoCommit, 0, 1)
		if err := httpGet(url, &parentCommits); err != nil {
			return nil, fmt.Errorf("get parent repository commits: %v", err)
		}
		if len(parentCommits) == 0 {
			return nil, fmt.Errorf("unexpected zero number of parent repository commits: %s", url)
		}

		if !forkCommits[0].Commit.Committer.Date.After(parentCommits[0].Commit.Committer.Date) {
			return nil, fmt.Errorf("commits of this fork repository are behind or equal to its parent: %s", repoInfo.Parent.FullName)
		}
	}

	// Check revision.
	var commit string
	if strings.HasPrefix(match["importPath"], "gopkg.in") {
		// FIXME: get commit ID of gopkg.in indepdently.
		var obj struct {
			Sha string `json:"sha"`
		}

		if err := com.HttpGetJSON(Client,
			com.Expand("https://gopm.io/api/v1/revision?pkgname={importPath}", match), &obj); err != nil {
			return nil, fmt.Errorf("get gopkg.in revision: %v", err)
		}

		commit = obj.Sha
		match["tag"] = commit
		log.Trace("Import path %q found commit: %s", match["importPath"], commit)
	} else {
		commit, err = getGithubRevision(com.Expand("github.com/{owner}/{repo}", match), match["tag"])
		if err != nil {
			return nil, fmt.Errorf("get revision: %v", err)
		}
		if commit == etag {
			return nil, ErrPackageNotModified
		}
	}

	// Get files.
	var tree struct {
		Tree []struct {
			Url  string
			Path string
			Type string
		}
		Url string
	}

	if err := httpGet(com.Expand("https://api.github.com/repos/{owner}/{repo}/git/trees/{tag}?recursive=1", match), &tree); err != nil {
		return nil, fmt.Errorf("get tree: %v", err)
	}

	// Because Github API URLs are case-insensitive, we need to check that the
	// userRepo returned from Github matches the one that we are requesting.
	if !strings.HasPrefix(tree.Url, com.Expand("https://api.github.com/repos/{owner}/{repo}/", match)) {
		return nil, errors.New("GitHub import path has incorrect case")
	}

	// Get source file data and subdirectories.
	dirPrefix := match["dir"]
	if dirPrefix != "" {
		dirPrefix = dirPrefix[1:] + "/"
	}
	dirLevel := len(strings.Split(dirPrefix, "/"))
	dirLength := len(dirPrefix)
	dirMap := make(map[string]bool)
	files := make([]com.RawFile, 0, 10)

	for _, node := range tree.Tree {
		// Skip directories and files in wrong directories, get them later.
		if node.Type != "blob" || !strings.HasPrefix(node.Path, dirPrefix) {
			continue
		}

		// Get files and check if directories have acceptable files.
		if d, f := path.Split(node.Path); base.IsDocFile(f) {
			// Check if file is in the directory that is corresponding to import path.
			if d == dirPrefix {
				files = append(files, &Source{
					SrcName:   f,
					BrowseUrl: com.Expand("github.com/{owner}/{repo}/blob/{tag}/{0}", match, node.Path),
					RawSrcUrl: com.Expand("https://raw.github.com/{owner}/{repo}/{tag}/{0}", match, node.Path),
				})
				continue
			}

			// Otherwise, check if it's a direct sub-directory of import path.
			if len(strings.Split(d, "/"))-dirLevel == 1 {
				dirMap[d[dirLength:len(d)-1]] = true
				continue
			}
		}
	}

	dirs := base.MapToSortedStrings(dirMap)

	if len(files) == 0 && len(dirs) == 0 {
		return nil, ErrPackageNoGoFile
	} else if err := com.FetchFiles(Client, files, githubRawHeader); err != nil {
		return nil, fmt.Errorf("fetch files: %v", err)
	}

	// Start generating data.
	// IsGoSubrepo check has been placed to crawl.getDynamic.
	w := &Walker{
		LineFmt: "#L%d",
		Pdoc: &Package{
			PkgInfo: &db.PkgInfo{
				ImportPath:  match["importPath"],
				ProjectPath: com.Expand("github.com/{owner}/{repo}", match),
				ViewDirPath: com.Expand("github.com/{owner}/{repo}/tree/{tag}/{importPath}", match),
				Etag:        commit,
				Subdirs:     strings.Join(dirs, "|"),
			},
		},
	}

	srcs := make([]*Source, 0, len(files))
	srcMap := make(map[string]*Source)
	for _, f := range files {
		s, _ := f.(*Source)
		srcs = append(srcs, s)

		if !strings.HasSuffix(f.Name(), "_test.go") {
			srcMap[f.Name()] = s
		}
	}

	pdoc, err := w.Build(&WalkRes{
		WalkDepth: WD_All,
		WalkType:  WT_Memory,
		WalkMode:  WM_All,
		Srcs:      srcs,
	})
	if err != nil {
		return nil, fmt.Errorf("error walking package: %v", err)
	}

	// Get stars.
	var repoTree struct {
		Stars int64 `json:"watchers"`
	}
	if err := httpGet(com.Expand("https://api.github.com/repos/{owner}/{repo}", match), &repoTree); err != nil {
		return nil, fmt.Errorf("get repoTree: %v", err)
	}
	pdoc.Stars = repoTree.Stars

	return pdoc, nil
}
