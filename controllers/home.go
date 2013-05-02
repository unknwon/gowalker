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

// Package controllers implemented controller methods of beego.

package controllers

import (
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

const (
	recentViewsProNum = 25
)

var (
	recentViewedPros []*recentPro // Recent viewed projects.
	langTypes        []*langType  // Languages that are supported.
)

type recentPro struct {
	Path, ViewedTime string
	IsGoRepo         bool
	Views            int64
}

type langType struct {
	Lang, Name string
}

func init() {
	// Initialized recent viewed projects.
	recentViewedPros = make([]*recentPro, 0, recentViewsProNum)
	proinfos, _ := models.GetRecentPros(recentViewsProNum)
	for _, p := range proinfos {
		recentViewedPros = append(recentViewedPros,
			&recentPro{
				Path:       p.Path,
				ViewedTime: p.ViewedTime,
				IsGoRepo:   p.ProName == "Go",
				Views:      p.Views,
			})
	}
}

func InitLangs(langs []string, names []string) {
	langTypes = make([]*langType, 0, len(langs))
	for i, v := range langs {
		langTypes = append(langTypes, &langType{
			Lang: v,
			Name: names[i],
		})
	}
}

type HomeController struct {
	beego.Controller
}

// Get implemented Get method for HomeController.
// It serves home page of Go Walker.
func (this *HomeController) Get() {
	// Get language version
	curLang, restLangs := getLangVer(this.Input().Get("lang"))

	// Get query field
	q := this.Input().Get("q")

	// Not empty query string shows search page
	if len(q) > 0 {
		// Show search page
		this.Redirect("/search?lang="+curLang.Lang+"&q="+q, 302)
		return
	}

	// Set properties
	this.Layout = "layout.html"
	this.TplNames = "home_" + curLang.Lang + ".html"

	// Recent projects
	this.Data["RecentPros"] = recentViewedPros
	pkgInfos, _ := models.GetPopularPros()
	this.Data["PopPros"] = pkgInfos

	this.Data["DataSrc"] = utils.GoRepoSet
	this.Data["Lang"] = curLang.Lang
	this.Data["CurLang"] = curLang.Name
	this.Data["RestLangs"] = restLangs
}

// getLangVer returns current language version and list of rest languages.
func getLangVer(lang string) (*langType, []*langType) {
	if len(lang) == 0 {
		lang = "en"
	}
	curLang := &langType{
		Lang: lang,
	}

	restLangs := make([]*langType, 0, len(langTypes)-1)
	for _, v := range langTypes {
		if lang != v.Lang {
			restLangs = append(restLangs, v)
		} else {
			curLang.Name = v.Name
		}
	}
	return curLang, restLangs
}
