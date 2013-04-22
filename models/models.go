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
	err = mg.CreateTableIfNotExists(new(pkgInfo))
	if err != nil {
		return err
	}

	return nil
}

func getDoc(path string) (*Package, error) {
	return nil, nil
}

func SearchDoc(key string) []pkgInfo {
	return nil
}
