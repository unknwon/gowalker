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
	"time"

	"github.com/astaxie/beego"
)

const (
	HUMAN_REQUEST = iota
	ROBOT_REQUEST
	REFRESH_REQUEST
)

const (
	_TIME_DAY      = 24 * time.Hour
	_FETCH_TIMEOUT = 10 * time.Second
)

// CheckDoc checks the package documentation from the database or from the version
// control system as needed.
func CheckDoc(path string, requestType int) error {
	// Get the package documentation
	pdoc, err := getDoc(path)
	if err != nil {
		return err
	}

	needsCrawl := false
	// Check if it is in database or needs to crawl
	switch requestType {
	case HUMAN_REQUEST:
		needsCrawl = (pdoc == nil)
	case ROBOT_REQUEST:
		needsCrawl = (pdoc != nil) && pdoc.Updated.Add(_TIME_DAY).Before(time.Now())
	}

	if needsCrawl {
		// Fetch package from VCS
		c := make(chan crawlResult, 1)
		go func() {
			pdoc, err := crawlDoc(path)
			c <- crawlResult{pdoc, err}
		}()

		var err error
		select {
		case cr := <-c:
			if cr.err == nil {
				pdoc = cr.pdoc
				// Recurse crawl import packages

				// Save to database

				// Generate static page
				//generatePage(pdoc)
			}
			err = cr.err
		case <-time.After(_FETCH_TIMEOUT):
			err = errUpdateTimeout
		}
		if err != nil {
			if pdoc != nil {
				beego.Error("Serving %q from database after error: %v", path, err)
				err = nil
			} else if err == errUpdateTimeout {
				// Handle timeout on packages never seen before as not found.
				beego.Error("Serving %q as not found after timeout", path)
				err = errors.New("Status not found")
			}
			return err
		}
	}

	return nil
}
