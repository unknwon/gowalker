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
	"strconv"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
	"github.com/coocood/qbs"
)

/*
	GetPopulars returns <num>
		1. Recent viewed
		2. Top rank
		3. Top viewed
		4. Rock this week
	projects and recent updated examples.
*/
func GetPopulars(proNum, exNum int) (error, []*PkgExam, []*hv.PkgInfo, []*hv.PkgInfo, []*hv.PkgInfo, []*hv.PkgInfo) {
	q := connDb()
	defer q.Close()

	var ruExs []*PkgExam
	err := q.OmitFields("examples", "created").
		Limit(exNum).OrderByDesc("created").FindAll(&ruExs)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}

	var rvPros, trPros, tvPros, rtwPros []*hv.PkgInfo
	var procks []*PkgRock
	err = q.OmitFields("ProName", "IsCmd", "Tags", "Created",
		"Etag", "Labels", "RefPids", "Note").
		Limit(proNum).OrderByDesc("viewed_time").FindAll(&rvPros)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}
	err = q.OmitFields("ProName", "IsCmd", "Tags", "ViewedTime", "Created",
		"Etag", "Labels", "RefPids", "Note").
		Limit(proNum).OrderByDesc("rank").FindAll(&trPros)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}
	err = q.OmitFields("ProName", "IsCmd", "Tags", "ViewedTime", "Created",
		"Etag", "Labels", "RefPids", "Note").
		Limit(proNum).OrderByDesc("views").FindAll(&tvPros)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}
	err = q.OmitFields("ProName", "IsCmd", "Tags", "ViewedTime", "Created",
		"Etag", "Labels", "RefPids", "Note").
		Limit(proNum).OrderByDesc("delta").FindAll(&procks)
	if err != nil {
		return err, nil, nil, nil, nil, nil
	}
	for _, pr := range procks {
		rtwPros = append(rtwPros, &hv.PkgInfo{
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

func updateImportInfo(q *qbs.Qbs, path string, pid, rank int, add bool) {
	spid := strconv.Itoa(pid)

	// Save package information.
	info := new(hv.PkgInfo)
	err := q.WhereEqual("import_path", path).Find(info)
	if err == nil {
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
			_, err = q.Save(info)
			if err != nil {
				beego.Error("models.updateImportInfo -> add:", path, err)
			}

		} else if i > -1 {
			// Delete operation
			refPids = append(refPids[:i], refPids[i+1:]...)

			info.RefPids = strings.Join(refPids, "|")
			info.RefNum = len(refPids)
			_, err = q.Save(info)
			if err != nil {
				beego.Error("models.updateImportInfo -> delete:", path, err)
			}
		}
		return
	}

	// Record imports.
	pimp := new(PkgImport)
	q.WhereEqual("path", path).Find(pimp)
	pimp.Path = path
	pimps := strings.Split(pimp.Imports, "|")
	i := getRefIndex(pimps, spid)
	if i == -1 {
		pimps = append(pimps, spid)
		pimp.Imports = strings.Join(pimps, "|")
		_, err = q.Save(pimp)
		if err != nil {
			beego.Error("models.updateImportInfo -> record import:", path, err)
		}
	}
}

// SaveProject saves package information, declaration and functions;
// update import information.
func SaveProject(pinfo *hv.PkgInfo, pdecl *PkgDecl, pfuncs []*PkgFunc, imports []string) error {
	q := connDb()
	defer q.Close()

	// Load package information(save after checked import information).
	info := new(hv.PkgInfo)
	err := q.WhereEqual("import_path", pinfo.ImportPath).Find(info)
	if err == nil {
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
			if checkImport(q, info.ImportPath, pid) {
				importPids = append(importPids, v)
			}
		}

		pinfo.RefPids = strings.Join(importPids, "|")
		pinfo.RefNum = len(importPids)
	}

	if isMaster {
		pimp := new(PkgImport)
		err := q.WhereEqual("path", pinfo.ImportPath).Find(pimp)
		if err == nil {
			importPids := strings.Split(pinfo.RefPids, "|")
			pimps := strings.Split(pimp.Imports, "|")
			for _, v := range pimps {
				if len(v) == 0 {
					continue
				}
				if i := getRefIndex(importPids, v); i == -1 {
					importPids = append(importPids, v)
				}
			}
			q.WhereEqual("id", pimp.Id).Delete(pimp)
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

	_, err = q.Save(pinfo)
	if err != nil {
		beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> Information2:", err)
	}

	// Don't need to check standard library and non-master projects.
	if imports != nil && isMaster && !utils.IsGoRepoPath(pinfo.ImportPath) {
		// Other packages.
		for _, v := range imports {
			if !utils.IsGoRepoPath(v) {
				// Only count non-standard library.
				updateImportInfo(q, v, int(pinfo.Id), int(pinfo.Rank), true)
			}
		}
	}
	// ------------- END ------------

	// Save package declaration.
	decl := new(PkgDecl)
	if pdecl != nil {
		cond := qbs.NewCondition("pid = ?", pinfo.Id).And("tag = ?", pdecl.Tag)
		err = q.Condition(cond).Find(decl)
		if err == nil {
			pdecl.Id = decl.Id
		}

		pdecl.Pid = pinfo.Id
		_, err = q.Save(pdecl)
		if err != nil {
			beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> Declaration:", err)
		}

		// ------------------------------
		// Save package tag.
		// ------------------------------

		i := strings.Index(pinfo.ImportPath, pinfo.ProjectName)
		proPath := pinfo.ImportPath[:i+len(pinfo.ProjectName)]
		if utils.IsGoRepoPath(pinfo.ImportPath) {
			proPath = "code.google.com/p/go"
		}
		pkgTag := new(PkgTag)
		cond = qbs.NewCondition("path = ?", proPath).And("tag = ?", pdecl.Tag)
		err = q.Condition(cond).Find(pkgTag)
		if err != nil {
			pkgTag.Path = proPath
			pkgTag.Tag = pdecl.Tag
		}
		pkgTag.Vcs = pinfo.Vcs
		pkgTag.Tags = pinfo.Tags

		_, err = q.Save(pkgTag)
		if err != nil {
			beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> PkgTag:", err)
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
			pfunc := new(pkgFunc)
			pfunc.IsOld = true
			_, err = q.WhereEqual("pid", pdecl.Id).Update(pfunc)
		}

		// Save new ones.
		for _, pf := range pfuncs {
			f := new(PkgFunc)
			cond := qbs.NewCondition("pid = ?", pdecl.Id).And("name = ?", pf.Name)
			err = q.Condition(cond).Find(f)
			if err == nil {
				pf.Id = f.Id
			}

			pf.Pid = pdecl.Id
			_, err = q.Save(pf)
			if err != nil {
				beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> Update function(", pf.Name, "):", err)
			}
		}

		if decl.Id > 0 {
			// Delete old ones if exist.
			cond := qbs.NewCondition("pid = ?", pdecl.Id).And("is_old = ?", true)
			_, err = q.Condition(cond).Delete(new(PkgFunc))
			if err != nil {
				beego.Error("models.SaveProject(", pinfo.ImportPath, ") -> Delete functions:", err)
			}
		}
	}

	// ------------- END ------------

	return nil
}

// checkImport returns true if the package(id) imports given package(path).
func checkImport(q *qbs.Qbs, path string, id int64) bool {
	pinfo := &hv.PkgInfo{
		Id: id,
	}
	err := q.Find(pinfo)
	if err != nil {
		return false
	}

	decl := new(PkgDecl)
	cond := qbs.NewCondition("pid = ?", pinfo.Id).And("tag = ?", "")
	err = q.Condition(cond).Find(decl)
	if err != nil {
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

	q := connDb()
	defer q.Close()

	pdecl := new(PkgDecl)
	cond := qbs.NewCondition("pid = ?", pid).And("tag = ?", tag)
	err := q.Condition(cond).Find(pdecl)
	return pdecl, err
}

// DeleteProject deletes everything of the package,
// and update import information.
func DeleteProject(path string) {
	// Check path length to reduce connect times(except launchpad.net).
	if path[0] != 'l' && len(strings.Split(path, "/")) <= 2 {
		beego.Error("models.DeleteProject(", path, ") -> Short path as not needed")
		return
	}

	q := connDb()
	defer q.Close()

	var i1, i2, i3, i4, i5 int64
	// Delete package information.
	// TODO: NEED TO DELETE ALL SUB-PEOJECTS.
	info := new(hv.PkgInfo)
	err := q.WhereEqual("import_path", path).Find(info)
	if err == nil {
		i1, err = q.WhereEqual("import_path", path).Delete(info)
		if err != nil {
			beego.Error("models.DeleteProject(", path, ") -> Information:", err)
		}
	}

	// Delete package declaration.
	if info.Id > 0 {
		// Find.
		var pdecls []*PkgDecl
		err = q.WhereEqual("pid", info.Id).FindAll(&pdecls)
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
						updateImportInfo(q, v, int(info.Id), 0, false)
					}
				}
			}
		}

		// Delete.
		i2, err = q.WhereEqual("pid", info.Id).Delete(new(PkgDecl))
		if err != nil {
			beego.Error("models.DeleteProject(", path, ") -> Delete declaration:", err)
		}
	}

	// Delete package documentation.
	i3, err = q.WhereEqual("path", path).Delete(new(PkgDoc))
	if err != nil {
		beego.Error("models.DeleteProject(", path, ") -> Documentation:", err)
	}

	// Delete package examples.
	i4, err = q.WhereEqual("path", path).Delete(new(PkgExam))
	if err != nil {
		beego.Error("models.DeleteProject(", path, ") -> Examples:", err)
	}

	// Delete package functions.
	if info.Id > 0 {
		i5, err = q.WhereEqual("path", path).Delete(new(PkgExam))
		if err != nil {
			beego.Error("models.DeleteProject(", path, ") -> Functions:", err)
		}
	}

	if i1+i2+i3+i4+i5 > 0 {
		beego.Info("models.DeleteProject(", path, i1, i2, i3, i4, i5, ")")
	}

	return
}

