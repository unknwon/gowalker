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
	"os"
	"reflect"
	"time"
)

type Value struct {
	Decl    Code
	FmtDecl string
	URL     string
	Doc     string
}

type Type struct {
	Doc      string
	Name     string
	Decl     Code
	FmtDecl  string
	URL      string
	Kind     reflect.Kind
	Consts   []*Value
	Vars     []*Value
	Funcs    []*Func
	Methods  []*Func
	Examples []*Example
	//Signatures []MethodSignature
}

type Func struct {
	Decl     Code
	FmtDecl  string
	URL      string
	Doc      string
	Name     string
	Recv     string
	Examples []*Example
}

type Note struct {
	URL  string
	UID  string
	Body string
}

type Example struct {
	Name   string
	Doc    string
	Code   Code
	Play   string
	Output string
}

type File struct {
	Name string
	URL  string
}

// PackageVersion is modified when previously stored packages are invalid.
const PackageVersion = "1"

type Package struct {
	// The import path for this package.
	ImportPath string

	// Import path prefix for all packages in the project.
	ProjectRoot string

	// Name of the project.
	ProjectName string

	// Project home page.
	ProjectURL string

	// Errors found when fetching or parsing this package.
	Errors []string

	// Packages referenced in README files.
	References []string

	// Version control system: git, hg, bzr, ...
	VCS string

	// The time this object was created.
	Updated time.Time

	// Package name or "" if no package for this import path. The proceeding
	// fields are set even if a package is not found for the import path.
	Name string

	// Synopsis and full documentation for package.
	Synopsis string
	Doc      string

	// Format this package as a command.
	IsCmd bool

	// True if package documentation is incomplete.
	Truncated bool

	// Environment
	GOOS, GOARCH string

	// Top-level declarations.
	Consts []*Value
	Funcs  []*Func
	Types  []*Type
	Vars   []*Value

	// Package examples
	Examples []*Example

	Notes map[string][]*Note
	Bugs  []string

	// Source.
	BrowseURL string
	Files     []*File
	TestFiles []*File

	// Source size in bytes.
	SourceSize     int
	TestSourceSize int

	// Imports
	Imports      []string
	TestImports  []string
	XTestImports []string
}

type source struct {
	name      string
	browseURL string
	rawURL    string
	data      []byte
}

func (s *source) Name() string       { return s.name }
func (s *source) Size() int64        { return int64(len(s.data)) }
func (s *source) Mode() os.FileMode  { return 0 }
func (s *source) ModTime() time.Time { return time.Time{} }
func (s *source) IsDir() bool        { return false }
func (s *source) Sys() interface{}   { return nil }
