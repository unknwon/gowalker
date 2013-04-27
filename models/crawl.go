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
	"encoding/xml"
	"errors"
	"flag"
	"io"
	"net"
	"net/http"
	"path"
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
		pkg := PkgInfo{
			Path:      pdoc.ImportPath,
			Synopsis:  pdoc.Synopsis,
			Generated: time.Now().Local(),
			ProName:   pdoc.ProjectName}

		/* TODO:WORKING */

		if err := savePkgInfo(&pkg); err != nil {
			beego.Error("ERROR savePkgInfo(", path, "):", err)
		}
	case isNotFound(err):

		/* TODO */

		if err := deletePkg(path); err != nil {
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
	{githubPattern, "github.com/", getGithubDoc},
	{googlePattern, "code.google.com/", getGoogleDoc},
	{bitbucketPattern, "bitbucket.org/", getBitbucketDoc},
	{launchpadPattern, "launchpad.net/", getLaunchpadDoc},
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
	match, err := fetchMeta(client, importPath)
	if err != nil {
		return nil, err
	}

	if match["projectRoot"] != importPath {
		rootMatch, err := fetchMeta(client, match["projectRoot"])
		if err != nil {
			return nil, err
		}
		if rootMatch["projectRoot"] != match["projectRoot"] {
			return nil, NotFoundError{"Project root mismatch."}
		}
	}

	pdoc, err = getStatic(client, expand("{repo}{dir}", match))
	if err == errNoMatch {
		pdoc, err = getVCSDoc(client, match, "")
	}
	if err != nil {
		return nil, err
	}

	if pdoc != nil {
		pdoc.ImportPath = importPath
		pdoc.ProjectRoot = match["projectRoot"]
		pdoc.ProjectName = match["projectName"]
		pdoc.ProjectURL = match["projectURL"]
	}

	return pdoc, err
}

func fetchMeta(client *http.Client, importPath string) (map[string]string, error) {
	uri := importPath
	if !strings.Contains(uri, "/") {
		// Add slash for root of domain.
		uri = uri + "/"
	}
	uri = uri + "?go-get=1"

	scheme := "https"
	resp, err := client.Get(scheme + "://" + uri)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			resp.Body.Close()
		}
		scheme = "http"
		resp, err = client.Get(scheme + "://" + uri)
		if err != nil {
			return nil, &RemoteError{strings.SplitN(importPath, "/", 2)[0], err}
		}
	}
	defer resp.Body.Close()
	return parseMeta(scheme, importPath, resp.Body)
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}

func parseMeta(scheme, importPath string, r io.Reader) (map[string]string, error) {
	var match map[string]string

	d := xml.NewDecoder(r)
	d.Strict = false
metaScan:
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			break metaScan
		}
		switch t := t.(type) {
		case xml.EndElement:
			if strings.EqualFold(t.Name.Local, "head") {
				break metaScan
			}
		case xml.StartElement:
			if strings.EqualFold(t.Name.Local, "body") {
				break metaScan
			}
			if !strings.EqualFold(t.Name.Local, "meta") ||
				attrValue(t.Attr, "name") != "go-import" {
				continue metaScan
			}
			f := strings.Fields(attrValue(t.Attr, "content"))
			if len(f) != 3 ||
				!strings.HasPrefix(importPath, f[0]) ||
				!(len(importPath) == len(f[0]) || importPath[len(f[0])] == '/') {
				continue metaScan
			}
			if match != nil {
				return nil, NotFoundError{"More than one <meta> found at " + scheme + "://" + importPath}
			}

			projectRoot, vcs, repo := f[0], f[1], f[2]

			repo = strings.TrimSuffix(repo, "."+vcs)
			i := strings.Index(repo, "://")
			if i < 0 {
				return nil, NotFoundError{"Bad repo URL in <meta>."}
			}
			proto := repo[:i]
			repo = repo[i+len("://"):]

			match = map[string]string{
				// Used in getVCSDoc, same as vcsPattern matches.
				"importPath": importPath,
				"repo":       repo,
				"vcs":        vcs,
				"dir":        importPath[len(projectRoot):],

				// Used in getVCSDoc
				"scheme": proto,

				// Used in getDynamic.
				"projectRoot": projectRoot,
				"projectName": path.Base(projectRoot),
				"projectURL":  scheme + "://" + projectRoot,
			}
		}
	}
	if match == nil {
		return nil, NotFoundError{"<meta> not found."}
	}
	return match, nil
}
