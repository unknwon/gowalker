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
	"time"

	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
)

/*
	GetPopulars returns <num>
		1. Recent viewed
		2. Top rank
		3. Top viewed
		4. Rock this week
	projects and recent updated examples.
*/
func GetPopulars(proNum, exNum int) (error, []PkgExam, []hv.PkgInfo, []hv.PkgInfo, []hv.PkgInfo, []hv.PkgInfo) {
	var ruExs []PkgExam
	err := x.Limit(exNum).Desc("created").Find(&ruExs)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}

	var rvPros, trPros, tvPros, rtwPros []hv.PkgInfo
	var procks []PkgRock
	err = x.Limit(proNum).Desc("viewed_time").Find(&rvPros)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}
	err = x.Limit(proNum).Desc("rank").Find(&trPros)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}
	err = x.Limit(proNum).Desc("views").Find(&tvPros)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}
	err = x.Limit(proNum).Desc("delta").Find(&procks)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}
	for _, pr := range procks {
		rtwPros = append(rtwPros, hv.PkgInfo{
			Id:         pr.Pid,
			ImportPath: pr.Path,
			Rank:       pr.Rank,
		})
	}
	return nil, ruExs, rvPros, trPros, tvPros, rtwPros
}

func getRefIndex(refPids []string, pid string) int {
	for i := range refPids {
		if refPids[i] == pid {
			return i
		}
	}
	return -1
}

func updateImportInfo(path string, pid, rank int, add bool) {
	spid := strconv.Itoa(pid)

	// Save package information.
	info := &hv.PkgInfo{ImportPath: path}
	has, err := x.Get(info)
	if err != nil {
		beego.Error("models.updateImportInfo(", path, ") -> Get hv.PkgInfo:", err)
		return
	}

	if has {
		// Check if pid exists in this project.
		refPids := strings.Split(info.RefPids, "|")
		i := getRefIndex(refPids, spid)

		if add {
			// Add operation.
			if i == -1 {
				refPids = append(refPids, spid)
				i = len(refPids) - 1
			}

			info.RefPids = strings.Join(refPids, "|")
			info.RefNum = len(refPids)
			if info.RefNum > 0 && strings.HasPrefix(info.RefPids, "|") {
				info.RefPids = info.RefPids[1:]
				info.RefNum--
			}

			_, err = x.Id(info.Id).Update(info)
			if err != nil {
				beego.Error("models.updateImportInfo(", path, ") -> add:", err)
			}

		} else if i > -1 {
			// Delete operation
			refPids = append(refPids[:i], refPids[i+1:]...)

			info.RefPids = strings.Join(refPids, "|")
			info.RefNum = len(refPids)
			if info.RefNum > 0 && strings.HasPrefix(info.RefPids, "|") {
				info.RefPids = info.RefPids[1:]
				info.RefNum--
			}

			_, err = x.Id(info.Id).Update(info)
			if err != nil {
				beego.Error("models.updateImportInfo(", path, ") -> delete:", err)
			}
		}
		return
	}

	if add {
		// Record imports.
		pimp := &PkgImport{Path: path}
		has, err := x.Get(pimp)
		if err != nil {
			beego.Error("models.updateImportInfo(", path, ") -> Get PkgImport:", err)
		}

		pimp.Path = path
		pimps := strings.Split(pimp.Imports, "|")
		i := getRefIndex(pimps, spid)
		if i == -1 {
			pimps = append(pimps, spid)
			pimp.Imports = strings.Join(pimps, "|")

			if has {
				_, err = x.Id(pimp.Id).Update(pimp)
			} else {
				_, err = x.Insert(pimp)
			}
			if err != nil {
				beego.Error("models.updateImportInfo(", path, ") -> record import:", err)
			}
		}
	}
}

