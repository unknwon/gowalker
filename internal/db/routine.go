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

package db

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	log "gopkg.in/clog.v1"

	"github.com/unknwon/gowalker/internal/setting"
	"github.com/unknwon/gowalker/internal/spaces"
)

func RefreshNumTotalPackages() {
	count, _ := x.Count(new(PkgInfo))
	atomic.StoreInt64(&numTotalPackages, count)
}

func ComposeSpacesObjectNames(importPath, etag string, numExtraFiles int) []string {
	names := make([]string, numExtraFiles+1)
	for i := range names {
		if i == 0 {
			names[i] = fmt.Sprintf("%s-%s.js", importPath, etag)
		} else {
			names[i] = fmt.Sprintf("%s-%s-%d.js", importPath, etag, i)
		}
	}
	return names
}

var distributeJSFilesStatus int32 = 0

// DistributeJSFiles uploads local JS files to DigitalOcean Spaces.
func DistributeJSFiles() {
	if !setting.DigitalOcean.Spaces.Enabled {
		return
	}

	if !atomic.CompareAndSwapInt32(&distributeJSFilesStatus, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&distributeJSFilesStatus, 0)

	log.Trace("Routine started: DistributeJSFiles")
	defer log.Trace("Routine ended: DistributeJSFiles")

	if err := x.Where("status = ?", JSFileStatusGenerated).Iterate(new(JSFile), func(idx int, bean interface{}) error {
		jsFile := bean.(*JSFile)

		// Gather package information
		pinfo, err := GetPkgInfoByID(jsFile.PkgID)
		if err != nil {
			if err == ErrPackageVersionTooOld {
				return nil
			}
			log.Error(2, "Failed to get package info by ID[%d]: %v", jsFile.PkgID, err)
			return nil
		}
		log.Trace("DistributeJSFiles[%d]: Distributing %q", jsFile.ID, pinfo.ImportPath)

		// Compose object names
		localJSPaths := pinfo.LocalJSPaths()
		objectNames := ComposeSpacesObjectNames(pinfo.ImportPath, jsFile.Etag, jsFile.NumExtraFiles)
		if len(objectNames) != len(localJSPaths) {
			log.Warn("DistributeJSFiles[%d]: Number of object names does not match local JS files: %d != %d",
				jsFile.ID, len(objectNames), len(localJSPaths))
			return nil
		}
		for i, localPath := range localJSPaths {
			if err = spaces.PutObject(localPath, objectNames[i]); err != nil {
				log.Error(2, "Failed to put object[%s]: %v", objectNames[i], err)
				return nil
			}
		}

		// Update database records and clean up local disk
		jsFile.Status = JSFileStatusDistributed
		if err = SaveJSFile(jsFile); err != nil {
			log.Error(2, "Failed to save JS file[%d]: %v", jsFile.ID, err)
			return nil
		}

		for i := range localJSPaths {
			os.Remove(localJSPaths[i])
		}

		log.Trace("DistributeJSFiles[%d]: Distributed %d files", jsFile.ID, len(objectNames))
		return nil
	}); err != nil {
		log.Error(2, "Failed to distribute JS files: %v", err)
	}
}

var recycleJSFilesStatus int32 = 0

// RecycleJSFiles deletes local or distributed JS files due to inactive status.
func RecycleJSFiles() {
	if !atomic.CompareAndSwapInt32(&recycleJSFilesStatus, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&recycleJSFilesStatus, 0)

	log.Trace("Routine started: RecycleJSFiles")
	defer log.Trace("Routine ended: RecycleJSFiles")

	outdated := time.Now().Add(-1 * time.Duration(setting.Maintenance.JSRecycleDays) * 24 * time.Hour).Unix()
	if err := x.Join("INNER", "pkg_info", "pkg_info.id = js_file.pkg_id").
		Where("js_file.status < ? AND pkg_info.last_viewed < ?", JSFileStatusRecycled, outdated).
		Iterate(new(JSFile), func(idx int, bean interface{}) error {
			jsFile := bean.(*JSFile)

			// Gather package information
			pinfo, err := GetPkgInfoByID(jsFile.PkgID)
			if err != nil {
				if err == ErrPackageVersionTooOld {
					return nil
				}
				log.Error(2, "Failed to get package info by ID[%d]: %v", jsFile.PkgID, err)
				return nil
			}

			var numFiles int
			switch jsFile.Status {
			case JSFileStatusGenerated:
				localJSPaths := pinfo.LocalJSPaths()
				for i := range localJSPaths {
					os.Remove(localJSPaths[i])
				}
				numFiles = len(localJSPaths)

			case JSFileStatusDistributed:
				if !setting.DigitalOcean.Spaces.Enabled {
					log.Warn("RecycleJSFiles[%d]: DigitalOcean Spaces is not enabled", jsFile.ID)
					return nil
				}

				objectNames := ComposeSpacesObjectNames(pinfo.ImportPath, jsFile.Etag, jsFile.NumExtraFiles)
				for i := range objectNames {
					if err = spaces.RemoveObject(objectNames[i]); err != nil {
						log.Error(2, "Failed to remove object[%s]: %v", objectNames[i], err)
						return nil
					}
				}
				numFiles = len(objectNames)

			default:
				log.Warn("RecycleJSFiles[%d]: Unexpected status %v", jsFile.ID, jsFile.Status)
				return nil
			}

			// FIXME: Database could be outdated if this operation fails, human must take action!
			jsFile.Status = JSFileStatusRecycled
			if err = SaveJSFile(jsFile); err != nil {
				log.Error(2, "Failed to save JS file[%d]: %v", jsFile.ID, err)
				return nil
			}

			log.Trace("RecycleJSFiles[%d]: Recycled %d files", jsFile.ID, numFiles)
			return nil
		}); err != nil {
		log.Error(2, "Failed to recycle JS files: %v", err)
	}
}
