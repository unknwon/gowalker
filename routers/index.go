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

package routers

import (
	"strconv"

	"github.com/Unknwon/gowalker/models"
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
	baseRouter
}

// Get implemented Get method for IndexRouter.
func (this *IndexRouter) Get() {
	this.Data["IsIndex"] = true
	this.TplNames = "index.html"

	// Calculate pages.
	pn, err := strconv.Atoi(this.Input().Get("p"))
	maxPageNum := int(indexStats.ProNum/100) + 1
	if err != nil || pn > maxPageNum {
		pn = 1
	}

	if pn < 10 {
		this.Data["BackPageNum"] = 1
	} else {
		this.Data["BackPageNum"] = pn - 10
	}

	if pn > maxPageNum-10 {
		this.Data["ForwardPageNum"] = maxPageNum
	} else {
		this.Data["ForwardPageNum"] = pn + 10
	}

	this.Data["IndexPkgs"] = models.GetIndexPkgs(pn)

	// Calculate page list.
	this.Data["PageList"] = calPageList(pn, maxPageNum)

	// Set properties
	this.Data["IndexStats"] = indexStats
}

type page struct {
	IsActive bool
	PageNum  int
}

// calPageList returns page lists.
func calPageList(p, maxPageNum int) []*page {
	listSize := 15
	hls := listSize / 2
	pl := make([]*page, 0, listSize)

	start, end := p-hls, p+hls
	if p < hls+1 {
		start, end = 1, listSize
	}

	if end > maxPageNum {
		end = maxPageNum
	}

	for i := start; i <= end; i++ {
		pl = append(pl, &page{
			IsActive: i == p,
			PageNum:  i,
		})
	}
	return pl
}
