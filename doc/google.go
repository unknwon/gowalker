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
	_ "strconv"
	"strings"

	"github.com/Unknwon/gowalker/utils"
)

var (
	googleRepoRe     = regexp.MustCompile(`id="checkoutcmd">(hg|git|svn)`)
	googleRevisionRe = regexp.MustCompile(`<h2>(?:[^ ]+ - )?Revision *([^:]+):`)
	googleEtagRe     = regexp.MustCompile(`^(hg|git|svn)-`)
	googleFileRe     = regexp.MustCompile(`<li><a href="([^"/]+)"`)
	googleDirRe      = regexp.MustCompile(`<li><a href="([^".]+)"`)
	googlePattern    = regexp.MustCompile(`^code\.google\.com/p/(?P<repo>[a-z0-9\-]+)(:?\.(?P<subrepo>[a-z0-9\-]+))?(?P<dir>/[a-z0-9A-Z_.\-/]+)?$`)
)

func getStandardDoc(client *http.Client, importPath string, savedEtag string) (pdoc *Package, err error) {
	p, err := httpGetBytes(client, "http://go.googlecode.com/hg-history/release/src/pkg/"+importPath+"/", nil)
	if err != nil {
		return nil, err
	}

	// Check revision tag.
	var etag string
	if m := googleRevisionRe.FindSubmatch(p); m == nil {
		return nil, errors.New("doc.getStandardDoc(): Could not find revision for " + importPath)
	} else {
		etag = string(m[1])
		if etag == savedEtag {
			return nil, errNotModified
		}
	}

	// Get source file data.
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

	// Fetch file from VCS.
	if err := fetchFiles(client, files, nil); err != nil {
		return nil, err
	}

	/* TODO:WORKING */

	/*w := &walker{
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

	return w.build(files)*/
	return nil, errors.New("doc.getStandardDoc(): Test Error.")
}
