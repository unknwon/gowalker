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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/Unknwon/log"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/modules/base"
	"github.com/Unknwon/gowalker/modules/context"
)

const (
	SEARCH base.TplName = "search"
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
		results []*models.PkgInfo
		err     error
	)
	switch q {
	case "gorepos":
		results, err = models.GetGoRepos()
	case "gosubrepos":
		results, err = models.GetGoSubepos()
	case "gaesdk":
		results, err = models.GetGAERepos()
	default:
		results, err = models.SearchPkgInfo(100, q)
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

type semanticSearchResultDefDocs struct {
	Data string
}

type semanticSearchResultDef struct {
	Exported bool
	Kind     string
	Unit     string
	Path     string
	Docs     []*semanticSearchResultDefDocs
}

type semanticSearchResult struct {
	Defs []*semanticSearchResultDef
}

// semanticSearch sends search request to sourcegraph.com.
func semanticSearch(ctx *context.Context, query string) {
	resp, err := http.Get("https://sourcegraph.com/.api/global-search?Query=golang+" + url.QueryEscape(query) + "&Limit=30&Fast=1")
	if err != nil {
		log.Error("semanticSearch.http.Get: %v", err)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("semanticSearch.ioutil.ReadAll: %v", err)
		return
	}

	var sgResults semanticSearchResult
	if err = json.Unmarshal(data, &sgResults); err != nil {
		log.Error("semanticSearch.json.Unmarshal: %v", err)
		return
	}

	maxResults := 7
	results := make([]*searchResult, 0, maxResults)
	for _, def := range sgResults.Defs {
		if !def.Exported {
			continue
		}

		var title, desc, url string
		switch def.Kind {
		case "package":
			title = def.Unit
		case "func":
			anchor := strings.Replace(def.Path, "/", "_", 1)
			title = def.Unit + "#" + anchor
		default:
			continue
		}

		if len(def.Docs) > 0 {
			if len(def.Docs[0].Data) > 100 {
				desc = def.Docs[0].Data[:100] + "..."
			} else {
				desc = def.Docs[0].Data
			}
		}
		url = "/" + title

		results = append(results, &searchResult{
			Title:       title,
			Description: desc,
			URL:         url,
		})

		if len(results) >= maxResults {
			break
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"results": results,
	})
}

func SearchJSON(ctx *context.Context) {
	q := ctx.Query("q")

	// Clean up keyword.
	q = strings.TrimFunc(q, func(c rune) bool {
		return unicode.IsSpace(c) || c == '"'
	})

	if ctx.Query("semantic_search") == "true" {
		semanticSearch(ctx, q)
		return
	}

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
			URL:         "/" + pinfos[i].ImportPath,
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"results": results,
	})
}
