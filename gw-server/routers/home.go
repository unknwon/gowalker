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
	"bytes"
	"encoding/base32"
	"fmt"
	godoc "go/doc"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

// A proInfo represents a project information.
type proInfo struct {
	Pid               int64
	Path, Synopsis    string
	IsGoRepo          bool
	Views, ViewedTime int64

	/*
		For recent and popular projects, Rank is the total rank value;
		for Rock projects, Rank is for rank value in this week.
	*/
	Rank int64
}

var (
	maxProInfoNum = 20
	maxExamNum    = 15

	recentUpdatedExs                                       []*models.PkgExam
	recentViewedPros, topRankPros, topViewedPros, RockPros []*proInfo

	labelList []string // Projects label list.
	labelSet  string   // Label data source.
	labels    []string
)

func init() {
	// Initialize project labels.
	labelList = strings.Split(utils.Cfg.MustValue("setting", "labels"), "|")
	for _, s := range labelList {
		labelSet += "&quot;" + s + "&quot;,"
	}
	labelSet = labelSet[:len(labelSet)-1]
	labels = strings.Split(utils.Cfg.MustValue("setting", "label_names"), "|")
}

// initPopPros initializes popular projects.
func initPopPros() {
	popPros := make([][]*models.PkgInfo, 4)
	var err error
	err, recentUpdatedExs, popPros[0], popPros[1], popPros[2], popPros[3] =
		models.GetPopulars(maxProInfoNum, maxExamNum)
	if err != nil {
		panic("initPopPros -> " + err.Error())
	}

	for i, ps := range popPros {
		tmpPros := make([]*proInfo, 0, maxProInfoNum)
		for _, p := range ps {
			tmpPros = append(tmpPros,
				&proInfo{
					Pid:      p.Id,
					Path:     p.Path,
					Synopsis: p.Synopsis,
					IsGoRepo: p.ProName == "Go" &&
						strings.Index(p.Path, ".") == -1,
					Views:      p.Views,
					ViewedTime: p.ViewedTime,
					Rank:       p.Rank,
				})
		}

		switch i {
		case 0:
			recentViewedPros = tmpPros
		case 1:
			topRankPros = tmpPros
		case 2:
			topViewedPros = tmpPros
		case 3:
			RockPros = tmpPros
		}
	}
}

// HomeRouter serves home and documentation pages.
type HomeRouter struct {
	beego.Controller
}

