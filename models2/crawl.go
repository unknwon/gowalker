// Copyright 2012 Gary Burd
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
	"flag"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/unknwon/gowalker/utils"
)

var nestedProjectPat = regexp.MustCompile(`/(?:github\.com|launchpad\.net|code\.google\.com/p|bitbucket\.org|labix\.org)/`)

// crawlDoc fetches the package documentation from the VCS and updates the database.
func crawlDoc(source string, path string, pdoc *Package, nextCrawl time.Time) (*Package, error) {
	message := []interface{}{source}
	defer func() {
		message = append(message, path)
		log.Println(message...)
	}()

	if !nextCrawl.IsZero() {
		d := time.Since(nextCrawl) / time.Hour
		if d > 0 {
			message = append(message, "late:", int64(d))
		}
	}

	etag := ""
	if pdoc != nil {
		etag = pdoc.Etag
		message = append(message, "etag:", etag)
	}

	now := time.Now()
	nextCrawl = now.Add(*maxAge)
	if strings.HasPrefix(path, "github.com/") {
		nextCrawl = now.Add(*maxAge * 7)
	}

	var err error
	if i := strings.Index(path, "/src/pkg/"); i > 0 && isGoRepoPath(path[i+len("/src/pkg/"):]) {
		// Go source tree mirror.
		pdoc = nil
		err = errors.New("Go source tree mirror.")
	} else if i := strings.Index(path, "/libgo/go/"); i > 0 && isGoRepoPath(path[i+len("/libgo/go/"):]) {
		// Go Frontend source tree mirror.
		pdoc = nil
		err = errors.New("Go Frontend source tree mirror.")
	} else if m := nestedProjectPat.FindStringIndex(path); m != nil && exists(path[m[0]+1:]) {
		pdoc = nil
		err = errors.New("Copy of other project.")
	} else {
		var pdocNew *Package
		pdocNew, err = getRepo(utils.HttpClient, path, etag)

		message = append(message, "fetch:", int64(time.Since(now)/time.Millisecond))

		// For timeout logic in client.go to work, we cannot leave connections idling. This is ugly.
		utils.HttpTransport.CloseIdleConnections()

		if err != ErrNotModified {
			pdoc = pdocNew
		}
	}

	switch {
	case err == nil:
		message = append(message, "save:", pdoc.Etag)
		pkg := PkgInfo{
			Path:     pdoc.ImportPath,
			Synopsis: pdoc.Synopsis,
			Updated:  nextCrawl}
		if err := savePkgInfo(&pkg); err != nil {
			log.Printf("ERROR savePkgInfo(%q): %v", path, err)
		}
	case err == ErrNotModified:
		message = append(message, "update")
		if err := setNextCrawl(path, nextCrawl); err != nil {
			log.Printf("ERROR setNextCrawl(%q): %v", path, err)
		}
	case IsNotFound(err):
		message = append(message, "notfound:", err)
		if err := deletePkg(path); err != nil {
			log.Printf("ERROR deletePkg(%q): %v", path, err)
		}
	default:
		message = append(message, "ERROR:", err)
		return nil, err
	}

	return pdoc, nil
}
