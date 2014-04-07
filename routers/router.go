// Copyright 2013-2014 Unknown
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

// Package routers implemented controller methods of beego.
package routers

import (
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego"

	"github.com/Unknwon/gowalker/hv"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
)

var (
	AppVer    string
	IsProMode bool
	IsBeta    bool
)

// baseRouter implemented global settings for all other routers.
type baseRouter struct {
	beego.Controller
}

// Prepare implemented Prepare method for baseRouter.
func (this *baseRouter) Prepare() {
	// Setting properties.
	this.Data["AppVer"] = AppVer
	this.Data["IsProMode"] = IsProMode
	this.Data["IsBeta"] = IsBeta

	// Setting language version.
	if len(utils.LangTypes) == 0 {
		// Initialize languages.
		langs := strings.Split(utils.Cfg.MustValue("lang", "types"), "|")
		names := strings.Split(utils.Cfg.MustValue("lang", "names"), "|")
		utils.LangTypes = make([]*utils.LangType, 0, len(langs))
		for i, v := range langs {
			utils.LangTypes = append(utils.LangTypes, &utils.LangType{
				Lang: v,
				Name: names[i],
			})
		}
	}

	var isNeedRedir bool
	isNeedRedir, this.Data["LangVer"] = utils.SetLangVer(this.Ctx, this.Input(), this.Data)
	if isNeedRedir {
		// Redirect to make URL clean.
		i := strings.Index(this.Ctx.Request.RequestURI, "?")
		redirUrl := this.Ctx.Request.RequestURI[:i]
		q := strings.TrimSpace(this.Input().Get("q"))
		if len(q) > 0 {
			redirUrl += "?q=" + q
		}
		this.Redirect(redirUrl, 302)
		return
	}
}

var (
	refreshCount int
	cacheTicker  *time.Ticker
	cachePros    []hv.PkgInfo
)

func InitRouter() {
	// Load max element numbers.
	num := utils.Cfg.MustInt("setting", "max_pro_info_num")
	if num > 0 {
		maxProInfoNum = num
	}

	num = utils.Cfg.MustInt("setting", "max_exam_num")
	if num > 0 {
		maxExamNum = num
	}
	beego.Trace(fmt.Sprintf("maxProInfoNum: %d; maxExamNum: %d",
		maxProInfoNum, maxExamNum))

	// Start cache ticker.
	cacheTicker = time.NewTicker(time.Minute)
	go cacheTickerCheck(cacheTicker.C)

	initPopPros()
	initIndexStats()
}

func cacheTickerCheck(cacheChan <-chan time.Time) {
	for {
		<-cacheChan
		refreshCount++

		// Check if reach the maximum limit of skip.
		if refreshCount >= utils.Cfg.MustInt("task", "max_skip_time") {
			// Yes.
			refreshCount = 0
		}

		// Check if need to flush cache.
		if refreshCount == 0 || len(cachePros) >= utils.Cfg.MustInt("task", "min_pro_num") {
			FlushCache()
			initPopPros()
			refreshCount = 0
		}

		initIndexStats()
	}
}

var cacheVisitIps = make(map[int64]map[string]bool)

func FlushCache() {
	// Flush cache projects.
	num := len(cachePros)
	// Set views increament.
	for _, pinfo := range cachePros {
		pinfo.Views += int64(len(cacheVisitIps[pinfo.Id]))
	}

	models.FlushCacheProjects(cachePros)
	beego.Info("FlushCacheProjects #", num)

	cachePros = make([]hv.PkgInfo, 0, num)
	cacheVisitIps = make(map[int64]map[string]bool)
}