// Get implemented Get method for HomeRouter.
func (this *HomeRouter) Get() {
	// Filter unusual User-Agent.
	ua := this.Ctx.Request.Header.Get("User-Agent")
	if len(ua) < 20 {
		beego.Warn("User-Agent:", this.Ctx.Request.Header.Get("User-Agent"))
		this.Ctx.WriteString("")
		return
	}

	// Set language version.
	curLang := globalSetting(this.Ctx, this.Input(), this.Data)

	// Get argument(s).
	q := strings.TrimRight(
		strings.TrimSpace(this.Input().Get("q")), "/")

	if path, ok := utils.IsBrowseURL(q); ok {
		q = path
	}

	// Get pure URL.
	reqUrl := this.Ctx.Request.RequestURI[1:]
	if i := strings.Index(reqUrl, "?"); i > -1 {
		reqUrl = reqUrl[:i]
		if path, ok := utils.IsBrowseURL(reqUrl); ok {
			reqUrl = path
		}
	}

	// Redirect to query string.
	if len(reqUrl) == 0 && len(q) > 0 {
		reqUrl = q
		this.Redirect("/"+reqUrl, 302)
		return
	}

	// User Recent projects.
	urpids, _ := this.Ctx.Request.Cookie("UserRecentPros")
	urpts, _ := this.Ctx.Request.Cookie("URPTimestamps")

	this.TplNames = "home_" + curLang.Lang + ".html"
	// Check to show home page or documentation page.
	if len(reqUrl) == 0 && len(q) == 0 {
		// Home.
		this.Data["IsHome"] = true

		// Global Recent projects.
		this.Data["GlobalRecentPros"] = recentViewedPros
		// User Recent projects.
		if urpids != nil && urpts != nil {
			upros := models.GetGroupPkgInfoById(strings.Split(urpids.Value, "|"))
			pts := strings.Split(urpts.Value, "|")
			for i, p := range upros {
				ts, _ := strconv.ParseInt(pts[i], 10, 64)
				p.ViewedTime = ts
			}
			this.Data["UserRecentPros"] = upros
		}

		// Popular projects and examples.
		this.Data["TopRankPros"] = topRankPros
		this.Data["TopViewedPros"] = topViewedPros
		this.Data["RockPros"] = RockPros
		this.Data["RecentExams"] = recentUpdatedExs
		// Standard library type-ahead.
		this.Data["DataSrc"] = utils.GoRepoSet
	} else {
		// Documentation.
		this.TplNames = "docs_" + curLang.Lang + ".html"
		broPath := reqUrl // Browse path.

		// Check if it's the standard library.
		if utils.IsGoRepoPath(broPath) {
			broPath = "code.google.com/p/go/source/browse/src/pkg/" + broPath
		}

		// Check if it's a remote path that can be used for 'go get', if not means it's a keyword.
		if !utils.IsValidRemotePath(broPath) {
			// Search.
			this.Redirect("/search?q="+reqUrl, 302)
			return
		}

		// Get tag field.
		tag := strings.TrimSpace(this.Input().Get("tag"))
		if tag == "master" || tag == "default" {
			tag = ""
		}

		// Check documentation of this import path, and update automatically as needed.
		pdoc, err := doc.CheckDoc(reqUrl, tag, doc.HUMAN_REQUEST)
		if err == nil {
			if pdoc != nil {
				pdoc.UserExamples = getUserExamples(pdoc.ImportPath)
				// Generate documentation page.
				if generatePage(this, pdoc, broPath, tag, curLang.Lang) {
					ps, ts := updateCacheInfo(pdoc, urpids, urpts)
					this.Ctx.SetCookie("UserRecentPros", ps, 9999999999, "/")
					this.Ctx.SetCookie("URPTimestamps", ts, 9999999999, "/")
					return
				}
			}
		} else {
			this.Data["IsHasError"] = true
			this.Data["ErrMsg"] = strings.Replace(err.Error(),
				doc.GetGithubCredentials(), "<githubCred>", 1)
			beego.Error("HomeRouter.Get ->", err)
		}

		this.Redirect("/search?q="+reqUrl, 302)
		return
	}
}

// getUserExamples returns user examples of given import path.
func getUserExamples(path string) []*doc.Example {
	gists, _ := models.GetPkgExams(path)
	// Doesn't have Gists.
	if len(gists) == 0 {
		return nil
	}

	pexams := make([]*doc.Example, 0, 5)
	for _, g := range gists {
		exams := convertDataFormatExample(g.Examples, "_"+g.Gist[strings.LastIndex(g.Gist, "/")+1:])
		pexams = append(pexams, exams...)
	}
	return pexams
}

