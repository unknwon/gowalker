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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/Unknwon/com"
	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
)

type byHash []byte

func (p byHash) Len() int { return len(p) / md5.Size }
func (p byHash) Less(i, j int) bool {
	return -1 == bytes.Compare(p[i*md5.Size:(i+1)*md5.Size], p[j*md5.Size:(j+1)*md5.Size])
}
func (p byHash) Swap(i, j int) {
	var temp [md5.Size]byte
	copy(temp[:], p[i*md5.Size:])
	copy(p[i*md5.Size:(i+1)*md5.Size], p[j*md5.Size:])
	copy(p[j*md5.Size:], temp[:])
}

var launchpadPattern = regexp.MustCompile(`^launchpad\.net/(?P<repo>(?P<project>[a-z0-9A-Z_.\-]+)(?P<series>/[a-z0-9A-Z_.\-]+)?|~[a-z0-9A-Z_.\-]+/(\+junk|[a-z0-9A-Z_.\-]+)/[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]+)*$`)

func getLaunchpadDoc(client *http.Client, match map[string]string, tag, savedEtag string) (*hv.Package, error) {

	if match["project"] != "" && match["series"] != "" {
		rc, err := com.HttpGet(client, com.Expand("https://code.launchpad.net/{project}{series}/.bzr/branch-format", match), nil)
		switch {
		case err == nil:
			rc.Close()
			// The structure of the import path is launchpad.net/{root}/{dir}.
		case isNotFound(err):
			// The structure of the import path is is launchpad.net/{project}/{dir}.
			match["repo"] = match["project"]
			match["dir"] = com.Expand("{series}{dir}", match)
		default:
			return nil, err
		}
	}

	// Scrape the repo browser to find the project revision and individual Go files.
	p, err := com.HttpGetBytes(client, com.Expand("https://bazaar.launchpad.net/+branch/{repo}/tarball", match), nil)
	if err != nil {
		return nil, err
	}

	// Get source file data.
	gzr, err := gzip.NewReader(bytes.NewReader(p))
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var hash []byte
	dirPrefix := com.Expand("+branch/{repo}{dir}/", match)
	preLen := len(dirPrefix)

	isGoPro := false // Indicates whether it's a Go project.
	isRootPath := match["importPath"] == utils.GetProjectPath(match["importPath"])
	dirs := make([]string, 0, 3)
	files := make([]com.RawFile, 0, 5)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Skip directories and files in wrong directories, get them later.
		if strings.HasSuffix(h.Name, "/") || !strings.HasPrefix(h.Name, dirPrefix) {
			continue
		}

		d, f := path.Split(h.Name)
		if utils.IsDocFile(f) && utils.FilterDirName(d) {
			// Check if it's a Go file.
			if isRootPath && !isGoPro && strings.HasSuffix(f, ".go") {
				isGoPro = true
			}

			// Get file from archive.
			b := make([]byte, h.Size)
			if _, err := io.ReadFull(tr, b); err != nil {
				return nil, err
			}

			m := md5.New()
			m.Write(b)
			hash = m.Sum(hash)

			// Check if file is in the directory that is corresponding to import path.
			if d == dirPrefix {
				// Yes.
				if !isRootPath && !isGoPro && strings.HasSuffix(f, ".go") {
					isGoPro = true
				}
				files = append(files, &hv.Source{
					SrcName:   f,
					BrowseUrl: com.Expand("bazaar.launchpad.net/+branch/{repo}/view/head:{dir}/{0}", match, f),
					SrcData:   b})
			} else {
				sd, _ := path.Split(d[preLen:])
				sd = strings.TrimSuffix(sd, "/")
				if !checkDir(sd, dirs) {
					dirs = append(dirs, sd)
				}
			}
		}
	}

	if !isGoPro {
		return nil, com.NotFoundError{"Cannot find Go files, it's not a Go project."}
	}

	if len(files) == 0 && len(dirs) == 0 {
		return nil, com.NotFoundError{"Directory tree does not contain Go files and subdirs."}
	}

	sort.Sort(byHash(hash))
	m := md5.New()
	m.Write(hash)
	hash = m.Sum(hash[:0])
	etag := hex.EncodeToString(hash)
	if etag == savedEtag {
		return nil, errNotModified
	}

	// Start generating data.
	w := &hv.Walker{
		LineFmt: "#L%d",
		Pdoc: &hv.Package{
			PkgInfo: &hv.PkgInfo{
				ImportPath:  match["importPath"],
				ProjectName: match["repo"],
				ProjectPath: com.Expand("bazaar.launchpad.net/+branch/{repo}/files", match),
				ViewDirPath: com.Expand("bazaar.launchpad.net/+branch/{repo}/files/head:{dir}/", match),
				Ptag:        etag,
				Vcs:         "Launchpad",
			},
			PkgDecl: &hv.PkgDecl{
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
		return nil, errors.New("doc.getLaunchpadDoc(" + match["importPath"] + ") -> Fail to build: " + err.Error())
	}

	if len(tag) == 0 && w.Pdoc.IsCmd {
		err = generateHv(match["importPath"], srcMap)
	}

	return pdoc, err
}
