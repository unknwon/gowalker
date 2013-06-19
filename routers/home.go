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
	"github.com/astaxie/beego"
)

// HomeRouter serves home and documentation pages.
type HomeRouter struct {
	beego.Controller
}

// Get implemented Get method for HomeRouter.
func (this *HomeRouter) Get() {
	// Link highlight.
	this.Data["IsHome"] = true
	this.Data["DataSrc"] = ""
	this.TplNames = "home_en.html"
}
