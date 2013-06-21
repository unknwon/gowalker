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
	"strings"

	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

// GetRecentPros gets recent viewed projects from database
func GetRecentPros(num int) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var pkgInfos []*PkgInfo
	err := q.Limit(num).OrderByDesc("viewed_time").FindAll(&pkgInfos)
	return pkgInfos, err
}

// GetPopulars gets <num> most viewed projects and examples from database.
func GetPopulars(proNum, examNum int) ([]*PkgInfo, []*PkgExam) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	var popPros []*PkgInfo
	var popExams []*PkgExam
	q.Limit(proNum).OrderByDesc("views").FindAll(&popPros)
	q.Limit(examNum).OrderByDesc("views").FindAll(&popExams)
	return popPros, popExams
}

// DeleteProject deletes everything about the path in database, and update import information.
func DeleteProject(path string) error {
	// Check path length to reduce connect times. (except launchpad.net)
	if path[0] != 'l' && len(strings.Split(path, "/")) <= 2 {
		return errors.New("models.DeleteProject -> Short path as not needed")
	}

	// Connect to database.
	q := connDb()
	defer q.Close()

	var i1, i2, i3, i4 int64
	// Delete package information.
	info := new(PkgInfo)
	err := q.WhereEqual("path", path).Find(info)
	if err == nil {
		i1, err = q.WhereEqual("path", path).Delete(info)
		if err != nil {
			beego.Error("models.DeleteProject -> Information:", err)
		}
	}

	// Delete package declaration.
	for {
		pdecl := new(PkgDecl)
		err = q.WhereEqual("path", path).Find(pdecl)
		if err != nil {
			// Not found, finish delete.
			break
		}

		i2, err = q.Delete(pdecl)
		if err != nil {
			beego.Error("models.DeleteProject -> Declaration:", err)
		} else if info.Id > 0 && !utils.IsGoRepoPath(path) {
			// Don't need to check standard library.
			// Update import information.
			imports := strings.Split(pdecl.Imports, "|")
			imports = imports[:len(imports)-1]
			for _, v := range imports {
				if !utils.IsGoRepoPath(v) {
					// Only count non-standard library.
					updateImportInfo(q, v, int(info.Id), false)
				}
			}
		}
	}

	// Delete package documentation.
	pdoc := new(PkgDoc)
	i3, err = q.WhereEqual("path", path).Delete(pdoc)
	if err != nil {
		beego.Error("models.DeleteProject -> Documentation:", err)
	}

	// Delete package examples.
	pexam := new(PkgExam)
	i4, err = q.WhereEqual("path", path).Delete(pexam)
	if err != nil {
		beego.Error("models.DeleteProject -> Example:", err)
	}

	if i1+i2+i3+i4 > 0 {
		beego.Info("models.DeleteProject(", path, i1, i2, i3, i4, ")")
	}

	return nil
}
