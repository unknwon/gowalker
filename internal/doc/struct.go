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

package doc

import (
	"go/ast"
	"go/doc"
	"go/token"
	"os"
	"time"

	"github.com/unknwon/gowalker/internal/db"
)

// Source represents a Source code file.
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

// Example represents function or method examples.
type Example struct {
	Name string
	Doc  string
	Code string
	//Play   string
	Output string
	IsUsed bool // Indicates if it's used by any kind object.
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

// PkgDecl is package declaration in database acceptable form.
type PkgDecl struct {
	Tag string // Current tag of project.
	Doc string // Package documentation(doc.go).

	File

	Examples             []*Example // Function or method example.
	Imports, TestImports []string   // Imports.
	Files, TestFiles     []*Source  // Source files.

	Notes []string // Source code notes.
	Dirs  []string // Subdirectories
}

// Package represents the full documentation and declaration of a project or package.
type Package struct {
	*db.PkgInfo

	Readme map[string][]byte

	*PkgDecl

	IsHasExport bool

	// Top-level declarations.
	IsHasConst, IsHasVar bool

	IsHasExample bool

	IsHasFile   bool
	IsHasSubdir bool
}

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
