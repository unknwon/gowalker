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

	// Check if it is a remote path, if not means it's a keyword
	if utils.IsValidRemotePath(q) {
		// Check documentation of this import path, and update automatically as needed

		/* TODO:WORKING */

		pdoc, err := models.CheckDoc(q, models.HUMAN_REQUEST)
		q = strings.Replace(q, "http://", "", 1)
		if err == nil {
			if pdoc != nil {
				// Generate static page

				/* TODO */

				generatePage(this, pdoc, q)
			}
			// Redirect to documentation page
			this.Redirect("/"+q+".html", 302)
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

func generatePage(this *SearchController, pdoc *models.Package, q string) {
	this.TplNames = "gene.html"
	this.Data["Content"] = pdoc
	// Create directories
	os.MkdirAll("./docs/"+q[:strings.LastIndex(q, "/")+1], os.ModePerm)
	// Create file
	f, _ := os.Create("./docs/" + q + ".html")

	/* TODO */

	s, _ := this.RenderString()
	f.WriteString(s)
	f.Close()
}
