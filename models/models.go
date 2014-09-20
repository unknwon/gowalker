// Copyright 2013 Unknwon
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
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/astaxie/beego"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"

	"github.com/Unknwon/gowalker/hv"
	"github.com/Unknwon/gowalker/modules/log"
	"github.com/Unknwon/gowalker/modules/setting"
)

// A PkgTag descriables the project revision tag for its sub-projects,
// any project that has this record means it's passed check of Go project,
// and do not need to download whole project archive when refresh.
type PkgTag struct {
	Id   int64
	Path string `xorm:"unique(pkg_tag_path_tag) index VARCHAR(150)"`
	Tag  string `xorm:"unique(pkg_tag_path_tag) VARCHAR(50)"`
	Vcs  string `xorm:"VARCHAR(50)"`
	Tags string `xorm:"TEXT"`
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
	Examples string    `xorm:"TEXT"`         // Examples.
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

	Imports, TestImports string `xorm:"TEXT"`

	IsHasFile   bool
	IsHasSubdir bool
}

// PkgFunc represents a package function.
type PkgFunc struct {
	Id    int64
	Pid   int64  `xorm:"index"` // Id of package documentation it belongs to.
	Path  string `xorm:"VARCHAR(150)"`
	Name  string `xorm:"index VARCHAR(100)"`
	Doc   string `xorm:"TEXT"`
	IsOld bool   // Indicates if the function no longer exists.
}

// PkgImport represents a package imports record.
type PkgImport struct {
	Id      int64
	Path    string `xorm:"index VARCHAR(150)"`
	Imports string `xorm:"TEXT"`
}

var x *xorm.Engine

func init() {
	var err error
	x, err = xorm.NewEngine("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		setting.Cfg.MustValue("database", "USER"),
		setting.Cfg.MustValue("database", "PASSWD"),
		setting.Cfg.MustValue("database", "HOST"),
		setting.Cfg.MustValue("database", "NAME")))
	if err != nil {
		log.Fatal(4, "Fail to init new engine: %v", err)
	} else if err = x.Sync(new(hv.PkgInfo), new(PkgTag), new(PkgRock), new(PkgExam),
		new(PkgDecl), new(PkgFunc), new(PkgImport)); err != nil {
		log.Fatal(4, "Fail to sync database: %v", err)
	}
}

func Ping() error {
	return x.Ping()
}

// GetGoRepo returns packages in go standard library.
func GetGoRepo() (pinfos []*hv.PkgInfo) {
	err := x.Where("is_go_repo = ?", true).Asc("import_path").Find(&pinfos)
	if err != nil {
		beego.Trace("models.GetGoRepo ->", err)
	}
	return pinfos
}

func GetGoSubrepo() (pinfos []*hv.PkgInfo) {
	err := x.Where("is_go_subrepo = ?", true).Asc("import_path").Find(&pinfos)
	if err != nil {
		beego.Trace("models.GetGoSubrepo ->", err)
	}
	return pinfos
}

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
	if err != nil {
		return errors.New(
			fmt.Sprintf("models.SavePkgExam( %s ) -> Get PkgExam: %s",
				gist.Path, err))
	}
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
		_, err = x.Id(gist.Id).Update(gist)
	} else {
		_, err = x.Insert(gist)
	}
	if err != nil {
		return errors.New(
			fmt.Sprintf("models.SavePkgExam( %s ) -> Save PkgExam: %s", gist.Path, err))
	}

	// Delete 'PkgDecl' in order to generate new page.
	_, err = x.Where("pid = ?", pinfo.Id).And("tag = ?", "").Delete(new(PkgDecl))
	if err != nil {
		return errors.New(
			fmt.Sprintf("models.SavePkgExam( %s ) -> Delete PkgDecl: %s", gist.Path, err))
	}

	beego.Trace("models.SavePkgExam(", gist.Path, ") -> Saved")
	return nil
}

func handleIllegalChars(data []byte) []byte {
	return bytes.Replace(data, []byte("<"), []byte("&lt;"), -1)
}

// GetIndexStats returns index page statistic information.
func GetIndexStats() (int64, int64, int64) {
	num1, err := x.Count(new(hv.PkgInfo))
	if err != nil {
		beego.Error("models.GetIndexStats -> Fail to count hv.PkgInfo:", err.Error())
	}

	num2, err := x.Count(new(PkgDecl))
	if err != nil {
		beego.Error("models.GetIndexStats -> Fail to count PkgDecl:", err.Error())
	}

	num3, err := x.Count(new(PkgFunc))
	if err != nil {
		beego.Error("models.GetIndexStats -> Fail to count PkgFunc:", err.Error())
	}
	return num1, num2, num3
}

// SearchFunc returns functions that name contains keyword.
func SearchFunc(key string) (pfuncs []PkgFunc) {
	err := x.Limit(200).Where("name like '%" + key + "%'").Find(&pfuncs)
	if err != nil {
		beego.Error("models.SearchFunc -> ", err.Error())
		return pfuncs
	}
	return pfuncs
}
