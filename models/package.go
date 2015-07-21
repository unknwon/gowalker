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
	"bytes"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Unknwon/com"

	"github.com/Unknwon/gowalker/modules/base"
	"github.com/Unknwon/gowalker/modules/setting"
)

var (
	ErrEmptyPackagePath     = errors.New("Package import path is empty")
	ErrPackageNotFound      = errors.New("Package does not found")
	ErrPackageVersionTooOld = errors.New("Package version is too old")
)

// PkgInfo represents the package information.
type PkgInfo struct {
	ID         int64  `xorm:"pk autoincr"`
	Name       string `xorm:"-"`
	ImportPath string `xorm:"UNIQUE"`
	Etag       string

	ProjectPath string
	ViewDirPath string
	Synopsis    string

	IsCmd       bool
	IsCgo       bool
	IsGoRepo    bool
	IsGoSubrepo bool
	IsGaeRepo   bool

	PkgVer int

	Priority int `xorm:" NOT NULL"`
	Views    int64
	// Indicate how many JS should be downloaded(JsNum=total num - 1)
	JsNum int

	ImportNum int64
	ImportIDs string `xorm:"import_ids LONGTEXT"`
	// Import num usually is small so save it to reduce a database query.
	ImportPaths string `xorm:"LONGTEXT"`

	RefNum int64
	RefIDs string `xorm:"ref_ids LONGTEXT"`

	Subdirs string `xorm:"TEXT"`

	LastView int64 `xorm:"-"`
	Created  int64
}

func (p *PkgInfo) JSPath() string {
	return path.Join(setting.DocsJsPath, p.ImportPath) + ".js"
}

// CanRefresh returns true if package is available to refresh.
func (p *PkgInfo) CanRefresh() bool {
	return time.Now().UTC().Add(-1*setting.RefreshInterval).Unix() > p.Created
}

// GetRefs returns a list of packages that import this one.
func (p *PkgInfo) GetRefs() []*PkgInfo {
	pinfos := make([]*PkgInfo, 0, p.RefNum)
	refIDs := strings.Split(p.RefIDs, "|")
	for i := range refIDs {
		if len(refIDs[i]) == 0 {
			continue
		}

		id := com.StrTo(refIDs[i][1:]).MustInt64()
		if pinfo, _ := GetPkgInfoById(id); pinfo != nil {
			pinfo.Name = path.Base(pinfo.ImportPath)
			pinfos = append(pinfos, pinfo)
		}
	}
	return pinfos
}

// PACKAGE_VER is modified when previously stored packages are invalid.
const PACKAGE_VER = 1

// PkgRef represents temporary reference information of a package.
type PkgRef struct {
	ID         int64  `xorm:"pk autoincr"`
	ImportPath string `xorm:"UNIQUE"`
	RefNum     int64
	RefIDs     string `xorm:"ref_ids LONGTEXT"`
}

func updatePkgRef(pid int64, refPath string) error {
	if base.IsGoRepoPath(refPath) ||
		refPath == "C" ||
		refPath[1] == '.' ||
		!base.IsValidRemotePath(refPath) {
		return nil
	}

	ref := new(PkgRef)
	has, err := x.Where("import_path=?", refPath).Get(ref)
	if err != nil {
		return fmt.Errorf("get PkgRef: %v", err)
	}

	queryStr := "$" + com.ToStr(pid) + "|"
	if !has {
		if _, err = x.Insert(&PkgRef{
			ImportPath: refPath,
			RefNum:     1,
			RefIDs:     queryStr,
		}); err != nil {
			return fmt.Errorf("insert PkgRef: %v", err)
		}
		return nil
	}

	i := strings.Index(ref.RefIDs, queryStr)
	if i > -1 {
		return nil
	}

	ref.RefIDs += queryStr
	ref.RefNum++
	_, err = x.Id(ref.ID).AllCols().Update(ref)
	return err
}

// checkRefs checks if given packages are still referencing this one.
func checkRefs(pinfo *PkgInfo) {
	var buf bytes.Buffer
	pinfo.RefNum = 0
	refIDs := strings.Split(pinfo.RefIDs, "|")
	for i := range refIDs {
		if len(refIDs[i]) == 0 {
			continue
		}

		fmt.Println(com.StrTo(refIDs[i][1:]).MustInt64())
		pkg, _ := GetPkgInfoById(com.StrTo(refIDs[i][1:]).MustInt64())
		if pkg == nil {
			continue
		}

		if strings.Index(pkg.ImportIDs, "$"+com.ToStr(pinfo.ID)+"|") == -1 {
			continue
		}

		buf.WriteString("$")
		buf.WriteString(com.ToStr(pkg.ID))
		buf.WriteString("|")
		pinfo.RefNum++
	}
	pinfo.RefIDs = buf.String()
}

