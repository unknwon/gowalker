// Copyright 2015 Unknwon
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

package doc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/Unknwon/com"
	"github.com/Unknwon/log"

	"github.com/Unknwon/gowalker/modules/httplib"
	"github.com/Unknwon/gowalker/modules/setting"
)

var (
	ErrInvalidRemotePath = errors.New("Invalid package remote path")
	ErrNoServiceMatch    = errors.New("Package remote path does not match any service")
)

type crawlResult struct {
	pdoc *Package
	err  error
}

// service represents a source code control service.
type service struct {
	pattern *regexp.Regexp
	prefix  string
	get     func(map[string]string, string) (*Package, error)
}

// services is the list of source code control services handled by gowalker.
var services = []*service{
	{githubPattern, "github.com/", getGithubDoc},
	// {googlePattern, "code.google.com/", getGoogleDoc},
	// {bitbucketPattern, "bitbucket.org/", getBitbucketDoc},
	// {launchpadPattern, "launchpad.net/", getLaunchpadDoc},
	// {oscPattern, "git.oschina.net/", getOSCDoc},
}

// getStatic gets a document from a statically known service.
// It returns ErrNoServiceMatch if the import path is not recognized.
func getStatic(importPath, etag string) (pdoc *Package, err error) {
	for _, s := range services {
		if s.get == nil || !strings.HasPrefix(importPath, s.prefix) {
			continue
		}
		m := s.pattern.FindStringSubmatch(importPath)
		if m == nil {
			if s.prefix != "" {
				log.Debug("Import path prefix matches known service, but regexp does not: %s", importPath)
				return nil, ErrInvalidRemotePath
			}
			continue
		}
		match := map[string]string{"importPath": importPath}
		for i, n := range s.pattern.SubexpNames() {
			if n != "" {
				match[n] = m[i]
			}
		}
		return s.get(match, etag)
	}
	return nil, ErrNoServiceMatch
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
				return nil, fmt.Errorf("More than one <meta> found at %s://%s", scheme, importPath)
			}

			projectRoot, vcs, repo := f[0], f[1], f[2]

			repo = strings.TrimSuffix(repo, "."+vcs)
			i := strings.Index(repo, "://")
			if i < 0 {
				return nil, errors.New("Bad repo URL in <meta>")
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
		return nil, errors.New("<meta> not found")
	}
	return match, nil
}

func fetchMeta(importPath string) (map[string]string, error) {
	uri := importPath
	if !strings.Contains(uri, "/") {
		// Add slash for root of domain.
		uri = uri + "/"
	}
	uri = uri + "?go-get=1"

	scheme := "https"
	resp, err := Client.Get(scheme + "://" + uri)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			resp.Body.Close()
		}
		scheme = "http"
		resp, err = Client.Get(scheme + "://" + uri)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()
	return parseMeta(scheme, importPath, resp.Body)
}

func getDynamic(importPath, etag string) (pdoc *Package, err error) {
	match, err := fetchMeta(importPath)
	if err != nil {
		return nil, err
	}

	if match["projectRoot"] != importPath {
		rootMatch, err := fetchMeta(match["projectRoot"])
		if err != nil {
			return nil, err
		}
		if rootMatch["projectRoot"] != match["projectRoot"] {
			return nil, errors.New("Project root mismatch")
		}
	}

	pdoc, err = getStatic(com.Expand("{repo}{dir}", match), etag)
	// if err == ErrNoServiceMatch {
	// 	pdoc, err = getVCSDoc(client, match, etag)
	// }
	if err != nil {
		return nil, err
	}

	if pdoc != nil {
		pdoc.ImportPath = importPath
		pdoc.ProjectPath = importPath
		// pdoc.ProjectName = match["projectName"]
	}

	return pdoc, err
}

func crawlDoc(importPath, etag string) (pdoc *Package, err error) {
	switch {
	case IsGoRepoPath(importPath):
		pdoc, err = getGolangDoc(importPath, etag)
	case IsValidRemotePath(importPath):
		pdoc, err = getStatic(importPath, etag)
		if err == ErrNoServiceMatch {
			pdoc, err = getDynamic(importPath, etag)
		}
	default:
		err = ErrInvalidRemotePath
	}

	if err != nil {
		return nil, err
	}

	// Render README.
	for name, content := range pdoc.Readme {
		p, err := httplib.Post("https://api.github.com/markdown/raw?"+setting.GitHubCredentials).
			Header("Content-Type", "text/plain").Body(content).Bytes()
		if err != nil {
			return nil, fmt.Errorf("error rendering README: %v", err)
		}
		pdoc.Readme[name] = p
	}

	return pdoc, nil
}
