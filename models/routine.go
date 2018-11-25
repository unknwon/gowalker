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

package models

import (
	"fmt"
	"github.com/Unknwon/gowalker/pkg/setting"
	"os"
	"sync/atomic"

	log "gopkg.in/clog.v1"

	"github.com/Unknwon/gowalker/pkg/spaces"
)

func RefreshNumTotalPackages() {
	numTotalPackages, _ = x.Count(new(PkgInfo))
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

		// Compose object names
		localJSPaths := pinfo.LocalJSPaths()
		objectNames := ComposeSpacesObjectNames(pinfo.ImportPath, jsFile.Etag, jsFile.NumExtraFiles)
		if len(objectNames) != len(localJSPaths) {
			log.Warn("DistributeJSFiles[%d]: Number of object names does not match local JS files: %d != %d",
				jsFile.ID, len(objectNames), len(localJSPaths))
			return nil
		}
		for i, localPath := range localJSPaths {
			if err = spaces.UploadFile(localPath, objectNames[i]); err != nil {
				log.Error(2, "Failed to upload object[%s]: %v", objectNames[i], err)
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
