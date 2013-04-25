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
	"strings"

	"github.com/astaxie/beego"
	"github.com/unknwon/gowalker/models"
	"github.com/unknwon/gowalker/utils"
)

type SearchController struct {
	beego.Controller
}

func (this *SearchController) Get() {
	// Get query field
	q := this.Input().Get("q")

	// Empty query string shows home page
	if len(q) == 0 {
		this.Redirect("/", 302)
	}

	// Set properties
	this.TplNames = "search.html"
	this.Layout = "layout.html"

	// Check if it is a browse URL, if not means it's a keyword or import path
	if path, ok := utils.IsBrowseURL(q); ok {
		q = path
	}

	// Check if it is a short path for standard library
	if utils.IsGoRepoPath(q) {
		q = "code.google.com/p/go/source/browse/src/pkg/" + q
	}

	// Check if it is a remote path, if not means it's a keyword
	if utils.IsValidRemotePath(q) {
		// Check documentation of this import path, and update automatically as needed

		/* TODO:WORKING */
		//os.Remove("./docs/" + strings.Replace(q, "http://", "", 1) + ".html")
		pdoc, err := models.CheckDoc(q, models.HUMAN_REQUEST)
		q = strings.Replace(q, "http://", "", 1)
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
	this.Data["proName"] = pdoc.Name
	pkgDocPath := pdoc.BrowseURL[7 : strings.Index(pdoc.BrowseURL, "/"+pdoc.Name)+1]
	this.Data["pkgSearch"] = pkgDocPath[:len(pkgDocPath)-1]
	this.Data["pkgDocPath"] = pkgDocPath
	this.Data["importPath"] = pdoc.ImportPath

	// Main introduction
	/*if synIndex := strings.Index(pdoc.Doc, "."); synIndex > -1 {
		pdoc.Doc = pdoc.Doc[synIndex+1:]
	}*/

	// Full introduction
	var buf bytes.Buffer
	doc.ToHTML(&buf, pdoc.Doc, nil)
	pkgInfo := buf.String()
	pkgInfo = strings.Replace(pkgInfo, "<p>", "<p><b>", 1)
	pkgInfo = strings.Replace(pkgInfo, "</p>", "</b></p>", 1)
	this.Data["pkgFullIntro"] = pkgInfo

	// Index
	this.Data["isHasConst"] = len(pdoc.Consts) > 0
	this.Data["isHasVar"] = len(pdoc.Vars) > 0
	this.Data["funcs"] = pdoc.Funcs
	for i, f := range pdoc.Funcs {
		buf.Reset()
		doc.ToHTML(&buf, f.Doc, nil)
		f.Doc = buf.String()
		pdoc.Funcs[i] = f
	}
	this.Data["types"] = pdoc.Types
	for i, t := range pdoc.Types {
		for j, f := range t.Funcs {
			buf.Reset()
			doc.ToHTML(&buf, f.Doc, nil)
			f.Doc = buf.String()
			t.Funcs[j] = f
		}
		for j, m := range t.Methods {
			buf.Reset()
			doc.ToHTML(&buf, m.Doc, nil)
			m.Doc = buf.String()
			t.Methods[j] = m
		}
		buf.Reset()
		doc.ToHTML(&buf, t.Doc, nil)
		t.Doc = buf.String()
		pdoc.Types[i] = t
	}

	// Constants
	this.Data["consts"] = pdoc.Consts
	// Variables
	this.Data["vars"] = pdoc.Vars
	// Files
	this.Data["files"] = pdoc.Files

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
