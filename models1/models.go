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

// Redis keys and types:
//
// maxPackageId string: next id to assign
// id:<path> string: id for given import path
// pkg:<id> hash
//      terms: space separated search terms
//      path: import path
//      synopsis: synopsis
//      gob: snappy compressed gob encoded doc.Package
//      rank: document search rank
//      etag:
//      kind: p=package, c=command, d=directory with no go files
// index:<term> set: package ids for given search term
// index:import:<path> set: packages with import path
// index:project:<root> set: packages in project with root
// crawl zset: package id, Unix time for next crawl
// block set: packages to block

// Package database manages storage for GoPkgDoc.
package models

import (
	"database/sql"
	"time"

	"github.com/coocood/qbs"
)

// GetPkgInfo gets the package documenation and sub-directories for the the given
// import path.
func GetPkgInfo(path string) (*Package, []DbPackage, time.Time, error) {
	q, err := ConnDb()
	defer q.Db.Close()
	if err != nil {
		return nil, nil, time.Time{}, err
	}

	pdoc, nextCrawl, err := getDocInfo(q, path)
	if err != nil {
		return nil, nil, time.Time{}, err
	}

	if pdoc != nil {
		// fixup for speclal "-" path.
		path = pdoc.ImportPath
	}

	subdirs, err := getSubdirs(q, path, pdoc)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	return pdoc, subdirs, nextCrawl, nil
}

func getDocInfo(q *qbs.Qbs, path string) (*Package, time.Time, error) {
	r, err := redis.Values(getDocScript.Do(c, path))
	if err == redis.ErrNil {
		return nil, time.Time{}, nil
	} else if err != nil {
		return nil, time.Time{}, err
	}

	var p []byte
	var t int64

	if _, err := redis.Scan(r, &p, &t); err != nil {
		return nil, time.Time{}, err
	}

	p, err = snappy.Decode(nil, p)
	if err != nil {
		return nil, time.Time{}, err
	}

	var pdoc doc.Package
	if err := gob.NewDecoder(bytes.NewReader(p)).Decode(&pdoc); err != nil {
		return nil, time.Time{}, err
	}

	nextCrawl := pdoc.Updated
	if t != 0 {
		nextCrawl = time.Unix(t, 0).UTC()
	}

	return &pdoc, nextCrawl, err
}

func getSubdirs(q *qbs.Qbs, path string, pdoc *Package) ([]DbPackage, error) {

}
