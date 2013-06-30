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
	"path"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

// Recent viewed project.
type recentPro struct {
	Path, Synopsis string
	IsGoRepo       bool
	ViewedTime     int64
}

var (
	recentViewedProNum = 20         // Maximum element number of recent viewed project list.
	recentViewedPros   []*recentPro // Recent viewed project list.

	labelList []string // Projects label list.
	labelSet  string   // Label data source.
)

func init() {
	// Initialized recent viewed project list.
	num, err := beego.AppConfig.Int("recentViewedProNum")
	if err == nil {
		recentViewedProNum = num
		beego.Trace("Loaded 'recentViewedProNum' -> value:", recentViewedProNum)
	} else {
		beego.Trace("Failed to load 'recentViewedProNum' -> Use default value:", recentViewedProNum)
	}

	recentViewedPros = make([]*recentPro, 0, recentViewedProNum)
	// Get recent viewed projects from database.
	proinfos, _ := models.GetRecentPros(recentViewedProNum)
	for _, p := range proinfos {
		// Only projects with import path length is less than 40 letters will be showed.
		if len(p.Path) < 40 {
			recentViewedPros = append(recentViewedPros,
				&recentPro{
					Path:       p.Path,
					Synopsis:   p.Synopsis,
					ViewedTime: p.ViewedTime,
					IsGoRepo: p.ProName == "Go" &&
						strings.Index(p.Path, ".") == -1,
				})
		}
	}

	// Initialize project tags.
	labelList = strings.Split(beego.AppConfig.String("labels"), "|")
	for _, s := range labelList {
		labelSet += "&quot;" + s + "&quot;,"
	}
	labelSet = labelSet[:len(labelSet)-1]
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
		return
	}

	// Set language version.
	curLang := setLangVer(this.Ctx, this.Input(), this.Data)

	// Get query field.
	q := strings.TrimSpace(this.Input().Get("q"))

	// Remove last "/".
	q = strings.TrimRight(q, "/")

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

	// Check show home page or documentation page.
	if len(reqUrl) == 0 && len(q) == 0 {
		// Home page.
		this.Data["IsHome"] = true
		this.TplNames = "home_" + curLang.Lang + ".html"

		// Recent projects
		this.Data["RecentPros"] = recentViewedPros
		// Get popular project and examples list from database.
		this.Data["PopPros"], this.Data["RecentExams"] = models.GetPopulars(20, 12)
		// Set standard library keyword type-ahead.
		this.Data["DataSrc"] = utils.GoRepoSet
	} else {
		// Documentation page.
		this.TplNames = "docs_" + curLang.Lang + ".html"
		broPath := reqUrl // Browse path.

		// Check if it is standard library.
		if utils.IsGoRepoPath(broPath) {
			broPath = "code.google.com/p/go/source/browse/src/pkg/" + broPath
		}

		// Check if it is a remote path that can be used for 'gopm get', if not means it's a keyword.
		if !utils.IsValidRemotePath(broPath) {
			// Show search page
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
			pdoc.UserExamples = getUserExamples(pdoc.ImportPath)
			// Generate documentation page.
			if generatePage(this, pdoc, broPath, tag, curLang.Lang) {
				// Update recent project list.
				updateRecentPros(pdoc)
				// Update project views.
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
				models.AddViews(pinfo)
				return
			}
		} else {
			beego.Error("HomeRouter.Get ->", err)
		}

		// Show search page
		this.Redirect("/search?q="+reqUrl, 302)
		return
	}
}

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
// it returns false when its a invaild(empty) project.
func generatePage(this *HomeRouter, pdoc *doc.Package, q, tag, lang string) bool {
	// Load project data from database.
	pdecl, err := models.LoadProject(pdoc.ImportPath, tag)
	if err != nil {
		beego.Error("HomeController.generatePage ->", err)
		return false
	}

	// Set properties.
	this.TplNames = "docs_" + lang + ".html"

	// Refresh (within 10 seconds).
	this.Data["IsRefresh"] = pdoc.Created.Add(10 * time.Second).UTC().After(time.Now().UTC())

	// Get VCS name, project name, project home page, and Upper level project URL.
	this.Data["VCS"], this.Data["ProName"], this.Data["ProPath"], this.Data["ProDocPath"] =
		getVCSInfo(q, tag, pdoc)

	if utils.IsGoRepoPath(pdoc.ImportPath) &&
		strings.Index(pdoc.ImportPath, ".") == -1 {
		this.Data["IsGoRepo"] = true
	}

	this.Data["Views"] = pdoc.Views + 1

	// Labels.
	this.Data["Labels"] = getLabels(pdoc.Labels)

	// Introduction.
	this.Data["ImportPath"] = pdoc.ImportPath
	byts, _ := base32.StdEncoding.DecodeString(
		models.LoadPkgDoc(pdoc.ImportPath, lang, "rm"))
	this.Data["PkgDoc"] = string(byts)
	byts, _ = base32.StdEncoding.DecodeString(pdecl.Doc)
	this.Data["PkgFullIntro"] = string(byts)

	var buf bytes.Buffer
	// Convert data format.
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

	for _, v := range pdoc.Imports {
		links = append(links, &utils.Link{
			Name: path.Base(v) + ".",
			Path: v,
		})
	}

	exportDataSrc := buf.String()
	if len(exportDataSrc) > 0 {
		this.Data["HasExports"] = true
		exportDataSrc = exportDataSrc[:len(exportDataSrc)-1]
		// Set export keyword type-ahead.
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
		buf.Reset()
		utils.FormatCode(&buf, &f.Code, links)
		f.Code = buf.String()
		if exs := getExamples(pdoc, "", f.Name); len(exs) > 0 {
			f.IsHasExam = true
			f.Exams = exs
		}
		totalNum++
		pdoc.Funcs[i] = f
	}

	this.Data["Types"] = pdoc.Types
	for i, t := range pdoc.Types {
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
			buf.Reset()
			utils.FormatCode(&buf, &f.Code, links)
			f.Code = buf.String()
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
			buf.Reset()
			utils.FormatCode(&buf, &m.Code, links)
			m.Code = buf.String()
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
	this.Data["IsHasSubdirs"] = len(pdoc.Dirs) > 0
	pinfos := make([]*models.PkgInfo, 0, len(pdoc.Dirs))
	for _, v := range pdoc.Dirs {
		v = pdoc.ImportPath + "/" + v
		if pinfo, err := models.GetPkgInfo(v, tag); err == nil {
			pinfos = append(pinfos, pinfo)
		} else {
			pinfos = append(pinfos, &models.PkgInfo{Path: v})
		}
	}
	this.Data["Subdirs"] = pinfos

	// Labels.
	this.Data["LabelDataSrc"] = labelSet

	this.Data["Files"] = pdoc.Files
	this.Data["ImportPkgs"] = pdecl.Imports
	this.Data["ImportPkgNum"] = len(pdoc.Imports) - 1
	this.Data["IsImported"] = pdoc.ImportedNum > 0
	this.Data["ImportPid"] = pdoc.ImportPid
	this.Data["ImportedNum"] = pdoc.ImportedNum
	this.Data["UtcTime"] = pdoc.Created
	return true
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

// getVCSInfo returns VCS name, project name, project home page, and Upper level project URL.
func getVCSInfo(q, tag string, pdoc *doc.Package) (vcs, proName, proPath, pkgDocPath string) {
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
		if proName != pdoc.ProjectName {
			// Not root.
			proName := utils.GetProjectPath(pdoc.ImportPath)
			proPath = strings.Replace(q, proName, proName+"/tree/"+tag, 1)
		} else {
			proPath = q + "/tree/" + tag
		}
	case strings.HasPrefix(q, "code.google.com"): // code.google.com
		vcs = "Google Code"
		if strings.Index(q, "source/") == -1 {
			proPath = strings.Replace(q, "/"+pdoc.ProjectName, "/"+pdoc.ProjectName+"/source/browse", 1)
		} else {
			proPath = q
			q = strings.Replace(q, "source/browse/", "", 1)
			lastIndex = strings.LastIndex(q, "/")
		}
		proPath += "?r=" + tag // Set tag.
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
	return vcs, proName, proPath, pkgDocPath
}

func getLabels(rawLabel string) []string {
	// Get labels.
	labels := strings.Split(beego.AppConfig.String("label_names"), "|")

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
							val2.Code = *codeDecode(&s2)
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
							val2.Code = *codeDecode(&s2)
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

	// Examples.
	pdoc.Examples = convertDataFormatExample(pdecl.Examples, "")

	// Dirs.
	pdoc.Dirs = strings.Split(pdecl.Dirs, "|")
	pdoc.Dirs = pdoc.Dirs[:len(pdoc.Dirs)-1]

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

func updateRecentPros(pdoc *doc.Package) {
	pdoc.ViewedTime = time.Now().UTC().Unix()

	// Only projects with import path length is less than 40 letters will be showed.
	if len(pdoc.ImportPath) < 40 {
		index := -1
		listLen := len(recentViewedPros)
		curPro := &recentPro{
			Path:       pdoc.ImportPath,
			Synopsis:   pdoc.Synopsis,
			ViewedTime: pdoc.ViewedTime,
			IsGoRepo: pdoc.ProjectName == "Go" &&
				strings.Index(pdoc.ImportPath, ".") == -1,
		}

		// Check if in the list
		for i, s := range recentViewedPros {
			if s.Path == curPro.Path {
				index = i
				break
			}
		}

		s := make([]*recentPro, 0, recentViewedProNum)
		s = append(s, curPro)
		switch {
		case index == -1 && listLen < recentViewedProNum:
			// Not found and list is not full
			s = append(s, recentViewedPros...)
		case index == -1 && listLen >= recentViewedProNum:
			// Not found but list is full
			s = append(s, recentViewedPros[:recentViewedProNum-1]...)
		case index > -1:
			// Found
			s = append(s, recentViewedPros[:index]...)
			s = append(s, recentViewedPros[index+1:]...)
		}
		recentViewedPros = s
	}
}
