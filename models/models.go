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
	"fmt"
	"regexp"

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
	} else if err = x.Sync(new(hv.PkgInfo), new(PkgTag),
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

var buildPicPattern = regexp.MustCompile(`\[+!+\[+([a-zA-Z ]*)+\]+\(+[a-zA-z]+://[^\s]*`)

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
