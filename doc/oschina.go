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
	"archive/zip"
	"bytes"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/Unknwon/gowalker/utils"
)

var (
	oscTagRe   = regexp.MustCompile(`/repository/archive\?ref=(.*)">`)
	oscPattern = regexp.MustCompile(`^git\.oschina\.net/(?P<owner>[a-z0-9A-Z_.\-]+)/(?P<repo>[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]*)?$`)
)

func getOSCDoc(client *http.Client, match map[string]string, tag, savedEtag string) (*Package, error) {
	if len(tag) == 0 {
		match["tag"] = "master"
	} else {
		match["tag"] = tag
	}

	match["projectRoot"] = utils.GetProjectPath(match["importPath"])
	// Download zip.
	p, err := httpGetBytes(client, expand("http://{projectRoot}/repository/archive?ref={tag}", match), nil)
	if err != nil {
		return nil, errors.New("doc.getOSCDoc(" + match["importPath"] + ") -> " + err.Error())
	}

	r, err := zip.NewReader(bytes.NewReader(p), int64(len(p)))
	if err != nil {
		return nil, errors.New("doc.getOSCDoc(" + match["importPath"] + ") -> create zip: " + err.Error())
	}

	commit := r.Comment
	// Get source file data and subdirectories.
	nameLen := len(match["repo"])
	dirLen := nameLen + len(match["dir"])
	dirs := make([]string, 0, 5)
	files := make([]*source, 0, 5)
	for _, f := range r.File {
		fileName := f.FileInfo().Name()
		if len(fileName) < dirLen {
			continue
		}

		// File.
		if utils.IsDocFile(fileName[dirLen+1:]) && strings.LastIndex(fileName, "/") == dirLen {
			// Get file from archive.
			rc, err := f.Open()
			if err != nil {
				return nil, errors.New("doc.getOSCDoc(" + match["importPath"] + ") -> open file: " + err.Error())
			}

			p := make([]byte, f.FileInfo().Size())
			rc.Read(p)
			if err != nil {
				return nil, errors.New("doc.getOSCDoc(" + match["importPath"] + ") -> read file: " + err.Error())
			}

			files = append(files, &source{
				name:      fileName[dirLen+1:],
				browseURL: expand("http://git.oschina.net/{owner}/{repo}/blob/{tag}/{0}", match, fileName[nameLen+1:]),
				rawURL:    expand("http://git.oschina.net/{owner}/{repo}/raw/{tag}/{0}", match, fileName[dirLen+1:]),
				data:      p,
			})
			continue
		}

		// Directory.
		if strings.HasSuffix(fileName, "/") && utils.FilterFileName(fileName[dirLen+1:]) {
			dirs = append(dirs, fileName[dirLen+1:])
		}
	}

	if len(files) == 0 && len(dirs) == 0 {
		return nil, NotFoundError{"Directory tree does not contain Go files and subdirs."}
	}

	// Get all tags.
	tags := getOSCTags(client, match["importPath"])

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
		},
	}
	return w.build(files)
}

func getOSCTags(client *http.Client, importPath string) []string {
	p, err := httpGetBytes(client, "http://"+utils.GetProjectPath(importPath)+"/repository/tags", nil)
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
