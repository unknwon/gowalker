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

	"github.com/coocood/qbs"
)

// GetProInfo returns package information from database.
func GetPkgInfo(path, tag string) (*PkgInfo, error) {
	// Check path length to reduce connect times.
	if len(path) == 0 {
		return nil, errors.New("models.GetPkgInfo -> Empty path as not found.")
	}

	// Connect to database.
	q := connDb()
	defer q.Close()

	pinfo := new(PkgInfo)
	err := q.WhereEqual("path", path).Find(pinfo)
	if err != nil {
		return pinfo, err
	}

	pdecl := new(PkgDecl)
	cond := qbs.NewCondition("path = ?", path).And("tag = ?", tag)
	err = q.Condition(cond).Find(pdecl)
	if err != nil {
		pinfo.Etag = ""
	}
	return pinfo, err
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

// GetGroupPkgInfoById returns group of package infomration by pid in order to reduce database connect times.
// The formatted pid looks like '$<pid>|', so we need to cut '$' here.
func GetGroupPkgInfoById(pids []string) ([]*PkgInfo, error) {
	// Connect to database.
	q := connDb()
	defer q.Close()

	pinfos := make([]*PkgInfo, 0, len(pids))
	for _, v := range pids {
		if len(v) > 1 {
			pid, err := strconv.Atoi(v[1:])
			if err == nil {
				pinfo := new(PkgInfo)
				err = q.WhereEqual("id", pid).Find(pinfo)
				if err == nil {
					pinfos = append(pinfos, pinfo)
				}
			}
		}
	}
	return pinfos, nil
}
