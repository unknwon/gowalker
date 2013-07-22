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

// Package models implemented database access funtions.
package models

import (
	"encoding/base32"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/coocood/qbs"
	"github.com/russross/blackfriday"
)

// PkgInfo is package information.
type PkgInfo struct {
	// Primary key.
	Id int64

	/*
		- Import path of package.
		eg.
			github.com/Unknwon/gowalker
		- Name of the project.
		eg.
			gowalker
		- Project synopsis.
		eg.
			Go Walker is a web server that generates Go projects API documentation with source code on the fly.
		- Indicates whether it's a command line tool or package.
		eg.
			True
	*/
	Path     string `qbs:"size:100,index"`
	ProName  string
	Synopsis string
	IsCmd    bool

	/*
		- All tags of project.
		eg.
			master|||v0.6.2.0718
		- Views of projects.
		eg.
			1342
		- User viewed time(Unix-timestamp).
		eg.
			1374127619
		- Time when information last updated(UTC).
		eg.
			2013-07-16 21:09:27.48932087
	*/
	Tags       string
	Views      int64 `qbs:"index"`
	ViewedTime int64
	Created    time.Time `qbs:"index"`

	/*
		- Rank is the benchmark of projects, it's based on BaseRank and views.
		eg.
			826
	*/
	Rank int64 `qbs:"index"`

	/*
		- Revision tag with package (structure) version.
		eg.
			9-ed5cfa7eb95cfa6787209e7d6bd5d3af0ed3679f
		- Project labels.
		eg.
			$tool|
	*/
	Etag, Labels string // Revision tag and project labels.

	/*
		- Number of packages that imports this project.
		eg.
			11
		- Ids of packages that imports this project.
		eg.
			$47|$89|$5464|$8586|$8595|$8787|$8789|$8790|$8918|$9134|$9139|
	*/
	ImportedNum int
	ImportPid   string

	/*
		- Addtional information.
		eg.
			Github: <forks>|<watchers>|
	*/
	Note string
}

// PkgDecl is package declaration in database acceptable form.
type PkgDecl struct {
	Id   int64
	Path string `qbs:"size:100,index"`
	Tag  string // Project tag.
	Doc  string // Package documentation.

	// Top-level declarations.
	Consts, Funcs, Types, Vars string

	// Internal declarations.
	Iconsts, Ifuncs, Itypes, Ivars string

	Examples         string // Examples.
	Notes            string // Source code notes.
	Files, TestFiles string // Source files.
	Dirs             string // Subdirectories.

	Imports, TestImports string // Imports.
}

func (*PkgDecl) Indexes(indexes *qbs.Indexes) {
	indexes.AddUnique("path", "tag")
}

// PkgDoc is package documentation for multi-language usage.
type PkgDoc struct {
	Id   int64
	Path string `qbs:"size:100,index"`
	Lang string // Documentation language.
	Type string
	Doc  string // Documentataion.
}

func (*PkgDoc) Indexes(indexes *qbs.Indexes) {
	indexes.AddUnique("path", "lang", "Type")
}

// PkgExam represents a package example.
type PkgExam struct {
	Id       int64
	Path     string    `qbs:"size:100,index"`
	Gist     string    // Gist path.
	Examples string    // Examples.
	Created  time.Time `qbs:"index"`
}

// PkgFunc represents a package function.
type PkgFunc struct {
	Id    int64
	Pid   int64 // Id of package documentation it belongs to.
	Path  string
	Name  string `qbs:"size:100,index"`
	Doc   string
	Code  string // Included field 'Decl', formatted.
	IsOld bool   // Indicates if the function no longer exists.
}

func connDb() *qbs.Qbs {
	// 'sql.Open' only returns error when unknown driver, so it's not necessary to check in other places.
	q, _ := qbs.GetQbs()
	return q
}

func setMg() (*qbs.Migration, error) {
	mg, err := qbs.GetMigration()
	return mg, err
}

func init() {
	// Initialize database.
	qbs.Register("mysql", fmt.Sprintf("%v:%v@%v/%v?charset=utf8&parseTime=true",
		beego.AppConfig.String("dbuser"), beego.AppConfig.String("dbpasswd"),
		beego.AppConfig.String("dbhost"), beego.AppConfig.String("dbname")),
		beego.AppConfig.String("dbname"), qbs.NewMysql())

	// Connect to database.
	q := connDb()
	defer q.Close()

	mg, err := setMg()
	if err != nil {
		beego.Error("models.init ->", err)
	}
	defer mg.Close()

	// Create data tables.
	mg.CreateTableIfNotExists(new(PkgInfo))
	mg.CreateTableIfNotExists(new(PkgDecl))
	mg.CreateTableIfNotExists(new(PkgDoc))
	mg.CreateTableIfNotExists(new(PkgExam))
	mg.CreateTableIfNotExists(new(PkgFunc))

	beego.Trace("Initialized database ->", beego.AppConfig.String("dbname"))
}

