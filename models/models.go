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
	"bytes"
	"encoding/base32"
	"errors"
	"fmt"
	"os/user"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
	//_ "github.com/coocood/mysql"
	"github.com/coocood/qbs"
)

// A PkgTag descriables the project revision tag for its sub-projects,
// any project that has this record means it's passed check of Go project,
// and do not need to download whole project archive when refresh.
type PkgTag struct {
	Id   int64
	Path string `qbs:"size:150,index"`
	Tag  string `qbs:"size:50"`
	Vcs  string `qbs:"size:50"`
	Tags string
}

func (*PkgTag) Indexes(indexes *qbs.Indexes) {
	indexes.AddUnique("path", "tag")
}

// A PkgRock descriables the trending rank of the project.
type PkgRock struct {
	Id    int64
	Pid   int64  `qbs:"index"`
	Path  string `qbs:"size:150"`
	Rank  int64
	Delta int64 `qbs:"index"`
}

// A PkgExam descriables the user example of the project.
type PkgExam struct {
	Id       int64
	Path     string    `qbs:"size:150,index"`
	Gist     string    `qbs:"size:150"` // Gist path.
	Examples string    // Examples.
	Created  time.Time `qbs:"index"`
}

// PkgDecl is package declaration in database acceptable form.
type PkgDecl struct {
	Id  int64
	Pid int64  `qbs:"index"`
	Tag string `qbs:"size:50"`

	// Indicate how many JS should be downloaded(JsNum=total num - 1)
	JsNum       int
	IsHasExport bool

	// Top-level declarations.
	IsHasConst, IsHasVar bool

	IsHasExample bool

	Imports, TestImports string

	IsHasFile   bool
	IsHasSubdir bool
}

