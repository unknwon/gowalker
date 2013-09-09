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
	"net/http"
	"strconv"
	"strings"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
)

var (
	maxProInfoNum = 20
	maxExamNum    = 15

	recentUpdatedExs                                       []*models.PkgExam
	recentViewedPros, topRankPros, topViewedPros, RockPros []*models.PkgInfo
)

// initPopPros initializes popular projects.
func initPopPros() {
	var err error
	err, recentUpdatedExs, recentViewedPros, topRankPros, topViewedPros, RockPros =
		models.GetPopulars(maxProInfoNum, maxExamNum)
	if err != nil {
		panic("initPopPros -> " + err.Error())
	}
}

// HomeRouter serves home page.
type HomeRouter struct {
	baseRouter
}

// Get implemented Get method for HomeRouter.
func (this *HomeRouter) Get() {
	this.Data["IsHome"] = true

	// Get argument(s).
	q := strings.TrimRight(
		strings.TrimSpace(this.Input().Get("q")), "/")

	if path, ok := utils.IsBrowseURL(q); ok {
		q = path
	}

	// Get pure URL.
	reqUrl := this.Ctx.Request.RequestURI[1:]
	if i := strings.Index(reqUrl, "?"); i > -1 {
		reqUrl = reqUrl[:i]
		if path, ok := utils.IsBrowseURL(reqUrl); ok {
			reqUrl = path
		}
	}

	// Redirect to query string.
	if len(reqUrl) == 0 && len(q) > 0 {
		reqUrl = q
		this.Redirect("/"+reqUrl, 302)
		return
	}

	// Get language.
	curLang, _ := this.Data["LangVer"].(utils.LangType)
	this.TplNames = "home_" + curLang.Lang + ".html"

	// User History.
	urpids, _ := this.Ctx.Request.Cookie("UserHistory")
	urpts, _ := this.Ctx.Request.Cookie("UHTimestamps")

	if len(reqUrl) == 0 && len(q) == 0 {
		serveHome(this, urpids, urpts)
	} else {

		this.Redirect("/search?q="+reqUrl, 302)
	}
}

func serveHome(this *HomeRouter, urpids, urpts *http.Cookie) {
	this.Data["IsHome"] = true

	// Global Recent projects.
	this.Data["GlobalHistory"] = recentViewedPros
	// User Recent projects.
	if urpids != nil && urpts != nil {
		upros := models.GetGroupPkgInfoById(strings.Split(urpids.Value, "|"))
		pts := strings.Split(urpts.Value, "|")
		for i, p := range upros {
			ts, _ := strconv.ParseInt(pts[i], 10, 64)
			p.ViewedTime = ts
		}
		this.Data["UserHistory"] = upros
	}

	// Popular projects and examples.
	this.Data["WeeklyStarPros"] = RockPros
	this.Data["TopRankPros"] = topRankPros
	this.Data["TopViewsPros"] = topViewedPros
	this.Data["RecentExams"] = recentUpdatedExs
	// Standard library type-ahead.
	this.Data["GoReposSrc"] = utils.GoRepoSet
}
