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

// SearchRouter serves labels pages.
type LabelsRouter struct {
	beego.Controller
}

// Get implemented Get method for LabelsRouter.
func (this *LabelsRouter) Get() {
	this.Data["IsLabels"] = true
	// Set language version.
	curLang := setLangVer(this.Ctx, this.Input(), this.Data)

	// Set properties
	this.TplNames = "labels_" + curLang.Lang + ".html"

	// Get index page data.
	this.Data["WFPros"], this.Data["ORMPros"], this.Data["DBDPros"],
		this.Data["GUIPros"], this.Data["NETPros"], this.Data["TOOLPros"],
		_ = models.GetLabelsPageInfo()
}
