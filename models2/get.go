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
	"log"
	"os"
	"strings"
	"time"
)

// generatePage generates documentation page for package
func generatePage(pkg *Package) error {
	name := strings.Replace(pkg.ImportPath, "/", ".", -1)
	fmt.Println(name)
	f, _ := os.Open(name + ".html")
	defer f.Close()
	f.Write([]byte(name))
	return nil
}

// SearchDoc searchs documentation in database
func SearchDoc(key string) []PkgInfo {
	return []PkgInfo{}
}