// SaveProject saves package information, declaration and functions;
// update import information.
func SaveProject(pinfo *hv.PkgInfo, pdecl *PkgDecl, pfuncs []PkgFunc, imports []string) error {
	// Load package information(save after checked import information).
	info := &hv.PkgInfo{ImportPath: pinfo.ImportPath}
	has, err := x.Get(info)
	if err != nil {
		return errors.New(
			fmt.Sprintf("models.SaveProject( %s ) -> Get hv.PkgInfo: %s",
				pinfo.ImportPath, err))
	}
	if has {
		pinfo.Id = info.Id
	}

	// ------------------------------
	// Update imported information.
	// ------------------------------

	isMaster := pdecl != nil && len(pdecl.Tag) == 0
	if info.Id > 0 {
		// Current package.
		importeds := strings.Split(info.RefPids, "|")
		importPids := make([]string, 0, len(importeds))
		for _, v := range importeds {
			pid, _ := strconv.ParseInt(v, 10, 64)
			if checkImport(info.ImportPath, pid) {
				importPids = append(importPids, v)
			}
		}

		pinfo.RefPids = strings.Join(importPids, "|")
		pinfo.RefNum = len(importPids)
	}

	if isMaster {
		pimp := &PkgImport{Path: pinfo.ImportPath}
		has, err := x.Get(pimp)
		if err != nil {
			return errors.New(
				fmt.Sprintf("models.SaveProject( %s ) -> Get PkgImport: %s",
					pinfo.ImportPath, err))
		}
		if has {
			importPids := strings.Split(pinfo.RefPids, "|")
			pimps := strings.Split(pimp.Imports, "|")
			for _, v := range pimps {
				if len(v) == 0 {
					continue
				}
				pid, _ := strconv.ParseInt(v, 10, 64)
				if i := getRefIndex(importPids, v); i == -1 &&
					checkImport(info.ImportPath, pid) {
					importPids = append(importPids, v)
				}
			}
			_, err := x.Id(pimp.Id).Delete(pimp)
			if err != nil {
				beego.Error("models.SaveProject(", pinfo.ImportPath,
					") -> Delete PkgImport:", err.Error())
			}

			pinfo.RefPids = strings.Join(importPids, "|")
			pinfo.RefNum = len(importPids)
			if pinfo.RefNum > 0 && strings.HasPrefix(pinfo.RefPids, "|") {
				pinfo.RefPids = pinfo.RefPids[1:]
				pinfo.RefNum--
			}
		}

	} else {
		pinfo.Ptag = info.Ptag
	}

	if has {
		_, err = x.Id(pinfo.Id).Update(pinfo)
	} else {
		_, err = x.Insert(pinfo)
	}
	if err != nil {
		beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> Information2:", err)
	}

	// Don't need to check standard library and non-master projects.
	if imports != nil && isMaster && !utils.IsGoRepoPath(pinfo.ImportPath) {
		// Other packages.
		for _, v := range imports {
			if !utils.IsGoRepoPath(v) {
				// Only count non-standard library.
				updateImportInfo(v, int(pinfo.Id), int(pinfo.Rank), true)
			}
		}
	}
	// ------------- END ------------

	// Save package declaration.
	decl := new(PkgDecl)
	if pdecl != nil {
		has, err := x.Where("pid = ?", pinfo.Id).And("tag = ?", pdecl.Tag).Get(decl)
		if err != nil {
			beego.Error("models.SaveProject(", pinfo.Id, pdecl.Tag,
				") -> Get PkgDecl:", err.Error())
		}
		if has {
			pdecl.Id = decl.Id
		}

		pdecl.Pid = pinfo.Id
		if has {
			_, err = x.Id(pdecl.Id).Update(pdecl)
		} else {
			_, err = x.Insert(pdecl)
		}
		if err != nil {
			beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> Declaration:", err)
		}

		// ------------------------------
		// Save package tag.
		// ------------------------------

		proPath := utils.GetProjectPath(pinfo.ImportPath)
		if utils.IsGoRepoPath(pinfo.ImportPath) {
			proPath = "code.google.com/p/go"
		}
		pkgTag := &PkgTag{
			Path: proPath,
			Tag:  pdecl.Tag,
		}
		has, err = x.Get(pkgTag)
		if err != nil {
			beego.Error("models.SaveProject(", proPath, pdecl.Tag, ") -> Get PkgTag:", err)
		}
		if !has {
			pkgTag.Path = proPath
			pkgTag.Tag = pdecl.Tag
		}
		pkgTag.Vcs = pinfo.Vcs
		pkgTag.Tags = pinfo.Tags

		if has {
			_, err = x.Id(pkgTag.Id).Update(pkgTag)
		} else {
			_, err = x.Insert(pkgTag)
		}
		if err != nil {
			beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> Save PkgTag:", err)
		}

		// ------------- END ------------
	}

	// ------------------------------
	// Save package functions.
	// ------------------------------

	if pfuncs != nil {
		// Old package need to clean old data.
		if decl.Id > 0 {
			// Update all old functions' 'IsOle' to be true.
			type pkgFunc struct {
				IsOld bool
			}
			pfunc := &pkgFunc{IsOld: true}
			_, err = x.Where("pid = ?", pdecl.Id).Update(pfunc)
			if err != nil {
				beego.Error("models.SaveProject(", pdecl.Id, ") -> Mark function old:", err)
			}
		}

		// Save new ones.
		for _, pf := range pfuncs {
			f := &PkgFunc{
				Pid:  pdecl.Id,
				Name: pf.Name,
			}
			has, err := x.Get(f)
			if err != nil {
				beego.Error("models.SaveProject(", pdecl.Id, ") -> Get PkgFunc:", err)
				continue
			}
			if has {
				pf.Id = f.Id
			}

			pf.Pid = pdecl.Id
			if has {
				_, err = x.Id(pf.Id).Update(pf)
			} else {
				_, err = x.Insert(pf)
			}
			if err != nil {
				beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> Update function(", pf.Name, "):", err)
			}
		}

		if decl.Id > 0 {
			// Delete old ones if exist.
			_, err := x.Where("pid = ?", pdecl.Id).And("is_old = ?", true).Delete(new(PkgFunc))
			if err != nil {
				beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> Delete functions:", err)
			}
		}
	}

	// ------------- END ------------

	return nil
}

