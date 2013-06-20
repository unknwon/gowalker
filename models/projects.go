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
