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
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/Unknwon/com"
	"github.com/Unknwon/ctw/packer"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
)

type crawlResult struct {
	pdoc *hv.Package
	err  error
}

// crawlDoc fetchs package from VCS and returns 'Package' by given import path and tag.
// It returns error when error occurs in the underlying functions.
func crawlDoc(path, tag string, pinfo *hv.PkgInfo) (pdoc *hv.Package, err error) {
	var pdocNew *hv.Package
	pdocNew, err = getRepo(packer.HttpClient, path, tag, pinfo.Ptag)

	if err != errNotModified && pdocNew != nil {
		pdoc = pdocNew
		pdoc.Views = pinfo.Views
		pdoc.Labels = pinfo.Labels
		pdoc.RefNum = pinfo.RefNum
		pdoc.RefPids = pinfo.RefPids
		pdoc.Rank = pinfo.Rank
	}

	switch {
	case err == nil:
		// Let upper level to render doc. page.
		return pdoc, nil
	case err == errNoMatch:
		beego.Error("doc.crawlDoc ->", err)
		err = nil
		pdoc = nil
	case isNotFound(err):
		// We do not need to delete standard library,
		// so here skip it by not found.

		// Only delete when server cannot find master branch
		// because sub-package(s) may not exist in old tag(s).
		if len(tag) == 0 {
			models.DeleteProject(path)
		}
	}
	return pdoc, err
}

// getRepo downloads package data and returns 'Package' by given import path and tag.
// It returns error when error occurs in the underlying functions.
func getRepo(client *http.Client, path, tag, ptag string) (pdoc *hv.Package, err error) {
	switch {
	case utils.IsGoRepoPath(path):
		pdoc, err = getStandardDoc(client, path, tag, ptag)
	case utils.IsValidRemotePath(path):
		pdoc, err = getStatic(client, path, tag, ptag)
		if err == errNoMatch {
			pdoc, err = getDynamic(client, path, tag, ptag)
		}
	default:
		return nil, errors.New(
			fmt.Sprintf("doc.getRepo -> No match( %s:%s )", path, tag))
	}

	if pdoc != nil {
		pdoc.PkgVer = hv.PACKAGE_VER
	}

	return pdoc, err
}

func RenderFuncs(pdoc *hv.Package) []*models.PkgFunc {
	pfuncs := make([]*models.PkgFunc, 0, len(pdoc.Funcs)+len(pdoc.Types)*3)

	links := getLinks(pdoc)

	// Functions.
	pfuncs = addFuncs(pfuncs, pdoc.Funcs, pdoc.ImportPath, links)
	pfuncs = addFuncs(pfuncs, pdoc.Ifuncs, pdoc.ImportPath, links)

	// Types.
	for _, v := range pdoc.Types {
		// Functions.
		for _, m := range v.Funcs {
			pfuncs = addFunc(pfuncs, m, pdoc.ImportPath, m.Name, links)
		}
		for _, m := range v.IFuncs {
			pfuncs = addFunc(pfuncs, m, pdoc.ImportPath, m.Name, links)
		}

		// Methods.
		for _, m := range v.Methods {
			pfuncs = addFunc(pfuncs, m, pdoc.ImportPath, v.Name+"_"+m.Name, links)
		}
		for _, m := range v.IMethods {
			pfuncs = addFunc(pfuncs, m, pdoc.ImportPath, v.Name+"_"+m.Name, links)
		}
	}

	return pfuncs
}

// SaveProject saves project information to database.
func SaveProject(pdoc *hv.Package, pfuncs []*models.PkgFunc) (int64, error) {
	// Save package information.
	pinfo := &hv.PkgInfo{
		ImportPath:  pdoc.ImportPath,
		ProjectName: pdoc.ProjectName,
		Synopsis:    pdoc.Synopsis,
		IsCmd:       pdoc.IsCmd,
		Tags:        pdoc.Tags,
		Views:       pdoc.Views,
		ViewedTime:  time.Now().UTC().Unix(),
		Created:     time.Now().UTC(),
		Rank:        pdoc.Rank,
		PkgVer:      pdoc.PkgVer,
		Ptag:        pdoc.Ptag,
		Labels:      pdoc.Labels,
		RefNum:      pdoc.RefNum,
		RefPids:     pdoc.RefPids,
		Note:        pdoc.Note,
	}
	_ = pinfo
	// Save package declaration and functions.
	pdecl := &hv.PkgDecl{
		Tag: pdoc.Tag,
		//JsNum:        pdoc.JsNum,
		// IsHasExport:  pdoc.IsHasExport,
		// IsHasConst:   pdoc.IsHasConst,
		// IsHasVar:     pdoc.IsHasVar,
		// IsHasExample: pdoc.IsHasExample,
		// IsHasFile:    pdoc.IsHasFile,
		// IsHasSubdir:  pdoc.IsHasSubdir,
	}

	// Imports.
	pdecl.Imports = pdoc.Imports
	pdecl.TestImports = pdoc.TestImports

	//err := models.SaveProject(pinfo, pdecl, pfuncs, pdoc.Imports)
	return 0, errors.New("doc.SaveProject -> test error")
}