// checkImport returns true if the package(id) imports given package(path).
func checkImport(path string, id int64) bool {
	pinfo := new(hv.PkgInfo)
	has, err := x.Id(id).Get(pinfo)
	if err != nil {
		beego.Error("models.checkImport(", path, id, ") -> Get hv.PkgInfo", err)
		// When error occurs, the best thing is keep the things same.
		return true
	}
	if !has {
		return false
	}

	decl := new(PkgDecl)
	has, err = x.Where("pid = ?", pinfo.Id).And("tag = ?", "").Get(decl)
	if err != nil {
		beego.Error("models.checkImport(", path, id, pinfo.Id, ") -> Get PkgDecl", err)
		return true
	}
	if !has {
		return false
	}

	if strings.Index(decl.Imports, path) == -1 {
		return false
	}
	return true
}

// LoadProject returns package declaration.
func LoadProject(pid int64, tag string) (*PkgDecl, error) {
	// Check path length to reduce connect times.
	if pid == 0 {
		return nil, errors.New("models.LoadProject -> Zero id.")
	}

	pdecl := &PkgDecl{
		Pid: pid,
		Tag: tag,
	}
	has, err := x.Get(pdecl)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("models.LoadProject -> Project does not exist")
	}
	return pdecl, err
}

// DeleteProject deletes everything of the package,
// and update import information.
func DeleteProject(path string) {
	// Check path length to reduce connect times(except launchpad.net).
	if path[0] != 'l' && len(strings.Split(path, "/")) <= 2 {
		beego.Trace("models.DeleteProject(", path, ") -> Short path as not needed")
		return
	}

	var i1, i2, i3, i4, i5 int64
	// Delete package information.
	// TODO: NEED TO DELETE ALL SUB-PEOJECTS.
	info := &hv.PkgInfo{ImportPath: path}
	has, err := x.Get(info)
	if err != nil {
		beego.Error("models.DeleteProject(", path, ") -> Get hv.PkgInfo", err)
		return
	}
	if has {
		i1, err = x.Where("import_path = ?", path).Delete(info)
		if err != nil {
			beego.Error("models.DeleteProject(", path, ") -> Information:", err)
		}
	}

	// Delete package declaration.
	if info.Id > 0 {
		// Find.
		var pdecls []PkgDecl
		err = x.Where("pid = ?", info.Id).Find(&pdecls)
		if err != nil {
			beego.Error("models.DeleteProject(", path, ") -> Find declaration:", err)
		}

		// Update.
		if !utils.IsGoRepoPath(path) {
			for _, pd := range pdecls {
				// Don't need to check standard library.
				// Update import information.
				imports := strings.Split(pd.Imports, "|")
				imports = imports[:len(imports)-1]
				for _, v := range imports {
					if !utils.IsGoRepoPath(v) {
						// Only count non-standard library.
						updateImportInfo(v, int(info.Id), 0, false)
					}
				}
			}
		}

		// Delete.
		i2, err = x.Where("pid = ?", info.Id).Delete(new(PkgDecl))
		if err != nil {
			beego.Error("models.DeleteProject(", path, ") -> Delete declaration:", err)
		}
	}

	// Delete package documentation.
	i3, err = x.Where("path = ?", path).Delete(new(PkgDoc))
	if err != nil {
		beego.Error("models.DeleteProject(", path, ") -> Documentation:", err)
	}

	// Delete package examples.
	i4, err = x.Where("path = ?", path).Delete(new(PkgExam))
	if err != nil {
		beego.Error("models.DeleteProject(", path, ") -> Examples:", err)
	}

	// Delete package functions.
	if info.Id > 0 {
		i5, err = x.Where("path = ?", path).Delete(new(PkgExam))
		if err != nil {
			beego.Error("models.DeleteProject(", path, ") -> Functions:", err)
		}
	}

	if i1+i2+i3+i4+i5 > 0 {
		beego.Info("models.DeleteProject(", path, i1, i2, i3, i4, i5, ")")
	}

	return
}

