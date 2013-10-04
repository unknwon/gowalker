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
	"github.com/Unknwon/hv"
)

var (
	bitbucketPattern = regexp.MustCompile(`^bitbucket\.org/(?P<owner>[a-z0-9A-Z_.\-]+)/(?P<repo>[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]*)?$`)
	bitbucketEtagRe  = regexp.MustCompile(`^(hg|git)-`)
)

func getBitbucketDoc(client *http.Client, match map[string]string, tag, savedEtag string) (*hv.Package, error) {

	if m := bitbucketEtagRe.FindStringSubmatch(savedEtag); m != nil {
		match["vcs"] = m[1]
	} else {
		var repo struct {
			Scm string
		}
		if err := com.HttpGetJSON(client, com.Expand("https://api.bitbucket.org/1.0/repositories/{owner}/{repo}", match), &repo); err != nil {
			return nil, errors.New("doc.getBitbucketDoc(" + match["importPath"] + ") -> " + err.Error())
		}
		match["vcs"] = repo.Scm
	}

	// Get master commit.
	var branches map[string]struct {
		Node string
	}
	if err := com.HttpGetJSON(client, com.Expand("https://api.bitbucket.org/1.0/repositories/{owner}/{repo}/branches", match), &branches); err != nil {
		return nil, errors.New("doc.getBitbucketDoc(" + match["importPath"] + ") -> get branches: " + err.Error())
	}
	match["commit"] = branches["default"].Node

	// Get all tags.
	tags := make([]string, 0, 5)
	var nodes map[string]struct {
		Node string
	}
	if err := com.HttpGetJSON(client, com.Expand("https://api.bitbucket.org/1.0/repositories/{owner}/{repo}/tags", match), &nodes); err != nil {
		return nil, errors.New("doc.getBitbucketDoc(" + match["importPath"] + ") -> get nodes: " + err.Error())
	}
	for k := range nodes {
		tags = append(tags, k)
	}
	if len(tags) > 0 {
		tags = append([]string{defaultTags[match["vcs"]]}, tags...)
	}

	var etag string
	if len(tag) == 0 {
		// Check revision tag.
		etag = match["commit"]
		if etag == savedEtag {
			return nil, errNotModified
		}

		match["tag"] = defaultTags[match["vcs"]]
	} else {
		match["tag"] = tag
	}

	// Get files and directories.
	var node struct {
		Files []struct {
			Path string
		}
		Directories []string
	}

	if err := com.HttpGetJSON(client, com.Expand("https://api.bitbucket.org/1.0/repositories/{owner}/{repo}/src/{tag}{dir}/", match), &node); err != nil {
		return nil, errors.New("doc.getBitbucketDoc(" + match["importPath"] + ") -> get trees: " + err.Error())
	}

	// Get source file data.
	files := make([]com.RawFile, 0, 5)
	for _, f := range node.Files {
		_, name := path.Split(f.Path)
		if utils.IsDocFile(name) {
			files = append(files, &hv.Source{
				SrcName:   name,
				BrowseUrl: com.Expand("bitbucket.org/{owner}/{repo}/src/{tag}/{0}", match, f.Path),
				RawSrcUrl: com.Expand("https://api.bitbucket.org/1.0/repositories/{owner}/{repo}/raw/{tag}/{0}", match, f.Path),
			})
		}
	}

	// Get subdirectories.
	dirs := make([]string, 0, len(node.Directories))
	for _, d := range node.Directories {
		if utils.FilterDirName(d) {
			dirs = append(dirs, d)
		}
	}

	if len(files) == 0 && len(dirs) == 0 {
		return nil, com.NotFoundError{"Directory tree does not contain Go files and subdirs."}
	}

	// Fetch file from VCS.
	if err := com.FetchFiles(client, files, nil); err != nil {
		return nil, err
	}

	// Start generating data.
	w := &hv.Walker{
		LineFmt: "#cl-%d",
		Pdoc: &hv.Package{
			PkgInfo: &hv.PkgInfo{
				ImportPath:  match["importPath"],
				ProjectName: match["repo"],
				ProjectPath: com.Expand("bitbucket.org/{owner}/{repo}/src/{tag}/", match),
				ViewDirPath: com.Expand("bitbucket.org/{owner}/{repo}/src/{tag}{dir}/", match),
				Tags:        strings.Join(tags, "|||"),
				Ptag:        etag,
				Vcs:         "BitBucket",
			},
			PkgDecl: &hv.PkgDecl{
				Tag:  tag,
				Dirs: dirs,
			},
		},
	}

	srcs := make([]*hv.Source, 0, len(files))
	for _, f := range files {
		s, _ := f.(*hv.Source)
		srcs = append(srcs, s)
	}

	return w.Build(&hv.WalkRes{
		WalkDepth: hv.WD_All,
		WalkType:  hv.WT_Memory,
		WalkMode:  hv.WM_All,
		Srcs:      srcs,
	})
}
