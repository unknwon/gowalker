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
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
	"github.com/coocood/qbs"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_NAME         = "./data/gowalker.db"
	_SQLITE3_DRIVER = "sqlite3"
)

// PkgInfo is package information.
type PkgInfo struct {
	Id          int64
	Path        string `qbs:"index"` // Import path of package.
	Synopsis    string
	Views       int64     `qbs:"index"`
	Created     time.Time `qbs:"index"` // Time when information last updated.
	ViewedTime  int64     // User viewed time(Unix-timestamp).
	ProName     string    // Name of the project.
	Etag, Tags  string    // Revision tag and project tags.
	ImportedNum int       // Number of packages that imports this project.
	ImportPid   string    // Packages id of packages that imports this project.
}

// PkgDecl is package declaration in database acceptable form.
type PkgDecl struct {
	Path      string `qbs:"pk,index"` // Import path of package.
	Doc       string // Package documentation.
	Truncated bool   // True if package documentation is incomplete.

	// Environment.
	Goos, Goarch string

	// Top-level declarations.
	Consts, Funcs, Types, Vars string

	// Internal declarations.
	Iconsts, Ifuncs, Itypes, Ivars string

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

func connDb() *qbs.Qbs {
	db := qbs.GetFreeDB()
	if db == nil {
		// 'sql.Open' only returns error when unknown driver, so it's not necessary to check in other places.
		db, _ = sql.Open(_SQLITE3_DRIVER, DB_NAME)
	}

	q := qbs.New(db, qbs.NewSqlite3())
	return q
}

func setMg() (*qbs.Migration, error) {
	db, err := sql.Open(_SQLITE3_DRIVER, DB_NAME)
	mg := qbs.NewMigration(db, DB_NAME, qbs.NewSqlite3())
	return mg, err
}

func init() {
	// Initialize database.
	os.Mkdir("./data", os.ModePerm)

	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	mg, err := setMg()
	if err != nil {
		beego.Error("models.init():", err)
	}
	defer mg.Db.Close()

	// Create data tables.
	mg.CreateTableIfNotExists(new(PkgInfo))
	mg.CreateTableIfNotExists(new(PkgDecl))
	mg.CreateTableIfNotExists(new(PkgDoc))

	beego.Trace("Initialized database ->", DB_NAME)
}

// GetProInfo returns package information from database.
func GetPkgInfo(path string) (*PkgInfo, error) {
	// Check path length to reduce connect times.
	if len(path) == 0 {
		return nil, errors.New("models.GetPkgInfo(): Empty path as not found.")
	}

	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	pinfo := new(PkgInfo)
	err := q.WhereEqual("path", path).Find(pinfo)

	return pinfo, err
}

// GetGroupPkgInfo returns group of package infomration in order to reduce database connect times.
func GetGroupPkgInfo(paths []string) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	pinfos := make([]*PkgInfo, 0, len(paths))
	for _, v := range paths {
		if len(v) > 0 {
			pinfo := new(PkgInfo)
			err := q.WhereEqual("path", v).Find(pinfo)
			if err == nil {
				pinfos = append(pinfos, pinfo)
			} else {
				pinfos = append(pinfos, &PkgInfo{Path: v})
			}
		}
	}
	return pinfos, nil
}

// GetPkgInfoById returns package information from database by pid.
func GetPkgInfoById(pid int) (*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	pinfo := new(PkgInfo)
	err := q.WhereEqual("id", pid).Find(pinfo)

	return pinfo, err
}

// GetGroupPkgInfoById returns group of package infomration by pid in order to reduce database connect times.
// The formatted pid looks like '$<pid>|', so we need to cut '$' here.
func GetGroupPkgInfoById(pids []string) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

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

// SaveProject save package information, declaration, documentation to database, and update import information.
func SaveProject(pinfo *PkgInfo, pdecl *PkgDecl, pdoc *PkgDoc, imports []string) error {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	// Save package information.
	info := new(PkgInfo)
	err := q.WhereEqual("path", pinfo.Path).Find(info)
	if err != nil {
		_, err = q.Save(pinfo)
	} else {
		pinfo.Id = info.Id
		_, err = q.Save(pinfo)
	}
	if err != nil {
		beego.Error("models.SaveProject(): Information:", err)
	}

	// Save package declaration
	if pdecl != nil {
		_, err = q.Save(pdecl)
		if err != nil {
			beego.Error("models.SaveProject(): Declaration:", err)
		}
	}

	// Save package documentation
	if pdoc != nil && len(pdoc.Doc) > 0 {
		_, err = q.Save(pdoc)
		if err != nil {
			beego.Error("models.SaveProject(): Documentation:", err)
		}
	}

	// Don't need to check standard library.
	if imports != nil && !utils.IsGoRepoPath(pinfo.Path) {
		// Update import information.
		for _, v := range imports {
			if !utils.IsGoRepoPath(v) {
				// Only count non-standard library.
				updateImportInfo(q, v, int(pinfo.Id), true)
			}
		}
	}
	return nil
}

func updateImportInfo(q *qbs.Qbs, path string, pid int, add bool) {
	// Save package information.
	info := new(PkgInfo)
	err := q.WhereEqual("path", path).Find(info)
	if err == nil {
		// Check if pid exists in this project.
		i := strings.Index(info.ImportPid, "$"+strconv.Itoa(pid)+"|")
		switch {
		case i == -1 && add: // Add operation and does not contain.
			info.ImportPid += "$" + strconv.Itoa(pid) + "|"
			info.ImportedNum++
			_, err = q.Save(info)
			if err != nil {
				beego.Error("models.updateImportInfo(): add:", path, err)
			}
		case i > -1 && !add: // Delete operation and contains.
			info.ImportPid = strings.Replace(info.ImportPid, "$"+strconv.Itoa(pid)+"|", "", 1)
			info.ImportedNum--
			_, err = q.Save(info)
			if err != nil {
				beego.Error("models.updateImportInfo(): delete:", path, err)
			}
		}
	}

	// Error means this project does not exist, simply skip.
}

