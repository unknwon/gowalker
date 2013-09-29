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
	"regexp"
	"strings"

	"github.com/Unknwon/com"
	"github.com/Unknwon/ctw/packer"
	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
)

var (
	googleRevisionRe = regexp.MustCompile(`<h2>(?:[^ ]+ - )?Revision *([^:]+):`)
	googleTagRe      = regexp.MustCompile(`<option value="([^"/]+)"`)
	googleEtagRe     = regexp.MustCompile(`^(hg|git|svn)-`)
	googleFileRe     = regexp.MustCompile(`<li><a href="([^"/]+)"`)
	googleDirRe      = regexp.MustCompile(`<li><a href="([^".]+)"`)
	googlePattern    = regexp.MustCompile(`^code\.google\.com/p/(?P<repo>[a-z0-9\-]+)(:?\.(?P<subrepo>[a-z0-9\-]+))?(?P<dir>/[a-z0-9A-Z_.\-/]+)?$`)
)

func getStandardDoc(client *http.Client, importPath, tag, ptag string) (pdoc *hv.Package, err error) {
	// hg-higtory: http://go.googlecode.com/hg-history/release/src/pkg/"+importPath+"/"
	stdout, _, err := com.ExecCmd("curl", "http://go.googlecode.com/hg/src/pkg/"+importPath+"/?r="+tag)
	if err != nil {
		return nil, errors.New("doc.getStandardDoc(" + importPath + ") -> " + err.Error())
	}
	p := []byte(stdout)

	// Check revision tag.
	var etag string
	if m := googleRevisionRe.FindSubmatch(p); m == nil {
		return nil, errors.New("doc.getStandardDoc(" + importPath + ") -> Could not find revision")
	} else {
		etag = string(m[1])
		if etag == ptag {
			return nil, errNotModified
		}
	}

	// Get source file data.
	ms := googleFileRe.FindAllSubmatch(p, -1)
	files := make([]com.RawFile, 0, len(ms))
	for _, m := range ms {
		fname := strings.Split(string(m[1]), "?")[0]
		if utils.IsDocFile(fname) {
			files = append(files, &hv.Source{
				SrcName:   fname,
				BrowseUrl: "code.google.com/p/go/source/browse/src/pkg/" + importPath + "/" + fname + "?r=" + tag,
				RawSrcUrl: "go.googlecode.com/hg/src/pkg/" + importPath + "/" + fname + "?r=" + tag,
			})
		}
	}

	// Get subdirectories.
	ms = googleDirRe.FindAllSubmatch(p, -1)
	dirs := make([]string, 0, len(ms))
	for _, m := range ms {
		dirName := strings.Split(string(m[1]), "?")[0]
		// Make sure we get directories.
		if strings.HasSuffix(dirName, "/") &&
			utils.FilterDirName(dirName) {
			dirs = append(dirs, strings.Replace(dirName, "/", "", -1))
		}
	}

	if len(files) == 0 && len(dirs) == 0 {
		return nil, com.NotFoundError{"Directory tree does not contain Go files and subdirs."}
	}

	// Fetch file from VCS.
	if err := com.FetchFilesCurl(files); err != nil {
		return nil, err
	}

	// Get all tags.
	tags := getGoogleTags("code.google.com/p/go/"+importPath, "default", true)

	// Start generating data.
	w := &hv.Walker{
		LineFmt: "#%d",
		Pdoc: &hv.Package{
			PkgInfo: &hv.PkgInfo{
				ImportPath:  importPath,
				ProjectName: "Go",
				ProjectPath: "code.google.com/p/go/source/browse/src/pkg/?r=" + tag,
				ViewDirPath: "code.google.com/p/go/source/browse/src/pkg/" + importPath + "/?r=" + tag,
				IsGoRepo:    true,
				Tags:        strings.Join(tags, "|||"),
				Ptag:        etag,
				Vcs:         "Google Code",
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

func getGoogleTags(importPath string, defaultBranch string, isGoRepo bool) []string {
	stdout, _, err := com.ExecCmd("curl", "http://"+utils.GetProjectPath(importPath)+"/source/browse")
	if err != nil {
		return nil
	}
	p := []byte(stdout)

	page := string(p)
	start := strings.Index(page, "<strong>Tag:</strong>")
	if start == -1 {
		return nil
	}

	m := googleTagRe.FindAllStringSubmatch(page[start:], -1)

	var tags []string
	if isGoRepo {
		tags = make([]string, 1, 20)
	} else {
		tags = make([]string, len(m)+1)
	}

	tags[0] = defaultBranch
	for i, v := range m {
		if isGoRepo {
			if strings.HasPrefix(v[1], "go") {
				tags = append(tags, v[1])
			}
			continue
		}
		tags[i+1] = v[1]
	}
	return tags
}

func getGoogleDoc(client *http.Client, match map[string]string, tag, ptag string) (*hv.Package, error) {
	packer.SetupGoogleMatch(match)
	if m := googleEtagRe.FindStringSubmatch(ptag); m != nil {
		match["vcs"] = m[1]
	} else if err := packer.GetGoogleVCS(match); err != nil {
		return nil, err
	}

	match["tag"] = tag
	// Scrape the repo browser to find the project revision and individual Go files.
	stdout, _, err := com.ExecCmd("curl", com.Expand("http://{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/?r={tag}", match))
	if err != nil {
		return nil, errors.New("doc.getGoogleDoc(" + match["importPath"] + ") -> " + err.Error())
	}
	p := []byte(stdout)

	// Check revision tag.
	var etag string
	if m := googleRevisionRe.FindSubmatch(p); m == nil {
		return nil, errors.New("doc.getGoogleDoc(" + match["importPath"] + ") -> Could not find revision")
	} else {
		etag = string(m[1])
		if etag == ptag {
			return nil, errNotModified
		}
	}

	match["browserUrlTpl"] = "code.google.com/p/{repo}/source/browse{dir}/{0}?repo={subrepo}&r={tag}"
	match["rawSrcUrlTpl"] = "{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/{0}?r={tag}"
	var isGoPro bool
	var files []com.RawFile
	var dirs []string
	// Unrecord and non-SVN project can download archive.
	if len(ptag) == 0 || match["vcs"] == "svn" {
		tmpTag := match["tag"]
		if len(tmpTag) == 0 {
			tmpTag = defaultTags[match["vcs"]]
		}

		isGoPro, _, files, dirs, err = getRepoByArchive(match,
			com.Expand("http://{subrepo}{dot}{repo}.googlecode.com/archive/{0}.zip", match, tmpTag))
		if err != nil {
			return nil, errors.New("doc.getGoogleDoc(" + match["importPath"] + ") -> Fail to download archive: " + err.Error())
		}
	} else {
		// Get source file data.
		ms := googleFileRe.FindAllSubmatch(p, -1)
		files = make([]com.RawFile, 0, len(ms))
		for _, m := range ms {
			fname := strings.Split(string(m[1]), "?")[0]
			if utils.IsDocFile(fname) {
				isGoPro = true
				files = append(files, &hv.Source{
					SrcName:   fname,
					BrowseUrl: com.Expand(match["browserUrlTpl"], match, fname),
					RawSrcUrl: com.Expand(match["rawSrcUrlTpl"], match, fname),
				})
			}
		}

		// Get subdirectories.
		ms = googleDirRe.FindAllSubmatch(p, -1)
		dirs = make([]string, 0, len(ms))
		for _, m := range ms {
			dirName := strings.Split(string(m[1]), "?")[0]
			// Make sure we get directories.
			if strings.HasSuffix(dirName, "/") &&
				utils.FilterDirName(dirName) {
				dirs = append(dirs, strings.Replace(dirName, "/", "", -1))
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
	if err := com.FetchFilesCurl(files); err != nil {
		return nil, err
	}

	// Get all tags.
	tags := getGoogleTags(match["importPath"], defaultTags[match["vcs"]], false)

	// Start generating data.
	w := &hv.Walker{
		LineFmt: "#%d",
		Pdoc: &hv.Package{
			PkgInfo: &hv.PkgInfo{
				ImportPath:  match["importPath"],
				ProjectName: com.Expand("{repo}{dot}{subrepo}", match),
				ProjectPath: com.Expand("code.google.com/p/{repo}/source/browse/?repo={subrepo}&r={tag}", match),
				ViewDirPath: com.Expand("code.google.com/p/{repo}/source/browse{dir}?repo={subrepo}&r={tag}", match),
				Tags:        strings.Join(tags, "|||"),
				Ptag:        etag,
				Vcs:         "Google Code",
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
	return nil, nil
}
