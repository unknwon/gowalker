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

	"github.com/Unknwon/com"
	"github.com/Unknwon/ctw/packer"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
	"github.com/coocood/qbs"
)

// SearchPkg returns packages that import path and synopsis contains keyword.
func SearchPkg(key string) []*hv.PkgInfo {
	q := connDb()
	defer q.Close()

	var pinfos []*hv.PkgInfo
	cond := qbs.NewCondition("path like ?", "%"+key+"%").Or("synopsis like ?", "%"+key+"%")
	q.OmitFields("ProName", "IsCmd", "Tags", "Views", "ViewedTime", "Created",
		"Etag", "Labels", "ImportedNum", "ImportPid", "Note").
		Limit(200).Condition(cond).OrderByDesc("rank").FindAll(&pinfos)
	return pinfos
}

func getPkgInfoWithQ(path, tag string, q *qbs.Qbs) (*hv.PkgInfo, error) {
	// Check path length to reduce connect times.
	if len(path) == 0 {
		return nil, errors.New("models.GetPkgInfo -> Empty path as not found.")
	}

	pinfo := new(hv.PkgInfo)
	q.WhereEqual("import_path", path).Find(pinfo)

	ptag := new(PkgTag)
	cond := qbs.NewCondition("path = ?", packer.GetProjectPath(path)).And("tag = ?", tag)
	err := q.Condition(cond).Find(ptag)
	if err != nil {
		return pinfo, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> 'PkgTag': %s", path, tag, err))
	}
	pinfo.Vcs = ptag.Vcs
	pinfo.Ptag = ptag.Ptag
	pinfo.Tags = ptag.Tags

	// Only 'PkgInfo' cannot prove that package exists,
	// we have to check 'PkgDecl' as well in case it was deleted by mistake.

	pdecl := new(PkgDecl)
	cond = qbs.NewCondition("pid = ?", pinfo.Id).And("tag = ?", tag)
	err = q.Condition(cond).Find(pdecl)
	if err != nil {
		// Basically, error means not found, so we set 'pinfo.PkgVer' to 0
		// because server uses it to decide whether force update.
		pinfo.PkgVer = 0
		return pinfo, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> 'PkgDecl': %s", path, tag, err))
	}

	docPath := path + packer.TagSuffix("-", tag)
	if !com.IsExist("./static/docs/" + docPath + ".js") {
		pinfo.PkgVer = 0
		return pinfo, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> JS: File not found", path, tag))
	}

	return pinfo, nil
}

// GetPkgInfo returns 'PkgInfo' by given import path and tag.
// It returns error when the package does not exist.
func GetPkgInfo(path, tag string) (*hv.PkgInfo, error) {
	q := connDb()
	defer q.Close()

	return getPkgInfoWithQ(path, tag, q)
}

// GetPkgInfoById returns package information from database by pid.
func GetPkgInfoById(pid int) (*hv.PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	pinfo := new(hv.PkgInfo)
	err := q.WhereEqual("id", pid).Find(pinfo)

	return pinfo, err
}

// GetGroupPkgInfo returns group of package infomration in order to reduce database connect times.
func GetGroupPkgInfo(paths []string) ([]*hv.PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	pinfos := make([]*hv.PkgInfo, 0, len(paths))
	for _, v := range paths {
		if len(v) > 0 {
			pinfo := new(hv.PkgInfo)
			err := q.WhereEqual("path", v).Find(pinfo)
			if err == nil {
				pinfos = append(pinfos, pinfo)
			} else {
				pinfos = append(pinfos, &hv.PkgInfo{ImportPath: v})
			}
		}
	}
	return pinfos, nil
}

// GetGroupPkgInfoById returns group of package infomration by pid.
func GetGroupPkgInfoById(pids []string) []*hv.PkgInfo {
	q := connDb()
	defer q.Close()

	pinfos := make([]*hv.PkgInfo, 0, len(pids))
	for _, v := range pids {
		pid, _ := strconv.ParseInt(v, 10, 64)
		if pid > 0 {
			pinfo := new(hv.PkgInfo)
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
func GetIndexPkgs(page int) (pkgs []*hv.PkgInfo) {
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

// GetSubPkgs returns sub-projects by given sub-directories.
func GetSubPkgs(importPath, tag string, dirs []string) []*hv.PkgInfo {
	q := connDb()
	defer q.Close()

	pinfos := make([]*hv.PkgInfo, 0, len(dirs))
	for _, v := range dirs {
		v = importPath + "/" + v
		if pinfo, err := getPkgInfoWithQ(v, tag, q); err == nil {
			pinfos = append(pinfos, pinfo)
		} else {
			pinfos = append(pinfos, &hv.PkgInfo{ImportPath: v})
		}
	}
	return pinfos
}
