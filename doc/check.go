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

// Package doc implemented fetch projects from VCS and genarate AST.
package doc

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/models"
	"github.com/astaxie/beego"
)

const (
	// Request type.
	HUMAN_REQUEST = iota
	REFRESH_REQUEST
)

const (
	_TIME_DAY      = 24 * time.Hour   // Time duration of one day.
	_FETCH_TIMEOUT = 10 * time.Second // Fetch package timeout duration.
	_REFRESH_LIMIT = 5 * time.Minute  // Package fresh time limitation.
)

// CheckDoc checks the project documentation from the database or from the version
// control system as needed.
func CheckDoc(path string, requestType int) (*Package, error) {
	// Package documentation and crawl sign.
	pdoc, needsCrawl := &Package{}, false

	// Reduce standard library path.
	if i := strings.Index(path, "/src/pkg/"); i > -1 {
		path = path[i+len("/src/pkg/"):]
	}

	// Get the package documentation from database.
	pinfo, err := models.GetPkgInfo(path)

	// If PACKAGE_VER does not match, refresh anyway.
	if err != nil || !strings.HasPrefix(pinfo.Etag, PACKAGE_VER) {
		// No package information in database.
		needsCrawl = true
	} else {
		// Check request type.
		switch requestType {
		case HUMAN_REQUEST:
			// Error means it does not exist.
			if err != nil {
				needsCrawl = true
			} else {
				// Check if the documentation is too old (1 day ago).
				needsCrawl = pinfo.Updated.Add(_TIME_DAY).UTC().Before(time.Now().UTC())
			}
		case REFRESH_REQUEST:
			// Check if the documentation is too frequently (within 5 minutes).
			needsCrawl = pinfo.Updated.Add(_REFRESH_LIMIT).UTC().Before(time.Now().UTC())
			if !needsCrawl {
				// Return error messages as limit time information.
				return nil, errors.New(pinfo.Updated.Add(_REFRESH_LIMIT).UTC().String())
			}
		}
	}

	if needsCrawl {
		// Fetch package from VCS.
		c := make(chan crawlResult, 1)
		go func() {
			pdoc, err = crawlDoc(path, pinfo)
			c <- crawlResult{pdoc, err}
		}()

		select {
		case cr := <-c:
			if cr.err == nil {
				pdoc = cr.pdoc
			}
			err = cr.err
		case <-time.After(_FETCH_TIMEOUT):
			err = errUpdateTimeout
		}

		if err != nil {
			switch {
			case err == errNotModified:
				beego.Info("Serving", path, "without modified")
				pdoc = &Package{}
				pinfo.Updated = time.Now().UTC()
				assginPkgInfo(pdoc, pinfo)
				return pdoc, nil
			case pdoc != nil:
				beego.Error("Serving", path, "with error:", err)
				return pdoc, nil
			case err == errUpdateTimeout:
				// Handle timeout on packages never seen before as not found.
				beego.Error("Serving", path, "as not found after timeout")
				return nil, errors.New("doc.CheckDoc -> Status not found")
			}
		}
	} else {
		assginPkgInfo(pdoc, pinfo)
	}

	return nil, errors.New(fmt.Sprint("doc.CheckDoc ->", err))
}

func assginPkgInfo(pdoc *Package, pinfo *models.PkgInfo) {
	// Assgin package information
	pdoc.ImportPath = pinfo.Path
	pdoc.Branchs = pinfo.Branchs
	pdoc.Tags = pinfo.Tags
	pdoc.IsCmd = pinfo.IsCmd
	pdoc.Synopsis = pinfo.Synopsis
	pdoc.Views = pinfo.Views
	pdoc.Updated = pinfo.Updated
	pdoc.ProjectName = pinfo.ProName
	pdoc.Etag = pinfo.Etag
	pdoc.Labels = pinfo.Labels
	pdoc.ImportedNum = pinfo.ImportedNum
	pdoc.ImportPid = pinfo.ImportPid
}
