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

	"github.com/astaxie/beego"
	"github.com/unknwon/gowalker/models"
)

const (
	recentViewsPkgNum = 24
)

var (
	recentViewedPkgs []string
)

func init() {
	recentViewedPkgs = make([]string, 0, recentViewsPkgNum)
	pkginfos, _ := models.GetRecentPkgs(recentViewsPkgNum)
	for _, p := range pkginfos {
		recentViewedPkgs = append(recentViewedPkgs, p.Path)
	}
}

type HomeController struct {
	beego.Controller
}

// Get implemented Get method for HomeController.
// It serves home page of Go Walker.
func (this *HomeController) Get() {
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
	if len(q) > 0 {
		// Show search page
		this.Redirect(lang+"/search?q="+q, 302)
		return
	}

	// Set properties
	this.TplNames = "home_" + lang + ".html"
	this.Layout = "layout.html"

	// Recent packages
	this.Data["RecentPkgs"] = recentViewedPkgs
	pkgInfos, _ := models.GetPopularPkgs()
	this.Data["PopularPkgs"] = pkgInfos
}

// isValidLanguage checks if URL has correct language version.
func isValidLanguage(reqUrl string) (string, bool) {
	var lang string

	if len(reqUrl) == 1 {
		return lang, false
	}
	if i := strings.LastIndex(reqUrl, "/"); i > 2 {
		lang = reqUrl[1:3]
	} else {
		return lang, false
	}

	return lang, true
}
