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
	"strings"

	"github.com/Unknwon/com"
	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
)

// SearchPkg returns packages that import path and synopsis contains keyword.
func SearchPkg(key string) (pinfos []hv.PkgInfo) {
	err := x.Limit(200).Desc("rank").Where("name like '%" + key + "%'").
		Or("synopsis like '%" + key + "%'").Find(&pinfos)
	if err != nil {
		beego.Error("models.SearchPkg -> ", err.Error())
		return pinfos
	}
	return pinfos
}

// GetPkgInfo returns 'PkgInfo' by given import path and tag.
// It returns error when the package does not exist.
func GetPkgInfo(path, tag string) (*hv.PkgInfo, error) {
	// Check path length to reduce connect times.
	if len(path) == 0 {
		return nil, errors.New("models.GetPkgInfo -> Empty path as not found.")
	}

	pinfo := &hv.PkgInfo{ImportPath: path}
	has, err := x.Get(pinfo)
	if !has || err != nil {
		return nil, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> Get hv.PkgInfo: %s",
				path, tag, err))
	}

	proPath := utils.GetProjectPath(path)
	if utils.IsGoRepoPath(path) {
		proPath = "code.google.com/p/go"
	}
	beego.Trace("models.GetPkgInfo -> proPath:", proPath)

	ptag := &PkgTag{
		Path: proPath,
		Tag:  tag,
	}
	has, err = x.Get(ptag)
	if !has || err != nil {
		pinfo.Ptag = "ptag"
		return nil, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> Get PkgTag: %s",
				path, tag, err))
	}

	pinfo.Vcs = ptag.Vcs
	pinfo.Tags = ptag.Tags

	// Only 'PkgInfo' cannot prove that package exists,
	// we have to check 'PkgDecl' as well in case it was deleted by mistake.

	pdecl := &PkgDecl{
		Pid: pinfo.Id,
		Tag: tag,
	}
	has, err = x.Get(pdecl)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> Get PkgDecl: %s", path, tag, err))
	}
	if !has {
		pinfo.PkgVer = 0
		pinfo.Ptag = "ptag"
		return pinfo, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> PkgDecl not exist: %s", path, tag, err))
	}

	docPath := path + utils.TagSuffix("-", tag)
	if !com.IsExist("." + utils.DocsJsPath + docPath + ".js") {
		pinfo.PkgVer = 0
		pinfo.Ptag = "ptag"
		return pinfo, errors.New(
			fmt.Sprintf("models.GetPkgInfo( %s:%s ) -> JS: File not found", path, tag))
	}

	return pinfo, nil
}

// GetPkgInfoById returns package information from database by pid.
func GetPkgInfoById(pid int) (pinfo *hv.PkgInfo, err error) {
	_, err = x.Id(int64(pid)).Get(pinfo)
	return pinfo, err
}

// GetGroupPkgInfo returns group of package infomration in order to reduce database connect times.
func GetGroupPkgInfo(paths []string) []hv.PkgInfo {
	pinfos := make([]hv.PkgInfo, 0, len(paths))
	for _, v := range paths {
		if len(v) > 0 {
			pinfo := &hv.PkgInfo{ImportPath: v}
			has, err := x.Get(pinfo)
			if err != nil {
				beego.Error("models.GetGroupPkgInfo(", v, ") -> Get PkgDoc:", err.Error())
			}
			if has {
				pinfos = append(pinfos, *pinfo)
			} else {
				pinfos = append(pinfos, hv.PkgInfo{ImportPath: v})
			}
		}
	}
	return pinfos
}

// GetGroupPkgInfoById returns group of package infomration by pid.
func GetGroupPkgInfoById(pids []string) []hv.PkgInfo {
	pinfos := make([]hv.PkgInfo, 0, len(pids))
	for _, v := range pids {
		pid, _ := strconv.ParseInt(v, 10, 64)
		if pid > 0 {
			pinfo := new(hv.PkgInfo)
			has, err := x.Id(pid).Get(pinfo)
			if err != nil {
				beego.Error("models.GetGroupPkgInfoById(", pid, ") -> Get hv.PkgInfo:", err.Error())
			}
			if has {
				pinfos = append(pinfos, *pinfo)
			} else {
				beego.Trace("models.GetGroupPkgInfoById -> Not exist:", pid)
			}
		}
	}
	return pinfos
}

// GetIndexPkgs returns package information in given page.
func GetIndexPkgs(page int) (pkgs []hv.PkgInfo) {
	err := x.Limit(100, (page-1)*100).Asc("rank").Find(&pkgs)
	if err != nil {
		beego.Error("models.GetIndexPkgs ->", err)
	}
	return pkgs
}

// GetSubPkgs returns sub-projects by given sub-directories.
func GetSubPkgs(importPath, tag string, dirs []string) []hv.PkgInfo {
	pinfos := make([]hv.PkgInfo, 0, len(dirs))
	for _, v := range dirs {
		v = importPath + "/" + v
		if pinfo, err := GetPkgInfo(v, tag); err == nil {
			pinfos = append(pinfos, *pinfo)
		} else {
			pinfos = append(pinfos, hv.PkgInfo{ImportPath: v})
		}
	}
	return pinfos
}

func GetImports(spid, tag string) []hv.PkgInfo {
	pid, _ := strconv.ParseInt(spid, 10, 64)
	pdecl := &PkgDecl{
		Pid: pid,
		Tag: tag,
	}
	has, err := x.Get(pdecl)
	if !has || err != nil {
		beego.Error("models.GetImports(", pid, tag, ") -> ", err.Error())
		return nil
	}

	return GetGroupPkgInfo(strings.Split(pdecl.Imports, "|"))
}

func GetRefs(spid string) []hv.PkgInfo {
	pid, _ := strconv.ParseInt(spid, 10, 64)
	pinfo := new(hv.PkgInfo)
	has, err := x.Id(pid).Get(pinfo)
	if !has || err != nil {
		beego.Error("models.GetRefs(", pid, ") -> ", err.Error())
		return nil
	}

	return GetGroupPkgInfoById(strings.Split(pinfo.RefPids, "|"))
}
