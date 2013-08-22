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
	"errors"
	"fmt"
	"strconv"

	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
	"github.com/coocood/qbs"
)

// SearchPkg returns packages that import path and synopsis contains keyword.
func SearchPkg(key string) []*PkgInfo {
	q := connDb()
	defer q.Close()

	var pinfos []*PkgInfo
	cond := qbs.NewCondition("path like ?", "%"+key+"%").Or("synopsis like ?", "%"+key+"%")
	q.OmitFields("ProName", "IsCmd", "Tags", "Views", "ViewedTime", "Created",
		"Etag", "Labels", "ImportedNum", "ImportPid", "Note").
		Limit(200).Condition(cond).OrderByDesc("rank").FindAll(&pinfos)
	return pinfos
}

// GetPkgInfo returns 'PkgInfo' by given import path and tag.
// It returns error when the package does not exist.
func GetPkgInfo(path, tag string) (*PkgInfo, error) {
	// Check path length to reduce connect times.
	if len(path) == 0 {
		return nil, errors.New("models.GetPkgInfo -> Empty path as not found.")
	}

	pinfo := new(PkgInfo)

	q := connDb()
	defer q.Close()

	err := q.WhereEqual("path", path).Find(pinfo)
	if err != nil {
		return pinfo, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> 'PkgInfo': %s", path, tag, err))
	}

	// Only 'PkgInfo' cannot prove that package exists,
	// we have to check 'PkgDecl' as well in case it was deleted by mistake.

	pdecl := new(PkgDecl)
	cond := qbs.NewCondition("pid = ?", pinfo.Id).And("tag = ?", tag)
	err = q.Condition(cond).Find(pdecl)
	if err != nil {
		// Basically, error means not found, so we set 'pinfo.Etag' to an empty string
		// because 'Etag' contains 'PACKAGE_VER' which server uses to decide force update.
		pinfo.Etag = ""
		return pinfo, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> 'PkgDecl': %s", path, tag, err))
	}

	docPath := path
	if len(tag) > 0 {
		docPath += "-" + tag
	}
	if !utils.IsExist("./static/docs/" + docPath + ".js") {
		pinfo.Etag = ""
		return pinfo, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> JS: File not found", path, tag))
	}

	return pinfo, nil
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

// GetGroupPkgInfo returns group of package infomration in order to reduce database connect times.
func GetGroupPkgInfo(paths []string) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

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

// GetGroupPkgInfoById returns group of package infomration by pid.
func GetGroupPkgInfoById(pids []string) []*PkgInfo {
	q := connDb()
	defer q.Close()

	pinfos := make([]*PkgInfo, 0, len(pids))
	for _, v := range pids {
		pid, _ := strconv.ParseInt(v, 10, 64)
		if pid > 0 {
			pinfo := new(PkgInfo)
			err := q.WhereEqual("id", pid).Find(pinfo)
			if err == nil {
				pinfos = append(pinfos, pinfo)
			} else {
				beego.Error("models.GetGroupPkgInfoById ->", err)
			}
		}
	}
	return pinfos
}

// GetIndexPkgs returns package information in given page.
func GetIndexPkgs(page int) (pkgs []*PkgInfo) {
	q := connDb()
	defer q.Close()

	err := q.OmitFields("pro_name", "is_cmd", "tags", "views", "viewd_time", "created",
		"etag", "labels", "imported_num", "import_pid", "note").
		Limit(100).Offset((page - 1) * 100).OrderByDesc("Rank").FindAll(&pkgs)
	if err != nil {
		beego.Error("models.GetIndexPkgs ->", err)
	}

	return pkgs
}
