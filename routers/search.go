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
	"strings"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/hv"
)

// SearchRouter serves search pages.
type SearchRouter struct {
	baseRouter
}

func checkSpecialUsage(this *SearchRouter, q, t, pid, tag string) bool {
	var pinfos []*hv.PkgInfo

	switch {
	case t == "imports":
		pinfos = models.GetImports(pid, tag)
	case t == "refs":
		pinfos = models.GetRefs(pid)
	case q == "gorepos":
		pinfos = models.GetGoRepo()
	case q == "gosubrepos":
		pinfos = models.GetGoSubrepo()
	}

	if len(pinfos) > 0 {
		this.Data["IsFindPro"] = true
		this.Data["Results"] = pinfos
		this.Data["ResultCount"] = len(pinfos)
		return true
	}

	return false
}

// Get implemented Get method for SearchRouter.
func (this *SearchRouter) Get() {
	this.TplNames = "search.html"

	// Get argument(s).
	q := strings.TrimSpace(this.Input().Get("q"))
	t := strings.TrimSpace(this.Input().Get("t"))
	pid := strings.TrimSpace(this.Input().Get("pid"))
	tag := strings.TrimSpace(this.Input().Get("tag"))
	if tag == "master" || tag == "default" {
		tag = ""
	}

	if len(q) == 0 {
		this.Redirect("/", 302)
		return
	}

	this.Data["Keyword"] = q

	if checkSpecialUsage(this, q, t, pid, tag) {
		return
	}

	pinfos := models.SearchPkg(q)
	if len(pinfos) > 0 {
		this.Data["IsFindPro"] = true
		this.Data["Results"] = pinfos
		this.Data["ResultCount"] = len(pinfos)
	}
}
