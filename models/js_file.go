// Copyright 2018 Unknwon
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
	"errors"
)

var ErrJSFileNotFound = errors.New("JS file does not exist")

type JSFileStatus int

const (
	JSFileStatusNone JSFileStatus = iota
	JSFileStatusGenerated
	JSFileStatusDistributed
	JSFileStatusRecycled
)

type JSFile struct {
	ID            int64
	PkgID         int64 `xorm:"INDEX"`
	Etag          string
	Status        JSFileStatus
	IsStale       bool // Indicates whether should be recycled
	NumExtraFiles int  // Indicates the number of extra JS files generated
}

func GetJSFile(pkgID int64, etag string) (*JSFile, error) {
	jsFile := new(JSFile)
	has, err := x.Where("pkg_id = ? AND etag = ?", pkgID, etag).Get(jsFile)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrJSFileNotFound
	}

	return jsFile, nil
}

func SaveJSFile(jsFile *JSFile) error {
	oldJSFile, err := GetJSFile(jsFile.PkgID, jsFile.Etag)
	if err != nil && err != ErrJSFileNotFound {
		return err
	}

	if err == ErrJSFileNotFound {
		_, err = x.InsertOne(jsFile)
		return err
	}

	jsFile.ID = oldJSFile.ID
	_, err = x.ID(jsFile.ID).AllCols().Update(jsFile)
	return err
}
