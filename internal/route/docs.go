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
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/unknwon/com"

	"github.com/unknwon/gowalker/internal/base"
	"github.com/unknwon/gowalker/internal/context"
	"github.com/unknwon/gowalker/internal/db"
	"github.com/unknwon/gowalker/internal/doc"
	"github.com/unknwon/gowalker/internal/setting"
)

const (
	DOCS         = "docs/docs"
	DOCS_IMPORTS = "docs/imports"
)

// updateHistory updates browser history.
func updateHistory(ctx *context.Context, id int64) {
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

func handleError(ctx *context.Context, err error) {
	importPath := ctx.Params("*")
	if err == doc.ErrInvalidRemotePath {
		ctx.Redirect("/search?q=" + importPath)
		return
	}

	if strings.Contains(err.Error(), "<meta> not found") ||
		strings.Contains(err.Error(), "resource not found") {
		db.DeletePackageByPath(importPath)
	}

	ctx.Flash.Error(importPath+": "+err.Error(), true)
	ctx.Flash.Info(ctx.Tr("form.click_to_search", importPath), true)
	Home(ctx)
}

func specialHandles(ctx *context.Context, pinfo *db.PkgInfo) bool {
	// Only show imports.
	if strings.HasSuffix(ctx.Req.RequestURI, "?imports") {
		ctx.Data["PageIsImports"] = true
		ctx.Data["Packages"] = db.GetPkgInfosByPaths(strings.Split(pinfo.ImportPaths, "|"))
		ctx.HTML(200, DOCS_IMPORTS)
		return true
	}

	// Only show references.
	if strings.HasSuffix(ctx.Req.RequestURI, "?refs") {
		ctx.Data["PageIsRefs"] = true
		ctx.Data["Packages"] = pinfo.GetRefs()
		ctx.HTML(200, DOCS_IMPORTS)
		return true
	}

	// Refresh documentation.
	if strings.HasSuffix(ctx.Req.RequestURI, "?refresh") {
		if !pinfo.CanRefresh() {
			ctx.Flash.Info(ctx.Tr("docs.refresh.too_often"))
		} else {
			importPath := ctx.Params("*")
			_, err := doc.CheckPackage(importPath, ctx.Render, doc.RequestTypeRefresh)
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

func Docs(c *context.Context) {
	importPath := c.Params("*")

	// Check if import path looks like a vendor directory
	if strings.Contains(importPath, "/vendor/") {
		handleError(c, errors.New("import path looks like is a vendor directory, don't try to fool me! :D"))
		return
	}

	if base.IsGAERepoPath(importPath) {
		c.Redirect("/google.golang.org/" + importPath)
		return
	}

	pinfo, err := doc.CheckPackage(importPath, c.Render, doc.RequestTypeHuman)
	if err != nil {
		handleError(c, err)
		return
	}

	c.PageIs("Docs")
	c.Title(pinfo.ImportPath)
	c.Data["ParentPath"] = path.Dir(pinfo.ImportPath)
	c.Data["ProjectName"] = path.Base(pinfo.ImportPath)
	c.Data["ProjectPath"] = pinfo.ProjectPath
	c.Data["NumStars"] = pinfo.Stars

	if specialHandles(c, pinfo) {
		return
	}

	if pinfo.IsGoRepo {
		c.Flash.Info(c.Tr("docs.turn_into_search", importPath), true)
	}

	c.Data["PkgDesc"] = pinfo.Synopsis

	// README
	lang := c.Data["Lang"].(string)[:2]
	readmePath := setting.DocsJSPath + pinfo.ImportPath + "_RM_" + lang + ".js"
	if com.IsFile(readmePath) {
		c.Data["IsHasReadme"] = true
		c.Data["ReadmePath"] = readmePath
	} else {
		readmePath := setting.DocsJSPath + pinfo.ImportPath + "_RM_en.js"
		if com.IsFile(readmePath) {
			c.Data["IsHasReadme"] = true
			c.Data["ReadmePath"] = readmePath
		}
	}

	// Documentation
	if pinfo.JSFile.Status == db.JSFileStatusDistributed {
		docJS := db.ComposeSpacesObjectNames(pinfo.ImportPath, pinfo.JSFile.Etag, pinfo.JSFile.NumExtraFiles)
		for i := range docJS {
			docJS[i] = setting.DigitalOcean.Spaces.BucketURL + docJS[i]
		}
		c.Data["DocJS"] = docJS

	} else {
		docJS := make([]string, 0, pinfo.JSFile.NumExtraFiles+1)
		docJS = append(docJS, "/"+setting.DocsJSPath+importPath+".js")
		for i := 1; i <= pinfo.JSFile.NumExtraFiles; i++ {
			docJS = append(docJS, fmt.Sprintf("/%s%s-%d.js", setting.DocsJSPath, importPath, i))
		}
		c.Data["DocJS"] = docJS
	}
	c.Data["Timestamp"] = pinfo.Created
	if time.Now().UTC().Add(-5*time.Second).Unix() < pinfo.Created {
		c.Flash.Success(c.Tr("docs.generate_success"), true)
	}

	// Subdirs
	if len(pinfo.Subdirs) > 0 {
		c.Data["IsHasSubdirs"] = true
		c.Data["ViewDirPath"] = pinfo.ViewDirPath
		c.Data["Subdirs"] = db.GetSubPkgs(pinfo.ImportPath, strings.Split(pinfo.Subdirs, "|"))
	}

	// Imports and references
	c.Data["ImportNum"] = pinfo.ImportNum
	c.Data["RefNum"] = pinfo.RefNum

	// Tools
	c.Data["TimeDuration"] = base.TimeSince(time.Unix(pinfo.Created, 0), c.Locale.Language())
	c.Data["CanRefresh"] = pinfo.CanRefresh()

	updateHistory(c, pinfo.ID)

	c.Success(DOCS)
}
