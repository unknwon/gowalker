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

package controllers

import (
	"bytes"
	godoc "go/doc"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

type SearchController struct {
	beego.Controller
}

// Get implemented Get method for SearchController.
// It serves search page of Go Walker.
func (this *SearchController) Get() {
	// Get language version.
	curLang, restLangs := getLangVer(this.Input().Get("lang"))

	// Get arguments.
	q := this.Input().Get("q")

	// Empty query string shows home page
	if len(q) == 0 {
		this.Redirect("/"+curLang.Lang+"/", 302)
		return
	}

	// Set properties.
	this.Layout = "layout.html"
	this.TplNames = "search_" + curLang.Lang + ".html"

	this.Data["DataSrc"] = utils.GoRepoSet
	this.Data["Keyword"] = q
	this.Data["Lang"] = curLang.Lang
	this.Data["CurLang"] = curLang.Name
	this.Data["RestLangs"] = restLangs

	if checkSpecialUsage(this, q) {
		return
	}

	rawPath := q // Raw path.
	rawPath = strings.Replace(rawPath, "http://", "", 1)
	rawPath = strings.Replace(rawPath, "https://", "", 1)

	if utils.IsGoRepoPath(q) {
		q = "code.google.com/p/go/source/browse/src/pkg/" + q
	}

	if path, ok := utils.IsBrowseURL(q); ok {
		q = path
	}

	// Check if it is a remote path, if not means it's a keyword.
	if utils.IsValidRemotePath(q) {
		// Check documentation of this import path, and update automatically as needed.

		/* TODO:WORKING */

		pdoc, err := doc.CheckDoc(rawPath, doc.HUMAN_REQUEST)
		if err == nil {
			// Generate documentation page.

			/* TODO */

			if pdoc != nil && generatePage(this, pdoc, q, curLang.Lang) {
				// Update recent projects
				updateRecentPros(pdoc)
				pinfo := &models.PkgInfo{
					Path:       pdoc.ImportPath,
					Synopsis:   pdoc.Synopsis,
					Created:    pdoc.Created,
					ProName:    pdoc.ProjectName,
					ViewedTime: pdoc.ViewedTime,
					Views:      pdoc.Views,
				}
				// Updated views
				models.AddViews(pinfo)
				return
			}
		} else {
			beego.Error("SearchController.Get():", err)
		}
	}

	// Returns a slice of results
	pkgInfos, _ := models.SearchDoc(rawPath)
	// Show results after searched
	if len(pkgInfos) > 0 {
		this.Data["IsFindPro"] = true
		this.Data["AllPros"] = pkgInfos
	}
}

// checkSpecialUsage checks special usage of keywords.
// It returns true if it is a special usage, false otherwise.
func checkSpecialUsage(this *SearchController, q string) bool {
	switch {
	case q == "gorepo":
		// Show list of standard library
		pkgInfos, _ := models.GetGoRepo()
		// Show results after searched
		if len(pkgInfos) > 0 {
			this.Data["IsFindPro"] = true
			this.Data["AllPros"] = pkgInfos
		}
		return true
	}

	return false
}

// generatePage genarates documentation page for project.
// it returns false when its a invaild(empty) project.
func generatePage(this *SearchController, pdoc *doc.Package, q string, lang string) bool {
	// Set properties
	this.TplNames = "docs_" + lang + ".html"

	// Refresh
	this.Data["IsRefresh"] = pdoc.Created.Add(10 * time.Second).UTC().After(time.Now().UTC())

	// Title
	this.Data["ProPath"] = q
	this.Data["Views"] = pdoc.Views + 1

	// Remove last "/"
	if urlLen := len(q); q[urlLen-1] == '/' {
		q = q[:urlLen-1]
	}

	lastIndex := strings.LastIndex(q, "/")
	if utils.IsGoRepoPath(pdoc.ImportPath) {
		this.Data["IsGoRepo"] = true
		proName := q[lastIndex+1:]
		if i := strings.Index(proName, "?"); i > -1 {
			proName = proName[:i]
		}
		this.Data["ProName"] = proName
	} else {
		this.Data["ProName"] = pdoc.ProjectName
	}
	pkgDocPath := q[:lastIndex]
	this.Data["ProDocPath"] = pkgDocPath

	// Introduction
	this.Data["ImportPath"] = pdoc.ImportPath
	pdecl, err := models.LoadProject(pdoc.ImportPath)
	if err != nil {
		beego.Error("SearchController.generatePage(): models.LoadProject()", err)
		return false
	}
	var buf bytes.Buffer
	godoc.ToHTML(&buf, pdecl.Doc, nil)
	pkgInfo := buf.String()
	pkgInfo = strings.Replace(pkgInfo, "<p>", "<p><b>", 1)
	pkgInfo = strings.Replace(pkgInfo, "</p>", "</b></p>", 1)
	this.Data["PkgFullIntro"] = pkgInfo

	err = ConvertDataFormat(pdoc, pdecl)
	if err != nil {
		beego.Error("SearchController.generatePage(): ConvertDataFormat", err)
		return false
	}

	links := make([]*utils.Link, 0, len(pdoc.Types)+len(pdoc.Imports))
	// Get all types and import packages
	for _, t := range pdoc.Types {
		links = append(links, &utils.Link{
			Name:    t.Name,
			Comment: t.Doc,
		})
	}

	// Index
	this.Data["IsHasConst"] = len(pdoc.Consts) > 0
	this.Data["IsHasVar"] = len(pdoc.Vars) > 0
	this.Data["Funcs"] = pdoc.Funcs
	for i, f := range pdoc.Funcs {
		buf.Reset()
		godoc.ToHTML(&buf, f.Doc, nil)
		f.Doc = buf.String()
		buf.Reset()
		utils.FormatCode(&buf, f.Decl, links)
		f.FmtDecl = buf.String()
		pdoc.Funcs[i] = f
	}
	this.Data["Types"] = pdoc.Types
	for i, t := range pdoc.Types {
		for j, f := range t.Funcs {
			buf.Reset()
			godoc.ToHTML(&buf, f.Doc, nil)
			f.Doc = buf.String()
			buf.Reset()
			utils.FormatCode(&buf, f.Decl, links)
			f.FmtDecl = buf.String()
			t.Funcs[j] = f
		}
		for j, m := range t.Methods {
			buf.Reset()
			godoc.ToHTML(&buf, m.Doc, nil)
			m.Doc = buf.String()
			buf.Reset()
			utils.FormatCode(&buf, m.Decl, links)
			m.FmtDecl = buf.String()
			t.Methods[j] = m
		}
		buf.Reset()
		godoc.ToHTML(&buf, t.Doc, nil)
		t.Doc = buf.String()
		buf.Reset()
		utils.FormatCode(&buf, t.Decl, links)
		t.FmtDecl = buf.String()
		pdoc.Types[i] = t
	}

	// Constants
	this.Data["Consts"] = pdoc.Consts
	for i, v := range pdoc.Consts {
		buf.Reset()
		utils.FormatCode(&buf, v.Decl, links)
		v.FmtDecl = buf.String()
		pdoc.Consts[i] = v
	}

	// Variables
	this.Data["Vars"] = pdoc.Vars
	for i, v := range pdoc.Vars {
		buf.Reset()
		utils.FormatCode(&buf, v.Decl, links)
		v.FmtDecl = buf.String()
		pdoc.Vars[i] = v
	}

	this.Data["Files"] = pdoc.Files
	this.Data["ImportPkgNum"] = len(pdoc.Imports)
	this.Data["UtcTime"] = pdoc.Created
	this.Data["GOOS"] = pdecl.Goos
	this.Data["GOARCH"] = pdecl.Goarch
	return true
}

// ConvertDataFormat converts data from database acceptable format to useable format.
func ConvertDataFormat(pdoc *doc.Package, pdecl *models.PkgDecl) error {
	// Consts
	pdoc.Consts = make([]*doc.Value, 0, 5)
	for _, v := range strings.Split(pdecl.Consts, "&$#") {
		val := new(doc.Value)
		for j, s := range strings.Split(v, "&V#") {
			switch j {
			case 0: // Name
				val.Name = s
			case 1: // Doc
				val.Doc = s
			case 2: // Decl
				val.Decl = s
			case 3: // URL
				val.URL = s
			}
		}
		pdoc.Consts = append(pdoc.Consts, val)
	}
	pdoc.Consts = pdoc.Consts[:len(pdoc.Consts)-1]

	// Variables
	pdoc.Vars = make([]*doc.Value, 0, 5)
	for _, v := range strings.Split(pdecl.Vars, "&$#") {
		val := new(doc.Value)
		for j, s := range strings.Split(v, "&V#") {
			switch j {
			case 0: // Name
				val.Name = s
			case 1: // Doc
				val.Doc = s
			case 2: // Decl
				val.Decl = s
			case 3: // URL
				val.URL = s
			}
		}
		pdoc.Vars = append(pdoc.Vars, val)
	}
	pdoc.Vars = pdoc.Vars[:len(pdoc.Vars)-1]

	// Functions
	pdoc.Funcs = make([]*doc.Func, 0, 10)
	for _, v := range strings.Split(pdecl.Funcs, "&$#") {
		val := new(doc.Func)
		for j, s := range strings.Split(v, "&F#") {
			switch j {
			case 0: // Name
				val.Name = s
			case 1: // Doc
				val.Doc = s
			case 2: // Decl
				val.Decl = s
			case 3: // URL
				val.URL = s
			case 4: // Code
				val.Code = s
			}
		}
		pdoc.Funcs = append(pdoc.Funcs, val)
	}
	pdoc.Funcs = pdoc.Funcs[:len(pdoc.Funcs)-1]

	// Types
	pdoc.Types = make([]*doc.Type, 0, 10)
	for _, v := range strings.Split(pdecl.Types, "&##") {
		val := new(doc.Type)
		for j, s := range strings.Split(v, "&$#") {
			switch j {
			case 0: // Type
				for y, s2 := range strings.Split(s, "&T#") {
					switch y {
					case 0: // Name
						val.Name = s2
					case 1: // Doc
						val.Doc = s2
					case 2: // Decl
						val.Decl = s2
					case 3: // URL
						val.URL = s2
					}
				}
			case 1: // Functions
				val.Funcs = make([]*doc.Func, 0, 2)
				for _, v2 := range strings.Split(s, "&M#") {
					val2 := new(doc.Func)
					for y, s2 := range strings.Split(v2, "&F#") {
						switch y {
						case 0: // Name
							val2.Name = s2
						case 1: // Doc
							val2.Doc = s2
						case 2: // Decl
							val2.Decl = s2
						case 3: // URL
							val2.URL = s2
						case 4: // Code
							val2.Code = s2
						}
					}
					val.Funcs = append(val.Funcs, val2)
				}
				val.Funcs = val.Funcs[:len(val.Funcs)-1]
			case 3: // Methods.
				val.Methods = make([]*doc.Func, 0, 5)
				for _, v2 := range strings.Split(s, "&M#") {
					val2 := new(doc.Func)
					for y, s2 := range strings.Split(v2, "&F#") {
						switch y {
						case 0: // Name
							val2.Name = s2
						case 1: // Doc
							val2.Doc = s2
						case 2: // Decl
							val2.Decl = s2
						case 3: // URL
							val2.URL = s2
						case 4: // Code
							val2.Code = s2
						}
					}
					val.Methods = append(val.Methods, val2)
				}
				val.Methods = val.Methods[:len(val.Methods)-1]
			}
		}
		pdoc.Types = append(pdoc.Types, val)
	}
	pdoc.Types = pdoc.Types[:len(pdoc.Types)-1]

	// Files.
	pdoc.Files = strings.Split(pdecl.Files, "&$#")

	// Imports.
	pdoc.Imports = strings.Split(pdecl.Imports, "&$#")
	return nil
}

func updateRecentPros(pdoc *doc.Package) {
	index := -1
	listLen := len(recentViewedPros)
	t := time.Now().UTC().String()
	curPro := &recentPro{
		Path:       pdoc.ImportPath,
		ViewedTime: t[:19],
		IsGoRepo:   pdoc.ProjectName == "Go",
		Views:      pdoc.Views,
	}

	pdoc.ViewedTime = curPro.ViewedTime

	// Check if in the list
	for i, s := range recentViewedPros {
		if s.Path == curPro.Path {
			index = i
			break
		}
	}

	s := make([]*recentPro, 0, recentViewsProNum)
	s = append(s, curPro)
	switch {
	case index == -1 && listLen < recentViewsProNum:
		// Not found and list is not full
		s = append(s, recentViewedPros...)
	case index == -1 && listLen >= recentViewsProNum:
		// Not found but list is full
		s = append(s, recentViewedPros[:recentViewsProNum-1]...)
	case index > -1:
		// Found
		s = append(s, recentViewedPros[:index]...)
		s = append(s, recentViewedPros[index+1:]...)
	}
	recentViewedPros = s
}
