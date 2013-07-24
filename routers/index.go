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
	"strconv"

	"github.com/Unknwon/gowalker/models"
	"github.com/astaxie/beego"
)

// indexStats represents statistic information.
var indexStats struct {
	ProNum, DeclNum, FuncNum int64
}

// initIndexStats initializes index page statistic information.
func initIndexStats() {
	indexStats.ProNum, indexStats.DeclNum, indexStats.FuncNum = models.GetIndexStats()
}

// IndexRouter serves index pages.
type IndexRouter struct {
	beego.Controller
}

// Get implemented Get method for IndexRouter.
func (this *IndexRouter) Get() {
	this.Data["IsIndex"] = true
	// Set language version.
	curLang := globalSetting(this.Ctx, this.Input(), this.Data)

	// Calculate pages.
	pn, err := strconv.Atoi(this.Input().Get("p"))
	maxPageNum := int(indexStats.ProNum/100) + 1
	if err != nil || pn > maxPageNum {
		pn = 1
	}

	if pn == 1 {
		this.Data["IsBackDisable"] = true
	} else {
		this.Data["BackPageNum"] = pn - 1
	}

	if pn == maxPageNum {
		this.Data["IsForwardDisable"] = true
	} else {
		this.Data["ForwardPageNum"] = pn + 1
	}

	this.Data["IndexPkgs"] = models.GetIndexPkgs(pn)

	// Calculate page list.
	this.Data["BackPageList"], this.Data["ForwardPageList"] = calPageList(pn, maxPageNum)

	// Set properties
	this.TplNames = "index_" + curLang.Lang + ".html"
	this.Data["IndexStats"] = indexStats
}

type page struct {
	IsActive bool
	PageNum  int
}

// calPageList returns back and forward page lists.
func calPageList(p, maxPageNum int) ([]*page, []*page) {
	listSize := 9
	bl := make([]*page, 0, listSize)
	fl := make([]*page, 0, listSize)

	start, end := p-listSize/2, p+listSize/2
	if p < listSize/2+1 {
		start, end = 1, listSize
	}

	if end > maxPageNum {
		end = maxPageNum
	}

	for i := start; i <= end; i++ {
		bl = append(bl, &page{
			IsActive: i == p,
			PageNum:  i,
		})
	}

	if maxPageNum > listSize {
		start, end = p-listSize/2, p+listSize/2
		if p > maxPageNum-3 {
			start, end = maxPageNum-listSize+1, maxPageNum
		}

		for i := start; i <= end; i++ {
			fl = append(fl, &page{
				IsActive: i == p,
				PageNum:  i,
			})
		}
	}
	return bl, fl
}
