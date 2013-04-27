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

package models

import (
	"database/sql"
	"time"

	"github.com/coocood/qbs"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_NAME         = "./data/gowalker.db"
	_SQLITE3_DRIVER = "sqlite3"
)

type PkgInfo struct {
	Id        int64
	Path      string `qbs:"index"`
	Synopsis  string
	Views     int64     `qbs:"index"`
	Generated time.Time `qbs:"index"`
	ProName   string
}

func connDb() (*qbs.Qbs, error) {
	db, err := sql.Open(_SQLITE3_DRIVER, DB_NAME)
	q := qbs.New(db, qbs.NewSqlite3())
	return q, err
}

func setMg() (*qbs.Migration, error) {
	db, err := sql.Open(_SQLITE3_DRIVER, DB_NAME)
	mg := qbs.NewMigration(db, DB_NAME, qbs.NewSqlite3())
	return mg, err
}

func InitDb() error {
	q, err := connDb()
	if err != nil {
		return err
	}
	defer q.Db.Close()

	mg, err := setMg()
	if err != nil {
		return err
	}
	defer mg.Db.Close()

	// Create data tables
	mg.CreateTableIfNotExists(new(PkgInfo))

	return nil
}

// getDoc returns package documentation in database
func getDoc(path string) (*Package, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	info := new(PkgInfo)
	err = q.WhereEqual("path", path).Find(info)

	pdoc := &Package{
		ImportPath:  info.Path,
		Synopsis:    info.Synopsis,
		Updated:     info.Generated,
		ProjectName: info.ProName,
	}
	return pdoc, err
}

// savePkgInfo saves package to database
func savePkgInfo(pkg *PkgInfo) error {
	q, err := connDb()
	if err != nil {
		return err
	}
	defer q.Db.Close()

	info := new(PkgInfo)
	err = q.WhereEqual("path", pkg.Path).Find(info)
	if err != nil {
		_, err = q.Save(pkg)
	} else {
		info.Synopsis = pkg.Synopsis
		info.Generated = pkg.Generated
		_, err = q.Save(info)
	}
	return err
}

// deletePkg removes package from database
func deletePkg(path string) error {
	return nil
	q, err := connDb()
	if err != nil {
		return err
	}
	defer q.Db.Close()

	pkg := PkgInfo{Path: path}
	_, err = q.Delete(&pkg)
	return err
}

// GetRecentPkgs gets recent updated packages from database
func GetRecentPkgs(num int) ([]*PkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	err = q.Where("views > ?", 0).Limit(num).OrderByDesc("generated").FindAll(&pkgInfos)
	return pkgInfos, err
}

// AddViews add views in database by 1 each time
func AddViews(pdoc *Package) error {
	q, err := connDb()
	if err != nil {
		return err
	}
	defer q.Db.Close()

	info := new(PkgInfo)
	err = q.WhereEqual("path", pdoc.ImportPath).Find(info)
	if err != nil {
		pkg := PkgInfo{
			Path:      pdoc.ImportPath,
			Synopsis:  pdoc.Synopsis,
			Generated: time.Now().Local(),
			ProName:   pdoc.ProjectName,
			Views:     1}
		err = savePkgInfo(&pkg)
		return err
	}

	info.Views++
	_, err = q.Save(info)
	return err
}

// GetPopularPkgs gets most viewed packages from database
func GetPopularPkgs() ([]*PkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	err = q.Where("views > ?", 0).Limit(15).OrderByDesc("views").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetAllPkgs gets all packages in database
func GetAllPkgs() ([]*PkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	err = q.Where("views > ?", 0).OrderBy("path").FindAll(&pkgInfos)
	return pkgInfos, err
}

// SearchDoc gets packages that contain keyword
func SearchDoc(key string) ([]*PkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	condition := qbs.NewCondition("path like ?", "%"+key+"%").And("views > ?", 0)
	err = q.Condition(condition).OrderBy("path").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetGoRepo gets go standard library
func GetGoRepo() ([]*PkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*PkgInfo
	condition := qbs.NewCondition("pro_name = ?", "Go").And("views > ?", 0)
	err = q.Condition(condition).OrderBy("path").FindAll(&pkgInfos)
	return pkgInfos, err
}
