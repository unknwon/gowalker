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
	"regexp"
	"strings"

	"github.com/Unknwon/com"

	"github.com/Unknwon/gowalker/hv"
)

var (
	oscTagRe   = regexp.MustCompile(`/repository/archive\?ref=(.*)">`)
	oscPattern = regexp.MustCompile(`^git\.oschina\.net/(?P<owner>[a-z0-9A-Z_.\-]+)/(?P<repo>[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]*)?$`)
)

func getOSCDoc(client *http.Client, match map[string]string, tag, ptag string) (*hv.Package, error) {
	if len(tag) == 0 {
		match["tag"] = "master"
	} else {
		match["tag"] = tag
	}

	// Force to lower case.
	match["importPath"] = strings.ToLower(match["importPath"])

	match["browserUrlTpl"] = "git.oschina.net/{owner}/{repo}/blob/{tag}{dir}/{0}"
	match["rawSrcUrlTpl"] = "git.oschina.net/{owner}/{repo}/raw/{tag}/{dir}{0}"
	var isGoPro bool
	var files []com.RawFile
	var dirs []string
	var commit string
	var err error

	// NOTE: in order to have commit, for now can only download whole archive.
	tmpTag := match["tag"]
	if len(tmpTag) == 0 {
		tmpTag = defaultTags[match["vcs"]]
	}
	isGoPro, commit, files, dirs, err = getRepoByArchive(match,
		com.Expand("http://git.oschina.net/{owner}/{repo}/repository/archive?ref={tag}", match, tmpTag))
	if err != nil {
		return nil, errors.New("doc.getOSCDoc(" + match["importPath"] + ") -> Fail to download archive: " + err.Error())
	}
	if commit == ptag {
		return nil, errNotModified
	}

	if !isGoPro {
		return nil, com.NotFoundError{"Cannot find Go files, it's not a Go project"}
	}

	if len(files) == 0 && len(dirs) == 0 {
		return nil, com.NotFoundError{"Directory tree does not contain Go files and subdirs."}
	}

	// Get all tags.
	tags := getOSCTags(client, com.Expand("http://git.oschina.net/{owner}/{repo}/repository/tags", match))

	// Start generating data.
	w := &hv.Walker{
		LineFmt: "#L%d",
		Pdoc: &hv.Package{
			PkgInfo: &hv.PkgInfo{
				ImportPath:  match["importPath"],
				ProjectName: match["repo"],
				ProjectPath: com.Expand("git.oschina.net/{owner}/{repo}/blob/{tag}/", match),
				ViewDirPath: com.Expand("git.oschina.net/{owner}/{repo}/blob/{tag}{dir}/", match),
				Tags:        strings.Join(tags, "|||"),
				Ptag:        commit,
				Vcs:         "Git@OSC",
			},
			PkgDecl: &hv.PkgDecl{
				Tag:  tag,
				Dirs: dirs,
			},
		},
	}

	srcs := make([]*hv.Source, 0, len(files))
	srcMap := make(map[string]*hv.Source)
	for _, f := range files {
		s, _ := f.(*hv.Source)
		srcs = append(srcs, s)

		if !strings.HasSuffix(f.Name(), "_test.go") {
			srcMap[f.Name()] = s
		}
	}

	pdoc, err := w.Build(&hv.WalkRes{
		WalkDepth: hv.WD_All,
		WalkType:  hv.WT_Memory,
		WalkMode:  hv.WM_All,
		Srcs:      srcs,
	})
	if err != nil {
		return nil, errors.New("doc.getOSCDoc(" + match["importPath"] + ") -> Fail to build: " + err.Error())
	}

	if len(tag) == 0 && w.Pdoc.IsCmd {
		err = generateHv(match["importPath"], srcMap)
	}

	return pdoc, err
}

func getOSCTags(client *http.Client, tagsPath string) []string {
	p, err := com.HttpGetBytes(client, tagsPath, nil)
	if err != nil {
		return nil
	}

	tags := make([]string, 1, 6)
	tags[0] = "master"

	page := string(p)
	start := strings.Index(page, "<ul class='bordered-list'>")
	if start > -1 {
		m := oscTagRe.FindAllStringSubmatch(page[start:], -1)
		for i, v := range m {
			tags = append(tags, v[1])
			if i == 4 {
				break
			}
		}
	}
	return tags
}
