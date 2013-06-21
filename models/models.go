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
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/coocood/qbs"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_NAME         = "data/gowalker.db"
	_SQLITE3_DRIVER = "sqlite3"
)

// PkgInfo is package information.
type PkgInfo struct {
	Id           int64
	Path         string `qbs:"index"` // Import path of package.
	Tags         string // All tags of project.
	IsCmd        bool
	Synopsis     string
	Views        int64     `qbs:"index"`
	Updated      time.Time `qbs:"index"` // Time when information last updated.
	ViewedTime   int64     // User viewed time(Unix-timestamp).
	ProName      string    // Name of the project.
	Etag, Labels string    // Revision tag and project labels.
	ImportedNum  int       // Number of packages that imports this project.
	ImportPid    string    // Packages id of packages that imports this project.
}

// PkgDecl is package declaration in database acceptable form.
type PkgDecl struct {
	Id   int64
	Path string `qbs:"index"` // Import path of package.
	Tag  string // Project tag.
	Doc  string // Package documentation.

	// Top-level declarations.
	Consts, Funcs, Types, Vars string

	// Internal declarations.
	Iconsts, Ifuncs, Itypes, Ivars string

	Examples         string // Function or method example.
	Notes            string // Source code notes.
	Files, TestFiles string // Source files.
	Dirs             string // Subdirectories

	Imports, TestImports string // Imports.
}

// PkgDoc is package documentation for multi-language usage.
type PkgDoc struct {
	Path string `qbs:"pk,index"` // Import path of package.
	Lang string // Documentation language.
	Doc  string // Documentataion.
}

// PkgExam represents a package example.
type PkgExam struct {
	Id    int64
	Path  string `qbs:"index"` // Import path of package.
	Views int64  `qbs:"index"`
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
	os.RemoveAll("data/")
	os.Mkdir("data", os.ModePerm)

	qbs.Register(_SQLITE3_DRIVER, DB_NAME, "", qbs.NewSqlite3())
	// Connect to database.
	q := connDb()
	defer q.Close()

	mg, err := setMg()
	if err != nil {
		beego.Error("models.init():", err)
	}
	defer mg.Close()

	// Create data tables.
	mg.CreateTableIfNotExists(new(PkgInfo))
	mg.CreateTableIfNotExists(new(PkgDecl))
	mg.CreateTableIfNotExists(new(PkgDoc))
	mg.CreateTableIfNotExists(new(PkgExam))

	beego.Trace("Initialized database ->", DB_NAME)
}

// GetPkgInfoById returns package information from database by pid.
func GetPkgInfoById(pid int) (*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	pinfo := new(PkgInfo)
	err := q.WhereEqual("id", pid).Find(pinfo)

	return pinfo, err
}

// GetGroupPkgInfoById returns group of package infomration by pid in order to reduce database connect times.
// The formatted pid looks like '$<pid>|', so we need to cut '$' here.
func GetGroupPkgInfoById(pids []string) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	pinfos := make([]*PkgInfo, 0, len(pids))
	for _, v := range pids {
		if len(v) > 1 {
			pid, err := strconv.Atoi(v[1:])
			if err == nil {
				pinfo := new(PkgInfo)
				err = q.WhereEqual("id", pid).Find(pinfo)
				if err == nil {
					pinfos = append(pinfos, pinfo)
				}
			}
		}
	}
	return pinfos, nil
}

// LoadProject gets package declaration from database.
func LoadProject(path string) (*PkgDecl, error) {
	// Check path length to reduce connect times.
	if len(path) == 0 {
		return nil, errors.New("models.LoadProject(): Empty path as not found.")
	}

	// Connect to database.
	q := connDb()
	defer q.Close()

	pdecl := &PkgDecl{Path: path}
	err := q.WhereEqual("path", path).Find(pdecl)
	return pdecl, err
}

// AddViews add views in database by 1 each time
func AddViews(pinfo *PkgInfo) error {
	// Connect to database.
	q := connDb()
	defer q.Close()

	pinfo.Views++

	info := new(PkgInfo)
	err := q.WhereEqual("path", pinfo.Path).Find(info)
	if err != nil {
		_, err = q.Save(pinfo)
	} else {
		pinfo.Id = info.Id
		_, err = q.Save(pinfo)
	}
	_, err = q.Save(pinfo)
	return err
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
	err := q.Condition(condition).Limit(200).OrderBy("path").FindAll(&pkgInfos)
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

// GetTagsPageInfo returns all data that used for tags page.
// One function is for reducing database connect times.
func GetTagsPageInfo() (WFPros, ORMPros, DBDPros, GUIPros, NETPros, TOOLPros []*PkgInfo, err error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	condition := qbs.NewCondition("tags like ?", "%$wf|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&WFPros)
	condition = qbs.NewCondition("tags like ?", "%$orm|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&ORMPros)
	condition = qbs.NewCondition("tags like ?", "%$dbd|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&DBDPros)
	condition = qbs.NewCondition("tags like ?", "%$gui|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&GUIPros)
	condition = qbs.NewCondition("tags like ?", "%$net|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&NETPros)
	condition = qbs.NewCondition("tags like ?", "%$tool|%")
	err = q.Limit(10).Condition(condition).OrderByDesc("views").FindAll(&TOOLPros)
	return WFPros, ORMPros, DBDPros, GUIPros, NETPros, TOOLPros, nil
}

// UpdateTagInfo updates prohect tag information, returns false if the project does not exist.
func UpdateTagInfo(path string, tag string, add bool) bool {
	// Connect to database.
	q := connDb()
	defer q.Close()

	info := new(PkgInfo)
	err := q.WhereEqual("path", path).Find(info)
	if err != nil {
		return false
	}

	i := strings.Index(info.Tags, "$"+tag+"|")
	switch {
	case i == -1 && add: // Add operation and does not contain.
		info.Tags += "$" + tag + "|"
		_, err = q.Save(info)
		if err != nil {
			beego.Error("models.UpdateTagInfo(): add:", path, err)
		}
	case i > -1 && !add: // Delete opetation and contains.
		info.Tags = strings.Replace(info.Tags, "$"+tag+"|", "", 1)
		_, err = q.Save(info)
		if err != nil {
			beego.Error("models.UpdateTagInfo(): add:", path, err)
		}
	}

	return true
}
