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
	"runtime"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
)

type requestType int

const (
	RT_Human requestType = iota
	RT_Refresh
)

const (
	_TIME_DAY      = 24 * time.Hour    // Time duration of one day.
	_FETCH_TIMEOUT = 300 * time.Second // Fetch package timeout duration.
	_REFRESH_LIMIT = 5 * time.Minute   // Package fresh time limitation.
)

// CheckDoc returns 'Package' by given import path and tag,
// or fetch from the VCS and render as needed.
// It returns error when error occurs in the underlying functions.
func CheckDoc(broPath, tag string, rt requestType) (*hv.Package, error) {
	// Package documentation and crawl sign.
	pdoc, needsCrawl := &hv.Package{}, false

	// Trim prefix of standard library path.
	broPath = strings.TrimPrefix(broPath, "code.google.com/p/go/source/browse/src/pkg/")

	// Get the package info.
	pinfo, err := models.GetPkgInfo(broPath, tag)
	switch {
	case err != nil:
		// Error means it does not exist.
		beego.Trace("doc.CheckDoc -> ", err)

		// Check if it's "Error 1040: Too many connections"
		if strings.Contains(err.Error(), "Error 1040:") {
			break
		}
		fallthrough
	case err != nil || pinfo.PkgVer != hv.PACKAGE_VER:
		// If PACKAGE_VER does not match, refresh anyway.
		pinfo.Ptag = ""
		needsCrawl = true
	default:
		// Check request type.
		switch rt {
		case RT_Human:
		case RT_Refresh:
			if len(tag) > 0 {
				break // Things of Tag will not be changed.
			}

			// Check if the refresh operation is too frequently (within 5 minutes).
			needsCrawl = pinfo.Created.Add(_REFRESH_LIMIT).UTC().Before(time.Now().UTC())
			needsCrawl = true
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
			// TODO
			pdoc, err = crawlDoc(broPath, tag, pinfo)
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

		if pdoc == nil {
			return nil, err
		}

		if err == nil {
			pdoc.IsNeedRender = true
			beego.Info("doc.CheckDoc(", pdoc.ImportPath, tag, "), Goroutine #", runtime.NumGoroutine())
		} else {
			switch {
			case err == errNotModified:
				beego.Info("Serving(", broPath, ")without modified")
				pdoc = &hv.Package{}
				pinfo.Created = time.Now().UTC()
				pdoc.PkgInfo = pinfo
				return pdoc, nil
			case pdoc != nil && len(pdoc.ImportPath) > 0:
				return pdoc, err
			case err == errUpdateTimeout:
				// Handle timeout on packages never seen before as not found.
				beego.Error("Serving(", broPath, ")as not found after timeout")
				return nil, errors.New("doc.CheckDoc -> " + err.Error())
			}
		}
	} else {
		pdoc.PkgInfo = pinfo
	}

	return pdoc, err
}
