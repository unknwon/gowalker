// Copyright 2014 Unknwon
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

	"github.com/Unknwon/com"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/modules/base"
	"github.com/Unknwon/gowalker/modules/context"
	"github.com/Unknwon/gowalker/modules/doc"
)

const (
	HOME base.TplName = "home"
)

// getHistory returns browse history.
func getHistory(ctx *context.Context) []*models.PkgInfo {
	pairs := strings.Split(ctx.GetCookie("user_history"), "|")
	pkgs := make([]*models.PkgInfo, 0, len(pairs))

	for _, pair := range pairs {
		infos := strings.Split(pair, ":")
		if len(infos) != 2 {
			continue
		}

		pid := com.StrTo(infos[0]).MustInt64()
		if pid == 0 {
			continue
		}

		pinfo, _ := models.GetPkgInfoById(pid)
		if pinfo == nil {
			continue
		}

		pinfo.LastView = com.StrTo(infos[1]).MustInt64()
		pkgs = append(pkgs, pinfo)
	}
	return pkgs
}

func Home(ctx *context.Context) {
	ctx.Data["PageIsHome"] = true
	ctx.Data["NumOfPackages"] = base.FormatNumString(models.NumOfPackages())
	ctx.Data["SearchContent"] = doc.SearchContent
	ctx.Data["BrowseHistory"] = getHistory(ctx)
	ctx.HTML(200, HOME)
}
