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
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Unknwon/com"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/modules/base"
	"github.com/Unknwon/gowalker/modules/doc"
	"github.com/Unknwon/gowalker/modules/middleware"
	"github.com/Unknwon/gowalker/modules/setting"
)

const (
	DOCS         base.TplName = "docs/docs"
	DOCS_IMPORTS base.TplName = "docs/imports"
)

// updateHistory updates browser history.
func updateHistory(ctx *middleware.Context, id int64) {
	pairs := make([]string, 1, 10)
	pairs[0] = com.ToStr(id) + ":" + com.ToStr(time.Now().UTC().Unix())

	count := 0
	for _, pair := range strings.Split(ctx.GetCookie("user_history"), "|") {
		infos := strings.Split(pair, ":")
		if len(infos) != 2 {
			continue
		}

		pid := com.StrTo(infos[0]).MustInt64()
		if pid == 0 || pid == id {
			continue
		}

		pairs = append(pairs, pair)

		count++
		if count == 9 {
			break
		}
	}
	ctx.SetCookie("user_history", strings.Join(pairs, "|"), 9999999)
}

func handleError(ctx *middleware.Context, err error) {
	importPath := ctx.Params("*")
	if err == doc.ErrInvalidRemotePath {
		ctx.Redirect("/search?q=" + importPath)
		return
	}
	ctx.Flash.Error(importPath+": "+err.Error(), true)
	ctx.Flash.Info(ctx.Tr("form.click_to_search", importPath), true)
	Home(ctx)
}

func specialHandles(ctx *middleware.Context, pinfo *models.PkgInfo) bool {
	// Only show imports.
	if strings.HasSuffix(ctx.Req.RequestURI, "?imports") {
		ctx.Data["Packages"] = models.GetPkgInfosByPaths(strings.Split(pinfo.ImportPaths, "|"))
		ctx.HTML(200, DOCS_IMPORTS)
		return true
	}

	// Refresh documentation.
	if strings.HasSuffix(ctx.Req.RequestURI, "?refresh") {
		if !pinfo.CanRefresh() {
			ctx.Flash.Info(ctx.Tr("docs.refresh.too_often"))
		} else {
			importPath := ctx.Params("*")
			_, err := doc.CheckPackage(importPath, ctx.Render, doc.REQUEST_TYPE_REFRESH)
			if err != nil {
				handleError(ctx, err)
				return true
			}
		}
		ctx.Redirect(ctx.Data["Link"].(string))
		return true
	}

	return false
}

func Docs(ctx *middleware.Context) {
	importPath := ctx.Params("*")
	pinfo, err := doc.CheckPackage(importPath, ctx.Render, doc.REQUEST_TYPE_HUMAN)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.Data["PageIsDocs"] = true
	ctx.Data["Title"] = pinfo.ImportPath
	ctx.Data["ParentPath"] = path.Dir(pinfo.ImportPath)
	ctx.Data["ProjectName"] = path.Base(pinfo.ImportPath)
	ctx.Data["ProjectPath"] = pinfo.ProjectPath

	if specialHandles(ctx, pinfo) {
		return
	}

	if pinfo.IsGoRepo {
		ctx.Flash.Info(ctx.Tr("docs.turn_into_search", importPath), true)
	}

	ctx.Data["PkgDesc"] = pinfo.Synopsis

	// README.
	lang := ctx.Data["Lang"].(string)[:2]
	readmePath := setting.DocsJsPath + pinfo.ImportPath + "_RM_" + lang + ".js"
	if com.IsFile(readmePath) {
		ctx.Data["IsHasReadme"] = true
		ctx.Data["ReadmePath"] = readmePath
	} else {
		readmePath := setting.DocsJsPath + pinfo.ImportPath + "_RM_en.js"
		if com.IsFile(readmePath) {
			ctx.Data["IsHasReadme"] = true
			ctx.Data["ReadmePath"] = readmePath
		}
	}

	// Documentation.
	docJS := make([]string, 0, pinfo.JsNum+1)
	docJS = append(docJS, setting.DocsJsPath+importPath+".js")
	for i := 1; i <= pinfo.JsNum; i++ {
		docJS = append(docJS, fmt.Sprintf("%s%s-%d.js", setting.DocsJsPath, importPath, i))
	}
	ctx.Data["DocJS"] = docJS
	ctx.Data["Timestamp"] = pinfo.Created
	if time.Now().UTC().Add(-5*time.Second).Unix() < pinfo.Created {
		ctx.Flash.Success(ctx.Tr("docs.generate_success"), true)
	}

	// Notes.
	if len(pinfo.Subdirs) > 0 {
		ctx.Data["IsHasSubdirs"] = true
		ctx.Data["ViewDirPath"] = pinfo.ViewDirPath
		ctx.Data["Subdirs"] = models.GetSubPkgs(pinfo.ImportPath, strings.Split(pinfo.Subdirs, "|"))
	}

	ctx.Data["ImportNum"] = len(strings.Split(pinfo.ImportPaths, "|"))

	// Tools.
	ctx.Data["CanRefresh"] = pinfo.CanRefresh()

	updateHistory(ctx, pinfo.Id)

	ctx.HTML(200, DOCS)
}
