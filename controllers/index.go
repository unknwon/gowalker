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

package controllers

import (
	"github.com/astaxie/beego"
)

type IndexController struct {
	beego.Controller
}

// Get implemented Get method for IndexController.
// It serves index page of Go Walker.
func (this *IndexController) Get() {
	// Get language version
	curLang, restLangs := getLangVer(this.Input().Get("lang"))

	// Set properties
	this.Layout = "layout.html"
	this.TplNames = "index_" + curLang.Lang + ".html"

	temp := []proSyno{}
	temp = append(temp, proSyno{
		Path:     "github.com/Unknwon/gowalker",
		Synopsis: "fucking aswone project",
	})
	temp = append(temp, proSyno{
		Path:     "github.com/coocood/qbs",
		Synopsis: "Package net/http is for http operations",
	})

	this.Data["AllPros"] = temp
	this.Data["ProNum"] = len(temp)
	this.Data["Lang"] = curLang.Lang
	this.Data["CurLang"] = curLang.Name
	this.Data["RestLangs"] = restLangs
}

type proSyno struct {
	Path, Synopsis string
}
