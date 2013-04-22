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
		err := models.CheckDoc(q, models.HUMAN_REQUEST)
		if err == nil {
			// Redirect to documentation page
			this.Redirect("/"+q, 302)
		} else {
			beego.Error("SearchController.Get:", err)
		}
	}

	// Search packages by the keyword
	this.Data["keyword"] = q
	// Returns a slice of results
	pkgs := models.SearchDoc(q)
	// Show results after searched
	if len(pkgs) > 0 {
		this.Data["showpkg"] = true
		this.Data["pkgs"] = pkgs
	}
}