// generatePage genarates documentation page for project.
// it returns false when it's a invaild(empty) project.
func generatePage(this *HomeRouter, pdoc *doc.Package, q, tag, lang string) bool {
	pdecl, err := models.LoadProject(pdoc.Id, tag)
	if err != nil {
		beego.Error("HomeController.generatePage ->", err)
		return false
	}

	// Set properties.
	this.TplNames = "docs_" + lang + ".html"

	// Refresh (within 10 seconds).
	this.Data["IsRefresh"] = pdoc.Created.UTC().Add(10 * time.Second).After(time.Now().UTC())

	// Get VCS name, project name, project home page, Upper level project URL, and project tag.
	this.Data["VCS"], this.Data["ProName"], this.Data["ProPath"], this.Data["ProDocPath"], this.Data["PkgTag"] =
		getVCSInfo(q, tag, pdoc)

	if utils.IsGoRepoPath(pdoc.ImportPath) &&
		strings.Index(pdoc.ImportPath, ".") == -1 {
		this.Data["IsGoRepo"] = true
	}

	// Introduction.
	this.Data["ImportPath"] = pdoc.ImportPath
	byts, _ := base32.StdEncoding.DecodeString(
		models.LoadPkgDoc(pdoc.ImportPath, lang, "rm"))
	if len(byts) > 0 {
		this.Data["IsHasReadme"] = true
		this.Data["PkgDoc"] = string(byts)
	}
	byts, _ = base32.StdEncoding.DecodeString(pdecl.Doc)
	this.Data["PkgFullIntro"] = string(byts)

	var buf bytes.Buffer
	err = ConvertDataFormat(pdoc, pdecl)
	if err != nil {
		beego.Error("HomeController.generatePage -> ConvertDataFormat:", err)
		return false
	}

	links := make([]*utils.Link, 0, len(pdoc.Types)+len(pdoc.Imports)+len(pdoc.Funcs)+10)
	// Get all types, functions and import packages
	for _, t := range pdoc.Types {
		links = append(links, &utils.Link{
			Name:    t.Name,
			Comment: template.HTMLEscapeString(t.Doc),
		})
		buf.WriteString("&quot;" + t.Name + "&quot;,")
	}

	for _, f := range pdoc.Funcs {
		links = append(links, &utils.Link{
			Name:    f.Name,
			Comment: template.HTMLEscapeString(f.Doc),
		})
		buf.WriteString("&quot;" + f.Name + "&quot;,")
	}

	for _, t := range pdoc.Types {
		for _, f := range t.Funcs {
			links = append(links, &utils.Link{
				Name:    f.Name,
				Comment: template.HTMLEscapeString(f.Doc),
			})
			buf.WriteString("&quot;" + f.Name + "&quot;,")
		}

		for _, m := range t.Methods {
			buf.WriteString("&quot;" + t.Name + "." + m.Name + "&quot;,")
		}
	}

	// Ignore C.
	for _, v := range pdoc.Imports {
		if v != "C" {
			links = append(links, &utils.Link{
				Name: path.Base(v) + ".",
				Path: v,
			})
		}
	}

	// Set exported objects type-ahead.
	exportDataSrc := buf.String()
	if len(exportDataSrc) > 0 {
		this.Data["IsHasExports"] = true
		exportDataSrc = exportDataSrc[:len(exportDataSrc)-1]
		this.Data["ExportDataSrc"] = exportDataSrc
	}

	// Commented and total objects number.
	var comNum, totalNum int

	// Index.
	this.Data["IsHasConst"] = len(pdoc.Consts) > 0
	this.Data["IsHasVar"] = len(pdoc.Vars) > 0

	// Constants.
	this.Data["Consts"] = pdoc.Consts
	for i, v := range pdoc.Consts {
		buf.Reset()
		v.Decl = template.HTMLEscapeString(v.Decl)
		v.Decl = strings.Replace(v.Decl, "&#34;", "\"", -1)
		utils.FormatCode(&buf, &v.Decl, links)
		v.FmtDecl = buf.String()
		pdoc.Consts[i] = v
	}

	// Variables.
	this.Data["Vars"] = pdoc.Vars
	for i, v := range pdoc.Vars {
		buf.Reset()
		utils.FormatCode(&buf, &v.Decl, links)
		v.FmtDecl = buf.String()
		pdoc.Vars[i] = v
	}

	this.Data["Funcs"] = pdoc.Funcs
	for i, f := range pdoc.Funcs {
		if len(f.Doc) > 0 {
			buf.Reset()
			godoc.ToHTML(&buf, f.Doc, nil)
			f.Doc = buf.String()
			comNum++
		}
		buf.Reset()
		utils.FormatCode(&buf, &f.Decl, links)
		f.FmtDecl = buf.String()
		if exs := getExamples(pdoc, "", f.Name); len(exs) > 0 {
			f.IsHasExam = true
			f.Exams = exs
		}
		totalNum++
		pdoc.Funcs[i] = f
	}

	this.Data["Types"] = pdoc.Types
	for i, t := range pdoc.Types {
		for j, v := range t.Consts {
			buf.Reset()
			v.Decl = template.HTMLEscapeString(v.Decl)
			v.Decl = strings.Replace(v.Decl, "&#34;", "\"", -1)
			utils.FormatCode(&buf, &v.Decl, links)
			v.FmtDecl = buf.String()
			t.Consts[j] = v
		}
		for j, v := range t.Vars {
			buf.Reset()
			utils.FormatCode(&buf, &v.Decl, links)
			v.FmtDecl = buf.String()
			t.Vars[j] = v
		}

		for j, f := range t.Funcs {
			if len(f.Doc) > 0 {
				buf.Reset()
				godoc.ToHTML(&buf, f.Doc, nil)
				f.Doc = buf.String()
				comNum++
			}
			buf.Reset()
			utils.FormatCode(&buf, &f.Decl, links)
			f.FmtDecl = buf.String()
			if exs := getExamples(pdoc, "", f.Name); len(exs) > 0 {
				f.IsHasExam = true
				f.Exams = exs
			}
			totalNum++
			t.Funcs[j] = f
		}
		for j, m := range t.Methods {
			if len(m.Doc) > 0 {
				buf.Reset()
				godoc.ToHTML(&buf, m.Doc, nil)
				m.Doc = buf.String()
				comNum++
			}
			buf.Reset()
			utils.FormatCode(&buf, &m.Decl, links)
			m.FmtDecl = buf.String()
			if exs := getExamples(pdoc, t.Name, m.Name); len(exs) > 0 {
				m.IsHasExam = true
				m.Exams = exs
			}
			totalNum++
			t.Methods[j] = m
		}
		if len(t.Doc) > 0 {
			buf.Reset()
			godoc.ToHTML(&buf, t.Doc, nil)
			t.Doc = buf.String()
			comNum++
		}
		buf.Reset()
		utils.FormatCode(&buf, &t.Decl, links)
		t.FmtDecl = buf.String()
		if exs := getExamples(pdoc, "", t.Name); len(exs) > 0 {
			t.IsHasExam = true
			t.Exams = exs
		}
		totalNum++
		pdoc.Types[i] = t
	}

	if !pdoc.IsCmd {
		// Calculate documentation complete %.
		this.Data["DocCPLabel"], this.Data["DocCP"] = calDocCP(comNum, totalNum)

		// Examples.
		this.Data["IsHasExams"] = len(pdoc.Examples)+len(pdoc.UserExamples) > 0
		this.Data["Exams"] = append(pdoc.Examples, pdoc.UserExamples...)

		// Tags.
		this.Data["IsHasTags"] = len(pdoc.Tags) > 1
		if len(tag) == 0 {
			tag = "master"
		}
		this.Data["CurTag"] = tag
		this.Data["Tags"] = pdoc.Tags
	} else {
		this.Data["IsCmd"] = true
	}

	// Dirs.
	pinfos := make([]*models.PkgInfo, 0, len(pdoc.Dirs))
	for _, v := range pdoc.Dirs {
		if len(v) > 0 {
			v = pdoc.ImportPath + "/" + v
			// TODO: Can be reduce to once database connection.
			// Note: This step will be deleted after served static pages.
			if pinfo, err := models.GetPkgInfo(v, tag); err == nil {
				pinfos = append(pinfos, pinfo)
			} else {
				pinfos = append(pinfos, &models.PkgInfo{Path: v})
			}
		}
	}
	this.Data["IsHasSubdirs"] = len(pinfos) > 0
	this.Data["Subdirs"] = pinfos

	// Files.
	if len(pdoc.Files) > 0 {
		this.Data["IsHasFiles"] = pdoc.Files
		this.Data["Files"] = pdoc.Files
	}

	// Examples.
	links = append(links, &utils.Link{
		Name: path.Base(pdoc.ImportPath) + ".",
	})

	for _, e := range pdoc.Examples {
		buf.Reset()
		utils.FormatCode(&buf, &e.Code, links)
		e.Code = buf.String()
	}
	for _, e := range pdoc.UserExamples {
		buf.Reset()
		utils.FormatCode(&buf, &e.Code, links)
		e.Code = buf.String()
	}

	this.Data["Pid"] = pdecl.Id
	this.Data["Rank"] = pdoc.Rank
	this.Data["Views"] = pdoc.Views + 1
	this.Data["Labels"] = getLabels(pdoc.Labels)
	this.Data["LabelDataSrc"] = labelSet
	this.Data["ImportPkgs"] = pdecl.Imports
	this.Data["ImportPkgNum"] = len(pdoc.Imports)
	this.Data["IsHasImports"] = len(pdoc.Imports) > 0
	this.Data["IsImported"] = pdoc.ImportedNum > 0
	this.Data["ImportPid"] = pdoc.ImportPid
	this.Data["ImportedNum"] = pdoc.ImportedNum
	this.Data["UtcTime"] = pdoc.Created
	this.Data["TimeSince"] = calTimeSince(pdoc.Created)
	return true
}

