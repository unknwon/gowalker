// Copyright 2011 Gary Burd
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

package doc

import (
	"bytes"
	"encoding/base32"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

func (w *walker) readDir(dir string) ([]os.FileInfo, error) {
	if dir != w.pdoc.ImportPath {
		panic("unexpected")
	}
	fis := make([]os.FileInfo, 0, len(w.srcs))
	for _, src := range w.srcs {
		fis = append(fis, src)
	}
	return fis, nil
}

func (w *walker) openFile(path string) (io.ReadCloser, error) {
	if strings.HasPrefix(path, w.pdoc.ImportPath+"/") {
		if src, ok := w.srcs[path[len(w.pdoc.ImportPath)+1:]]; ok {
			return ioutil.NopCloser(bytes.NewReader(src.data)), nil
		}
	}
	panic("unexpected")
}

func simpleImporter(imports map[string]*ast.Object, path string) (*ast.Object, error) {
	pkg := imports[path]
	if pkg == nil {
		// Guess the package name without importing it. Start with the last
		// element of the path.
		name := path[strings.LastIndex(path, "/")+1:]

		// Trim commonly used prefixes and suffixes containing illegal name
		// runes.
		name = strings.TrimSuffix(name, ".go")
		name = strings.TrimSuffix(name, "-go")
		name = strings.TrimPrefix(name, "go.")
		name = strings.TrimPrefix(name, "go-")
		name = strings.TrimPrefix(name, "biogo.")

		// It's also common for the last element of the path to contain an
		// extra "go" prefix, but not always. TODO: examine unresolved ids to
		// detect when trimming the "go" prefix is appropriate.

		pkg = ast.NewObj(ast.Pkg, name)
		pkg.Data = ast.NewScope(nil)
		imports[path] = pkg
	}
	return pkg, nil
}

func (b *walker) printNode(node interface{}) string {
	b.buf = b.buf[:0]
	err := (&printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}).Fprint(sliceWriter{&b.buf}, b.fset, node)
	if err != nil {
		return err.Error()
	}
	return string(b.buf)
}

func (w *walker) printDecl(decl ast.Node) string {
	var d Code
	d, w.buf = printDecl(decl, w.fset, w.buf)
	return d.Text
}

func (w *walker) printPos(pos token.Pos) string {
	position := w.fset.Position(pos)
	src := w.srcs[position.Filename]
	if src == nil || src.browseURL == "" {
		// src can be nil when line comments are used (//line <file>:<line>).
		return ""
	}
	return src.browseURL + fmt.Sprintf(w.lineFmt, position.Line)
}

