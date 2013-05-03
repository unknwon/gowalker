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
	"github.com/Unknwon/gowalker/doc"
	"github.com/astaxie/beego"
)

type RefreshController struct {
	beego.Controller
}

// Get implemented Get method for RefreshController.
// It serves refresh page of Go Walker.
func (this *RefreshController) Get() {
	// Get language version
	curLang, restLangs := getLangVer(this.Input().Get("lang"))

	// Get query field
	q := this.Input().Get("q")

	// Empty query string shows home page
	if len(q) == 0 {
		this.Redirect("/"+curLang.Lang+"/", 302)
		return
	}

	// Set properties
	this.Layout = "layout.html"
	this.TplNames = "refresh_" + curLang.Lang + ".html"

	this.Data["Lang"] = curLang.Lang
	this.Data["CurLang"] = curLang.Name
	this.Data["RestLangs"] = restLangs

	_, err := doc.CheckDoc(q, doc.REFRESH_REQUEST)
	if err == nil {
		// Show search page
		this.Redirect("/search?lang="+curLang.Lang+"&q="+q, 302)
		return
	}

	// Set data
	this.Data["Path"] = q
	this.Data["LimitTime"] = err.Error()
}