// calTimeSince returns time interval from documentation generated to now with friendly format.
// TODO: Chinese.
func calTimeSince(created time.Time) string {
	mins := int(time.Since(created).Minutes())

	switch {
	case mins < 0:
		return fmt.Sprintf("in %d minutes later", -mins)
	case mins < 1:
		return "less than 1 minute"
	case mins < 60:
		return fmt.Sprintf("%d minutes ago", mins)
	case mins < 60*24:
		return fmt.Sprintf("%d hours ago", mins/(60))
	case mins < 60*24*30:
		return fmt.Sprintf("%d days ago", mins/(60*24))
	case mins < 60*24*365:
		return fmt.Sprintf("%d months ago", mins/(60*24*30))
	default:
		return fmt.Sprintf("%d years ago", mins/(60*24*365))
	}
}

// calDocCP returns label style name and percentage string according to commented and total pbjects number.
func calDocCP(comNum, totalNum int) (label, perStr string) {
	if totalNum == 0 {
		totalNum = 1
	}
	per := comNum * 100 / totalNum
	perStr = strings.Replace(
		fmt.Sprintf("%dPER(%d/%d)", per, comNum, totalNum), "PER", "%", 1)
	switch {
	case per > 80:
		label = "success"
	case per > 50:
		label = "warning"
	default:
		label = "important"
	}
	return label, perStr
}

