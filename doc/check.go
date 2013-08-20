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

// CheckDoc returns 'Package' by given import path and tag,
// or fetch from the VCS as needed.
// It returns error when error occurs in the underlying functions.
func CheckDoc(path, tag string, requestType int) (*Package, error) {
	// Package documentation and crawl sign.
	pdoc, needsCrawl := &Package{}, false

	// Trim prefix of standard library path.
	if i := strings.Index(path, "/src/pkg/"); i > -1 {
		path = path[i+len("/src/pkg/"):]
	}

	// For code.google.com.
	path = strings.Replace(path, "source/browse/", "", 1)

	// Get the package info.
	pinfo, err := models.GetPkgInfo(path, tag)
	switch {
	case err != nil:
		// Error means it does not exist.
		beego.Trace("doc.CheckDoc -> ", err)
		fallthrough
	case err != nil || !strings.HasPrefix(pinfo.Etag, PACKAGE_VER):
		// If PACKAGE_VER does not match, refresh anyway.
		needsCrawl = true
	default:
		// Check request type.
		switch requestType {
		case HUMAN_REQUEST:
		case REFRESH_REQUEST:
			if len(tag) > 0 {
				break // Things of Tag will not be changed.
			}

			// Check if the refresh operation is too frequently (within 5 minutes).
			needsCrawl = pinfo.Created.Add(_REFRESH_LIMIT).UTC().Before(time.Now().UTC())
			if !needsCrawl {
				// Return limit time information as error message.
				return nil, errors.New(pinfo.Created.Add(_REFRESH_LIMIT).UTC().String())
			}
		}
	}

	if needsCrawl {
		// Fetch package from VCS.
		c := make(chan crawlResult, 1)
		go func() {
			pdoc, err = crawlDoc(path, tag, pinfo)
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
				beego.Info("Serving(", path, ")without modified")
				pdoc = &Package{}
				pinfo.Created = time.Now().UTC()
				assginPkgInfo(pdoc, pinfo)
				return pdoc, nil
			case pdoc != nil && len(pdoc.ImportPath) > 0:
				return pdoc, err
			case err == errUpdateTimeout:
				// Handle timeout on packages never seen before as not found.
				beego.Error("Serving(", path, ")as not found after timeout")
				return nil, errors.New("doc.CheckDoc -> " + err.Error())
			}
		}
	} else {
		assginPkgInfo(pdoc, pinfo)
	}

	return pdoc, err
}

func assginPkgInfo(pdoc *Package, pinfo *models.PkgInfo) {
	// Assgin package information
	pdoc.Id = pinfo.Id
	pdoc.ImportPath = pinfo.Path
	pdoc.ProjectName = pinfo.ProName
	pdoc.Synopsis = pinfo.Synopsis
	pdoc.IsCmd = pinfo.IsCmd
	pdoc.Tags = strings.Split(pinfo.Tags, "|||")
	pdoc.Views = pinfo.Views
	pdoc.Created = pinfo.Created
	pdoc.Rank = pinfo.Rank
	pdoc.Etag = pinfo.Etag
	pdoc.Labels = pinfo.Labels
	pdoc.ImportedNum = pinfo.ImportedNum
	pdoc.ImportPid = pinfo.ImportPid
	pdoc.Note = pinfo.Note
}
