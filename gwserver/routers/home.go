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
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Unknwon/com"
	"github.com/Unknwon/ctw/packer"
	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
)

var (
	maxProInfoNum = 20
	maxExamNum    = 15

	recentUpdatedExs                                       []*models.PkgExam
	recentViewedPros, topRankPros, topViewedPros, RockPros []*hv.PkgInfo
)

// initPopPros initializes popular projects.
func initPopPros() {
	var err error
	err, recentUpdatedExs, recentViewedPros, topRankPros, topViewedPros, RockPros =
		models.GetPopulars(maxProInfoNum, maxExamNum)
	if err != nil {
		panic("initPopPros -> " + err.Error())
	}
}

// HomeRouter serves home page.
type HomeRouter struct {
	baseRouter
}

func serveHome(this *HomeRouter, urpids, urpts *http.Cookie) {
	this.Data["IsHome"] = true
	this.TplNames = "home.html"

	// Global Recent projects.
	this.Data["GlobalHistory"] = recentViewedPros
	// User Recent projects.
	if urpids != nil && urpts != nil {
		upros := models.GetGroupPkgInfoById(strings.Split(urpids.Value, "|"))
		pts := strings.Split(urpts.Value, "|")
		for i, p := range upros {
			ts, _ := strconv.ParseInt(pts[i], 10, 64)
			p.ViewedTime = ts
		}
		this.Data["UserHistory"] = upros
	}

	// Popular projects and examples.
	this.Data["WeeklyStarPros"] = RockPros
	this.Data["TopRankPros"] = topRankPros
	this.Data["TopViewsPros"] = topViewedPros
	this.Data["RecentExams"] = recentUpdatedExs
}

func updateCacheInfo(pdoc *hv.Package, urpids, urpts *http.Cookie) (string, string) {
	pdoc.ViewedTime = time.Now().UTC().Unix()

	updateCachePros(pdoc)
	updateProInfos(pdoc)
	return updateUrPros(pdoc, urpids, urpts)
}

func updateCachePros(pdoc *hv.Package) {
	pdoc.Views++

	for _, p := range cachePros {
		if p.Id == pdoc.Id {
			p = pdoc.PkgInfo
			p.Rank = int64(pdoc.RefProNum*30) + pdoc.Views
			return
		}
	}

	pinfo := pdoc.PkgInfo
	pinfo.Rank = int64(pdoc.RefProNum*30) + pdoc.Views
	cachePros = append(cachePros, pinfo)
}