// getExamples returns index of function example if it exists.
func getExamples(pdoc *doc.Package, typeName, name string) (exams []*doc.Example) {
	matchName := name
	if len(typeName) > 0 {
		matchName = typeName + "_" + name
	}

	for i, v := range pdoc.Examples {
		// Already used or doesn't match.
		if v.IsUsed || !strings.HasPrefix(v.Name, matchName) {
			continue
		}

		// Check if it has right prefix.
		index := strings.Index(v.Name, "_")
		// Not found "_", name length shoule be equal.
		if index == -1 && (len(v.Name) != len(name)) {
			continue
		}

		// Found "_", prefix length shoule be equal.
		if index > -1 && len(typeName) == 0 && (index > len(name)) {
			continue
		}

		pdoc.Examples[i].IsUsed = true
		exams = append(exams, v)
	}

	for i, v := range pdoc.UserExamples {
		// Already used or doesn't match.
		if v.IsUsed || !strings.HasPrefix(v.Name, matchName) {
			continue
		}

		pdoc.UserExamples[i].IsUsed = true
		exams = append(exams, v)
	}
	return exams
}

// getVCSInfo returns VCS name, project name, project home page, Upper level project URL and package tag.
func getVCSInfo(q, tag string, pdoc *doc.Package) (vcs, proName, proPath, pkgDocPath, pkgTag string) {
	// pkgTag is only for Google Code which needs tag information as GET argument.
	// Get project name.
	lastIndex := strings.LastIndex(q, "/")
	proName = q[lastIndex+1:]
	if i := strings.Index(proName, "?"); i > -1 {
		proName = proName[:i]
	}

	// Project VCS home page.
	switch {
	case strings.HasPrefix(q, "github.com"): // github.com
		vcs = "Github"
		if len(tag) == 0 {
			tag = "master" // Set tag.
		}
		proName := utils.GetProjectPath(pdoc.ImportPath)
		proPath = strings.Replace(q, proName, proName+"/tree/"+tag, 1)
	case strings.HasPrefix(q, "code.google.com"): // code.google.com
		vcs = "Google Code"
		if strings.Index(q, "source/") == -1 {
			proPath = strings.Replace(q, "/"+pdoc.ProjectName, "/"+pdoc.ProjectName+"/source/browse", 1)
		} else {
			proPath = q
			q = strings.Replace(q, "source/browse/", "", 1)
			lastIndex = strings.LastIndex(q, "/")
		}
		pkgTag = "?r=" + tag // Set tag.
	case q[0] == 'b': // bitbucket.org
		vcs = "BitBucket"
		if len(tag) == 0 {
			tag = "default" // Set tag.
		}
		if proName != pdoc.ProjectName {
			// Not root.
			proPath = strings.Replace(q, "/"+pdoc.ProjectName, "/"+pdoc.ProjectName+"/src/"+tag, 1)
		} else {
			proPath = q + "/src/" + tag
		}
	case q[0] == 'l': // launchpad.net
		vcs = "Launchpad"
		proPath = "bazaar." + strings.Replace(q, "/"+pdoc.ProjectName, "/+branch/"+pdoc.ProjectName+"/view/head:/", 1)
	case strings.HasPrefix(q, "git.oschina.net"): // git.oschina.net
		vcs = "Git @ OSC"
		if len(tag) == 0 {
			tag = "master" // Set tag.
		}
		if proName != pdoc.ProjectName {
			// Not root.
			proName := utils.GetProjectPath(pdoc.ImportPath)
			proPath = strings.Replace(q, proName, proName+"/tree/"+tag, 1)
		} else {
			proPath = q + "/tree/" + tag
		}
	case strings.HasPrefix(q, "code.csdn.net"): // code.csdn.net
		vcs = "CSDN Code"
		if len(tag) == 0 {
			tag = "master" // Set tag.
		}
		if proName != pdoc.ProjectName {
			// Not root.
			proName := utils.GetProjectPath(pdoc.ImportPath)
			proPath = strings.Replace(q, proName, proName+"/tree/"+tag, 1)
		} else {
			proPath = q + "/tree/" + tag
		}
	}

	pkgDocPath = q[:lastIndex]
	return vcs, proName, proPath, pkgDocPath, pkgTag
}

