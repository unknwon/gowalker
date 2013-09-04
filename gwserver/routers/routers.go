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

// Package routers implemented controller methods of beego.
package routers

import (
	"strings"

	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

var (
	AppVer    string
	IsProMode bool
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
	// Redirect to make URL clean.
	if isNeedRedir {
		i := strings.Index(this.Ctx.Request.RequestURI, "?")
		this.Redirect(this.Ctx.Request.RequestURI[:i], 301)
		return
	}
}
