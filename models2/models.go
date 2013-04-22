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

// CheckPkg checks if package is in database
func CheckPkg(path string) (*PkgInfo, time.Time, error) {
	q, err := ConnDb()
	if err != nil {
		return nil, time.Time{}, err
	}
	defer q.Db.Close()

	// Check package in database
	pkg := new(PkgInfo)
	err = q.WhereEqual("path", path).Find(pkg)
	if err != nil {
		return nil, time.Time{}, err
	}

	// Get package updated time
	nextCrawl := pkg.Updated
	sec := nextCrawl.Second()
	if sec != 0 {
		nextCrawl = time.Unix(int64(sec), 0).UTC()
	}

	return pkg, nextCrawl, nil
}

// savePkgInfo saves package to database
func savePkgInfo(pkg *PkgInfo) error {
	q, err := ConnDb()
	if err != nil {
		return err
	}
	defer q.Db.Close()

	_, err = q.Save(pkg)
	return err
}

// setNextCrawl updates next crawl time of package
func setNextCrawl(path string, nextCrawl time.Time) error {
	q, err := ConnDb()
	if err != nil {
		return err
	}
	defer q.Db.Close()

	pkg := new(PkgInfo)
	q.WhereEqual("path", path).Find(pkg)
	pkg.Updated = nextCrawl

	_, err = q.Save(pkg)
	return err
}

// deletePkg removes package from database
func deletePkg(path string) error {
	q, err := ConnDb()
	if err != nil {
		return err
	}
	defer q.Db.Close()

	pkg := PkgInfo{Path: path}
	_, err = q.Delete(&pkg)
	return err
}