// getLabels retuens corresponding label name.
func getLabels(rawLabel string) []string {
	rawLabels := strings.Split(rawLabel, "|")
	rawLabels = rawLabels[:len(rawLabels)-1] // The last element is always empty.
	// Remove first character '$' in every label.
	for i := range rawLabels {
		rawLabels[i] = rawLabels[i][1:]
		// Reassign label name.
		for j, v := range labelList {
			if rawLabels[i] == v {
				rawLabels[i] = labels[j]
				break
			}
		}
	}
	return rawLabels
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
				val.Code = *codeDecode(&s)
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
			case 1: // Constants.
				val.Consts = make([]*doc.Value, 0, 1)
				for _, v2 := range strings.Split(s, "&M#") {
					val2 := new(doc.Value)
					for z, s2 := range strings.Split(v2, "&V#") {
						switch z {
						case 0: // Name
							val2.Name = s2
						case 1: // Doc
							val2.Doc = s2
						case 2: // Decl
							val2.Decl = s2
						case 3: // URL
							val2.URL = s2
						}
					}
					val.Consts = append(val.Consts, val2)
				}
				val.Consts = val.Consts[:len(val.Consts)-1]
			case 2: // Variables.
				val.Vars = make([]*doc.Value, 0, 1)
				for _, v2 := range strings.Split(s, "&M#") {
					val2 := new(doc.Value)
					for z, s2 := range strings.Split(v2, "&V#") {
						switch z {
						case 0: // Name
							val2.Name = s2
						case 1: // Doc
							val2.Doc = s2
						case 2: // Decl
							val2.Decl = s2
						case 3: // URL
							val2.URL = s2
						}
					}
					val.Vars = append(val.Vars, val2)
				}
				val.Vars = val.Vars[:len(val.Vars)-1]
			case 3: // Functions.
				val.Funcs = make([]*doc.Func, 0, 3)
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
							val2.Code = *codeDecode(&s2)
						}
					}
					val.Funcs = append(val.Funcs, val2)
				}
				val.Funcs = val.Funcs[:len(val.Funcs)-1]
			case 5: // Methods.
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
							val2.Code = *codeDecode(&s2)
						}
					}
					val2.FullName = val.Name + "_" + val2.Name
					val.Methods = append(val.Methods, val2)
				}
				val.Methods = val.Methods[:len(val.Methods)-1]
			}
		}
		pdoc.Types = append(pdoc.Types, val)
	}
	pdoc.Types = pdoc.Types[:len(pdoc.Types)-1]

	// Examples.
	pdoc.Examples = convertDataFormatExample(pdecl.Examples, "")

	// Dirs.
	pdoc.Dirs = strings.Split(pdecl.Dirs, "|")

	// Imports.
	pdoc.Imports = strings.Split(pdecl.Imports, "|")

	// Files.
	pdoc.Files = strings.Split(pdecl.Files, "|")
	return nil
}

