// Copyright 2015 Unknwon
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
	"path"
	"time"

	"github.com/Unknwon/com"

	"github.com/Unknwon/gowalker/modules/setting"
)

var (
	ErrEmptyPackagePath     = errors.New("Package import path is empty")
	ErrPackageNotFound      = errors.New("Package does not found")
	ErrPackageVersionTooOld = errors.New("Package version is too old")
)

// PkgInfo represents the package information.
type PkgInfo struct {
	Id         int64
	ImportPath string `xorm:"UNIQUE"`
	Etag       string

	ProjectPath string
	ViewDirPath string
	Synopsis    string

	IsCmd       bool
	IsCgo       bool
	IsGoRepo    bool
	IsGoSubrepo bool

	PkgVer int

	Views  int64
	RefNum int64
	// Indicate how many JS should be downloaded(JsNum=total num - 1)
	JsNum int

	LastView int64 `xorm:"-"`
	Created  int64
}

func (p *PkgInfo) JSPath() string {
	return path.Join(setting.DocsJsPath, p.ImportPath) + ".js"
}

// PACKAGE_VER is modified when previously stored packages are invalid.
const PACKAGE_VER = 1

// SavePkgInfo saves package information.
func SavePkgInfo(pinfo *PkgInfo) (err error) {
	pinfo.PkgVer = PACKAGE_VER
	pinfo.Created = time.Now().UTC().Unix()

	if pinfo.Id == 0 {
		pinfo.Views = 1
		_, err = x.Insert(pinfo)
	} else {
		_, err = x.Id(pinfo.Id).AllCols().Update(pinfo)
	}
	return err
}

// GetPkgInfo returns package information by given import path.
func GetPkgInfo(importPath string) (*PkgInfo, error) {
	if len(importPath) == 0 {
		return nil, ErrEmptyPackagePath
	}

	pinfo := new(PkgInfo)
	has, err := x.Where("import_path=?", importPath).Get(pinfo)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrPackageNotFound
	} else if pinfo.PkgVer < PACKAGE_VER {
		pinfo.Etag = ""
		return pinfo, ErrPackageVersionTooOld
	}

	if !com.IsFile(pinfo.JSPath()) {
		pinfo.Etag = ""
		return pinfo, ErrPackageVersionTooOld
	}

	return pinfo, nil
}

// GetPkgInfoById returns package information by given ID.
func GetPkgInfoById(id int64) (*PkgInfo, error) {
	pinfo := new(PkgInfo)
	has, err := x.Id(id).Get(pinfo)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrPackageNotFound
	} else if pinfo.PkgVer < PACKAGE_VER {
		return pinfo, ErrPackageVersionTooOld
	}

	if !com.IsFile(pinfo.JSPath()) {
		return pinfo, ErrPackageVersionTooOld
	}

	return pinfo, nil
}

// SearchPkgInfo searches package information by given keyword.
func SearchPkgInfo(keyword string) ([]*PkgInfo, error) {
	if len(keyword) == 0 {
		return nil, nil
	}
	pkgs := make([]*PkgInfo, 0, 100)
	return pkgs, x.Limit(100).Where("import_path like ?", "%"+keyword+"%").Find(&pkgs)
}
