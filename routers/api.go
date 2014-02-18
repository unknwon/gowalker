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
	"github.com/Unknwon/gowalker/hv"
	"github.com/Unknwon/gowalker/models"
)

// ApiRouter serves API service.
type ApiRouter struct {
	baseRouter
}

func (this *ApiRouter) Docs() {
	this.TplNames = "api_docs.html"
}

// Badge redirector.
func (this *ApiRouter) Badge() {
	this.Redirect("http://b.repl.ca/v1/Go_Walker-API_Documentation-green.png", 302)
}

func parseParaToInt(val string) int {
	if val == "true" {
		return 1
	} else if val == "false" {
		return 0
	}
	return -1
}

func (this *ApiRouter) Search() {
	var result struct {
		Packages []hv.PkgInfo `json:"packages"`
	}
	result.Packages = models.SearchPkg(this.GetString("key"),
		parseParaToInt(this.GetString("cmd")), parseParaToInt(this.GetString("cgo")),
		parseParaToInt(this.GetString("gorepo")), parseParaToInt(this.GetString("gosubrepo")),
		false)
	this.Data["json"] = &result
	this.ServeJson(true)
}