func (w *walker) printCode(decl ast.Node) string {
	pos := decl.Pos()
	posPos := w.fset.Position(pos)
	src := w.srcs[posPos.Filename]
	if src == nil || src.browseURL == "" {
		// src can be nil when line comments are used (//line <file>:<line>).
		return ""
	}

	var code []string
	code, ok := w.srcLines[posPos.Filename]
	// Check source file line arrays.
	if !ok {
		w.srcLines[posPos.Filename] = strings.Split(string(src.data), "\n")
		code = w.srcLines[posPos.Filename]
	}

	// Get code.
	var buf bytes.Buffer
	l := len(code)
CutCode:
	for i := posPos.Line; i < l; i++ {
		// Check end of code block.
		switch {
		case len(code[i]) > 0 && code[i][0] == '}': // Normal end.
			break CutCode
		case len(code[i-1]) > 4 && code[i-1][:4] == "func" &&
			code[i-1][len(code[i-1])-1] == '}': // One line functions.
			line := code[i-1]
			buf.WriteString("       ")
			buf.WriteString(line[strings.Index(line, "{")+1 : len(line)-1])
			buf.WriteByte('\n')
			break CutCode
		}

		buf.WriteString(code[i])
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (w *walker) values(vdocs []*doc.Value) []*Value {
	var result []*Value
	for _, d := range vdocs {
		result = append(result, &Value{
			Decl: w.printDecl(d.Decl),
			URL:  w.printPos(d.Decl.Pos()),
			Doc:  d.Doc,
		})
	}
	return result
}

func (w *walker) funcs(fdocs []*doc.Func) []*Func {
	var result []*Func
	for _, d := range fdocs {
		/*	var exampleName string
			switch {
			case d.Recv == "":
				exampleName = d.Name
			case d.Recv[0] == '*':
				exampleName = d.Recv[1:] + "_" + d.Name
			default:
				exampleName = d.Recv + "_" + d.Name
			}*/
		result = append(result, &Func{
			Decl: w.printDecl(d.Decl),
			URL:  w.printPos(d.Decl.Pos()),
			Doc:  d.Doc,
			Name: d.Name,
			Code: w.printCode(d.Decl),
			//Recv: d.Recv,
			//Examples: w.getExamples(exampleName),
		})
	}
	return result
}

func (w *walker) types(tdocs []*doc.Type) []*Type {
	var result []*Type
	for _, d := range tdocs {
		result = append(result, &Type{
			Doc:  d.Doc,
			Name: d.Name,
			Decl: w.printDecl(d.Decl),
			URL:  w.printPos(d.Decl.Pos()),
			//Consts:  w.values(d.Consts),
			//Vars:    w.values(d.Vars),
			Funcs:   w.funcs(d.Funcs),
			Methods: w.funcs(d.Methods),
			//Examples: w.getExamples(d.Name),
		})
	}
	return result
}

var exampleOutputRx = regexp.MustCompile(`(?i)//[[:space:]]*output:`)

func (w *walker) getExamples(name string) []*Example {
	var docs []*Example
	for _, e := range w.examples {
		// if !strings.HasPrefix(e.Name, name) {
		// 	continue
		// }

		output := e.Output
		code := w.printNode(&printer.CommentedNode{
			Node:     e.Code,
			Comments: e.Comments,
		})

		// additional formatting if this is a function body
		if i := len(code); i >= 2 && code[0] == '{' && code[i-1] == '}' {
			// remove surrounding braces
			code = code[1 : i-1]
			// unindent
			code = strings.Replace(code, "\n    ", "\n", -1)
			// remove output comment
			if j := exampleOutputRx.FindStringIndex(code); j != nil {
				code = strings.TrimSpace(code[:j[0]])
			}
		} else {
			// drop output, as the output comment will appear in the code
			output = ""
		}

		// play := ""
		// if e.Play != nil {
		// 	w.buf = w.buf[:0]
		// 	if err := format.Node(sliceWriter{&w.buf}, w.fset, e.Play); err != nil {
		// 		play = err.Error()
		// 	} else {
		// 		play = string(w.buf)
		// 	}
		// }

		docs = append(docs, &Example{
			Name:   e.Name,
			Doc:    e.Doc,
			Code:   code,
			Output: output,
		})
		//Play:   play
	}
	return docs
}

// build generates data from source files.
func (w *walker) build(srcs []*source) (*Package, error) {
	// Set created time.
	w.pdoc.Created = time.Now().UTC()

	// Add source files to walker, I skipped references here.
	w.srcs = make(map[string]*source)
	for _, src := range srcs {
		srcName := strings.ToLower(src.name) // For readme comparation.
		switch {
		case strings.HasSuffix(src.name, ".go"):
			w.srcs[src.name] = src
		case len(w.pdoc.Tag) > 0:
			continue // Only save latest readme.
		case strings.HasPrefix(srcName, "readme_zh") || strings.HasPrefix(srcName, "readme_cn"):
			models.SavePkgDoc(w.pdoc.ImportPath, "zh", src.data)
		case strings.HasPrefix(srcName, "readme"):
			models.SavePkgDoc(w.pdoc.ImportPath, "en", src.data)
		}
	}

	w.fset = token.NewFileSet()

	// Find the package and associated files.
	ctxt := build.Context{
		GOOS:          runtime.GOOS,
		GOARCH:        runtime.GOARCH,
		CgoEnabled:    true,
		JoinPath:      path.Join,
		IsAbsPath:     path.IsAbs,
		SplitPathList: func(list string) []string { return strings.Split(list, ":") },
		IsDir:         func(path string) bool { panic("unexpected") },
		HasSubdir:     func(root, dir string) (rel string, ok bool) { panic("unexpected") },
		ReadDir:       func(dir string) (fi []os.FileInfo, err error) { return w.readDir(dir) },
		OpenFile:      func(path string) (r io.ReadCloser, err error) { return w.openFile(path) },
		Compiler:      "gc",
	}

	bpkg, err := ctxt.ImportDir(w.pdoc.ImportPath, 0)
	// Continue if there are no Go source files; we still want the directory info.
	_, nogo := err.(*build.NoGoError)
	if err != nil {
		if nogo {
			err = nil
			beego.Info("doc.walker.build -> No Go Source file")
		} else {
			return w.pdoc, errors.New("doc.walker.build -> " + err.Error())
		}
	}

	// Parse the Go files
	files := make(map[string]*ast.File)
	for _, name := range append(bpkg.GoFiles, bpkg.CgoFiles...) {
		file, err := parser.ParseFile(w.fset, name, w.srcs[name].data, parser.ParseComments)
		if err != nil {
			beego.Error("doc.walker.build -> parse go files:", err)
			continue
		}
		w.pdoc.Files = append(w.pdoc.Files, name)
		//w.pdoc.SourceSize += len(w.srcs[name].data)
		files[name] = file
	}

	apkg, _ := ast.NewPackage(w.fset, files, simpleImporter, nil)

	// Find examples in the test files.
	for _, name := range append(bpkg.TestGoFiles, bpkg.XTestGoFiles...) {
		file, err := parser.ParseFile(w.fset, name, w.srcs[name].data, parser.ParseComments)
		if err != nil {
			beego.Error("doc.walker.build -> find examples:", err)
			continue
		}
		//w.pdoc.TestFiles = append(w.pdoc.TestFiles, &File{Name: name, URL: w.srcs[name].browseURL})
		//w.pdoc.TestSourceSize += len(w.srcs[name].data)
		w.examples = append(w.examples, doc.Examples(file)...)
	}

	//w.vetPackage(apkg)

	mode := doc.Mode(0)
	if w.pdoc.ImportPath == "builtin" {
		mode |= doc.AllDecls
	}

	pdoc := doc.New(apkg, w.pdoc.ImportPath, mode)

	w.pdoc.Synopsis = utils.Synopsis(pdoc.Doc)
	pdoc.Doc = strings.TrimRight(pdoc.Doc, " \t\n\r")
	var buf bytes.Buffer
	doc.ToHTML(&buf, pdoc.Doc, nil)
	w.pdoc.Doc = w.pdoc.Doc + "<br />" + buf.String()
	w.pdoc.Doc = strings.Replace(w.pdoc.Doc, "<p>", "<p><b>", 1)
	w.pdoc.Doc = strings.Replace(w.pdoc.Doc, "</p>", "</b></p>", 1)
	w.pdoc.Doc = base32.StdEncoding.EncodeToString([]byte(w.pdoc.Doc))

	w.pdoc.Examples = w.getExamples("")
	w.pdoc.IsCmd = bpkg.IsCommand()
	w.srcLines = make(map[string][]string)
	w.pdoc.Consts = w.values(pdoc.Consts)
	w.pdoc.Funcs = w.funcs(pdoc.Funcs)
	w.pdoc.Types = w.types(pdoc.Types)
	w.pdoc.Vars = w.values(pdoc.Vars)
	//w.pdoc.Notes = w.notes(pdoc.Notes)

	w.pdoc.Imports = bpkg.Imports
	w.pdoc.TestImports = bpkg.TestImports
	//w.pdoc.XTestImports = bpkg.XTestImports

	beego.Info("doc.walker.build(", pdoc.ImportPath, "), Goroutine #", runtime.NumGoroutine())
	return w.pdoc, err
}
