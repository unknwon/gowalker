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
	"time"

	"github.com/Unknwon/gowalker/models"
	"github.com/astaxie/beego"
)

const (
	HUMAN_REQUEST = iota
	REFRESH_REQUEST
)

const (
	_TIME_DAY      = 24 * time.Hour
	_FETCH_TIMEOUT = 10 * time.Second
)

// CheckDoc checks the project documentation from the database or from the version
// control system as needed.
func CheckDoc(path string, requestType int) (*Package, error) {
	// Package documentation and crawl sign.
	pdoc, needsCrawl := &Package{}, false

	// Get the package documentation from database.
	pinfo, err := models.GetPkgInfo(path)

	if err != nil {
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
				needsCrawl = pinfo.Created.Add(_TIME_DAY).Local().Before(time.Now().Local())
			}
		case REFRESH_REQUEST:
			// Check if the documentation is too frequently (within 1 hour).
			needsCrawl = pinfo.Created.Add(_TIME_DAY).Local().Before(time.Now().Local())
			if !needsCrawl {
				return &Package{}, errors.New("doc.CheckDoc(): Package cannot be refreshed until" +
					pdoc.Created.Add(time.Hour).Local().String())
			}
		}
	}

	if needsCrawl {
		// Fetch package from VCS.
		c := make(chan crawlResult, 1)
		go func() {

			/* TODO:WORKING */

			pdoc, err = crawlDoc(path, pinfo.Etag)
			c <- crawlResult{pdoc, err}
		}()

		var err error
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
			if pdoc != nil {
				beego.Error("Serving", path, "from database after error: ", err)
				err = nil
			} else if err == errUpdateTimeout {
				// Handle timeout on packages never seen before as not found.
				beego.Error("Serving", path, "as not found after timeout")
				err = errors.New("Status not found")
			}
			return nil, err
		}
	}

	return pdoc, nil
}
