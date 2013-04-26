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
	"flag"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/unknwon/gowalker/utils"
)

var (
	dialTimeout  = flag.Duration("dial_timeout", 5*time.Second, "Timeout for dialing an HTTP connection.")
	readTimeout  = flag.Duration("read_timeout", 10*time.Second, "Timeoout for reading an HTTP response.")
	writeTimeout = flag.Duration("write_timeout", 5*time.Second, "Timeout writing an HTTP request.")
)

type timeoutConn struct {
	net.Conn
}

func (c *timeoutConn) Read(p []byte) (int, error) {
	return c.Conn.Read(p)
}

func (c *timeoutConn) Write(p []byte) (int, error) {
	// Reset timeouts when writing a request.
	c.Conn.SetWriteDeadline(time.Now().Add(*readTimeout))
	c.Conn.SetWriteDeadline(time.Now().Add(*writeTimeout))
	return c.Conn.Write(p)
}

func timeoutDial(network, addr string) (net.Conn, error) {
	c, err := net.DialTimeout(network, addr, *dialTimeout)
	if err != nil {
		return nil, err
	}
	return &timeoutConn{Conn: c}, nil
}

var (
	httpTransport = &http.Transport{Dial: timeoutDial}
	httpClient    = &http.Client{Transport: httpTransport}
)

type crawlResult struct {
	pdoc *Package
	err  error
}

var nestedProjectPat = regexp.MustCompile(`/(?:github\.com|launchpad\.net|code\.google\.com/p|bitbucket\.org|labix\.org)/`)

func crawlDoc(path string) (*Package, error) {
	var pdoc *Package
	var err error

	if i := strings.Index(path, "/libgo/go/"); i > 0 && utils.IsGoRepoPath(path[i+len("/libgo/go/"):]) {
		// Go Frontend source tree mirror.
		pdoc = nil
		err = errors.New("Go Frontend source tree mirror.")
	} else {
		var pdocNew *Package

		/* TODO:WORKING */

		pdocNew, err = getRepo(httpClient, path)

		// For timeout logic in client.go to work, we cannot leave connections idling. This is ugly.
		httpTransport.CloseIdleConnections()

		if err != errNotModified {
			pdoc = pdocNew
		}
	}

	// if i := strings.Index(path, "/src/pkg/"); i > 0 && utils.IsGoRepoPath(path[i+len("/src/pkg/"):]) {
	// 	// Go source tree mirror.
	// 	pdoc = nil
	// 	err = errors.New("Go source tree mirror.")
	// } else if m := nestedProjectPat.FindStringIndex(path); m != nil && exists(path[m[0]+1:]) {
	// 	pdoc = nil
	// 	err = errors.New("Copy of other project.")
	// }

	switch {
	case err == nil:
		pkg := pkgInfo{
			Path:     pdoc.ImportPath,
			Synopsis: pdoc.Synopsis,
			Updated:  time.Now()}

		/* TODO */

		if err := savePkgInfo(&pkg); err != nil {

			/* TODO */

			beego.Error("ERROR savePkgInfo(", path, "):", err)
		}
	case isNotFound(err):

		/* TODO */

		if err := deletePkg(path); err != nil {

			/* TODO */

			beego.Error("ERROR deletePkg(", path, "):", err)
		}
	}

	return pdoc, err
}

func getRepo(client *http.Client, importPath string) (pdoc *Package, err error) {
	const VER_PREFIX = PackageVersion + "-"
	i := strings.Index(importPath, "/src/pkg/")

	switch {
	case utils.IsGoRepoPath(importPath[i+len("/src/pkg/"):]):
		pdoc, err = getStandardDoc(client, importPath[i+len("/src/pkg/"):])
	case utils.IsValidRemotePath(importPath):
		pdoc, err = getStatic(client, importPath)
		if err == errNoMatch {

			/* TODO */

			pdoc, err = getDynamic(client, importPath)
		}
	default:
		err = errNoMatch
	}

	if err == errNoMatch {
		err = NotFoundError{"Import path not valid:"}
	}

	return pdoc, err
}

// service represents a source code control service.
type service struct {
	pattern *regexp.Regexp
	prefix  string
	get     func(*http.Client, map[string]string, string) (*Package, error)
}

// services is the list of source code control services handled by gopkgdoc.
var services = []*service{
	{googlePattern, "code.google.com/", getGoogleDoc},
	{vcsPattern, "", getVCSDoc},
}

// getStatic gets a document from a statically known service. getStatic
// returns errNoMatch if the import path is not recognized.
func getStatic(client *http.Client, importPath string) (pdoc *Package, err error) {
	for _, s := range services {
		if s.get == nil || !strings.HasPrefix(importPath, s.prefix) {
			continue
		}
		m := s.pattern.FindStringSubmatch(importPath)
		if m == nil {
			if s.prefix != "" {
				return nil, NotFoundError{"Import path prefix matches known service, but regexp does not."}
			}
			continue
		}
		match := map[string]string{"importPath": importPath}
		for i, n := range s.pattern.SubexpNames() {
			if n != "" {
				match[n] = m[i]
			}
		}
		return s.get(client, match, "")
	}
	return nil, errNoMatch
}

func getDynamic(client *http.Client, importPath string) (pdoc *Package, err error) {
	return pdoc, errors.New("Test Error: getDynamic")
}
