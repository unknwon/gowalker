// Copyright 2011 Gary Burd
// Copyright 2013 Unknown
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
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/Unknwon/gowalker/utils"
)

var (
	githubRawHeader = http.Header{"Accept": {"application/vnd.github-blob.raw"}}
	githubPattern   = regexp.MustCompile(`^github\.com/(?P<owner>[a-z0-9A-Z_.\-]+)/(?P<repo>[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]*)?$`)
	githubCred      string
)

func SetGithubCredentials(id, secret string) {
	githubCred = "client_id=" + id + "&client_secret=" + secret
}

func getGithubDoc(client *http.Client, match map[string]string, savedEtag string) (*Package, error) {
	SetGithubCredentials("1862bcb265171f37f36c", "308d71ab53ccd858416cfceaed52d5d5b7d53c5f")
	match["cred"] = githubCred

	var refs []*struct {
		Object struct {
			Type string
			Sha  string
			Url  string
		}
		Ref string
		Url string
	}

	err := httpGetJSON(client, expand("https://api.github.com/repos/{owner}/{repo}/git/refs?{cred}", match), &refs)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]string)
	for _, ref := range refs {
		switch {
		case strings.HasPrefix(ref.Ref, "refs/heads/"):
			tags[ref.Ref[len("refs/heads/"):]] = ref.Object.Sha
		case strings.HasPrefix(ref.Ref, "refs/tags/"):
			tags[ref.Ref[len("refs/tags/"):]] = ref.Object.Sha
		}
	}

	// Check revision tag.
	var commit string
	match["tag"], commit, err = bestTag(tags, "master")
	if err != nil {
		return nil, err
	}

	if commit == savedEtag {
		return nil, errNotModified
	}

	var tree struct {
		Tree []struct {
			Url  string
			Path string
			Type string
		}
		Url string
	}

	err = httpGetJSON(client, expand("https://api.github.com/repos/{owner}/{repo}/git/trees/{tag}?recursive=1&{cred}", match), &tree)
	if err != nil {
		return nil, err
	}

	// Because Github API URLs are case-insensitive, we need to check that the
	// userRepo returned from Github matches the one that we are requesting.
	if !strings.HasPrefix(tree.Url, expand("https://api.github.com/repos/{owner}/{repo}/", match)) {
		return nil, NotFoundError{"Github import path has incorrect case."}
	}

	inTree := false
	dirPrefix := match["dir"]
	if dirPrefix != "" {
		dirPrefix = dirPrefix[1:] + "/"
	}
	preLen := len(dirPrefix)

	// Get source file data.
	dirs := make([]string, 0, 5)
	files := make([]*source, 0, 5)
	for _, node := range tree.Tree {
		if node.Type != "blob" || !strings.HasPrefix(node.Path, dirPrefix) {
			if len(dirPrefix) > 0 && strings.HasPrefix(node.Path, dirPrefix) {
				p := node.Path[preLen:]
				dirs = append(dirs, p)
			} else if len(dirPrefix) == 0 && strings.Index(node.Path, "/") == -1 {
				dirs = append(dirs, node.Path)
			}
			continue
		}
		inTree = true
		if d, f := path.Split(node.Path); d == dirPrefix && utils.IsDocFile(f) {
			files = append(files, &source{
				name:      f,
				browseURL: expand("https://github.com/{owner}/{repo}/blob/{tag}/{0}", match, node.Path),
				rawURL:    node.Url + "?" + githubCred,
			})
		}
	}

	if !inTree || len(files) == 0 {
		return nil, NotFoundError{"Directory tree does not contain Go files."}
	}

	// Fetch file from VCS.
	if err := fetchFiles(client, files, githubRawHeader); err != nil {
		return nil, err
	}

	/*browseURL := expand("https://github.com/{owner}/{repo}", match)
	if match["dir"] != "" {
		browseURL = expand("https://github.com/{owner}/{repo}/tree/{tag}{dir}", match)
	}*/

	// Start generating data.
	w := &walker{
		lineFmt: "#L%d",
		pdoc: &Package{
			ImportPath:  match["importPath"],
			ProjectName: match["repo"],
			Etag:        commit,
			Dirs:        dirs,
		},
	}
	return w.build(files)
}
