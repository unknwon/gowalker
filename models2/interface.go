// Copyright 2013 Gary Burd
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

// Method signature bugs:
// - Embedded interfaces are not expanded.
// - Inline interface methods are not sorted to a canonical order.
// - Array size expressions are not evaluated.
// - Unnecessary use of () are not removed.

package models

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go/ast"
	"sort"
	"strconv"
)

type aborted struct{ err error }

func abort(err error) { panic(aborted{err}) }

func handleAbort(err *error) {
	if r := recover(); r != nil {
		if a, ok := r.(aborted); ok {
			*err = a.err
		} else {
			panic(r)
		}
	}
}

const (
	exportedMask = 1 << 7
	embeddedMask = 1 << 6
	allMask      = exportedMask | embeddedMask
)

// MethodSignature represents the name, parameter types and result types of a
// method.
type MethodSignature [16]byte

func (ms MethodSignature) String() string            { return hex.EncodeToString(ms[:]) }
func (ms MethodSignature) IsEmbeddedInterface() bool { return ms[0]&embeddedMask != 0 }
func (ms MethodSignature) IsExported() bool          { return ms[0]&exportedMask != 0 }

type byMethodSignature []MethodSignature

func (p byMethodSignature) Len() int           { return len(p) }
func (p byMethodSignature) Less(i, j int) bool { return -1 == bytes.Compare(p[i][:], p[j][:]) }
func (p byMethodSignature) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// methodWriter writes a canonical representation of a method to its buffer.
type methodWriter struct {
	buf  []byte
	path string
}

func (w *methodWriter) nodePath(n ast.Node) string {
	if n, _ := n.(*ast.Ident); n != nil {
		if obj := n.Obj; obj != nil && obj.Kind == ast.Pkg {
			if spec, _ := obj.Decl.(*ast.ImportSpec); spec != nil {
				return spec.Path.Value
			}
		}
	}
	return "UNKNOWN"
}

func (w *methodWriter) writeFunc(name string, n *ast.FuncType) {
	w.buf = append(w.buf, name...)
	w.writeParams(n.Params, true)
	w.writeParams(n.Results, n.Results != nil && n.Results.NumFields() > 1)
}

func (w *methodWriter) writeParams(list *ast.FieldList, paren bool) {
	var sep bool
	if paren {
		w.buf = append(w.buf, '(')
	}
	if list != nil {
		for _, field := range list.List {
			m := len(field.Names)
			if m == 0 {
				m = 1
			}
			for i := 0; i < m; i++ {
				if sep {
					w.buf = append(w.buf, ',')
				} else {
					sep = true
				}
				w.writeNode(field.Type)
			}
		}
	}
	if paren {
		w.buf = append(w.buf, ')')
	}
}

var anonymousNames = []*ast.Ident{nil}

func (w *methodWriter) writeStruct(s *ast.StructType) {
	w.buf = append(w.buf, "struct{"...)
	var sep bool
	if s.Fields != nil {
		for _, field := range s.Fields.List {
			names := field.Names
			if len(names) == 0 {
				names = anonymousNames
			}
			for _, name := range names {
				if sep {
					w.buf = append(w.buf, ';')
				} else {
					sep = true
				}
				if name != nil {
					w.buf = append(w.buf, name.Name...)
					w.buf = append(w.buf, ' ')
				}
				w.writeNode(field.Type)
				if field.Tag != nil {
					tag, err := strconv.Unquote(field.Tag.Value)
					if err != nil {
						abort(err)
					}
					w.buf = append(w.buf, ' ')
					w.buf = append(w.buf, strconv.Quote(tag)...)
				}
			}
		}
	}
	w.buf = append(w.buf, '}')
}

func (w *methodWriter) writeInterface(s *ast.InterfaceType) {
	w.buf = append(w.buf, "interface{"...)
	var sep bool
	if s.Methods != nil {
		for _, field := range s.Methods.List {
			names := field.Names
			if len(names) == 0 {
				names = anonymousNames
			}
			for _, name := range names {
				if sep {
					w.buf = append(w.buf, ';')
				} else {
					sep = true
				}
				switch n := field.Type.(type) {
				case *ast.Ident:
					w.writeNode(n)
				case *ast.SelectorExpr:
					w.writeNode(n)
				case *ast.FuncType:
					w.writeFunc(name.Name, field.Type.(*ast.FuncType))
				default:
					abort(fmt.Errorf("Unexpected %T in InterfaceType", n))
				}
			}
		}
	}
	w.buf = append(w.buf, '}')
}

