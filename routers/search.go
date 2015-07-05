// Copyright 2015 Unknwon
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
	"github.com/Unknwon/log"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/modules/base"
	"github.com/Unknwon/gowalker/modules/doc"
	"github.com/Unknwon/gowalker/modules/middleware"
)

const (
	SEARCH base.TplName = "search"
)

func Search(ctx *middleware.Context) {
	q := ctx.Query("q")

	if ctx.Query("auto_redirect") == "true" &&
		(doc.IsGoRepoPath(q) || doc.IsValidRemotePath(q)) {
		ctx.Redirect("/" + q)
		return
	}

	results, err := models.SearchPkgInfo(100, q)
	if err != nil {
		ctx.Flash.Error(err.Error(), true)
	} else {
		ctx.Data["Results"] = results
	}

	ctx.Data["Keyword"] = q
	ctx.HTML(200, SEARCH)
}

type searchResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func SearchJSON(ctx *middleware.Context) {
	q := ctx.QueryEscape("q")
	pinfos, err := models.SearchPkgInfo(7, q)
	if err != nil {
		log.ErrorD(4, "SearchPkgInfo '%s': %v", q, err)
		return
	}

	results := make([]*searchResult, len(pinfos))
	for i := range pinfos {
		results[i] = &searchResult{
			Title:       pinfos[i].ImportPath,
			Description: pinfos[i].Synopsis,
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"results": results,
	})
}
