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
)

var (
	googleRevisionRe = regexp.MustCompile(`<h2>(?:[^ ]+ - )?Revision *([^:]+):`)
	googleTagRe      = regexp.MustCompile(`<option value="([^"/]+)"`)
	googleEtagRe     = regexp.MustCompile(`^(hg|git|svn)-`)
	googleFileRe     = regexp.MustCompile(`<li><a href="([^"/]+)"`)
	googleDirRe      = regexp.MustCompile(`<li><a href="([^".]+)"`)
	googlePattern    = regexp.MustCompile(`^code\.google\.com/p/(?P<repo>[a-z0-9\-]+)(:?\.(?P<subrepo>[a-z0-9\-]+))?(?P<dir>/[a-z0-9A-Z_.\-/]+)?$`)
)

func getStandardDoc(client *http.Client, importPath, tag, savedEtag string) (pdoc *Package, err error) {
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
		if etag == savedEtag {
			return nil, errNotModified
		}
	}

	// Check if source files saved locally.
	if 

	// Get source file data.
	files := make([]com.RawFile, 0, 5)
	for _, m := range googleFileRe.FindAllSubmatch(p, -1) {
		fname := strings.Split(string(m[1]), "?")[0]
		if utils.IsDocFile(fname) {
			files = append(files, &source{
				name:      fname,
				browseURL: "http://code.google.com/p/go/source/browse/src/pkg/" + importPath + "/" + fname + "?r=" + tag,
				rawURL:    "http://go.googlecode.com/hg/src/pkg/" + importPath + "/" + fname + "?r=" + tag,
			})
		}
	}

	dirs := make([]string, 0, 5)
	// Get subdirectories.
	for _, m := range googleDirRe.FindAllSubmatch(p, -1) {
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
	if err := com.FetchFiles(client, files, nil); err != nil {
		return nil, err
	}

	// Get all tags.
	tags := getGoogleTags(client, "code.google.com/p/go/"+importPath)

	// Start generating data.
	w := &walker{
		lineFmt: "#%d",
		pdoc: &Package{
			ImportPath:  importPath,
			ProjectName: "Go",
			Tags:        tags,
			Tag:         tag,
			Etag:        etag,
			Dirs:        dirs,
		},
	}
	return w.build(files)
}

func getGoogleTags(client *http.Client, importPath string) []string {
	p, err := com.HttpGetBytes(client, "http://"+utils.GetProjectPath(importPath)+"/source/browse", nil)
	if err != nil {
		return nil
	}

	tags := make([]string, 1, 6)
	tags[0] = "master"

	page := string(p)
	start := strings.Index(page, "<strong>Tag:</strong>")
	if start > -1 {
		m := googleTagRe.FindAllStringSubmatch(page[start:], -1)
		for i, v := range m {
			tags = append(tags, v[1])
			if i == 4 {
				break
			}
		}
	}
	return tags
}

func getGoogleDoc(client *http.Client, match map[string]string, tag, savedEtag string) (*Package, error) {
	packer.SetupGoogleMatch(match)
	if m := googleEtagRe.FindStringSubmatch(savedEtag); m != nil {
		match["vcs"] = m[1]
	} else if err := packer.GetGoogleVCS(client, match); err != nil {
		return nil, err
	}

	match["tag"] = tag
	// Scrape the repo browser to find the project revision and individual Go files.
	p, err := com.HttpGetBytes(client, com.Expand("http://{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/?r={tag}", match), nil)
	if err != nil {
		return nil, errors.New("doc.getGoogleDoc(" + match["importPath"] + ") -> " + err.Error())
	}

	// Check revision tag.
	var etag string
	if m := googleRevisionRe.FindSubmatch(p); m == nil {
		return nil, errors.New("doc.getGoogleDoc(" + match["importPath"] + ") -> Could not find revision")
	} else {
		etag = com.Expand("{vcs}-{0}", match, string(m[1]))
		if etag == savedEtag {
			return nil, errNotModified
		}
	}

	// Get source file data.
	files := make([]com.RawFile, 0, 5)
	for _, m := range googleFileRe.FindAllSubmatch(p, -1) {
		fname := string(m[1])
		if utils.IsDocFile(fname) {
			files = append(files, &source{
				name:      fname,
				browseURL: com.Expand("http://code.google.com/p/{repo}/source/browse{dir}/{0}{query}?r={tag}", match, fname),
				rawURL:    com.Expand("http://{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/{0}?r={tag}", match, fname),
			})
		}
	}

	dirs := make([]string, 0, 5)
	// Get subdirectories.
	for _, m := range googleDirRe.FindAllSubmatch(p, -1) {
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
	if err := com.FetchFiles(client, files, nil); err != nil {
		return nil, err
	}

	// Get all tags.
	tags := getGoogleTags(client, match["importPath"])

	// Start generating data.
	w := &walker{
		lineFmt: "#%d",
		pdoc: &Package{
			ImportPath:  match["importPath"],
			ProjectName: com.Expand("{repo}{dot}{subrepo}", match),
			Tags:        tags,
			Tag:         tag,
			Etag:        etag,
			Dirs:        dirs,
		},
	}
	return w.build(files)
}