// getLinks returns exported objects with its jump link.
func getLinks(pdoc *hv.Package) []*utils.Link {
	links := make([]*utils.Link, 0, len(pdoc.Types)+len(pdoc.Imports)+len(pdoc.Funcs)+10)
	// Get all types, functions and import packages
	for _, t := range pdoc.Types {
		links = append(links, &utils.Link{
			Name:    t.Name,
			Comment: template.HTMLEscapeString(t.Doc),
		})
	}

	for _, f := range pdoc.Funcs {
		links = append(links, &utils.Link{
			Name:    f.Name,
			Comment: template.HTMLEscapeString(f.Doc),
		})
	}

	for _, t := range pdoc.Types {
		for _, f := range t.Funcs {
			links = append(links, &utils.Link{
				Name:    f.Name,
				Comment: template.HTMLEscapeString(f.Doc),
			})
		}
	}

	for _, v := range pdoc.Imports {
		if v != "C" {
			links = append(links, &utils.Link{
				Name: path.Base(v) + ".",
				Path: v,
			})
		}
	}
	return links
}

func addValues(buf *bytes.Buffer, pvals *string, vals []*hv.Value) {
	buf.Reset()
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

// addFuncs appends functions to 'pfuncs'.
// NOTE: it can be only use for pure functions(not belong to any type), not methods.
func addFuncs(pfuncs []*models.PkgFunc, fs []*hv.Func, path string, links []*utils.Link) []*models.PkgFunc {
	for _, f := range fs {
		pfuncs = addFunc(pfuncs, f, path, f.Name, links)
	}
	return pfuncs
}

// addFunc appends a function to 'pfuncs'.
func addFunc(pfuncs []*models.PkgFunc, f *hv.Func, path, name string, links []*utils.Link) []*models.PkgFunc {
	var buf bytes.Buffer
	f.FullName = name
	f.Code = f.Decl + " {\n" + f.Code + "}"
	utils.FormatCode(&buf, &f.Code, links)
	f.Code = buf.String()
	return append(pfuncs, &models.PkgFunc{
		Name: name,
		Path: path,
		Doc:  f.Doc,
	})
}

func CodeEncode(code *string) *string {
	str := new(string)
	*str = base32.StdEncoding.EncodeToString([]byte(*code))
	return str
}

// service represents a source code control service.
type service struct {
	pattern *regexp.Regexp
	prefix  string
	get     func(*http.Client, map[string]string, string, string) (*hv.Package, error)
}

// services is the list of source code control services handled by gowalker.
var services = []*service{
	{googlePattern, "code.google.com/", getGoogleDoc},
	{githubPattern, "github.com/", getGithubDoc},
	{bitbucketPattern, "bitbucket.org/", getBitbucketDoc},
	{launchpadPattern, "launchpad.net/", getLaunchpadDoc},
	{oscPattern, "git.oschina.net/", getOSCDoc},
	//{csdnPattern, "code.csdn.net/", getCSDNDoc},
}

// getStatic gets a document from a statically known service. getStatic
// returns errNoMatch if the import path is not recognized.
func getStatic(client *http.Client, importPath, tag, etag string) (pdoc *hv.Package, err error) {
	for _, s := range services {
		if s.get == nil || !strings.HasPrefix(importPath, s.prefix) {
			continue
		}
		m := s.pattern.FindStringSubmatch(importPath)
		if m == nil {
			if s.prefix != "" {
				return nil, com.NotFoundError{"Import path prefix matches known service, but regexp does not."}
			}
			continue
		}
		match := map[string]string{"importPath": importPath}
		for i, n := range s.pattern.SubexpNames() {
			if n != "" {
				match[n] = m[i]
			}
		}
		return s.get(client, match, tag, etag)
	}
	return nil, errNoMatch
}

func getDynamic(client *http.Client, importPath, tag, etag string) (pdoc *hv.Package, err error) {
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
			return nil, com.NotFoundError{"Project root mismatch."}
		}
	}

	pdoc, err = getStatic(client, com.Expand("{repo}{dir}", match), tag, etag)
	if err == errNoMatch {
		//pdoc, err = getVCSDoc(client, match, etag)
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
			return nil, &com.RemoteError{strings.SplitN(importPath, "/", 2)[0], err}
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
				return nil, com.NotFoundError{"More than one <meta> found at " + scheme + "://" + importPath}
			}

			projectRoot, vcs, repo := f[0], f[1], f[2]

			repo = strings.TrimSuffix(repo, "."+vcs)
			i := strings.Index(repo, "://")
			if i < 0 {
				return nil, com.NotFoundError{"Bad repo URL in <meta>."}
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
		return nil, com.NotFoundError{"<meta> not found."}
	}
	return match, nil
}
