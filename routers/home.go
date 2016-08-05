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
	"fmt"
	"strings"

	"github.com/Unknwon/com"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/modules/base"
	"github.com/Unknwon/gowalker/modules/context"
)

const (
	HOME base.TplName = "home"
)

func getBrowsingHistory(ctx *context.Context) []*models.PkgInfo {
	rawInfos := strings.Split(ctx.GetCookie("user_history"), "|")
	pkgIDs := make([]int64, 0, len(rawInfos)) // ID -> Unix
	lastViewedTimes := make(map[int64]int64)
	for _, rawInfo := range rawInfos {
		fields := strings.Split(rawInfo, ":")
		if len(fields) != 2 {
			continue
		}

		pkgID := com.StrTo(fields[0]).MustInt64()
		if pkgID == 0 {
			continue
		}
		pkgIDs = append(pkgIDs, pkgID)

		lastViewedTimes[pkgID] = com.StrTo(fields[1]).MustInt64()
	}

	// Get all package info in one single query.
	pkgInfos, err := models.GetPkgInfosByIDs(pkgIDs)
	if err != nil {
		ctx.Flash.Error(fmt.Sprintf("Cannot get browsing history: %v", err), true)
		return nil
	}

	pkgInfosSet := make(map[int64]*models.PkgInfo)
	for i := range pkgInfos {
		pkgInfosSet[pkgInfos[i].ID] = pkgInfos[i]
	}

	// Assign package info in the same order they stored in cookie.
	pkgInfos = make([]*models.PkgInfo, 0, len(pkgIDs))
	for i := range pkgIDs {
		if pkgInfosSet[pkgIDs[i]] == nil {
			continue
		}

		pkgInfosSet[pkgIDs[i]].LastViewed = lastViewedTimes[pkgIDs[i]]
		pkgInfos = append(pkgInfos, pkgInfosSet[pkgIDs[i]])
	}

	return pkgInfos
}

func Home(ctx *context.Context) {
	ctx.Data["PageIsHome"] = true
	ctx.Data["NumTotalPackages"] = base.FormatNumString(models.NumTotalPackages())
	ctx.Data["BrowsingHistory"] = getBrowsingHistory(ctx)
	ctx.HTML(200, HOME)
}
