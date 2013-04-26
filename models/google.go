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

package models

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/unknwon/gowalker/utils"
)

var (
	googleRepoRe     = regexp.MustCompile(`id="checkoutcmd">(hg|git|svn)`)
	googleRevisionRe = regexp.MustCompile(`<h2>(?:[^ ]+ - )?Revision *([^:]+):`)
	googleEtagRe     = regexp.MustCompile(`^(hg|git|svn)-`)
	googleFileRe     = regexp.MustCompile(`<li><a href="([^"/]+)"`)
	googleDirRe      = regexp.MustCompile(`<li><a href="([^".]+)"`)
	googlePattern    = regexp.MustCompile(`^code\.google\.com/p/(?P<repo>[a-z0-9\-]+)(:?\.(?P<subrepo>[a-z0-9\-]+))?(?P<dir>/[a-z0-9A-Z_.\-/]+)?$`)
)

func getStandardDoc(client *http.Client, importPath string) (pdoc *Package, err error) {
	p, err := httpGetBytes(client, "http://go.googlecode.com/hg-history/release/src/pkg/"+importPath+"/", nil)
	if err != nil {
		return nil, err
	}

	var files []*source
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

	if err := fetchFiles(client, files, nil); err != nil {
		return nil, err
	}

	w := &walker{
		lineFmt: "#%d",
		pdoc: &Package{
			ImportPath:  importPath,
			ProjectRoot: "http://code.google.com/p/go/source/browse/src/pkg/",
			ProjectName: "Go",
			ProjectURL:  "https://code.google.com/p/go/",
			BrowseURL:   "http://code.google.com/p/go/source/browse/src/pkg/" + importPath + "?name=release",
			VCS:         "hg",
		},
	}

	return w.build(files)
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

	var etag string
	if m := googleRevisionRe.FindSubmatch(p); m == nil {
		return nil, errors.New("Could not find revision for " + match["importPath"])
	} else {
		etag = expand("{vcs}-{0}", match, string(m[1]))
		if etag == savedEtag {
			return nil, errNotModified
		}
	}

	var files []*source
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

	if err := fetchFiles(client, files, nil); err != nil {
		return nil, err
	}

	w := &walker{
		lineFmt: "#%d",
		pdoc: &Package{
			ImportPath:  match["importPath"],
			ProjectRoot: expand("code.google.com/p/{repo}{dot}{subrepo}", match),
			ProjectName: expand("{repo}{dot}{subrepo}", match),
			ProjectURL:  expand("https://code.google.com/p/{repo}/", match),
			BrowseURL:   expand("http://code.google.com/p/{repo}/source/browse{dir}/{query}", match),
			VCS:         match["vcs"],
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

// expand replaces {k} in template with match[k] or subs[atoi(k)] if k is not in match.
func expand(template string, match map[string]string, subs ...string) string {
	var p []byte
	var i int
	for {
		i = strings.Index(template, "{")
		if i < 0 {
			break
		}
		p = append(p, template[:i]...)
		template = template[i+1:]
		i = strings.Index(template, "}")
		if s, ok := match[template[:i]]; ok {
			p = append(p, s...)
		} else {
			j, _ := strconv.Atoi(template[:i])
			p = append(p, subs[j]...)
		}
		template = template[i+1:]
	}
	p = append(p, template...)
	return string(p)
}
