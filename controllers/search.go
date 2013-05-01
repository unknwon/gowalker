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
	"strings"

	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

type SearchController struct {
	beego.Controller
}

// Get implemented Get method for SearchController.
// It serves search page of Go Walker.
func (this *SearchController) Get() {
	// Get language version.
	curLang, restLangs := getLangVer(this.Input().Get("lang"))

	// Get arguments.
	q := this.Input().Get("q")

	// Set properties.
	this.Layout = "layout.html"
	this.TplNames = "search_" + curLang.Lang + ".html"

	this.Data["DataSrc"] = utils.GoRepoSet
	this.Data["Keyword"] = q
	this.Data["Lang"] = curLang.Lang
	this.Data["CurLang"] = curLang.Name
	this.Data["RestLangs"] = restLangs

	if checkSpecialUsage(this, q) {
		return
	}

	rawPath := q // Raw path.
	rawPath = strings.Replace(rawPath, "http://", "", 1)
	rawPath = strings.Replace(rawPath, "https://", "", 1)

	if utils.IsGoRepoPath(q) {
		q = "code.google.com/p/go/source/browse/src/pkg/" + q
	}

	if path, ok := utils.IsBrowseURL(q); ok {
		q = path
	}

	// Check if it is a remote path, if not means it's a keyword.
	if utils.IsValidRemotePath(q) {
		// Check documentation of this import path, and update automatically as needed.

		/* TODO:WORKING */

		pdoc, err := doc.CheckDoc(q, doc.HUMAN_REQUEST)
		if err == nil {
			// Generate documentation page.

			/* TODO */

			if generatePage(pdoc) {
				return
			}
		} else {
			beego.Error("SearchController.Get():", err)
		}
	}

	// Returns a slice of results.

	/* TODO */

}

// checkSpecialUsage checks special usage of keywords.
// It returns true if it is a special usage, false otherwise.
func checkSpecialUsage(this *SearchController, q string) bool {
	switch {
	case q == "gorepo":
		// Show list of standard library

		/* TODO */

		return true
	}

	return false
}

// generatePage genarates documentation page for project.
// it returns false when its a invaild(empty) project.
func generatePage(pdoc *doc.Package) bool {
	return true
}
