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
	"strings"

	"github.com/astaxie/beego"
)

type AboutController struct {
	beego.Controller
}

// Get implemented Get method for AboutController.
// It serves about page of Go Walker.
func (this *AboutController) Get() {
	// Check language version
	reqUrl := this.Ctx.Request.RequestURI
	if len(reqUrl) == 1 {
		// English is default language version
		this.Redirect("/en/about", 302)
	}

	lang := ""
	if i := strings.LastIndex(reqUrl, "/"); i > 2 {
		lang = reqUrl[1:i]
	} else {
		this.Redirect("/en/about", 302)
	}

	// Set properties
	this.TplNames = "about_" + lang + ".html"
	this.Layout = "layout.html"
}
