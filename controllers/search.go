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

package controllers

import (
	"bytes"
	"go/doc"
	"os"
	"runtime"
	"strings"

	"github.com/astaxie/beego"
	"github.com/unknwon/gowalker/models"
	"github.com/unknwon/gowalker/utils"
)

type SearchController struct {
	beego.Controller
}

func (this *SearchController) Get() {
	// Check language version
	lang, ok := isValidLanguage(this.Ctx.Request.RequestURI)
	if !ok {
		// English is default language version
		this.Redirect("/en/", 302)
		return
	}

	// Get query field
	q := this.Input().Get("q")

	// Empty query string shows home page
	if len(q) == 0 {
		this.Redirect("/"+lang+"/", 302)
	}

	// Set properties
	this.TplNames = "search_" + lang + ".html"
	this.Layout = "layout.html"

	// Check if it is a import path for standard library
	if utils.IsGoRepoPath(q) {
		q = "code.google.com/p/go/source/browse/src/pkg/" + q
	}

	if path, ok := utils.IsBrowseURL(q); ok {
		q = path
	}

	q = strings.Replace(q, "http://", "", 1)
	// Check if it is a remote path, if not means it's a keyword
	if utils.IsValidRemotePath(q) {
		// This is for regenerating documentation every time in develop mode
		//os.Remove("./docs/" + strings.Replace(q, "http://", "", 1) + ".html")

		// Check documentation of this import path, and update automatically as needed

		/* TODO:WORKING */

		pdoc, err := models.CheckDoc(q, models.HUMAN_REQUEST)
		if err == nil {
			// Generate static page

			/* TODO */

			if generatePage(this, pdoc, q) {
				// Redirect to documentation page
				this.Redirect("/"+q+".html", 302)
			}
		} else {
			beego.Error("SearchController.Get:", err)
		}
	}

	// Search packages by the keyword
	this.Data["keyword"] = q
	// Returns a slice of results

	/* TODO */

	pkgs := models.SearchDoc(q)
	// Show results after searched
	if len(pkgs) > 0 {
		this.Data["showpkg"] = true
		this.Data["pkgs"] = pkgs
	}
}

func generatePage(this *SearchController, pdoc *models.Package, q string) bool {
	if pdoc == nil || len(pdoc.Name) == 0 {
		return utils.IsExist("./docs/" + q + ".html")
	}

	// Set properties
	this.TplNames = "docs.html"

	// Set data
	// Introduction
	this.Data["proPath"] = pdoc.BrowseURL

	if urlLen := len(pdoc.BrowseURL); pdoc.BrowseURL[urlLen-1] == '/' {
		pdoc.BrowseURL = pdoc.BrowseURL[:urlLen-1]
	}
	lastIndex := strings.LastIndex(pdoc.BrowseURL, "/")
	pkgName := pdoc.BrowseURL[lastIndex+1:]
	if i := strings.Index(pkgName, "?"); i > -1 {
		pkgName = pkgName[:i]
	}
	this.Data["proName"] = pkgName
	cutIndex := strings.Index(pdoc.BrowseURL, "://")
	pkgDocPath := pdoc.BrowseURL[cutIndex+3 : lastIndex+1]
	this.Data["pkgSearch"] = pkgDocPath[:len(pkgDocPath)]
	this.Data["pkgDocPath"] = pkgDocPath
	this.Data["importPath"] = pdoc.ImportPath

	// Full introduction
	var buf bytes.Buffer
	doc.ToHTML(&buf, pdoc.Doc, nil)
	pkgInfo := buf.String()
	pkgInfo = strings.Replace(pkgInfo, "<p>", "<p><b>", 1)
	pkgInfo = strings.Replace(pkgInfo, "</p>", "</b></p>", 1)
	this.Data["pkgFullIntro"] = pkgInfo

	links := make([]*utils.Link, 0, len(pdoc.Types)+len(pdoc.Imports))
	// Get all types and import packages
	for _, t := range pdoc.Types {
		links = append(links, &utils.Link{
			Name:    t.Name,
			Comment: t.Doc,
		})
	}

	// Index
	this.Data["isHasConst"] = len(pdoc.Consts) > 0
	this.Data["isHasVar"] = len(pdoc.Vars) > 0
	this.Data["funcs"] = pdoc.Funcs
	for i, f := range pdoc.Funcs {
		buf.Reset()
		doc.ToHTML(&buf, f.Doc, nil)
		f.Doc = buf.String()
		buf.Reset()
		utils.FormatCode(&buf, f.Decl.Text, links)
		f.FmtDecl = buf.String()
		pdoc.Funcs[i] = f
	}
	this.Data["types"] = pdoc.Types
	for i, t := range pdoc.Types {
		for j, f := range t.Funcs {
			buf.Reset()
			doc.ToHTML(&buf, f.Doc, nil)
			f.Doc = buf.String()
			buf.Reset()
			utils.FormatCode(&buf, f.Decl.Text, links)
			f.FmtDecl = buf.String()
			t.Funcs[j] = f
		}
		for j, m := range t.Methods {
			buf.Reset()
			doc.ToHTML(&buf, m.Doc, nil)
			m.Doc = buf.String()
			buf.Reset()
			utils.FormatCode(&buf, m.Decl.Text, links)
			m.FmtDecl = buf.String()
			t.Methods[j] = m
		}
		buf.Reset()
		doc.ToHTML(&buf, t.Doc, nil)
		t.Doc = buf.String()
		buf.Reset()
		utils.FormatCode(&buf, t.Decl.Text, links)
		t.FmtDecl = buf.String()
		pdoc.Types[i] = t
	}

	// Constants
	this.Data["consts"] = pdoc.Consts
	for i, v := range pdoc.Consts {
		buf.Reset()
		utils.FormatCode(&buf, v.Decl.Text, links)
		v.FmtDecl = buf.String()
		pdoc.Consts[i] = v
	}
	// Variables
	this.Data["vars"] = pdoc.Vars
	for i, v := range pdoc.Vars {
		buf.Reset()
		utils.FormatCode(&buf, v.Decl.Text, links)
		v.FmtDecl = buf.String()
		pdoc.Vars[i] = v
	}
	// Files
	this.Data["files"] = pdoc.Files

	// Import packages
	this.Data["ImportPkgNum"] = len(pdoc.Imports)
	// Generated time
	this.Data["UtcTime"] = pdoc.Updated.Local()
	// System info
	this.Data["GOOS"] = runtime.GOOS
	this.Data["GOARCH"] = runtime.GOARCH

	// Create directories
	os.MkdirAll("./docs/"+q[:strings.LastIndex(q, "/")+1], os.ModePerm)
	// Create file
	f, _ := os.Create("./docs/" + q + ".html")
	// Render content
	s, _ := this.RenderString()
	f.WriteString(s)
	f.Close()
	return true
}
