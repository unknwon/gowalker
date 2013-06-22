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

	"github.com/Unknwon/gowalker/utils"
)

var (
	googleRepoRe     = regexp.MustCompile(`id="checkoutcmd">(hg|git|svn)`)
	googleRevisionRe = regexp.MustCompile(`<h2>(?:[^ ]+ - )?Revision *([^:]+):`)
	googleTagRe      = regexp.MustCompile(`<option value="([^"/]+)"`)
	googleEtagRe     = regexp.MustCompile(`^(hg|git|svn)-`)
	googleFileRe     = regexp.MustCompile(`<li><a href="([^"/]+)"`)
	googleDirRe      = regexp.MustCompile(`<li><a href="([^".]+)"`)
	googlePattern    = regexp.MustCompile(`^code\.google\.com/p/(?P<repo>[a-z0-9\-]+)(:?\.(?P<subrepo>[a-z0-9\-]+))?(?P<dir>/[a-z0-9A-Z_.\-/]+)?$`)
)

func getStandardDoc(client *http.Client, importPath, tag, savedEtag string) (pdoc *Package, err error) {
	// hg-higtory: http://go.googlecode.com/hg-history/release/src/pkg/"+importPath+"/"
	p, err := httpGetBytes(client, "http://go.googlecode.com/hg/src/pkg/"+importPath+"/?r="+tag, nil)
	if err != nil {
		return nil, errors.New("doc.getStandardDoc(" + importPath + ") -> " + err.Error())
	}

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

	// Get source file data.
	files := make([]*source, 0, 5)
	for _, m := range googleFileRe.FindAllSubmatch(p, -1) {
		fname := strings.Split(string(m[1]), "?")[0]
		if utils.IsDocFile(fname) {
			files = append(files, &source{
				name:      fname,
				browseURL: "http://code.google.com/p/go/source/browse/src/pkg/" + importPath + "/" + fname + "?name=release",
				rawURL:    "http://go.googlecode.com/hg-history/release/src/pkg/" + importPath + "/" + fname,
			})
		}
	}

	dirs := make([]string, 0, 5)
	// Get subdirectories.
	for _, m := range googleDirRe.FindAllSubmatch(p, -1) {
		dirName := strings.Split(string(m[1]), "?")[0]
		// Make sure we get directories.
		if strings.HasSuffix(dirName, "/") &&
			utils.FilterFileName(dirName) {
			dirs = append(dirs, strings.Replace(dirName, "/", "", -1))
		}
	}

	srcDirs := make([]string, 0, len(dirs))
	c := make([]chan bool, len(dirs))
	// Filter direcotires.
	for i, d := range dirs {
		go func() {
			c[i] <- checkGoogleGoFile(client, importPath+"/"+d, tag)
		}()
	}
	for i, d := range dirs {
		if ok := <-c[i]; ok {
			srcDirs = append(srcDirs, d)
		}
	}

	if len(files) == 0 && len(srcDirs) == 0 {
		return nil, NotFoundError{"Directory tree does not contain Go files and subdirs."}
	}

	// Fetch file from VCS.
	if err := fetchFiles(client, files, nil); err != nil {
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
			Dirs:        srcDirs,
		},
	}
	return w.build(files)
}

func getGoogleTags(client *http.Client, importPath string) []string {
	p, err := httpGetBytes(client, "http://"+utils.GetProjectPath(importPath)+"/source/browse", nil)
	if err != nil {
		return nil
	}

	page := string(p)
	start := strings.Index(page, "<strong>Tag:</strong>")
	m := googleTagRe.FindAllStringSubmatch(page[start:], -1)

	tags := make([]string, 1, 6)
	tags[0] = "master"
	for i, v := range m {
		tags = append(tags, v[1])
		if i == 4 {
			break
		}
	}
	return tags
}

func checkGoogleGoFile(client *http.Client, path, tag string) bool {
	p, err := httpGetBytes(client, "http://go.googlecode.com/hg/src/pkg/"+path+"/?r="+tag, nil)
	if err != nil {
		return false
	}
	return len(googleFileRe.FindAllSubmatch(p, -1)) > 0
}

func getGoogleDoc(client *http.Client, match map[string]string, savedEtag string) (*Package, error) {
	setupGoogleMatch(match)
	if m := googleEtagRe.FindStringSubmatch(savedEtag); m != nil {
		match["vcs"] = m[1]
	} else if err := getGoogleVCS(client, match); err != nil {
		return nil, err
	}

	// Scrape the repo browser to find the project revision and individual Go files.
	p, err := httpGetBytes(client, expand("http://{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/", match), nil)
	if err != nil {
		return nil, err
	}

	// Check revision tag.
	var etag string
	if m := googleRevisionRe.FindSubmatch(p); m == nil {
		return nil, errors.New("doc.getGoogleDoc(): Could not find revision for " + match["importPath"])
	} else {
		etag = expand("{vcs}-{0}", match, string(m[1]))
		if etag == savedEtag {
			return nil, errNotModified
		}
	}

	// Get source file data.
	files := make([]*source, 0, 5)
	for _, m := range googleFileRe.FindAllSubmatch(p, -1) {
		fname := string(m[1])
		if utils.IsDocFile(fname) {
			files = append(files, &source{
				name:      fname,
				browseURL: expand("http://code.google.com/p/{repo}/source/browse{dir}/{0}{query}", match, fname),
				rawURL:    expand("http://{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/{0}", match, fname),
			})
		}
	}

	if len(files) == 0 {
		return nil, NotFoundError{"Directory tree does not contain Go files."}
	}

	dirs := make([]string, 0, 3)
	// Get subdirectories.
	for _, m := range googleDirRe.FindAllSubmatch(p, -1) {
		dirName := strings.Split(string(m[1]), "?")[0]
		if strings.HasSuffix(dirName, "/") {
			dirs = append(dirs, strings.Replace(dirName, "/", "", -1))
		}
	}

	// Fetch file from VCS.
	if err := fetchFiles(client, files, nil); err != nil {
		return nil, err
	}

	// Start generating data.
	w := &walker{
		lineFmt: "#%d",
		pdoc: &Package{
			ImportPath:  match["importPath"],
			ProjectName: expand("{repo}{dot}{subrepo}", match),
			Etag:        etag,
			Dirs:        dirs,
		},
	}
	return w.build(files)
}

func setupGoogleMatch(match map[string]string) {
	if s := match["subrepo"]; s != "" {
		match["dot"] = "."
		match["query"] = "?repo=" + s
	} else {
		match["dot"] = ""
		match["query"] = ""
	}
}

func getGoogleVCS(client *http.Client, match map[string]string) error {
	// Scrape the HTML project page to find the VCS.
	p, err := httpGetBytes(client, expand("http://code.google.com/p/{repo}/source/checkout", match), nil)
	if err != nil {
		return err
	}
	m := googleRepoRe.FindSubmatch(p)
	if m == nil {
		return NotFoundError{"Could not VCS on Google Code project page."}
	}
	match["vcs"] = string(m[1])
	return nil
}