func calRefRanks(q *qbs.Qbs, refPids []string) int64 {
	refRank := 0
	for _, spid := range refPids {
		pid, _ := strconv.Atoi(spid)
		info := new(hv.PkgInfo)
		err := q.WhereEqual("id", pid).Find(info)
		if err == nil {
			refRank += int(info.Rank) * 10 / 100
		} else {
			beego.Error("models.calRefRanks ->", err)
		}
	}
	return int64(refRank)
}

// FlushCacheProjects saves cache data to database.
func FlushCacheProjects(pinfos []*hv.PkgInfo) {
	q := connDb()
	defer q.Close()

	procks := make([]*PkgRock, 0, len(pinfos))
	// Update project data.
	for _, p := range pinfos {
		info := new(hv.PkgInfo)
		err := q.WhereEqual("import_path", p.ImportPath).Find(info)
		if err == nil {
			// Shoule always be nil, just in case not exist.
			p.Id = info.Id
			// Limit 10 views each period.
			if p.Views-info.Views > 10 {
				p.Views = info.Views + 10
			}
		}

		// Update rank.
		p.Rank = calRefRanks(q, strings.Split(p.RefPids, "|")) + p.Views
		_, err = q.Save(p)
		if err != nil {
			beego.Error("models.FlushCacheProjects(", p.ImportPath, ") ->", err)
		}

		procks = append(procks, &PkgRock{
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
		_, err := q.Where("id > ?", int64(0)).Delete(new(PkgRock))
		if err != nil {
			beego.Error("models.FlushCacheProjects -> Reset rock table:", err)
		}
	} else if time.Now().UTC().Weekday() != time.Monday && !utils.Cfg.MustBool("task", "rock_reset") {
		utils.Cfg.SetValue("task", "rock_reset", "1")
		utils.SaveConfig()
	}

	for _, pr := range procks {
		r := new(PkgRock)
		err := q.WhereEqual("path", pr.Path).Find(r)
		if err == nil {
			pr.Id = r.Id
			r.Delta += pr.Rank - r.Rank
			pr.Delta = r.Delta
		}
		q.Save(pr)
	}
}