func (w *methodWriter) writeNode(n ast.Node) {
	switch n := n.(type) {
	case *ast.Ellipsis:
		w.buf = append(w.buf, "..."...)
		w.writeNode(n.Elt)
	case *ast.MapType:
		w.buf = append(w.buf, "map["...)
		w.writeNode(n.Key)
		w.buf = append(w.buf, ']')
		w.writeNode(n.Value)
	case *ast.ArrayType:
		w.buf = append(w.buf, '[')
		if n.Len != nil {
			w.writeNode(n.Len)
		}
		w.buf = append(w.buf, ']')
		w.writeNode(n.Elt)
	case *ast.ChanType:
		if n.Dir == ast.RECV {
			w.buf = append(w.buf, "<-"...)
		}
		w.buf = append(w.buf, "chan"...)
		if n.Dir == ast.SEND {
			w.buf = append(w.buf, "<-"...)
		}
		w.buf = append(w.buf, ' ')
		w.writeNode(n.Value)
	case *ast.ParenExpr:
		w.buf = append(w.buf, '(')
		w.writeNode(n.X)
		w.buf = append(w.buf, ')')
	case *ast.BinaryExpr:
		w.writeNode(n.X)
		w.buf = append(w.buf, n.Op.String()...)
		w.writeNode(n.Y)
	case *ast.BasicLit:
		w.buf = append(w.buf, n.Value...)
	case *ast.StarExpr:
		w.buf = append(w.buf, '*')
		w.writeNode(n.X)
	case *ast.FuncDecl:
		w.writeFunc(n.Name.Name, n.Type)
	case *ast.FuncType:
		w.writeFunc("func", n)
	case *ast.InterfaceType:
		w.writeInterface(n)
	case *ast.StructType:
		w.writeStruct(n)
	case *ast.SelectorExpr:
		w.buf = append(w.buf, w.nodePath(n.X)...)
		w.buf = append(w.buf, '.')
		w.buf = append(w.buf, n.Sel.Name...)
	case *ast.Ident:
		if n.Obj != nil || predeclared[n.Name] != predeclaredType {
			w.buf = append(w.buf, w.path...)
			w.buf = append(w.buf, '.')
		}
		w.buf = append(w.buf, n.Name...)
	default:
		abort(fmt.Errorf("Unexpected %T in method declaration", n))
	}
}

func (w *methodWriter) writeCanonicalMethodDecl(name string, n *ast.FuncType) (err error) {
	defer handleAbort(&err)
	w.buf = w.buf[:0]
	w.writeFunc(name, n)
	return err
}

func makeSignature(p []byte, exported, embedded bool) MethodSignature {
	h := md5.New()
	h.Write(p)
	var sig MethodSignature
	h.Sum(sig[:])
	sig[0] &^= allMask
	if exported {
		sig[0] |= exportedMask
	}
	if embedded {
		sig[0] |= embeddedMask
	}
	return sig
}

func (w *methodWriter) interfaceSignatures(pkg *ast.Package) (map[string][]MethodSignature, error) {
	result := make(map[string][]MethodSignature)
	for name, obj := range pkg.Scope.Objects {
		if !ast.IsExported(name) {
			continue
		}
		spec, ok := obj.Decl.(*ast.TypeSpec)
		if !ok {
			continue
		}
		itf, ok := spec.Type.(*ast.InterfaceType)
		if !ok {
			continue
		}
		var sigs []MethodSignature
		for _, field := range itf.Methods.List {
			names := field.Names
			if len(names) == 0 {
				names = anonymousNames
			}
			for _, name := range names {
				switch n := field.Type.(type) {
				case *ast.Ident:
					switch n.Name {
					case "error":
						sigs = append(sigs, makeSignature([]byte(`Error()string`), true, false))
					default:
						sigs = append(sigs, makeSignature([]byte(w.path+"."+n.Name), ast.IsExported(n.Name), true))
					}
				case *ast.SelectorExpr:
					sigs = append(sigs, makeSignature([]byte(w.nodePath(n.X)+"."+n.Sel.Name), true, true))
				case *ast.FuncType:
					if err := w.writeCanonicalMethodDecl(name.Name, n); err != nil {
						return nil, err
					}
					sigs = append(sigs, makeSignature(w.buf, ast.IsExported(name.Name), false))
				default:
					return nil, fmt.Errorf("Unexpected %T in InterfaceType", n)
				}
			}
		}
		sort.Sort(byMethodSignature(sigs))
		result[name] = sigs
	}
	return result, nil
}

/*
func unexportedMethodSignatures(m map[string][]MethodSignature) map[MethodSignature]bool {
    for _, sigs := range m {
        for _, sig := range sigs {

        if !

*/
