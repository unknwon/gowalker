// Copyright 2011 Gary Burd
// Copyright 2013-2014 Unknown
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

package hv

import (
	"go/ast"
	"go/doc"
	"go/token"
	"os"
	"time"
)

// A PkgInfo describles a project information.
type PkgInfo struct {
	Id         int64
	ImportPath string `xorm:"index VARCHAR(150)"`

	ProjectName string `xorm:"VARCHAR(50)"`
	ProjectPath string `xorm:"VARCHAR(120)"`
	ViewDirPath string `xorm:"VARCHAR(120)"`
	Synopsis    string `xorm:"VARCHAR(300)"`

	/*
		- Indicates whether it's a command line tool or package.
		- Indicates whether it uses cgo / os/user or not.
		- Indicates whether it belongs to Go standard library.
		- Indicates whether it's developed by Go team.
	*/
	IsCmd, IsCgo          bool
	IsGoRepo, IsGoSubrepo bool

	/*
		- All tags of project.
		eg.
			master|||v0.6.2.0718
		- Views of projects.
		eg.
			1342
		- User viewed time(Unix-timestamp).
		eg.
			1374127619
		- Time when information last updated(UTC).
		eg.
			2013-07-16 21:09:27.48932087
	*/
	Tags       string `xorm:"-"`
	Views      int64  `xorm:"index"`
	ViewedTime int64
	Created    int64 `xorm:"index"`

	/*
		- Rank is the benchmark of projects, it's based on BaseRank and views.
		eg.
			826
	*/
	Rank int64 `xorm:"index"`

	/*
		- Package (structure) version.
		eg.
			9
		- Project revision.
		eg.
			8976ce8b2848
		- Project labels.
		eg.
			$tool|
	*/
	PkgVer int
	Ptag   string `xorm:"VARCHAR(50)"`
	Labels string `xorm:"TEXT"`

	/*
		- Number of projects that import this project.
		eg.
			11
		- Ids of projects that import this project.
		eg.
			$47|$89|$5464|$8586|$8595|$8787|$8789|$8790|$8918|$9134|$9139|
	*/
	RefNum  int
	RefPids string `xorm:"TEXT"`

	/*
		- Addtional information.
	*/
	Vcs      string `xorm:"-"`
	Homepage string `xorm:"VARCHAR(100)"`
	ForkUrl  string `xorm:"VARCHAR(150)"`

	Issues, Stars, Forks int
	Note                 string `xorm:"TEXT"`

	SourceSize int64
}

// Value represents constants and variable
type Value struct {
	Name          string // Value name.
	Doc           string
	Decl, FmtDecl string // Normal and formatted form of declaration.
	URL           string // VCS URL.
}

// Func represents functions
type Func struct {
	Name, FullName string
	Doc            string
	Decl, FmtDecl  string
	URL            string // VCS URL.
	Code           string // Included field 'Decl', formatted.
	Examples       []*Example
}

// Type represents structs and interfaces.
type Type struct {
	Name          string // Type name.
	Doc           string
	Decl, FmtDecl string // Normal and formatted form of declaration.
	URL           string // VCS URL.

	Consts, Vars []*Value
	Funcs        []*Func // Exported functions that return this type.
	Methods      []*Func // Exported methods.

	IFuncs   []*Func // Internal functions that return this type.
	IMethods []*Func // Internal methods.

	Examples []*Example
}

// Example represents function or method examples.
type Example struct {
	Name string
	Doc  string
	Code string
	//Play   string
	Output string
	IsUsed bool // Indicates if it's used by any kind object.
}

// Gist represents a Gist.
type Gist struct {
	ImportPath string     `xorm:"index VARCHAR(150)"`
	Gist       string     // Gist path.
	Examples   []*Example // Examples.
}

// PkgDecl is package declaration in database acceptable form.
type PkgDecl struct {
	Tag string // Current tag of project.
	Doc string // Package documentation(doc.go).

	File

	Examples, UserExamples []*Example // Function or method example.
	Imports, TestImports   []string   // Imports.
	Files, TestFiles       []*Source  // Source files.

	Notes []string // Source code notes.
	Dirs  []string // Subdirectories

	// Indicate how many JS should be downloaded(JsNum=total num - 1)
	JsNum int
}

// PACKAGE_VER is modified when previously stored packages are invalid.
const PACKAGE_VER = 1

// A Package describles the full documentation and declaration of a project or package.
type Package struct {
	*PkgInfo

	Readme map[string][]byte

	*PkgDecl

	IsNeedRender bool

	IsHasExport bool

	// Top-level declarations.
	IsHasConst, IsHasVar bool

	IsHasExample bool

	IsHasFile   bool
	IsHasSubdir bool
}

// A Source describles a Source code file.
type Source struct {
	SrcName   string
	BrowseUrl string
	RawSrcUrl string
	SrcData   []byte
}

func (s *Source) Name() string       { return s.SrcName }
func (s *Source) Size() int64        { return int64(len(s.RawSrcUrl)) }
func (s *Source) Mode() os.FileMode  { return 0 }
func (s *Source) ModTime() time.Time { return time.Time{} }
func (s *Source) IsDir() bool        { return false }
func (s *Source) Sys() interface{}   { return nil }
func (s *Source) RawUrl() string     { return s.RawSrcUrl }
func (s *Source) Data() []byte       { return s.SrcData }
func (s *Source) SetData(p []byte)   { s.SrcData = p }

// Walker holds the state used when building the documentation.
type Walker struct {
	LineFmt  string
	Pdoc     *Package
	apkg     *ast.Package
	Examples []*doc.Example // Function or method example.
	Fset     *token.FileSet
	SrcLines map[string][]string // Source file line slices.
	SrcFiles map[string]*Source
	Buf      []byte // scratch space for printNode method.
}

// A File describles declaration of file.
type File struct {
	// Top-level declarations.
	Consts []*Value
	Funcs  []*Func
	Types  []*Type
	Vars   []*Value

	// Internal declarations.
	Ifuncs []*Func
	Itypes []*Type
}