func convertDataFormatExample(examStr, suffix string) []*doc.Example {
	exams := make([]*doc.Example, 0, 5)
	for _, v := range strings.Split(examStr, "&$#") {
		val := new(doc.Example)
		for j, s := range strings.Split(v, "&E#") {
			switch j {
			case 0: // Name
				val.Name = s + suffix
				if len(val.Name) == 0 {
					val.Name = "Package"
				}
			case 1: // Doc
				val.Doc = s
			case 2: // Code
				val.Code = *codeDecode(&s)
			case 3: // Output
				val.Output = s
				if len(s) > 0 {
					val.IsHasOutput = true
				}
			}
		}
		exams = append(exams, val)
	}
	exams = exams[:len(exams)-1]
	return exams
}

func codeDecode(code *string) *string {
	str := new(string)
	byts, _ := base32.StdEncoding.DecodeString(*code)
	*str = string(byts)
	return str
}

func updateCacheInfo(pdoc *doc.Package, urpids, urpts *http.Cookie) (string, string) {
	pdoc.ViewedTime = time.Now().UTC().Unix()

	updateCachePros(pdoc)
	updateProInfos(pdoc)
	return updateUrPros(pdoc, urpids, urpts)
}

func updateCachePros(pdoc *doc.Package) {
	for _, p := range cachePros {
		if p.Id == pdoc.Id {
			p.ProName = pdoc.ProjectName
			p.Synopsis = pdoc.Synopsis
			p.IsCmd = pdoc.IsCmd
			p.Tags = strings.Join(pdoc.Tags, "|||")
			p.Views++
			p.ViewedTime = pdoc.ViewedTime
			p.Created = pdoc.Created
			p.Rank = int64(pdoc.ImportedNum*30) + p.Views
			p.Etag = pdoc.Etag
			p.Labels = pdoc.Labels
			p.ImportedNum = pdoc.ImportedNum
			p.ImportPid = pdoc.ImportPid
			p.Note = pdoc.Note
			return
		}
	}

	pdoc.Views++
	cachePros = append(cachePros, &models.PkgInfo{
		Id:          pdoc.Id,
		Path:        pdoc.ImportPath,
		ProName:     pdoc.ProjectName,
		Synopsis:    pdoc.Synopsis,
		IsCmd:       pdoc.IsCmd,
		Tags:        strings.Join(pdoc.Tags, "|||"),
		Views:       pdoc.Views,
		ViewedTime:  pdoc.ViewedTime,
		Created:     pdoc.Created,
		Rank:        int64(pdoc.ImportedNum*30) + pdoc.Views,
		Etag:        pdoc.Etag,
		Labels:      pdoc.Labels,
		ImportedNum: pdoc.ImportedNum,
		ImportPid:   pdoc.ImportPid,
		Note:        pdoc.Note,
	})
}

