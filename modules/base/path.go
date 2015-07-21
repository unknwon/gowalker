// Copyright 2013 The Go Authors. All rights reserved.
// Copyright 2015 Unknwon
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

package base

import (
	"path"
	"regexp"
	"strings"
)

var validHost = regexp.MustCompile(`^[-a-z0-9]+(?:\.[-a-z0-9]+)+$`)
var validPathElement = regexp.MustCompile(`^[-A-Za-z0-9~+][-A-Za-z0-9_.]*$`)

func isValidPathElement(s string) bool {
	return validPathElement.MatchString(s) && s != "testdata"
}

// IsValidRemotePath returns true if importPath is structurally valid for "go get".
func IsValidRemotePath(importPath string) bool {
	parts := strings.Split(importPath, "/")
	if len(parts) <= 1 {
		// Import path must contain at least one "/".
		return false
	}
	if !validTLDs[path.Ext(parts[0])] {
		return false
	}
	if !validHost.MatchString(parts[0]) {
		return false
	}
	for _, part := range parts[1:] {
		if !isValidPathElement(part) {
			return false
		}
	}
	return true
}

// IsGoRepoPath returns true if path is in $GOROOT/src.
func IsGoRepoPath(path string) bool {
	return PathFlag(path)&goRepoPath != 0
}

// IsGAERepoPath returns true if path is from appengine SDK.
func IsGAERepoPath(path string) bool {
	return PathFlag(path)&gaeRepoPath != 0
}

// IsValidPath returns true if importPath is structurally valid.
func IsValidPath(importPath string) bool {
	return PathFlag(importPath)&packagePath != 0 || IsValidRemotePath(importPath)
}

func IsDocFile(n string) bool {
	if strings.HasSuffix(n, ".go") && n[0] != '_' && n[0] != '.' {
		return true
	}
	return strings.HasPrefix(strings.ToLower(n), "readme")
}

var filterDirNames = []string{
	"static", "docs", "views", "js", "assets", "public", "img", "css"}

// FilterDirName guess the file or directory is or contains Go source files.
func FilterDirName(name string) bool {
	for _, v := range filterDirNames {
		if strings.Contains(strings.ToLower(name), "/"+v+"/") {
			return false
		}
	}
	return true
}
