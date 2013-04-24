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

package utils

import (
	"os"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
)

// IsExist returns if a file or directory exists
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

var readmePat = regexp.MustCompile(`^[Rr][Ee][Aa][Dd][Mm][Ee](?:$|\.)`)

func IsDocFile(n string) bool {
	if strings.HasSuffix(n, ".go") && n[0] != '_' && n[0] != '.' {
		return true
	}
	return readmePat.MatchString(n)
}

// FormatDoc formats documentation in HTML format
func FormatDoc(pkgDoc string) (pkgHTML string) {
	cutIndex := -1
	prefix := true // Indicates if should be replaced by start tag
	for {
		index := strings.Index(pkgDoc[cutIndex+1:], "\n\n") + 1
		beego.Info("index", cutIndex)
		if index > -1 {
			cutIndex += index + 1
			if prefix {
				pkgDoc += pkgDoc[:cutIndex] + strings.Replace(pkgDoc[cutIndex:], "\n\n", "<p>", 1)
				prefix = false
			} else {
				pkgDoc += pkgDoc[:cutIndex] + strings.Replace(pkgDoc[cutIndex:], "\n\n", "</p>", 1)
				prefix = true
			}
			continue
		}

		break
	}

	pkgHTML = pkgDoc
	return pkgHTML
}