func updateProInfos(pdoc *doc.Package) {
	index := -1
	listLen := len(recentViewedPros)
	curPro := &proInfo{
		Path:     pdoc.ImportPath,
		Synopsis: pdoc.Synopsis,
		IsGoRepo: pdoc.ProjectName == "Go" &&
			strings.Index(pdoc.ImportPath, ".") == -1,
		Views:      pdoc.Views,
		ViewedTime: pdoc.ViewedTime,
		Rank:       pdoc.Rank,
	}

	// Check if in the list
	for i, s := range recentViewedPros {
		if s.Path == curPro.Path {
			index = i
			break
		}
	}

	s := make([]*proInfo, 0, maxProInfoNum)
	s = append(s, curPro)
	switch {
	case index == -1 && listLen < maxProInfoNum:
		// Not found and list is not full
		s = append(s, recentViewedPros...)
	case index == -1 && listLen >= maxProInfoNum:
		// Not found but list is full
		s = append(s, recentViewedPros[:maxProInfoNum-1]...)
	case index > -1:
		// Found
		s = append(s, recentViewedPros[:index]...)
		s = append(s, recentViewedPros[index+1:]...)
	}
	recentViewedPros = s
}

// updateUrPros returns strings of user recent viewd projects and timestamps.
func updateUrPros(pdoc *doc.Package, urpids, urpts *http.Cookie) (string, string) {
	if pdoc.Id == 0 {
		return urpids.Value, urpts.Value
	}

	var urPros, urTs []string
	if urpids != nil && urpts != nil {
		urPros = strings.Split(urpids.Value, "|")
		urTs = strings.Split(urpts.Value, "|")
		if len(urTs) != len(urPros) {
			urTs = strings.Split(
				strings.Repeat(strconv.Itoa(int(time.Now().UTC().Unix()))+"|", len(urPros)), "|")
			urTs = urTs[:len(urTs)-1]
		}
	}

	index := -1
	listLen := len(urPros)

	// Check if in the list
	for i, s := range urPros {
		pid, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return urpids.Value, urpts.Value
		}
		if pid == pdoc.Id {
			index = i
			break
		}
	}

	s := make([]string, 0, maxProInfoNum)
	ts := make([]string, 0, maxProInfoNum)
	s = append(s, strconv.Itoa(int(pdoc.Id)))
	ts = append(ts, strconv.Itoa(int(time.Now().UTC().Unix())))

	switch {
	case index == -1 && listLen < maxProInfoNum:
		// Not found and list is not full
		s = append(s, urPros...)
		ts = append(ts, urTs...)
	case index == -1 && listLen >= maxProInfoNum:
		// Not found but list is full
		s = append(s, urPros[:maxProInfoNum-1]...)
		ts = append(ts, urTs[:maxProInfoNum-1]...)
	case index > -1:
		// Found
		s = append(s, urPros[:index]...)
		s = append(s, urPros[index+1:]...)
		ts = append(ts, urTs[:index]...)
		ts = append(ts, urTs[index+1:]...)
	}
	return strings.Join(s, "|"), strings.Join(ts, "|")
}
