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

package doc

import (
	"bytes"
	"encoding/base32"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

type crawlResult struct {
	pdoc *Package
	err  error
}

// crawlDoc fetchs package from VCS.
func crawlDoc(path string, pinfo *models.PkgInfo) (pdoc *Package, err error) {
	// I have no idea what the fuck does this mean.
	if i := strings.Index(path, "/libgo/go/"); i > 0 && utils.IsGoRepoPath(path[i+len("/libgo/go/"):]) {
		// Go Frontend source tree mirror.
		pdoc = nil
		err = errors.New("Go Frontend source tree mirror.")
	} else {
		var pdocNew *Package
		pdocNew, err = getRepo(httpClient, path, pinfo.Etag)

		// For timeout logic in client.go to work, we cannot leave connections idling. This is ugly.
		httpTransport.CloseIdleConnections()

		if err != errNotModified && pdocNew != nil {
			pdoc = pdocNew
			pdoc.Views = pinfo.Views
			pdoc.Tags = pinfo.Tags
			pdoc.ImportPid = pinfo.ImportPid
			pdoc.ImportedNum = pinfo.ImportedNum
		}
	}

	switch {
	case err == nil:
		if err = SaveProject(pdoc, pinfo); err != nil {
			beego.Error("doc.SaveProject(", path, "):", err)
		}
	case isNotFound(err):
		// We do not need to delete standard library, so here is fine.
		if err = models.DeleteProject(path); err != nil {
			beego.Error("doc.DeleteProject(", path, "):", err)
		}
	}
	return pdoc, err
}

// getRepo downloads package data.
func getRepo(client *http.Client, importPath string, etag string) (pdoc *Package, err error) {
	const VER_PREFIX = PACKAGE_VER + "-"

	// Check version prefix.
	if strings.HasPrefix(etag, VER_PREFIX) {
		etag = etag[len(VER_PREFIX):]
	} else {
		etag = ""
	}

	switch {
	case utils.IsGoRepoPath(importPath):
		pdoc, err = getStandardDoc(client, importPath, etag)
	case utils.IsValidRemotePath(importPath):
		pdoc, err = getStatic(client, importPath, etag)
		if err == errNoMatch {
			pdoc, err = getDynamic(client, importPath, etag)
		}
	default:
		return nil, errors.New("doc.getRepo(): No match: " + importPath)
	}

	// Save revision tag.
	if pdoc != nil {
		pdoc.Etag = VER_PREFIX + pdoc.Etag
	}

	return pdoc, err
}

// SaveProject saves project information to database.
func SaveProject(pdoc *Package, info *models.PkgInfo) error {

	// Save package information.
	pinfo := &models.PkgInfo{
		Path:        pdoc.ImportPath,
		Synopsis:    pdoc.Synopsis,
		Created:     time.Now().UTC(),
		ViewedTime:  time.Now().UTC().Unix(),
		ProName:     pdoc.ProjectName,
		Views:       info.Views,
		Etag:        pdoc.Etag,
		Tags:        pdoc.Tags,
		ImportedNum: info.ImportedNum,
		ImportPid:   info.ImportPid,
	}

	// Save package declaration.
	pdecl := &models.PkgDecl{
		Path:      pdoc.ImportPath,
		Doc:       pdoc.Doc,
		Truncated: pdoc.Truncated,
		Goos:      pdoc.GOOS,
		Goarch:    pdoc.GOARCH,
	}
	var buf bytes.Buffer

	// Consts.
	addValues(&buf, &pdecl.Consts, pdoc.Consts)
	buf.Reset()
	addValues(&buf, &pdecl.Iconsts, pdoc.Iconsts)

	// Variables.
	buf.Reset()
	addValues(&buf, &pdecl.Vars, pdoc.Vars)
	buf.Reset()
	addValues(&buf, &pdecl.Ivars, pdoc.Ivars)

	// Functions.
	buf.Reset()
	addFuncs(&buf, &pdecl.Funcs, pdoc.Funcs)
	buf.Reset()
	addFuncs(&buf, &pdecl.Ifuncs, pdoc.Ifuncs)

	// Types.
	buf.Reset()
	for _, v := range pdoc.Types {
		buf.WriteString(v.Name)
		buf.WriteString("&T#")
		buf.WriteString(v.Doc)
		buf.WriteString("&T#")
		buf.WriteString(v.Decl)
		buf.WriteString("&T#")
		buf.WriteString(v.URL)
		buf.WriteString("&$#")
		// Functions.
		for _, m := range v.Funcs {
			buf.WriteString(m.Name)
			buf.WriteString("&F#")
			buf.WriteString(m.Doc)
			buf.WriteString("&F#")
			buf.WriteString(m.Decl)
			buf.WriteString("&F#")
			buf.WriteString(m.URL)
			buf.WriteString("&F#")
			buf.WriteString(*codeEncode(&m.Code))
			buf.WriteString("&M#")
		}
		buf.WriteString("&$#")
		for _, m := range v.IFuncs {
			buf.WriteString(m.Name)
			buf.WriteString("&F#")
			buf.WriteString(m.Doc)
			buf.WriteString("&F#")
			buf.WriteString(m.Decl)
			buf.WriteString("&F#")
			buf.WriteString(m.URL)
			buf.WriteString("&F#")
			buf.WriteString(*codeEncode(&m.Code))
			buf.WriteString("&M#")
		}
		buf.WriteString("&$#")

		// Methods.
		for _, m := range v.Methods {
			buf.WriteString(m.Name)
			buf.WriteString("&F#")
			buf.WriteString(m.Doc)
			buf.WriteString("&F#")
			buf.WriteString(m.Decl)
			buf.WriteString("&F#")
			buf.WriteString(m.URL)
			buf.WriteString("&F#")
			buf.WriteString(*codeEncode(&m.Code))
			buf.WriteString("&M#")
		}
		buf.WriteString("&$#")
		for _, m := range v.IMethods {
			buf.WriteString(m.Name)
			buf.WriteString("&F#")
			buf.WriteString(m.Doc)
			buf.WriteString("&F#")
			buf.WriteString(m.Decl)
			buf.WriteString("&F#")
			buf.WriteString(m.URL)
			buf.WriteString("&F#")
			buf.WriteString(*codeEncode(&m.Code))
			buf.WriteString("&M#")
		}
		buf.WriteString("&##")
	}
	pdecl.Types = buf.String()

	// Notes.
	buf.Reset()
	for _, s := range pdoc.Notes {
		buf.WriteString(s)
		buf.WriteString("|")
	}
	pdecl.Notes = buf.String()

	// Dirs.
	buf.Reset()
	for _, s := range pdoc.Dirs {
		buf.WriteString(s)
		buf.WriteString("|")
	}
	pdecl.Dirs = buf.String()

	// Imports.
	buf.Reset()
	for _, s := range pdoc.Imports {
		buf.WriteString(s)
		buf.WriteString("|")
	}
	pdecl.Imports = buf.String()

	buf.Reset()
	for _, s := range pdoc.TestImports {
		buf.WriteString(s)
		buf.WriteString("|")
	}
	pdecl.TestImports = buf.String()

	// Files.
	buf.Reset()
	for _, s := range pdoc.Files {
		buf.WriteString(s)
		buf.WriteString("|")
	}
	pdecl.Files = buf.String()

	buf.Reset()
	for _, s := range pdoc.TestFiles {
		buf.WriteString(s)
		buf.WriteString("|")
	}
	pdecl.TestFiles = buf.String()

	// Save package documentation.
	doc := &models.PkgDoc{
		Path: pdoc.ImportPath,
		Lang: "zh",
	}

	// Documentataion.
	buf.Reset()
	for k, v := range pdoc.TDoc {
		buf.WriteString(k)
		buf.WriteByte(':')
		buf.WriteString(v)
		buf.WriteString("&$#")
	}

	for k, v := range pdoc.IDoc {
		buf.WriteString(k)
		buf.WriteByte(':')
		buf.WriteString(v)
		buf.WriteString("&$#")
	}
	doc.Doc = buf.String()

	err := models.SaveProject(pinfo, pdecl, doc, pdoc.Imports)
	return err
}

func addValues(buf *bytes.Buffer, pvals *string, vals []*Value) {
	for _, v := range vals {
		buf.WriteString(v.Name)
		buf.WriteString("&V#")
		buf.WriteString(v.Doc)
		buf.WriteString("&V#")
		buf.WriteString(v.Decl)
		buf.WriteString("&V#")
		buf.WriteString(v.URL)
		buf.WriteString("&$#")
	}
	*pvals = buf.String()
}

func addFuncs(buf *bytes.Buffer, pfuncs *string, funcs []*Func) {
	for _, v := range funcs {
		buf.WriteString(v.Name)
		buf.WriteString("&F#")
		buf.WriteString(v.Doc)
		buf.WriteString("&F#")
		buf.WriteString(v.Decl)
		buf.WriteString("&F#")
		buf.WriteString(v.URL)
		buf.WriteString("&F#")
		buf.WriteString(*codeEncode(&v.Code))
		buf.WriteString("&$#")
	}
	*pfuncs = buf.String()
}

func codeEncode(code *string) *string {
	str := new(string)
	*str = base32.StdEncoding.EncodeToString([]byte(*code))
	return str
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
}

// getStatic gets a document from a statically known service. getStatic
// returns errNoMatch if the import path is not recognized.
func getStatic(client *http.Client, importPath string, etag string) (pdoc *Package, err error) {
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
		return s.get(client, match, etag)
	}
	return nil, errNoMatch
}

func getDynamic(client *http.Client, importPath string, etag string) (pdoc *Package, err error) {
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

	pdoc, err = getStatic(client, expand("{repo}{dir}", match), etag)
	if err == errNoMatch {
		pdoc, err = getVCSDoc(client, match, etag)
	}
	if err != nil {
		return nil, err
	}

	if pdoc != nil {
		pdoc.ImportPath = importPath
		pdoc.ProjectName = match["projectName"]
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
