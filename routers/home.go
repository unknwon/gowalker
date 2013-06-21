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

package routers

import (
	"fmt"
	"strings"

	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

// Recent viewed project.
type recentPro struct {
	Path, Synopsis string
	IsGoRepo       bool
	ViewedTime     int64
}

var (
	recentViewedProNum = 20         // Maximum element number of recent viewed project list.
	recentViewedPros   []*recentPro // Recent viewed project list.

	tagList []string // Projects tag list.
	tagSet  string   // Tags data source.
)

func init() {
	// Initialized recent viewed project list.
	num, err := beego.AppConfig.Int("recentViewedProNum")
	if err == nil {
		recentViewedProNum = num
		beego.Trace("Loaded 'recentViewedProNum' -> value:", recentViewedProNum)
	} else {
		beego.Trace("Failed to load 'recentViewedProNum' -> Use default value:", recentViewedProNum)
	}

	recentViewedPros = make([]*recentPro, 0, recentViewedProNum)
	// Get recent viewed projects from database.
	proinfos, _ := models.GetRecentPros(recentViewedProNum)
	for _, p := range proinfos {
		// Only projects with import path length is less than 40 letters will be showed.
		if len(p.Path) < 40 {
			recentViewedPros = append(recentViewedPros,
				&recentPro{
					Path:       p.Path,
					Synopsis:   p.Synopsis,
					ViewedTime: p.ViewedTime,
					IsGoRepo: p.ProName == "Go" &&
						strings.Index(p.Path, ".") == -1,
				})
		}
	}

	// Initialize project tags.
	tagList = strings.Split(beego.AppConfig.String("tags"), "|")
	for _, s := range tagList {
		tagSet += "&quot;" + s + "&quot;,"
	}
	tagSet = tagSet[:len(tagSet)-1]
}

// HomeRouter serves home and documentation pages.
type HomeRouter struct {
	beego.Controller
}

// Get implemented Get method for HomeRouter.
func (this *HomeRouter) Get() {
	// Filter unusual User-Agent.
	ua := this.Ctx.Request.Header.Get("User-Agent")
	if len(ua) < 20 {
		beego.Warn("User-Agent:", this.Ctx.Request.Header.Get("User-Agent"))
		return
	}

	// Set language version.
	curLang := setLangVer(this.Ctx, this.Input(), this.Data)

	// Get query field.
	q := strings.TrimSpace(this.Input().Get("q"))

	// Remove last "/".
	q = strings.TrimRight(q, "/")
	if path, ok := utils.IsBrowseURL(q); ok {
		q = path
	}

	// Get pure URL.
	reqUrl := this.Ctx.Request.RequestURI[1:]
	if i := strings.Index(reqUrl, "?"); i > -1 {
		reqUrl = reqUrl[:i]
	}

	// Redirect to query string.
	if len(reqUrl) == 0 && len(q) > 0 {
		reqUrl = q
		this.Redirect("/"+reqUrl, 302)
		return
	}

	// Check show home page or documentation page.
	if len(reqUrl) == 0 && len(q) == 0 {
		// Home page.
		this.Data["IsHome"] = true
		this.TplNames = "home_" + curLang.Lang + ".html"

		// Recent projects
		this.Data["RecentPros"] = recentViewedPros
		// Get popular project and examples list from database.
		this.Data["PopPros"], this.Data["PopExams"] = models.GetPopulars(20, 12)
		// Set standard library keyword type-ahead.
		this.Data["DataSrc"] = utils.GoRepoSet
	} else {
		// Documentation page.
		broPath := reqUrl // Browse path.

		// Check if it is standard library.
		if utils.IsGoRepoPath(broPath) {
			broPath = "code.google.com/p/go/source/browse/src/pkg/" + broPath
		}

		// Check if it is a remote path that can be used for 'gopm get', if not means it's a keyword.
		if !utils.IsValidRemotePath(broPath) {
			// Show search page
			this.Redirect("/search?q="+reqUrl, 302)
			return
		}

		// Check documentation of this import path, and update automatically as needed.
		pdoc, err := doc.CheckDoc(reqUrl, doc.HUMAN_REQUEST)
		if err == nil {
			// Generate documentation page.
			fmt.Println("Generate documentation page.")
			_ = pdoc
			return
		} else {
			beego.Error("HomeRouter.Get ->", err)
		}

		// Show search page
		this.Redirect("/search?q="+reqUrl, 302)
		return
	}
}
