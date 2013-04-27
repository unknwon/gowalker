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
	"github.com/unknwon/gowalker/models"
)

type RefreshController struct {
	beego.Controller
}

// Get implemented Get method for RefreshController.
// It serves refresh page of Go Walker.
func (this *RefreshController) Get() {
	// Check language version
	lang, ok := isValidLanguage(this.Ctx.Request.RequestURI)
	if !ok {
		// English is default language version
		this.Redirect("/en/", 302)
		return
	}

	// Get query field
	q := this.Input().Get("q")

	// Empty query string shows home page
	if len(q) == 0 {
		this.Redirect("/"+lang+"/", 302)
	}

	_, err := models.CheckDoc(q, models.REFRESH_REQUEST)
	if err == nil {
		this.Redirect("/"+lang+"/search?q="+q, 302)
		return
	}

	// Set properties
	this.TplNames = "refresh_" + lang + ".html"
	this.Layout = "layout.html"

	// Set data
	this.Data["Path"] = q
	this.Data["LimitTime"] = err.Error()
}
