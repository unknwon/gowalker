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

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

type SearchController struct {
	beego.Controller
}

// Get implemented Get method for SearchController.
// It serves search page of Go Walker.
func (this *SearchController) Get() {
	// Set language version.
	curLang, restLangs := setLangVer(this.Ctx.Request, this.Input())

	// Save language information in cookies.
	this.Ctx.SetCookie("lang", curLang.Lang+";path=/", 0)

	// Get arguments.
	q := strings.TrimSpace(this.Input().Get("q"))

	// Empty query string shows home page.
	if len(q) == 0 {
		this.Redirect("/", 302)
		return
	}

	// Set properties.
	this.Layout = "layout_" + curLang.Lang + ".html"
	this.TplNames = "search_" + curLang.Lang + ".html"

	this.Data["Keyword"] = q
	// Set standard library keyword type-ahead.
	this.Data["DataSrc"] = utils.GoRepoSet
	// Set language properties.
	this.Data["Lang"] = curLang.Lang
	this.Data["CurLang"] = curLang.Name
	this.Data["RestLangs"] = restLangs

	if checkSpecialUsage(this, q) {
		return
	}

	if path, ok := utils.IsBrowseURL(q); ok {
		q = path
	}

	// Returns a slice of results.
	pkgInfos, _ := models.SearchDoc(q)
	// Show results after searched.
	if len(pkgInfos) > 0 {
		this.Data["IsFindPro"] = true
		this.Data["AllPros"] = pkgInfos
	}
}

// checkSpecialUsage checks special usage of keywords.
// It returns true if it is a special usage, false otherwise.
func checkSpecialUsage(this *SearchController, q string) bool {
	switch {
	case q == "gorepo": // Show list of standard library.
		pkgInfos, _ := models.GetGoRepo()
		// Show results after searched.
		if len(pkgInfos) > 0 {
			this.Data["IsFindPro"] = true
			this.Data["AllPros"] = pkgInfos
		}
		return true
	case q == "imports": // Show imports package list.
		pkgs := strings.Split(this.Input().Get("pkgs"), "|")
		pinfos, _ := models.GetGroupPkgInfo(pkgs)
		if len(pinfos) > 0 {
			this.Data["IsFindPro"] = true
			this.Data["AllPros"] = pinfos
		}
		return true
	case q == "imported": // Show packages that import this project.
		pkgs := strings.Split(this.Input().Get("pkgs"), "|")
		pinfos, _ := models.GetGroupPkgInfoById(pkgs)
		if len(pinfos) > 0 {
			this.Data["IsFindPro"] = true
			this.Data["AllPros"] = pinfos
		}
		return true
	case strings.Index(q, ":tag=") > -1: // Add tag(s) to the project.
		// Get tag(s).
		i := strings.Index(q, ":tag=")
		if utils.IsGoRepoPath(q[:i]) {
			this.Redirect("/"+q[:i], 302)
			return true
		}

		// Verify tags.
		inputs := strings.Split(q[i+5:], ":")
		tags := verifyTags(inputs)

		if len(tags) > 0 && models.UpdateTagInfo(q[:i], tags, true) {
			this.Redirect("/"+q[:i], 302)
		}
		return true
	case strings.Index(q, ":rtag=") > -1: // Remove tag(s) to the project.
		// Get tag(s).
		i := strings.Index(q, ":rtag=")
		if utils.IsGoRepoPath(q[:i]) {
			this.Redirect("/"+q[:i], 302)
			return true
		}

		// Verify tags.
		inputs := strings.Split(q[i+6:], ":")
		tags := verifyTags(inputs)

		if len(tags) > 0 && models.UpdateTagInfo(q[:i], tags, false) {
			this.Redirect("/"+q[:i], 302)
		}
		return true
	}

	return false
}

func verifyTags(inputs []string) []string {
	tags := make([]string, 0, len(inputs))
	for _, v := range inputs {
		if len(v) > 0 {
			for _, t := range tagList {
				if t == v {
					tags = append(tags, v)
					break
				}
			}
		}
	}
	return tags
}