func (*PkgDecl) Indexes(indexes *qbs.Indexes) {
	indexes.AddUnique("pid", "tag")
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

// PkgFunc represents a package function.
type PkgFunc struct {
	Id    int64
	Pid   int64  `qbs:"index"` // Id of package documentation it belongs to.
	Path  string `qbs:"size:150"`
	Name  string `qbs:"size:100,index"`
	Doc   string
	IsOld bool // Indicates if the function no longer exists.
}

// PkgImport represents a package imports record.
type PkgImport struct {
	Id      int64
	Path    string `qbs:"size:150,index"`
	Imports string
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

// InitDb initializes the database.
func InitDb() {
	dbName := utils.Cfg.MustValue("db", "name")
	dbPwd := utils.Cfg.MustValue("db", "pwd_"+runtime.GOOS)

	if runtime.GOOS == "darwin" {
		u, err := user.Current()
		if err != nil {
			panic("models.init -> fail to get user: " + err.Error())
		}
		dbPwd = utils.Cfg.MustValue("db", "pwd_"+runtime.GOOS+"_"+u.Username)
	}

	// Register database.
	qbs.Register("mysql", fmt.Sprintf("%v:%v@%v/%v?charset=utf8&parseTime=true",
		utils.Cfg.MustValue("db", "user"), dbPwd,
		utils.Cfg.MustValue("db", "host"), dbName),
		dbName, qbs.NewMysql())

	// Connect to database.
	q := connDb()
	defer q.Close()

	mg, err := setMg()
	if err != nil {
		panic("models.init -> " + err.Error())
	}
	defer mg.Close()

	// Create data tables.
	mg.CreateTableIfNotExists(new(hv.PkgInfo))
	mg.CreateTableIfNotExists(new(PkgTag))
	mg.CreateTableIfNotExists(new(PkgRock))
	mg.CreateTableIfNotExists(new(PkgExam))
	mg.CreateTableIfNotExists(new(PkgDecl))
	mg.CreateTableIfNotExists(new(PkgDoc))
	mg.CreateTableIfNotExists(new(PkgFunc))
	mg.CreateTableIfNotExists(new(PkgImport))

	beego.Trace("Initialized database ->", dbName)
}

// GetGoRepo returns packages in go standard library.
func GetGoRepo() []*hv.PkgInfo {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pinfos []*hv.PkgInfo
	err := q.WhereEqual("is_go_repo", true).OrderBy("import_path").FindAll(&pinfos)
	if err != nil {
		beego.Trace("models.GetGoRepo ->", err)
	}
	return pinfos
}

func GetGoSubrepo() []*hv.PkgInfo {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pinfos []*hv.PkgInfo
	err := q.WhereEqual("is_go_subrepo", true).OrderBy("import_path").FindAll(&pinfos)
	if err != nil {
		beego.Trace("models.GetGoSubrepo ->", err)
	}
	return pinfos
}

// SearchRawDoc returns results for raw page,
// which are package that import path and synopsis contains keyword.
func SearchRawDoc(key string, isMatchSub bool) (pkgInfos []*hv.PkgInfo, err error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	// TODO: need to use q.OmitFields to speed up.
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

// GetPkgExams returns user examples.
func GetPkgExams(path string) ([]*PkgExam, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pkgExams []*PkgExam
	err := q.WhereEqual("path", path).FindAll(&pkgExams)
	return pkgExams, err
}

// GetAllExams returns all user examples.
func GetAllExams() ([]*PkgExam, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pkgExams []*PkgExam
	err := q.OmitFields("Examples", "Created").OrderBy("path").FindAll(&pkgExams)
	return pkgExams, err
}

// GetLabelsPageInfo returns all data that used for labels page.
// One function is for reducing database connect times.
func GetLabelsPageInfo() (WFPros, ORMPros, DBDPros, GUIPros, NETPros, TOOLPros []*hv.PkgInfo, err error) {
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

	info := new(hv.PkgInfo)
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

// SavePkgExam saves user examples.
func SavePkgExam(gist *PkgExam) error {
	q := connDb()
	defer q.Close()

	// Check if corresponding package exists.
	pinfo := new(hv.PkgInfo)
	err := q.WhereEqual("import_path", gist.Path).Find(pinfo)
	if err != nil {
		return errors.New(
			fmt.Sprintf("models.SavePkgExam( %s ) -> Package does not exist", gist.Path))
	}

	pexam := new(PkgExam)
	cond := qbs.NewCondition("path = ?", gist.Path).And("gist = ?", gist.Gist)
	err = q.Condition(cond).Find(pexam)
	if err == nil {
		// Check if refresh too frequently(within in 5 minutes).
		if pexam.Created.Add(5 * time.Minute).UTC().After(time.Now().UTC()) {
			return errors.New(
				fmt.Sprintf("models.SavePkgExam( %s ) -> Refresh too frequently(within in 5 minutes)", gist.Path))
		}
		gist.Id = pexam.Id
	}
	gist.Created = time.Now().UTC()

	_, err = q.Save(gist)
	if err != nil {
		return errors.New(
			fmt.Sprintf("models.SavePkgExam( %s ) -> %s", gist.Path, err))
	}

	// Delete 'PkgDecl' in order to generate new page.
	cond = qbs.NewCondition("pid = ?", pinfo.Id).And("tag = ?", "")
	q.Condition(cond).Delete(new(PkgDecl))

	return nil
}

// SavePkgDoc saves readered readme.md file data.
func SavePkgDoc(path string, readmes map[string][]byte) {
	q := connDb()
	defer q.Close()

	for lang, data := range readmes {
		if len(data) == 0 {
			continue
		}

		if data[0] == '\n' {
			data = data[1:]
		}

		pdoc := new(PkgDoc)
		cond := qbs.NewCondition("path = ?", path).And("lang = ?", lang).And("type = ?", "rm")
		q.Condition(cond).Find(pdoc)
		pdoc.Path = path
		pdoc.Lang = lang
		pdoc.Type = "rm"
		pdoc.Doc = base32.StdEncoding.EncodeToString(handleIllegalChars(data))
		_, err := q.Save(pdoc)
		if err != nil {
			beego.Error("models.SavePkgDoc -> readme:", err)
		}
	}
}

func handleIllegalChars(data []byte) []byte {
	return bytes.Replace(data, []byte("<"), []byte("&lt;"), -1)
}

// LoadPkgDoc loads project introduction documentation.
func LoadPkgDoc(path, lang, docType string) (doc string) {
	q := connDb()
	defer q.Close()

	if len(lang) < 2 {
		return ""
	}

	lang = lang[:2]

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

// GetIndexStats returns index page statistic information.
func GetIndexStats() (int64, int64, int64) {
	q := connDb()
	defer q.Close()

	return q.Count(new(hv.PkgInfo)), q.Count(new(PkgDecl)), q.Count(new(PkgFunc))
}

// SearchFunc returns functions that name contains keyword.
func SearchFunc(key string) []*PkgFunc {
	q := connDb()
	defer q.Close()

	var pfuncs []*PkgFunc
	q.Limit(200).Where("name like ?", "%"+key+"%").FindAll(&pfuncs)
	return pfuncs
}