func calRefRanks(refPids []string) int64 {
	refRank := 0
	for _, spid := range refPids {
		pid, _ := strconv.Atoi(spid)
		if pid == 0 {
			continue
		}
		info := new(hv.PkgInfo)
		has, err := x.Id(int64(pid)).Get(info)
		if err != nil {
			beego.Error("models.calRefRanks(", pid, ") -> Get hv.PkgInfo:", err)
			continue
		}
		if has {
			refRank += int(info.Rank) * 10 / 100
		} else {
			beego.Trace("models.calRefRanks(", pid, ") ->", err)
		}
	}
	return int64(refRank)
}

// FlushCacheProjects saves cache data to database.
func FlushCacheProjects(pinfos []hv.PkgInfo) {
	procks := make([]PkgRock, 0, len(pinfos))
	// Update project data.
	for _, p := range pinfos {
		info := &hv.PkgInfo{ImportPath: p.ImportPath}
		has, err := x.Get(info)
		if err != nil {
			beego.Error("models.FlushCacheProjects(", p.ImportPath, ") -> Get hv.PkgInfo:", err)
			continue
		}
		if has {
			// Shoule always be nil, just in case not exist.
			p.Id = info.Id
			// Limit 10 views each period.
			if p.Views-info.Views > 10 {
				p.Views = info.Views + 10
			}
		}

		// Update rank.
		p.Rank = calRefRanks(strings.Split(p.RefPids, "|")) + p.Views
		if p.Rank > 2*p.Views {
			p.Rank = 2 * p.Views
		}

		if has {
			_, err = x.Id(p.Id).Update(p)
		} else {
			_, err = x.Insert(p)
		}
		if err != nil {
			beego.Error("models.FlushCacheProjects(", p.ImportPath,
				") -> Save hv.PkgInfo:", err)
			continue
		}

		procks = append(procks, PkgRock{
			Pid:  p.Id,
			Path: p.ImportPath,
			Rank: p.Rank,
		})
	}

	// Update rock this week.
	if time.Now().UTC().Weekday() == time.Monday && utils.Cfg.MustBool("task", "rock_reset") {
		utils.Cfg.SetValue("task", "rock_reset", "0")
		utils.SaveConfig()
		// Reset rock table.
		_, err := x.Where("id > ?", int64(0)).Delete(new(PkgRock))
		if err != nil {
			beego.Error("models.FlushCacheProjects -> Reset rock table:", err)
		}
	} else if time.Now().UTC().Weekday() != time.Monday && !utils.Cfg.MustBool("task", "rock_reset") {
		utils.Cfg.SetValue("task", "rock_reset", "1")
		utils.SaveConfig()
	}

	for _, pr := range procks {
		r := &PkgRock{Path: pr.Path}
		has, err := x.Get(r)
		if err != nil {
			beego.Error("models.FlushCacheProjects(", pr.Path, ") -> Get PkgRock:", err)
			continue
		}
		if has {
			pr.Id = r.Id
			r.Delta += pr.Rank - r.Rank
			pr.Delta = r.Delta
			_, err = x.Id(pr.Id).Update(pr)
		} else {
			_, err = x.Insert(pr)
		}
		if err != nil {
			beego.Error("models.FlushCacheProjects(", pr.Path, ") -> Save PkgRock:", err)
		}
	}
}