func updateProInfos(pdoc *hv.Package) {
	index := -1
	listLen := len(recentViewedPros)
	curPro := pdoc.PkgInfo

	// Check if in the list
	for i, s := range recentViewedPros {
		if s.ImportPath == curPro.ImportPath {
			index = i
			break
		}
	}

	s := make([]*hv.PkgInfo, 0, maxProInfoNum)
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
func updateUrPros(pdoc *hv.Package, urpids, urpts *http.Cookie) (string, string) {
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

// Get implemented Get method for HomeRouter.
func (this *HomeRouter) Get() {
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

	// User History.
	urpids, _ := this.Ctx.Request.Cookie("UserHistory")
	urpts, _ := this.Ctx.Request.Cookie("UHTimestamps")

	if len(reqUrl) == 0 && len(q) == 0 {
		serveHome(this, urpids, urpts)
		return
	}

	// Documentation page.
	broPath := reqUrl // Browse path.

	// Check if it's the standard library.
	if utils.IsGoRepoPath(broPath) {
		broPath = "code.google.com/p/go/source/browse/src/pkg/" + broPath
	} else if utils.IsGoSubrepoPath(broPath) {
		broPath = "code.google.com/p/" + broPath
		reqUrl = broPath
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

	// Check documentation of current import path, update automatically as needed.
	pdoc, err := doc.CheckDoc(reqUrl, tag, doc.RT_Human)
	if err == nil {
		// errNoMatch leads to pdoc == nil.
		if pdoc != nil {
			// Generate documentation page.
			if generatePage(this, pdoc, broPath, tag) {
				ps, ts := updateCacheInfo(pdoc, urpids, urpts)
				this.Ctx.SetCookie("UserRecentPros", ps, 9999999999, "/")
				this.Ctx.SetCookie("URPTimestamps", ts, 9999999999, "/")
				return
			}
		}
	}

	// TODO
	beego.Error(err)
	//this.Redirect("/search?q="+reqUrl, 302)
	return
}

func codeDecode(code *string) *string {
	str := new(string)
	byts, _ := base32.StdEncoding.DecodeString(*code)
	*str = string(byts)
	return str
}

func convertDataFormatExample(examStr, suffix string) []*hv.Example {
	exams := make([]*hv.Example, 0, 5)
	for _, v := range strings.Split(examStr, "&$#") {
		val := new(hv.Example)
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
			}
		}
		exams = append(exams, val)
	}
	exams = exams[:len(exams)-1]
	return exams
}

// getUserExamples returns user examples of given import path.
func getUserExamples(path string) []*hv.Example {
	gists, _ := models.GetPkgExams(path)
	// Doesn't have Gists.
	if len(gists) == 0 {
		return nil
	}

	pexams := make([]*hv.Example, 0, 5)
	for _, g := range gists {
		exams := convertDataFormatExample(g.Examples, "_"+g.Gist[strings.LastIndex(g.Gist, "/")+1:])
		pexams = append(pexams, exams...)
	}
	return pexams
}

// getExamples returns index of function example if it exists.
func getExamples(pdoc *hv.Package, typeName, name string) (exams []*hv.Example) {
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

// getVCSInfo returns VCS name, project name, project home page, Upper level project URL and package tag.
func getVCSInfo(q, tag string, pdoc *hv.Package) (vcs, proName, proPath, pkgDocPath, pkgTag string) {
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
	case strings.HasPrefix(q, "gitcafe.com"): // code.csdn.net
		vcs = "GitCafe"
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

// saveDocPage saves doc. content to JS file(s),
// it returns max index of JS file(s);
// it returns -1 when error occurs.
func saveDocPage(docPath string, data []byte) int {
	os.MkdirAll(path.Dir("./static/docs/"+docPath), os.ModePerm)

	buf := new(bytes.Buffer)

	count := 0
	d := string(data)
	l := len(d)
	if l < 80000 {
		buf.WriteString("document.write(\"")
		buf.Write(data)
		buf.WriteString("\")")

		if _, err := com.SaveFile("./static/docs/"+docPath+".js", buf.Bytes()); err != nil {
			beego.Error("saveDocPage(", docPath, ") ->", err)
			return -1
		}
	} else {
		// Too large, need to sperate.
		start := 0
		end := start + 40000
		for {
			if end >= l {
				end = l
			} else {
				// Need to break in space.
				for {
					if d[end-3:end] == "/b>" {
						break
					}
					end += 1

					if end >= l {
						break
					}
				}
			}

			buf.WriteString("document.write(\"")
			buf.WriteString(d[start:end])
			buf.WriteString("\")\n")

			p := docPath
			if count != 0 {
				p += fmt.Sprintf("-%d", count)
			}

			if _, err := com.SaveFile("./static/docs/"+p+".js", buf.Bytes()); err != nil {
				beego.Error("saveDocPage(", p, ") ->", err)
				return -1
			}

			if end >= l {
				break
			}

			buf.Reset()
			start = end
			end += 204800
			count++
		}
	}

	return count
}

func renderDoc(this *HomeRouter, pdoc *hv.Package, q, tag, docPath string) bool {
	this.Data["PkgFullIntro"] = pdoc.Doc

	var buf bytes.Buffer
	links := make([]*utils.Link, 0, len(pdoc.Types)+len(pdoc.Imports)+len(pdoc.TestImports)+
		len(pdoc.Funcs)+10)
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
	for _, v := range append(pdoc.Imports, pdoc.TestImports...) {
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
		pdoc.IsHasExport = true
		this.Data["IsHasExports"] = true
		exportDataSrc = exportDataSrc[:len(exportDataSrc)-1]
		this.Data["ExportDataSrc"] = exportDataSrc
	}

	pdoc.UserExamples = getUserExamples(pdoc.ImportPath)

	pdoc.IsHasConst = len(pdoc.Consts) > 0
	pdoc.IsHasVar = len(pdoc.Vars) > 0
	if len(pdoc.Examples)+len(pdoc.UserExamples) > 0 {
		pdoc.IsHasExample = true
		this.Data["IsHasExample"] = pdoc.IsHasExample
		this.Data["Examples"] = append(pdoc.Examples, pdoc.UserExamples...)
	}

	// Commented and total objects number.
	var comNum, totalNum int

	// Constants.
	this.Data["IsHasConst"] = pdoc.IsHasConst
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
	this.Data["IsHasVar"] = pdoc.IsHasVar
	this.Data["Vars"] = pdoc.Vars
	for i, v := range pdoc.Vars {
		buf.Reset()
		utils.FormatCode(&buf, &v.Decl, links)
		v.FmtDecl = buf.String()
		pdoc.Vars[i] = v
	}

	// Dirs.
	pinfos := models.GetSubPkgs(pdoc.ImportPath, tag, pdoc.Dirs)
	if len(pinfos) > 0 {
		pdoc.IsHasSubdir = true
		this.Data["IsHasSubdirs"] = pdoc.IsHasSubdir
		this.Data["Subdirs"] = pinfos
	}

	// Files.
	if len(pdoc.Files) > 0 {
		pdoc.IsHasFile = true
		this.Data["IsHasFiles"] = pdoc.IsHasFile
		this.Data["Files"] = pdoc.Files
	}

	var err error
	pfuncs := doc.RenderFuncs(pdoc)

	this.Data["ImportPkgs"] = strings.Join(pdoc.Imports, "|")

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
		f.FmtDecl = buf.String() + " {"
		if exs := getExamples(pdoc, "", f.Name); len(exs) > 0 {
			f.Examples = exs
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
			f.FmtDecl = buf.String() + " {"
			if exs := getExamples(pdoc, "", f.Name); len(exs) > 0 {
				f.Examples = exs
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
			m.FmtDecl = buf.String() + " {"
			if exs := getExamples(pdoc, t.Name, m.Name); len(exs) > 0 {
				m.Examples = exs
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
			t.Examples = exs
		}
		totalNum++
		pdoc.Types[i] = t
	}

	if !pdoc.IsCmd {
		// Calculate documentation complete %.
		this.Data["DocCPLabel"], this.Data["DocCP"] = calDocCP(comNum, totalNum)
	} else {
		this.Data["IsCmd"] = true
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

	// Get VCS name, project name, project home page, Upper level project URL, and project tag.
	this.Data["VCS"], this.Data["ProName"], this.Data["ProPath"], this.Data["ProDocPath"], this.Data["PkgTag"] =
		getVCSInfo(q, tag, pdoc)

	this.TplNames = "tpl/docs.tpl"
	data, err := this.RenderBytes()
	if err != nil {
		beego.Error("generatePage(", pdoc.ImportPath, ") -> RenderBytes:", err)
		return false
	}

	n := saveDocPage(docPath, com.Html2JS(data))
	if n == -1 {
		return false
	}
	pdoc.JsNum = n
	pdoc.Id, err = doc.SaveProject(pdoc, pfuncs)
	if err != nil {
		beego.Error("generatePage(", pdoc.ImportPath, ") -> SaveProject:", err)
		return false
	}
	return true
}

// ConvertDataFormat converts data from database acceptable format to useable format.
func ConvertDataFormat(pdoc *hv.Package, pdecl *models.PkgDecl) error {
	if pdoc.PkgDecl == nil {
		pdoc.PkgDecl = &hv.PkgDecl{}
	}

	pdoc.JsNum = pdecl.JsNum
	pdoc.IsHasExport = pdecl.IsHasExport
	pdoc.IsHasConst = pdecl.IsHasConst
	pdoc.IsHasVar = pdecl.IsHasVar
	pdoc.IsHasExample = pdecl.IsHasExample
	pdoc.IsHasFile = pdecl.IsHasFile
	pdoc.IsHasSubdir = pdecl.IsHasSubdir

	// Imports.
	pdoc.Imports = strings.Split(pdecl.Imports, "|")
	if len(pdoc.Imports) == 1 && len(pdoc.Imports[0]) == 0 {
		// No import.
		pdoc.Imports = nil
	}
	return nil
}

// getLabels retuens corresponding label name.
// func getLabels(rawLabel string) []string {
// 	rawLabels := strings.Split(rawLabel, "|")
// 	rawLabels = rawLabels[:len(rawLabels)-1] // The last element is always empty.
// 	// Remove first character '$' in every label.
// 	for i := range rawLabels {
// 		rawLabels[i] = rawLabels[i][1:]
// 		// Reassign label name.
// 		for j, v := range labelList {
// 			if rawLabels[i] == v {
// 				rawLabels[i] = labels[j]
// 				break
// 			}
// 		}
// 	}
// 	return rawLabels
// }

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

// generatePage genarates documentation page for project.
// it returns false when it's a invaild(empty) project.
func generatePage(this *HomeRouter, pdoc *hv.Package, q, tag string) bool {
	docPath := pdoc.ImportPath + packer.TagSuffix("-", tag)

	if pdoc.IsNeedRender {
		if !renderDoc(this, pdoc, q, tag, docPath) {
			return false
		}
	} else {
		pdecl, err := models.LoadProject(pdoc.Id, tag)
		if err != nil {
			beego.Error("HomeController.generatePage ->", err)
			return false
		}
		this.Data["ImportPkgs"] = pdecl.Imports

		err = ConvertDataFormat(pdoc, pdecl)
		if err != nil {
			beego.Error("HomeController.generatePage -> ConvertDataFormat:", err)
			return false
		}
	}

	// Set properties.
	this.TplNames = "docs.html"

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
	lang := this.Data["Lang"].(string)
	byts, _ := base32.StdEncoding.DecodeString(
		models.LoadPkgDoc(pdoc.ImportPath, lang, "rm"))
	if len(byts) > 0 {
		this.Data["IsHasReadme"] = true
		this.Data["PkgDoc"] = string(byts)
	}

	// Index.
	this.Data["IsHasExports"] = pdoc.IsHasExport
	this.Data["IsHasConst"] = pdoc.IsHasConst
	this.Data["IsHasVar"] = pdoc.IsHasVar

	if !pdoc.IsCmd {
		this.Data["IsHasExams"] = pdoc.IsHasExample

		// Tags.
		if len(tag) == 0 {
			tag = "master"
		}
		this.Data["CurTag"] = tag
		this.Data["Tags"] = strings.Split(pdoc.Tags, "|||")
	} else {
		this.Data["IsCmd"] = true
	}

	this.Data["Rank"] = pdoc.Rank
	this.Data["Views"] = pdoc.Views + 1
	//this.Data["Labels"] = getLabels(pdoc.Labels)
	//this.Data["LabelDataSrc"] = labelSet
	this.Data["ImportPkgNum"] = len(pdoc.Imports)
	this.Data["IsHasSubdirs"] = pdoc.IsHasSubdir
	this.Data["IsHasFiles"] = pdoc.IsHasFile
	this.Data["IsHasImports"] = len(pdoc.Imports) > 0
	this.Data["IsImported"] = pdoc.RefNum > 0
	this.Data["ImportPid"] = pdoc.RefPids
	this.Data["ImportedNum"] = pdoc.RefNum
	this.Data["UtcTime"] = pdoc.Created
	this.Data["TimeSince"] = calTimeSince(pdoc.Created)

	docJS := make([]string, 0, pdoc.JsNum+1)
	docJS = append(docJS, "/static/docs/"+docPath+".js")

	for i := 1; i <= pdoc.JsNum; i++ {
		docJS = append(docJS, fmt.Sprintf(
			"/static/docs/%s-%d.js", docPath, i))
	}
	this.Data["DocJS"] = docJS
	return true
}