// updateRef updates or crates corresponding reference import information.
func updateRef(pid int64, refPath string) (int64, error) {
	pinfo, err := GetPkgInfo(refPath)
	if err != nil && pinfo == nil {
		if err == ErrPackageNotFound ||
			err == ErrPackageVersionTooOld {
			// Package hasn't existed yet, save to temporary place.
			return 0, updatePkgRef(pid, refPath)
		}
		return 0, fmt.Errorf("GetPkgInfo(%s): %v", refPath, err)
	}

	// Check if reference information has beed recorded.
	queryStr := "$" + com.ToStr(pid) + "|"
	i := strings.Index(pinfo.RefIDs, queryStr)
	if i > -1 {
		return pinfo.ID, nil
	}

	// Add new as needed.
	pinfo.RefIDs += queryStr
	pinfo.RefNum++
	_, err = x.Id(pinfo.ID).AllCols().Update(pinfo)
	return pinfo.ID, err
}

// SavePkgInfo saves package information.
func SavePkgInfo(pinfo *PkgInfo, updateRefs bool) (err error) {
	if len(pinfo.Synopsis) > 255 {
		pinfo.Synopsis = pinfo.Synopsis[:255]
	}

	pinfo.PkgVer = PACKAGE_VER

	switch {
	case pinfo.IsGaeRepo:
		pinfo.Priority = 70
	case pinfo.IsGoSubrepo:
		pinfo.Priority = 80
	case pinfo.IsGoRepo:
		pinfo.Priority = 99
	}

	// When package is not created, there is no ID so check will certainly fail.
	var ignoreCheckRefs bool

	// Create or update package info itself.
	// Note(Unknwon): do this because we need ID field later.
	if pinfo.ID == 0 {
		ignoreCheckRefs = true
		pinfo.Views = 1

		// First time created, check PkgRef.
		ref := new(PkgRef)
		has, err := x.Where("import_path=?", pinfo.ImportPath).Get(ref)
		if err != nil {
			return fmt.Errorf("get PkgRef: %v", err)
		} else if has {
			pinfo.RefNum = ref.RefNum
			pinfo.RefIDs = ref.RefIDs
			if _, err = x.Id(ref.ID).Delete(ref); err != nil {
				return fmt.Errorf("delete PkgRef: %v", err)
			}
		}

		_, err = x.Insert(pinfo)
	} else {
		_, err = x.Id(pinfo.ID).AllCols().Update(pinfo)
	}
	if err != nil {
		return fmt.Errorf("update package info: %v", err)
	}

	// Update package import references.
	// Note(Unknwon): I just don't see the value of who imports STD
	//	when you don't even import and uses what objects.
	if updateRefs && !pinfo.IsGoRepo {
		var buf bytes.Buffer
		paths := strings.Split(pinfo.ImportPaths, "|")
		for i := range paths {
			if base.IsGoRepoPath(paths[i]) {
				continue
			}

			refID, err := updateRef(pinfo.ID, paths[i])
			if err != nil {
				return fmt.Errorf("updateRef: %v", err)
			} else if refID == 0 {
				continue
			}
			buf.WriteString("$")
			buf.WriteString(com.ToStr(refID))
			buf.WriteString("|")
		}
		pinfo.ImportIDs = buf.String()

		if !ignoreCheckRefs {
			// Check packages who import this is still importing.
			checkRefs(pinfo)
		}
		_, err = x.Id(pinfo.ID).AllCols().Update(pinfo)
		return err
	}
	return nil
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

// GetSubPkgs returns sub-projects by given sub-directories.
func GetSubPkgs(importPath string, dirs []string) []*PkgInfo {
	pinfos := make([]*PkgInfo, 0, len(dirs))
	for _, dir := range dirs {
		if len(dir) == 0 {
			continue
		}

		fullPath := importPath + "/" + dir
		if pinfo, err := GetPkgInfo(fullPath); err == nil {
			pinfo.Name = dir
			pinfos = append(pinfos, pinfo)
		} else {
			pinfos = append(pinfos, &PkgInfo{
				Name:       dir,
				ImportPath: fullPath,
			})
		}
	}
	return pinfos
}

// GetPkgInfosByPaths returns a list of packages by given import paths.
func GetPkgInfosByPaths(paths []string) []*PkgInfo {
	pinfos := make([]*PkgInfo, 0, len(paths))
	for _, p := range paths {
		if len(p) == 0 {
			continue
		}

		if pinfo, err := GetPkgInfo(p); err == nil {
			pinfo.Name = path.Base(p)
			pinfos = append(pinfos, pinfo)
		} else {
			pinfos = append(pinfos, &PkgInfo{
				Name:       path.Base(p),
				ImportPath: p,
			})
		}
	}
	return pinfos
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

func getRepos(trueCondition string) ([]*PkgInfo, error) {
	pkgs := make([]*PkgInfo, 0, 100)
	return pkgs, x.Desc("views").Where(trueCondition+"=?", true).Find(&pkgs)
}

func GetGoRepos() ([]*PkgInfo, error) {
	return getRepos("is_go_repo")
}

func GetGoSubepos() ([]*PkgInfo, error) {
	return getRepos("is_go_subrepo")
}

func GetGAERepos() ([]*PkgInfo, error) {
	return getRepos("is_gae_repo")
}

// SearchPkgInfo searches package information by given keyword.
func SearchPkgInfo(limit int, keyword string) ([]*PkgInfo, error) {
	if len(keyword) == 0 {
		return nil, nil
	}
	pkgs := make([]*PkgInfo, 0, limit)
	return pkgs, x.Limit(limit).Desc("priority").Desc("views").Where("import_path like ?", "%"+keyword+"%").Find(&pkgs)
}
