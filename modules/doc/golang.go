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

	"github.com/Unknwon/com"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/modules/setting"
)

var (
	ErrPackageNotModified = errors.New("Package has not been modified")
	ErrPackageNoGoFile    = errors.New("Package does not contain Go file")
)

func getGolangDoc(importPath, etag string) (*Package, error) {
	match := map[string]string{"cred": setting.GitHubCredentials}

	// Check revision.
	commit, err := getGithubRevision("github.com/golang/go")
	if err != nil {
		return nil, fmt.Errorf("error getting revision: %v", err)
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

	if err := com.HttpGetJSON(Client,
		com.Expand("https://api.github.com/repos/golang/go/git/trees/master?recursive=1&{cred}", match), &tree); err != nil {
		return nil, fmt.Errorf("error getting tree: %v", err)
	}

	dirPrefix := "src/" + importPath + "/"
	files := make([]com.RawFile, 0, 10)
	for _, node := range tree.Tree {
		// Skip directories and files in irrelevant directories.
		if node.Type != "blob" || !strings.HasPrefix(node.Path, dirPrefix) {
			continue
		}

		// Get files and check if directories have acceptable files.
		if d, f := path.Split(node.Path); IsDocFile(f) {
			// Check if file is in the directory that is corresponding to import path.
			if d == dirPrefix {
				files = append(files, &Source{
					SrcName:   f,
					BrowseUrl: com.Expand("github.com/golang/go/blob/master/{0}", nil, node.Path),
					RawSrcUrl: com.Expand("https://raw.github.com/golang/go/master/{0}?{1}", nil, node.Path, setting.GitHubCredentials),
				})
			}
		}
	}

	if len(files) == 0 {
		return nil, ErrPackageNoGoFile
	} else if err := com.FetchFiles(Client, files, githubRawHeader); err != nil {
		return nil, fmt.Errorf("error fetching files: %v", err)
	}

	// Start generating data.
	w := &Walker{
		LineFmt: "#L%d",
		Pdoc: &Package{
			PkgInfo: &models.PkgInfo{
				ImportPath:  importPath,
				ProjectPath: "github.com/golang/go",
				ViewDirPath: "github.com/golang/go/tree/master/" + importPath,
				Etag:        commit,
				IsGoRepo:    true,
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

	return pdoc, nil
}
