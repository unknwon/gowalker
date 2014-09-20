// Copyright 2013 Unknwon
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
	"time"

	"github.com/astaxie/beego"

	"github.com/Unknwon/gowalker/hv"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/modules/log"
	"github.com/Unknwon/gowalker/modules/setting"
)

// baseRouter implemented global settings for all other routers.
type baseRouter struct {
	beego.Controller
}

var (
	refreshCount int
	cacheTicker  *time.Ticker
	cachePros    []hv.PkgInfo
)

func init() {
	// Load max element numbers.
	num := setting.Cfg.MustInt("setting", "MAX_PRO_INFO_NUM")
	if num > 0 {
		maxProInfoNum = num
	}

	num = setting.Cfg.MustInt("setting", "MAX_EXAM_NUM")
	if num > 0 {
		maxExamNum = num
	}
	log.Trace("maxProInfoNum: %d; maxExamNum: %d", maxProInfoNum, maxExamNum)

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
		if refreshCount >= setting.Cfg.MustInt("task", "MAX_SKIP_TIME") {
			// Yes.
			refreshCount = 0
		}

		// Check if need to flush cache.
		if refreshCount == 0 || len(cachePros) >= setting.Cfg.MustInt("task", "MIN_PRO_NUM") {
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
