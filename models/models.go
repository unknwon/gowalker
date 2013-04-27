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

type pkgInfo struct {
	Id       int64
	Path     string `qbs:"index"`
	Synopsis string
	Views    int64     `qbs:"index"`
	Updated  time.Time `qbs:"index"`
	ProName  string
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
	mg.CreateTableIfNotExists(new(pkgInfo))

	return nil
}

// getDoc returns package documentation in database
func getDoc(path string) (*Package, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	pdoc := new(Package)
	err = q.WhereEqual("path", path).Find(pdoc)

	return pdoc, nil
}

// savePkgInfo saves package to database
func savePkgInfo(pkg *pkgInfo) error {
	q, err := connDb()
	if err != nil {
		return err
	}
	defer q.Db.Close()

	info := new(pkgInfo)
	err = q.WhereEqual("path", pkg.Path).Find(info)
	if info.Synopsis != pkg.Synopsis {
		_, err = q.Save(pkg)
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

	pkg := pkgInfo{Path: path}
	_, err = q.Delete(&pkg)
	return err
}

// GetRecentPkgs gets recent updated packages from database
func GetRecentPkgs() ([]*pkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*pkgInfo
	err = q.Where("views > ?", 0).Limit(25).OrderByDesc("updated").FindAll(&pkgInfos)
	return pkgInfos, err
}

// AddViews add views in database by 1 each time
func AddViews(path string) error {
	q, err := connDb()
	if err != nil {
		return err
	}
	defer q.Db.Close()

	info := new(pkgInfo)
	err = q.WhereEqual("path", path).Find(info)
	if err != nil {
		return err
	}

	info.Views++
	_, err = q.Save(info)
	return err
}

// GetPopularPkgs gets most viewed packages from database
func GetPopularPkgs() ([]*pkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*pkgInfo
	err = q.Where("views > ?", 0).Limit(15).OrderByDesc("views").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetAllPkgs gets all packages in database
func GetAllPkgs() ([]*pkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*pkgInfo
	err = q.Where("views > ?", 0).OrderBy("path").FindAll(&pkgInfos)
	return pkgInfos, err
}

// SearchDoc gets packages that contain keyword
func SearchDoc(key string) ([]*pkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*pkgInfo
	condition := qbs.NewCondition("path like ?", "%"+key+"%").And("views > ?", 0)
	err = q.Condition(condition).OrderBy("path").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetGoRepo gets go standard library
func GetGoRepo() ([]*pkgInfo, error) {
	q, err := connDb()
	if err != nil {
		return nil, err
	}
	defer q.Db.Close()

	var pkgInfos []*pkgInfo
	condition := qbs.NewCondition("pro_name = ?", "Go").And("views > ?", 0)
	err = q.Condition(condition).OrderBy("path").FindAll(&pkgInfos)
	return pkgInfos, err
}
