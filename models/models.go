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
	"os"
	"os/user"
	"regexp"
	"runtime"
	//"strings"
	"time"

	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
	"github.com/lunny/xorm"

	"github.com/Unknwon/gowalker/utils"
)

// A PkgTag descriables the project revision tag for its sub-projects,
// any project that has this record means it's passed check of Go project,
// and do not need to download whole project archive when refresh.
type PkgTag struct {
	Id   int64
	Path string `xorm:"unique(pkg_tag_path_tag) index VARCHAR(150)"`
	Tag  string `xorm:"unique(pkg_tag_path_tag) VARCHAR(50)"`
	Vcs  string `xorm:"VARCHAR(50)"`
	Tags string
}

// A PkgRock descriables the trending rank of the project.
type PkgRock struct {
	Id    int64
	Pid   int64  `xorm:"index"`
	Path  string `xorm:"VARCHAR(150)"`
	Rank  int64
	Delta int64 `xorm:"index"`
}

// A PkgExam descriables the user example of the project.
type PkgExam struct {
	Id       int64
	Path     string    `xorm:"index VARCHAR(150)"`
	Gist     string    `xorm:"VARCHAR(150)"` // Gist path.
	Examples string    // Examples.
	Created  time.Time `xorm:"index"`
}

// PkgDecl is package declaration in database acceptable form.
type PkgDecl struct {
	Id  int64
	Pid int64  `xorm:"unique(pkg_decl_pid_tag) index"`
	Tag string `xorm:"unique(pkg_decl_pid_tag) VARCHAR(50)"`

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

// PkgDoc is package documentation for multi-language usage.
type PkgDoc struct {
	Id   int64
	Path string `xorm:"unique(pkg_decl_path_lang_type) index VARCHAR(100)"`
	Lang string `xorm:"unique(pkg_decl_path_lang_type)"` // Documentation language.
	Type string `xorm:"unique(pkg_decl_path_lang_type)"`
	Doc  string // Documentataion.
}

// PkgFunc represents a package function.
type PkgFunc struct {
	Id    int64
	Pid   int64  `xorm:"index"` // Id of package documentation it belongs to.
	Path  string `xorm:"VARCHAR(150)"`
	Name  string `xorm:"index VARCHAR(100)"`
	Doc   string
	IsOld bool // Indicates if the function no longer exists.
}

// PkgImport represents a package imports record.
type PkgImport struct {
	Id      int64
	Path    string `xorm:"index VARCHAR(150)"`
	Imports string
}

var x *xorm.Engine

func setEngine() {
	dbName := utils.Cfg.MustValue("db", "name")
	dbPwd := utils.Cfg.MustValue("db", "pwd_"+runtime.GOOS)

	if runtime.GOOS == "darwin" {
		u, err := user.Current()
		if err != nil {
			beego.Critical("models.init -> fail to get user:", err.Error())
			os.Exit(2)
		}
		dbPwd = utils.Cfg.MustValue("db", "pwd_"+runtime.GOOS+"_"+u.Username)
	}

	var err error
	x, err = xorm.NewEngine("mysql", fmt.Sprintf("%v:%v@%v/%v?charset=utf8",
		utils.Cfg.MustValue("db", "user"), dbPwd,
		utils.Cfg.MustValue("db", "host"), dbName))
	if err != nil {
		beego.Critical("models.init -> fail to conntect database:", err.Error())
		os.Exit(2)
	}

	x.ShowDebug = true
	x.ShowErr = true
	x.ShowSQL = true

	beego.Trace("Initialized database ->", dbName)
}

// InitDb initializes the database.
func InitDb() {
	setEngine()
	x.Sync(new(hv.PkgInfo), new(PkgTag), new(PkgRock), new(PkgExam),
		new(PkgDecl), new(PkgDoc), new(PkgFunc), new(PkgImport))
}

// GetGoRepo returns packages in go standard library.
func GetGoRepo() (pinfos []hv.PkgInfo) {
	err := x.Where("is_go_repo = ?", true).Asc("import_path").Find(&pinfos)
	if err != nil {
		beego.Trace("models.GetGoRepo ->", err)
	}
	return pinfos
}

func GetGoSubrepo() (pinfos []hv.PkgInfo) {
	err := x.Where("is_go_subrepo = ?", true).Asc("import_path").Find(&pinfos)
	if err != nil {
		beego.Trace("models.GetGoSubrepo ->", err)
	}
	return pinfos
}

// SearchRawDoc returns results for raw page,
// which are package that import path and synopsis contains keyword.
// func SearchRawDoc(key string, isMatchSub bool) (pkgInfos []*hv.PkgInfo, err error) {
// 	// Connect to database.
// 	q := connDb()
// 	defer q.Close()

// 	// TODO: need to use q.OmitFields to speed up.
// 	// Check if need to match sub-packages.
// 	if isMatchSub {
// 		condition := qbs.NewCondition("pro_name != ?", "Go")
// 		condition2 := qbs.NewCondition("path like ?", "%"+key+"%").Or("synopsis like ?", "%"+key+"%")
// 		err = q.Condition(condition).Condition(condition2).Limit(50).OrderByDesc("views").FindAll(&pkgInfos)
// 		return pkgInfos, err
// 	}

// 	condition := qbs.NewCondition("pro_name like ?", "%"+key+"%").Or("synopsis like ?", "%"+key+"%")
// 	err = q.Condition(condition).Limit(50).OrderByDesc("views").FindAll(&pkgInfos)
// 	return pkgInfos, err
// }

// GetPkgExams returns user examples.
func GetPkgExams(path string) (pkgExams []PkgExam, err error) {
	err = x.Where("path = ?", path).Find(&pkgExams)
	return pkgExams, err
}

// GetAllExams returns all user examples.
func GetAllExams() (pkgExams []PkgExam, err error) {
	err = x.Asc("path").Find(&pkgExams)
	return pkgExams, err
}

// GetLabelsPageInfo returns all data that used for labels page.
// One function is for reducing database connect times.
// func GetLabelsPageInfo() (WFPros, ORMPros, DBDPros, GUIPros, NETPros, TOOLPros []*hv.PkgInfo, err error) {
// 	// Connect to database.
// 	q := connDb()
// 	defer q.Close()

// 	condition := qbs.NewCondition("labels like ?", "%$wf|%")
// 	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&WFPros)
// 	condition = qbs.NewCondition("labels like ?", "%$orm|%")
// 	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&ORMPros)
// 	condition = qbs.NewCondition("labels like ?", "%$dbd|%")
// 	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&DBDPros)
// 	condition = qbs.NewCondition("labels like ?", "%$gui|%")
// 	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&GUIPros)
// 	condition = qbs.NewCondition("labels like ?", "%$net|%")
// 	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&NETPros)
// 	condition = qbs.NewCondition("labels like ?", "%$tool|%")
// 	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&TOOLPros)
// 	return WFPros, ORMPros, DBDPros, GUIPros, NETPros, TOOLPros, nil
// }

// UpdateLabelInfo updates project label information, returns false if the project does not exist.
// func UpdateLabelInfo(path string, label string, add bool) bool {
// 	info := new(hv.PkgInfo)
// 	err := q.WhereEqual("path", path).Find(info)
// 	if err != nil {
// 		return false
// 	}

// 	i := strings.Index(info.Labels, "$"+label+"|")
// 	switch {
// 	case i == -1 && add: // Add operation and does not contain.
// 		info.Labels += "$" + label + "|"
// 		_, err = q.Save(info)
// 		if err != nil {
// 			beego.Error("models.UpdateLabelInfo -> add:", path, err)
// 		}
// 	case i > -1 && !add: // Delete opetation and contains.
// 		info.Labels = strings.Replace(info.Labels, "$"+label+"|", "", 1)
// 		_, err = q.Save(info)
// 		if err != nil {
// 			beego.Error("models.UpdateLabelInfo -> delete:", path, err)
// 		}
// 	}

// 	return true
// }

var buildPicPattern = regexp.MustCompile(`\[+!+\[+([a-zA-Z ]*)+\]+\(+[a-zA-z]+://[^\s]*`)

// SavePkgExam saves user examples.
func SavePkgExam(gist *PkgExam) error {
	pinfo := &hv.PkgInfo{ImportPath: gist.Path}
	has, err := x.Get(pinfo)
	if !has || err != nil {
		return errors.New(
			fmt.Sprintf("models.SavePkgExam( %s ) -> Package does not exist: %s",
				gist.Path, err))
	}

	pexam := &PkgExam{
		Path: gist.Path,
		Gist: gist.Gist,
	}
	has, err = x.Get(pexam)
	if has {
		// Check if refresh too frequently(within in 5 minutes).
		if pexam.Created.Add(5 * time.Minute).UTC().After(time.Now().UTC()) {
			return errors.New(
				fmt.Sprintf("models.SavePkgExam( %s ) -> Refresh too frequently(within in 5 minutes)", gist.Path))
		}
		gist.Id = pexam.Id
	}
	gist.Created = time.Now().UTC()

	if has {
		_, err = x.Id(gist.Id).Update(pexam)
	} else {
		_, err = x.Insert(pexam)
	}
	if err != nil {
		return errors.New(
			fmt.Sprintf("models.SavePkgExam( %s ) -> %s", gist.Path, err))
	}

	// Delete 'PkgDecl' in order to generate new page.
	x.Where("pid = ?", pinfo.Id).And("tag = ?", "").Delete(new(PkgDecl))
	return nil
}

// SavePkgDoc saves readered readme.md file data.
func SavePkgDoc(path string, readmes map[string][]byte) {
	for lang, data := range readmes {
		if len(data) == 0 {
			continue
		}

		if data[0] == '\n' {
			data = data[1:]
		}

		pdoc := &PkgDoc{
			Path: path,
			Lang: lang,
			Type: "rm",
		}
		has, _ := x.Get(pdoc)
		pdoc.Path = path
		pdoc.Lang = lang
		pdoc.Type = "rm"
		pdoc.Doc = base32.StdEncoding.EncodeToString(handleIllegalChars(data))

		var err error
		if has {
			_, err = x.Id(pdoc.Id).Update(pdoc)
		} else {
			_, err = x.Insert(pdoc)
		}
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
	if len(lang) < 2 {
		return ""
	}

	lang = lang[:2]

	pdoc := &PkgDoc{
		Path: path,
		Lang: lang,
		Type: docType,
	}

	if has, _ := x.Get(pdoc); has {
		return pdoc.Doc
	}

	pdoc.Lang = "en"
	if has, _ := x.Get(pdoc); has {
		return pdoc.Doc
	}
	return doc
}

// GetIndexStats returns index page statistic information.
func GetIndexStats() (int64, int64, int64) {
	num1, _ := x.Count(new(hv.PkgInfo))
	num2, _ := x.Count(new(PkgDecl))
	num3, _ := x.Count(new(PkgFunc))
	return num1, num2, num3
}

// SearchFunc returns functions that name contains keyword.
func SearchFunc(key string) (pfuncs []PkgFunc) {
	x.Where("name like ?", "%"+key+"%").Find(&pfuncs)
	return pfuncs
}
