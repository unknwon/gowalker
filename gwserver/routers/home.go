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
	"time"

	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
)

var (
	maxProInfoNum = 20
	maxExamNum    = 15

	recentUpdatedExs                                       []*models.PkgExam
	recentViewedPros, topRankPros, topViewedPros, RockPros []*hv.PkgInfo
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

func serveHome(this *HomeRouter, urpids, urpts *http.Cookie) {
	this.Data["IsHome"] = true
	this.TplNames = "home.html"

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
}

func updateCacheInfo(pdoc *hv.Package, urpids, urpts *http.Cookie) (string, string) {
	pdoc.ViewedTime = time.Now().UTC().Unix()

	updateCachePros(pdoc)
	updateProInfos(pdoc)
	return updateUrPros(pdoc, urpids, urpts)
}

func updateCachePros(pdoc *hv.Package) {
	pdoc.Views++

	for _, p := range cachePros {
		if p.Id == pdoc.Id {
			p = pdoc.PkgInfo
			p.Rank = int64(pdoc.RefProNum*30) + pdoc.Views
			return
		}
	}

	pinfo := pdoc.PkgInfo
	pinfo.Rank = int64(pdoc.RefProNum*30) + pdoc.Views
	cachePros = append(cachePros, pinfo)
}

func updateProInfos(pdoc *hv.Package) {
	index := -1
	listLen := len(recentViewedPros)
	curPro := pdoc.PkgInfo

	// Check if in the list
	for i, s := range recentViewedPros {
		if s.ImportPath == curPro.ImportPath {
			index = i
			break
		}
	}

	s := make([]*hv.PkgInfo, 0, maxProInfoNum)
	s = append(s, curPro)
	switch {
	case index == -1 && listLen < maxProInfoNum:
		// Not found and list is not full
		s = append(s, recentViewedPros...)
	case index == -1 && listLen >= maxProInfoNum:
		// Not found but list is full
		s = append(s, recentViewedPros[:maxProInfoNum-1]...)
	case index > -1:
		// Found
		s = append(s, recentViewedPros[:index]...)
		s = append(s, recentViewedPros[index+1:]...)
	}
	recentViewedPros = s
}

// updateUrPros returns strings of user recent viewd projects and timestamps.
func updateUrPros(pdoc *hv.Package, urpids, urpts *http.Cookie) (string, string) {
	if pdoc.Id == 0 {
		return urpids.Value, urpts.Value
	}

	var urPros, urTs []string
	if urpids != nil && urpts != nil {
		urPros = strings.Split(urpids.Value, "|")
		urTs = strings.Split(urpts.Value, "|")
		if len(urTs) != len(urPros) {
			urTs = strings.Split(
				strings.Repeat(strconv.Itoa(int(time.Now().UTC().Unix()))+"|", len(urPros)), "|")
			urTs = urTs[:len(urTs)-1]
		}
	}

	index := -1
	listLen := len(urPros)

	// Check if in the list
	for i, s := range urPros {
		pid, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return urpids.Value, urpts.Value
		}
		if pid == pdoc.Id {
			index = i
			break
		}
	}

	s := make([]string, 0, maxProInfoNum)
	ts := make([]string, 0, maxProInfoNum)
	s = append(s, strconv.Itoa(int(pdoc.Id)))
	ts = append(ts, strconv.Itoa(int(time.Now().UTC().Unix())))

	switch {
	case index == -1 && listLen < maxProInfoNum:
		// Not found and list is not full
		s = append(s, urPros...)
		ts = append(ts, urTs...)
	case index == -1 && listLen >= maxProInfoNum:
		// Not found but list is full
		s = append(s, urPros[:maxProInfoNum-1]...)
		ts = append(ts, urTs[:maxProInfoNum-1]...)
	case index > -1:
		// Found
		s = append(s, urPros[:index]...)
		s = append(s, urPros[index+1:]...)
		ts = append(ts, urTs[:index]...)
		ts = append(ts, urTs[index+1:]...)
	}
	return strings.Join(s, "|"), strings.Join(ts, "|")
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

	// User History.
	urpids, _ := this.Ctx.Request.Cookie("UserHistory")
	urpts, _ := this.Ctx.Request.Cookie("UHTimestamps")

	if len(reqUrl) == 0 && len(q) == 0 {
		serveHome(this, urpids, urpts)
		return
	}

	// Documentation page.
	broPath := reqUrl // Browse path.

	// Check if it's the standard library.
	if utils.IsGoRepoPath(broPath) {
		broPath = "code.google.com/p/go/source/browse/src/pkg/" + broPath
	} else if utils.IsGoSubrepoPath(broPath) {
		broPath = "code.google.com/p/" + broPath
		reqUrl = broPath
	}

	// Check if it's a remote path that can be used for 'go get', if not means it's a keyword.
	if !utils.IsValidRemotePath(broPath) {
		// Search.
		this.Redirect("/search?q="+reqUrl, 302)
		return
	}

	// Get tag field.
	tag := strings.TrimSpace(this.Input().Get("tag"))
	if tag == "master" || tag == "default" {
		tag = ""
	}

	// Check documentation of current import path, update automatically as needed.
	pdoc, err := doc.CheckDoc(reqUrl, tag, doc.RT_Human)
	if err == nil {
		// errNoMatch leads to pdoc == nil.
		if pdoc != nil {
			// Generate documentation page.
			if generatePage(this, pdoc, broPath, tag) {
				ps, ts := updateCacheInfo(pdoc, urpids, urpts)
				this.Ctx.SetCookie("UserRecentPros", ps, 9999999999, "/")
				this.Ctx.SetCookie("URPTimestamps", ts, 9999999999, "/")
				return
			}
		}
	}

	// TODO
	beego.Error(err)
	//this.Redirect("/search?q="+reqUrl, 302)
	return
}

// generatePage genarates documentation page for project.
// it returns false when it's a invaild(empty) project.
func generatePage(this *HomeRouter, pdoc *hv.Package, q, tag string) bool {
	return false
}
