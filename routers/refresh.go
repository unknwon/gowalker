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

	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/models"
	"github.com/astaxie/beego"
)

// RefreshRouter serves search pages.
type RefreshRouter struct {
	beego.Controller
}

// Get implemented Get method for RefreshRouter.
func (this *RefreshRouter) Get() {
	// Set language version.
	curLang := setLangVer(this.Ctx, this.Input(), this.Data)

	// Get query field
	q := this.Input().Get("q")

	// Empty query string shows home page
	if len(q) == 0 {
		this.Redirect("/", 302)
		return
	}

	// Set properties
	this.TplNames = "refresh_" + curLang.Lang + ".html"

	pdoc, err := doc.CheckDoc(q, "", doc.REFRESH_REQUEST)
	if err == nil {
		pinfo := &models.PkgInfo{
			Path:        pdoc.ImportPath,
			Synopsis:    pdoc.Synopsis,
			Created:     pdoc.Created,
			ProName:     pdoc.ProjectName,
			ViewedTime:  pdoc.ViewedTime,
			Views:       pdoc.Views,
			IsCmd:       pdoc.IsCmd,
			Etag:        pdoc.Etag,
			Labels:      pdoc.Labels,
			Tags:        strings.Join(pdoc.Tags, "|||"),
			ImportedNum: pdoc.ImportedNum,
			ImportPid:   pdoc.ImportPid,
		}
		models.SaveProject(pinfo, nil, nil, nil)
		// Show search page
		this.Redirect("/"+q, 302)
		return
	}

	// Set data
	this.Data["Path"] = q
	this.Data["LimitTime"] = err.Error()
}
