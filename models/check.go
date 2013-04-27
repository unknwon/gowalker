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
	"os"
	"time"

	"github.com/astaxie/beego"
	"github.com/unknwon/gowalker/utils"
)

const (
	HUMAN_REQUEST = iota
	REFRESH_REQUEST
)

const (
	_TIME_DAY      = 24 * time.Hour
	_FETCH_TIMEOUT = 10 * time.Second
)

// CheckDoc checks the package documentation from the database or from the version
// control system as needed.
func CheckDoc(path string, requestType int) (pdoc *Package, err error) {
	needsCrawl := false
	// Get the package documentation from database
	pdoc, err = getDoc(path)
	if utils.IsGoRepoPath(path) {
		path = "code.google.com/p/go/source/browse/src/pkg/" + path
	}

	switch requestType {
	case HUMAN_REQUEST:
		if err != nil {
			needsCrawl = true
			os.Remove("./docs/" + path + ".html")
		} else {
			// Check static file
			needsCrawl = !utils.IsExist("./docs/" + path + ".html")
		}
	case REFRESH_REQUEST:
		if err != nil {
			needsCrawl = true
		} else {
			needsCrawl = (pdoc != nil) && pdoc.Updated.Add(time.Hour).Before(time.Now().UTC())
			if !needsCrawl {
				return nil, errors.New(pdoc.Updated.Add(time.Hour).String())
			}
			os.Remove("./docs/" + path + ".html")
		}
	}

	if needsCrawl {
		// Fetch package from VCS
		c := make(chan crawlResult, 1)
		go func() {

			/* TODO:WORKING */

			pdoc, err = crawlDoc(path)
			c <- crawlResult{pdoc, err}
		}()

		var err error
		select {
		case cr := <-c:
			if cr.err == nil {
				pdoc = cr.pdoc

				/* TODO */

				// Recurse crawl import packages

				// Save to database
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
				beego.Error("Serving ", path, "as not found after timeout")
				err = errors.New("Status not found")
			}
			return nil, err
		}
	}

	return pdoc, nil
}
