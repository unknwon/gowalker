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
	"flag"
	"log"
	"time"
)

const (
	HUMAN_REQUEST = iota
	ROBOT_REQUEST
	QUERY_REQUEST
	REFRESH_REQUEST
)

var (
	robot           = flag.Bool("robot", false, "Robot mode")
	baseDir         = flag.String("base", defaultBase("github.com/garyburd/gopkgdoc/gddo-server"), "Base directory for templates and static files.")
	gzBaseDir       = flag.String("gzbase", "", "Base directory for compressed static files.")
	presentBaseDir  = flag.String("presentBase", defaultBase("code.google.com/p/go.talks/present"), "Base directory for templates and static files.")
	getTimeout      = flag.Duration("get_timeout", 8*time.Second, "Time to wait for package update from the VCS.")
	firstGetTimeout = flag.Duration("first_get_timeout", 5*time.Second, "Time to wait for first fetch of package from the VCS.")
	maxAge          = flag.Duration("max_age", 24*time.Hour, "Update package documents older than this age.")
	httpAddr        = flag.String("http", ":8080", "Listen for HTTP connections on this address")
	crawlInterval   = flag.Duration("crawl_interval", 0, "Package updater sleeps for this duration between package updates. Zero disables updates.")
	githubInterval  = flag.Duration("github_interval", 0, "Github updates crawler sleeps for this duration between fetches. Zero disables the crawler.")
	popularInterval = flag.Duration("popular_interval", 0, "Google Analytics fetcher sleeps for this duration between updates. Zero disables updates.")
	secretsPath     = flag.String("secrets", "secrets.json", "Path to file containing application ids and credentials for other services.")
	secrets         struct {
		UserAgent             string
		GithubId              string
		GithubSecret          string
		GAAccount             string
		ServiceAccountSecrets struct {
			Web struct {
				ClientEmail string `json:"client_email"`
				TokenURI    string `json:"token_uri"`
			}
		}
		ServiceAccountPEM      []string
		serviceAccountPEMBytes []byte
	}
)

type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}

type crawlResult struct {
	pdoc *Package
	err  error
}

func defaultBase(path string) string {
	p, err := build.Default.Import(path, "", build.FindOnly)
	if err != nil {
		return "."
	}
	return p.Dir
}

// GetDoc gets the package documentation from the database or from the version
// control system as needed.
func GetDoc(path string, requestType int) (*Package, []Package, error) {
	pdoc, pkgs, nextCrawl, err := GetPkgInfo(path)
	if err != nil {
		return nil, nil, err
	}

	needsCrawl := false
	switch requestType {
	case QUERY_REQUEST:
		needsCrawl = nextCrawl.IsZero() && len(pkgs) == 0
	case HUMAN_REQUEST:
		needsCrawl = nextCrawl.Before(time.Now())
	case ROBOT_REQUEST:
		needsCrawl = nextCrawl.IsZero() && len(pkgs) > 0
	}

	if needsCrawl {
		c := make(chan crawlResult, 1)
		go func() {
			pdoc, err := crawlDoc("web  ", path, pdoc, len(pkgs) > 0, nextCrawl)
			c <- crawlResult{pdoc, err}
		}()
		var err error
		timeout := *getTimeout
		if pdoc == nil {
			timeout = *firstGetTimeout
		}
		select {
		case rr := <-c:
			if rr.err == nil {
				pdoc = rr.pdoc
			}
			err = rr.err
		case <-time.After(timeout):
			err = errUpdateTimeout
		}
		if err != nil {
			if pdoc != nil {
				log.Printf("Serving %q from database after error: %v", path, err)
				err = nil
			} else if err == errUpdateTimeout {
				// Handle timeout on packages never seeen before as not found.
				log.Printf("Serving %q as not found after timeout", path)
				err = &web.Error{Status: web.StatusNotFound}
			}
		}
	}
	return pdoc, pkgs, err
}
