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
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/unknwon/com"

	"github.com/unknwon/gowalker/internal/base"
	"github.com/unknwon/gowalker/internal/db"
	"github.com/unknwon/gowalker/internal/httplib"
	"github.com/unknwon/gowalker/internal/setting"
)

var (
	ErrPackageNotModified = errors.New("Package has not been modified")
	ErrPackageNoGoFile    = errors.New("Package does not contain Go file")
)

func getGolangDoc(importPath, etag string) (*Package, error) {
	httpGet := func(url string, v interface{}) error {
		return httplib.Get(url).
			SetBasicAuth(setting.GitHub.ClientID, setting.GitHub.ClientSecret).
			ToJson(v)
	}
	// Check revision.
	commit, err := getGithubRevision("github.com/golang/go", "master")
	if err != nil {
		return nil, fmt.Errorf("get revision: %v", err)
	}
	if commit == etag {
		return nil, ErrPackageNotModified
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

	if err := httpGet("https://api.github.com/repos/golang/go/git/trees/master?recursive=1", &tree); err != nil {
		return nil, fmt.Errorf("get tree: %v", err)
	}

	dirPrefix := "src/" + importPath + "/"
	dirLevel := len(strings.Split(dirPrefix, "/"))
	dirLength := len(dirPrefix)
	dirMap := make(map[string]bool)
	files := make([]com.RawFile, 0, 10)

	for _, node := range tree.Tree {
		// Skip directories and files in irrelevant directories.
		if node.Type != "blob" || !strings.HasPrefix(node.Path, dirPrefix) {
			continue
		}

		// Get files and check if directories have acceptable files.
		if d, f := path.Split(node.Path); base.IsDocFile(f) {
			// Check if file is in the directory that is corresponding to import path.
			if d == dirPrefix {
				files = append(files, &Source{
					SrcName:   f,
					BrowseUrl: com.Expand("github.com/golang/go/blob/master/{0}", nil, node.Path),
					RawSrcUrl: com.Expand("https://raw.github.com/golang/go/master/{0}", nil, node.Path),
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
	w := &Walker{
		LineFmt: "#L%d",
		Pdoc: &Package{
			PkgInfo: &db.PkgInfo{
				ImportPath:  importPath,
				ProjectPath: "github.com/golang/go",
				ViewDirPath: "github.com/golang/go/tree/master/src/" + importPath,
				Etag:        commit,
				IsGoRepo:    true,
				Subdirs:     strings.Join(dirs, "|"),
			},
		},
	}

	srcs := make([]*Source, 0, len(files))
	srcMap := make(map[string]*Source)
	for _, f := range files {
		s := f.(*Source)
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
		return nil, fmt.Errorf("walk package: %v", err)
	}

	return pdoc, nil
}
