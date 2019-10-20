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

package route

import (
	"strings"
	"unicode"

	log "gopkg.in/clog.v1"

	"github.com/unknwon/gowalker/internal/base"
	"github.com/unknwon/gowalker/internal/context"
	"github.com/unknwon/gowalker/internal/db"
)

const (
	SEARCH = "search"
)

func Search(ctx *context.Context) {
	q := ctx.Query("q")

	// Clean up keyword.
	q = strings.TrimFunc(q, func(c rune) bool {
		return unicode.IsSpace(c) || c == '"'
	})

	if ctx.Query("auto_redirect") == "true" &&
		(base.IsGoRepoPath(q) || base.IsGAERepoPath(q) ||
			base.IsValidRemotePath(q)) {
		ctx.Redirect("/" + q)
		return
	}

	var (
		results []*db.PkgInfo
		err     error
	)
	switch q {
	case "gorepos":
		results, err = db.GetGoRepos()
	case "gosubrepos":
		results, err = db.GetGoSubepos()
	case "gaesdk":
		results, err = db.GetGAERepos()
	default:
		results, err = db.SearchPkgInfo(100, q)
	}
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
	URL         string `json:"url"`
}

func SearchJSON(ctx *context.Context) {
	q := ctx.Query("q")

	// Clean up keyword
	q = strings.TrimFunc(q, func(c rune) bool {
		return unicode.IsSpace(c) || c == '"'
	})

	pinfos, err := db.SearchPkgInfo(7, q)
	if err != nil {
		log.Error(2, "SearchPkgInfo '%s': %v", q, err)
		return
	}

	results := make([]*searchResult, len(pinfos))
	for i := range pinfos {
		results[i] = &searchResult{
			Title:       pinfos[i].ImportPath,
			Description: pinfos[i].Synopsis,
			URL:         "/" + pinfos[i].ImportPath,
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"results": results,
	})
}