// DeleteProject deletes everything about the path in database, and update import information.
func DeleteProject(path string) error {
	// Check path length to reduce connect times. (except launchpad.net)
	if path[0] != 'l' && len(strings.Split(path, "/")) <= 2 {
		return errors.New("models.DeleteProject(): Short path as not needed.")
	}

	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	var i1, i2, i3 int64
	// Delete package information.
	info := new(PkgInfo)
	err := q.WhereEqual("path", path).Find(info)
	if err == nil {
		i1, err = q.Delete(info)
		if err != nil {
			beego.Error("models.DeleteProject(): Information:", err)
		}
	}

	// Delete package declaration
	pdecl := new(PkgDecl)
	err = q.WhereEqual("path", path).Find(pdecl)
	if err == nil {
		i2, err = q.Delete(pdecl)
		if err != nil {
			beego.Error("models.DeleteProject(): Declaration:", err)
		} else if info.Id > 0 && !utils.IsGoRepoPath(path) {
			// Don't need to check standard library.
			// Update import information.
			imports := strings.Split(pdecl.Imports, "|")
			imports = imports[:len(imports)-1]
			for _, v := range imports {
				if !utils.IsGoRepoPath(v) {
					// Only count non-standard library.
					updateImportInfo(q, v, int(info.Id), false)
				}
			}
		}
	}

	// Delete package documentation
	pdoc := &PkgDoc{Path: path}
	i3, err = q.Delete(pdoc)
	if err != nil {
		beego.Error("models.DeleteProject(): Documentation:", err)
	}

	if i1+i2+i3 > 0 {
		beego.Info("models.DeleteProject(", path, i1, i2, i3, ")")
	}

	return nil
}

// LoadProject gets package declaration from database.
func LoadProject(path string) (*PkgDecl, error) {
	// Check path length to reduce connect times.
	if len(path) == 0 {
		return nil, errors.New("models.LoadProject(): Empty path as not found.")
	}

	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	pdecl := &PkgDecl{Path: path}
	err := q.WhereEqual("path", path).Find(pdecl)
	return pdecl, err
}

// GetRecentPros gets recent viewed projects from database
func GetRecentPros(num int) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	err := q.Limit(num).OrderByDesc("viewed_time").FindAll(&pkgInfos)
	return pkgInfos, err
}

// AddViews add views in database by 1 each time
func AddViews(pinfo *PkgInfo) error {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

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

// GetPopularPros gets <num> most viewed projects from database with offset.
func GetPopularPros(offset, num int) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	err := q.Offset(offset).Limit(num).OrderByDesc("views").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetGoRepo returns packages in go standard library.
func GetGoRepo() ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	condition := qbs.NewCondition("pro_name = ?", "Go")
	err := q.Condition(condition).OrderBy("path").FindAll(&pkgInfos)
	return pkgInfos, err
}

// SearchDoc returns packages that import path contains keyword.
func SearchDoc(key string) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	condition := qbs.NewCondition("path like ?", "%"+key+"%")
	err := q.Condition(condition).OrderBy("path").FindAll(&pkgInfos)
	return pkgInfos, err
}

// SearchRawDoc returns results for raw page,
// which are package that import path and synopsis contains keyword.
func SearchRawDoc(key string) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	condition := qbs.NewCondition("path like ?", "%"+key+"%").Or("synopsis like ?", "%"+key+"%")
	err := q.Condition(condition).OrderBy("path").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetAllPkgs returns all packages information in database
func GetAllPkgs() ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	err := q.OrderBy("pro_name").OrderByDesc("views").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetIndexPageInfo returns all data that used for index page.
// One function is for reducing database connect times.
func GetIndexPageInfo() (totalNum int64, popPkgs, importedPkgs, WFPros, ORMPros, DBDPros, GUIPros, NETPros []*PkgInfo, err error) {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	totalNum = q.Count(&PkgInfo{})
	err = q.Offset(25).Limit(39).OrderByDesc("views").FindAll(&popPkgs)
	if err != nil {
		beego.Error("models.GetIndexPageInfo(): popPkgs:", err)
	}
	err = q.Limit(20).OrderByDesc("imported_num").OrderByDesc("views").FindAll(&importedPkgs)

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
	return totalNum, popPkgs, importedPkgs, WFPros, ORMPros, DBDPros, GUIPros, NETPros, nil
}

// UpdateTagInfo updates prohect tag information, returns false if the project does not exist.
func UpdateTagInfo(path string, tags []string, add bool) bool {
	// Connect to database.
	q := connDb()
	defer q.Db.Close()

	info := new(PkgInfo)
	err := q.WhereEqual("path", path).Find(info)
	if err != nil {
		return false
	}

	for _, v := range tags {
		i := strings.Index(info.Tags, "$"+v+"|")
		switch {
		case i == -1 && add: // Add operation and does not contain.
			info.Tags += "$" + v + "|"
			_, err = q.Save(info)
			if err != nil {
				beego.Error("models.UpdateTagInfo(): add:", path, err)
			}
		case i > -1 && !add: // Delete opetation and contains.
			info.Tags = strings.Replace(info.Tags, "$"+v+"|", "", 1)
			_, err = q.Save(info)
			if err != nil {
				beego.Error("models.UpdateTagInfo(): add:", path, err)
			}
		}
	}

	return true
}
