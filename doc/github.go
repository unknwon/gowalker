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
	"errors"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/Unknwon/com"
	"github.com/Unknwon/gowalker/utils"
)

var (
	githubRawHeader = http.Header{"Accept": {"application/vnd.github-blob.raw"}}
	githubPattern   = regexp.MustCompile(`^github\.com/(?P<owner>[a-z0-9A-Z_.\-]+)/(?P<repo>[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]*)?$`)
	githubCred      string
)

func init() {
	setGithubCredentials(utils.Cfg.MustValue("github", "client_id"),
		utils.Cfg.MustValue("github", "client_secret"))
}

func setGithubCredentials(id, secret string) {
	githubCred = "client_id=" + id + "&client_secret=" + secret
}

func GetGithubCredentials() string {
	return githubCred
}

func getGithubDoc(client *http.Client, match map[string]string, tag, savedEtag string) (*Package, error) {
	match["cred"] = githubCred

	// Get master commit.
	var refs []*struct {
		Ref    string
		Url    string
		Object struct {
			Sha  string
			Type string
			Url  string
		}
	}

	err := com.HttpGetJSON(client, com.Expand("https://api.github.com/repos/{owner}/{repo}/git/refs?{cred}", match), &refs)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Resource not found") {
			return nil, com.NotFoundError{"doc.getGithubDoc(" + match["importPath"] + ") -> " + err.Error()}
		}
		return nil, errors.New("doc.getGithubDoc(" + match["importPath"] + ") -> " + err.Error())
	}

	var commit string
	// Get all tags.
	tags := make([]string, 0, 5)
	for _, ref := range refs {
		switch {
		case strings.HasPrefix(ref.Ref, "refs/heads/master"):
			commit = ref.Object.Sha
		case strings.HasPrefix(ref.Ref, "refs/tags/"):
			tags = append(tags, ref.Ref[len("refs/tags/"):])
		}
	}

	if len(tags) > 5 {
		tags = tags[len(tags)-5:]
	}
	tags = append([]string{"master"}, tags...)

	if len(tag) == 0 {
		// Check revision tag.
		if commit == savedEtag {
			return nil, errNotModified
		}

		match["tag"] = "master"
	} else {
		match["tag"] = tag
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

	err = com.HttpGetJSON(client, com.Expand("https://api.github.com/repos/{owner}/{repo}/git/trees/{tag}?recursive=1&{cred}", match), &tree)
	if err != nil {
		return nil, errors.New("doc.getGithubDoc(" + match["importPath"] + ") -> get trees: " + err.Error())
	}

	// Because Github API URLs are case-insensitive, we need to check that the
	// userRepo returned from Github matches the one that we are requesting.
	if !strings.HasPrefix(tree.Url, com.Expand("https://api.github.com/repos/{owner}/{repo}/", match)) {
		return nil, errors.New("Github import path has incorrect case.")
	}

	// Get source file data and subdirectories.
	dirPrefix := match["dir"]
	if dirPrefix != "" {
		dirPrefix = dirPrefix[1:] + "/"
	}
	preLen := len(dirPrefix)

	isGoPro := false // Indicates whether it's a Go project.
	isRootPath := match["importPath"] == utils.GetProjectPath(match["importPath"])
	dirs := make([]string, 0, 5)
	files := make([]com.RawFile, 0, 5)
	for _, node := range tree.Tree {
		// Skip directories and files in wrong directories, get them later.
		if node.Type != "blob" || !strings.HasPrefix(node.Path, dirPrefix) {
			continue
		}

		// Get files and check if directories have acceptable files.
		if d, f := path.Split(node.Path); utils.IsDocFile(f) &&
			utils.FilterDirName(d) {
			// Check if it's a Go file.
			if isRootPath && !isGoPro && strings.HasSuffix(f, ".go") {
				isGoPro = true
			}

			// Check if file is in the directory that is corresponding to import path.
			if d == dirPrefix {
				// Yes.
				if !isRootPath && !isGoPro && strings.HasSuffix(f, ".go") {
					isGoPro = true
				}
				files = append(files, &source{
					name:      f,
					browseURL: com.Expand("https://github.com/{owner}/{repo}/blob/{tag}/{0}", match, node.Path),
					rawURL:    node.Url + "?" + githubCred,
				})
			} else {
				sd, _ := path.Split(d[preLen:])
				sd = strings.TrimSuffix(sd, "/")
				if !checkDir(sd, dirs) {
					dirs = append(dirs, sd)
				}
			}
		}
	}

	if !isGoPro {
		return nil, com.NotFoundError{"Cannot find Go files, it's not a Go project."}
	}

	if len(files) == 0 && len(dirs) == 0 {
		return nil, com.NotFoundError{"Directory tree does not contain Go files and subdirs."}
	}

	// Fetch file from VCS.
	if err := com.FetchFiles(client, files, githubRawHeader); err != nil {
		return nil, err
	}

	// Get addtional information: forks, watchers.
	// var note struct {
	// 	Forks    int
	// 	Watchers int `json:"watchers_count"`
	// }

	// err = httpGetJSON(client, expand("https://api.github.com/repos/{owner}/{repo}?{cred}", match), &note)
	// if err != nil {
	// 	return nil, errors.New("doc.getGithubDoc(" + match["importPath"] + ") -> get note: " + err.Error())
	// }

	// Start generating data.
	w := &walker{
		lineFmt: "#L%d",
		pdoc: &Package{
			ImportPath:  match["importPath"],
			ProjectName: match["repo"],
			Tags:        tags,
			Tag:         tag,
			Etag:        commit,
			Dirs:        dirs,
			//Note: strconv.Itoa(note.Forks) + "|" +
			//	strconv.Itoa(note.Watchers) + "|",
		},
	}
	return w.build(files)
}

// checkDir checks if directory has been appended to slice.
func checkDir(dir string, dirs []string) bool {
	for _, d := range dirs {
		if dir == d {
			return true
		}
	}
	return false
}
