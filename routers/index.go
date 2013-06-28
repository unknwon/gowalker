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
	"github.com/Unknwon/gowalker/models"
	"github.com/astaxie/beego"
)

// IndexRouter serves index pages.
type IndexRouter struct {
	beego.Controller
}

// Get implemented Get method for IndexRouter.
func (this *IndexRouter) Get() {
	this.Data["IsIndex"] = true
	// Set language version.
	curLang := setLangVer(this.Ctx, this.Input(), this.Data)

	// Set properties
	this.TplNames = "index_" + curLang.Lang + ".html"

	// Get index page data.
	this.Data["ProNum"], this.Data["PopPros"], this.Data["ImportedPros"],
		_ = models.GetIndexPageInfo()
}