// GetGoRepo returns packages in go standard library.
func GetGoRepo() ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pkgInfos []*PkgInfo
	condition := qbs.NewCondition("pro_name = ?", "Go")
	err := q.Condition(condition).OrderBy("path").FindAll(&pkgInfos)
	infos := make([]*PkgInfo, 0, 30)
	for _, v := range pkgInfos {
		if strings.Index(v.Path, ".") == -1 {
			infos = append(infos, v)
		}
	}
	return infos, err
}

// SearchDoc returns packages that import path contains keyword.
func SearchDoc(key string) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pkgInfos []*PkgInfo
	condition := qbs.NewCondition("path like ?", "%"+key+"%").Or("synopsis like ?", "%"+key+"%")
	err := q.Condition(condition).Limit(200).OrderByDesc("views").FindAll(&pkgInfos)
	return pkgInfos, err
}

// SearchRawDoc returns results for raw page,
// which are package that import path and synopsis contains keyword.
func SearchRawDoc(key string, isMatchSub bool) (pkgInfos []*PkgInfo, err error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	// Check if need to match sub-packages.
	if isMatchSub {
		condition := qbs.NewCondition("pro_name != ?", "Go")
		condition2 := qbs.NewCondition("path like ?", "%"+key+"%").Or("synopsis like ?", "%"+key+"%")
		err = q.Condition(condition).Condition(condition2).Limit(50).OrderByDesc("views").FindAll(&pkgInfos)
		return pkgInfos, err
	}

	condition := qbs.NewCondition("pro_name like ?", "%"+key+"%").Or("synopsis like ?", "%"+key+"%")
	err = q.Condition(condition).Limit(50).OrderByDesc("views").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetAllPkgs returns all packages information in database
func GetAllPkgs() ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pkgInfos []*PkgInfo
	err := q.OrderBy("pro_name").OrderByDesc("views").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetPkgExams returns user examples from database.
func GetPkgExams(path string) ([]*PkgExam, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pkgExams []*PkgExam
	err := q.WhereEqual("path", path).FindAll(&pkgExams)
	return pkgExams, err
}

func GetAllExams() ([]*PkgExam, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pkgExams []*PkgExam
	err := q.OrderBy("path").FindAll(&pkgExams)
	return pkgExams, err
}

// GetIndexPageInfo returns all data that used for index page.
// One function is for reducing database connect times.
func GetIndexPageInfo() (totalNum int64, popPkgs, importedPkgs []*PkgInfo, err error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	totalNum = q.Count(&PkgInfo{})
	err = q.Offset(25).Limit(39).OrderByDesc("views").FindAll(&popPkgs)
	if err != nil {
		beego.Error("models.GetIndexPageInfo(): popPkgs:", err)
	}
	err = q.Limit(20).OrderByDesc("imported_num").OrderByDesc("views").FindAll(&importedPkgs)
	return totalNum, popPkgs, importedPkgs, nil
}

// GetLabelsPageInfo returns all data that used for labels page.
// One function is for reducing database connect times.
func GetLabelsPageInfo() (WFPros, ORMPros, DBDPros, GUIPros, NETPros, TOOLPros []*PkgInfo, err error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	condition := qbs.NewCondition("labels like ?", "%$wf|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&WFPros)
	condition = qbs.NewCondition("labels like ?", "%$orm|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&ORMPros)
	condition = qbs.NewCondition("labels like ?", "%$dbd|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&DBDPros)
	condition = qbs.NewCondition("labels like ?", "%$gui|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&GUIPros)
	condition = qbs.NewCondition("labels like ?", "%$net|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&NETPros)
	condition = qbs.NewCondition("labels like ?", "%$tool|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&TOOLPros)
	return WFPros, ORMPros, DBDPros, GUIPros, NETPros, TOOLPros, nil
}

// UpdateLabelInfo updates project label information, returns false if the project does not exist.
func UpdateLabelInfo(path string, label string, add bool) bool {
	// Connect to database.
	q := connDb()
	defer q.Close()

	info := new(PkgInfo)
	err := q.WhereEqual("path", path).Find(info)
	if err != nil {
		return false
	}

	i := strings.Index(info.Labels, "$"+label+"|")
	switch {
	case i == -1 && add: // Add operation and does not contain.
		info.Labels += "$" + label + "|"
		_, err = q.Save(info)
		if err != nil {
			beego.Error("models.UpdateLabelInfo -> add:", path, err)
		}
	case i > -1 && !add: // Delete opetation and contains.
		info.Labels = strings.Replace(info.Labels, "$"+label+"|", "", 1)
		_, err = q.Save(info)
		if err != nil {
			beego.Error("models.UpdateLabelInfo -> delete:", path, err)
		}
	}

	return true
}

var buildPicPattern = regexp.MustCompile(`\[+!+\[+([a-zA-Z ]*)+\]+\(+[a-zA-z]+://[^\s]*`)

// SavePkgExam saves user examples to database.
func SavePkgExam(gist *PkgExam) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	pexam := new(PkgExam)
	cond := qbs.NewCondition("path = ?", gist.Path).And("gist = ?", gist.Gist)
	err := q.Condition(cond).Find(pexam)
	if err == nil {
		gist.Id = pexam.Id
	}
	gist.Created = time.Now().UTC()

	_, err = q.Save(gist)
	if err != nil {
		beego.Error("models.SavePkgExam -> ", gist.Path, err)
	}
}

// SavePkgDoc saves readered readme.md file data.
func SavePkgDoc(path, lang string, docBys []byte) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	// Reader readme.
	doc := string(docBys)
	if len(doc) == 0 {
		return
	}

	if doc[0] == '\n' {
		doc = doc[1:]
	}
	// Remove title and `==========`.
	doc = doc[strings.Index(doc, "\n")+1:]
	if len(doc) == 0 {
		return
	}

	if doc[0] == '=' {
		doc = doc[strings.Index(doc, "\n")+1:]
	}
	// Find all picture path of build system. HAVE BUG!!!
	for _, m := range buildPicPattern.FindAllString(doc, -1) {
		start := strings.Index(m, "http")
		end := strings.Index(m, ")")
		if (start > -1) && (end > -1) && (start < end) {
			picPath := m[start:end]
			doc = strings.Replace(doc, m, "![]("+picPath+")", 1)
		}
	}
	doc = string(blackfriday.MarkdownCommon([]byte(doc)))
	doc = strings.Replace(doc, "h3>", "h5>", -1)
	doc = strings.Replace(doc, "h2>", "h4>", -1)
	doc = strings.Replace(doc, "h1>", "h3>", -1)
	doc = strings.Replace(doc, "<center>", "", -1)
	doc = strings.Replace(doc, "</center>", "", -1)
	doc = "<div style='display:block; padding: 3px; border:1px solid #4F4F4F;'>" + doc + "</div>"

	pdoc := new(PkgDoc)
	cond := qbs.NewCondition("path = ?", path).And("lang = ?", lang).And("type = ?", "rm")
	q.Condition(cond).Find(pdoc)
	pdoc.Path = path
	pdoc.Lang = lang
	pdoc.Type = "rm"
	pdoc.Doc = base32.StdEncoding.EncodeToString([]byte(doc))
	_, err := q.Save(pdoc)
	if err != nil {
		beego.Error("models.SavePkgDoc -> readme:", err)
	}
}

// LoadPkgDoc loads project introduction documentation.
func LoadPkgDoc(path, lang, docType string) (doc string) {
	q := connDb()
	defer q.Close()

	pdoc := new(PkgDoc)
	cond := qbs.NewCondition("path = ?", path).And("lang = ?", lang).And("type = ?", docType)
	err := q.Condition(cond).Find(pdoc)
	if err == nil {
		return pdoc.Doc
	}

	cond = qbs.NewCondition("path = ?", path).And("lang = ?", "en").And("type = ?", docType)
	err = q.Condition(cond).Find(pdoc)
	if err == nil {
		return pdoc.Doc
	}
	return doc
}

// GetPkgFunc returns code of given function name and pid.
func GetPkgFunc(name, pidS string) string {
	q := connDb()
	defer q.Close()

	pfunc := new(PkgFunc)
	pid, _ := strconv.Atoi(pidS)
	cond := qbs.NewCondition("pid = ?", int64(pid)).And("name = ?", name)
	err := q.Condition(cond).Find(pfunc)
	if err != nil {
		pfunc.Code = "Function not found:\n" +
			"Pid: " + pidS +
			"\nName: " + name
	}

	return pfunc.Code
}
