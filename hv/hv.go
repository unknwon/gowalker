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

// Package Hacker View provides APIs to analyze Go projects and generate AST-based source code view.
package hv

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"path"
	"strings"

	"github.com/Unknwon/com"
)

// Render renders source code to HTML.
// Note that you have to call Build before you call this method.
func (w *Walker) Render() (map[string][]byte, error) {
	htmls := make(map[string][]byte)
	files := make(map[string]File)
	srcs := w.SrcFiles
	imports := make([]string, 0, 10)

	for name, src := range srcs {
		pkg, err := w.Build(&WalkRes{
			WalkDepth: WD_All,
			WalkType:  WT_Memory,
			WalkMode:  WM_NoReadme | WM_NoExample,
			Srcs:      []*Source{src},
			BuildAll:  true,
		})

		if err != nil {
			return nil, errors.New("hv.Walker.Render -> " + err.Error())
		}

		files[name] = pkg.File

		for _, v := range w.Pdoc.Imports {
			imports = com.AppendStr(imports, v)
		}
	}

	r := Render{
		Links:     GetLinks(w.Pdoc.ImportPath, imports, files),
		FilteList: make(map[string]bool),
	}

	for name := range srcs {
		htmls[name] = r.Render(name, srcs[name].Data())
		//com.SaveFile(name+".hmtl", htmls[name])
	}
	return htmls, nil
}

// A link describes the (HTML) link information for an identifier.
// The zero value of a link represents "no link".
type Link struct {
	Path, Name, Comment string
}

type localVar struct {
	name, tp string
}

// A Render describles a code render.
type Render struct {
	recv       localVar
	blockLevel int

	Links     []*Link
	FilteList map[string]bool
}

// Render highlights code.
func (r *Render) Render(name string, data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	code := string(data)
	l := len(code)

	buf := new(bytes.Buffer)
	//buf.WriteString("<pre>")

	strTag := uint8(0)
	isComment := false
	isBlockComment := false
	isString := false
	isFuncDecl := false
	isFuncBlock := false
	isHasRecv := false
	isTypeDecl := false
	last := 0
	pos := 0

	for {
	CutWords:
		for {
			curChar := code[pos]
			if !com.IsLetter(curChar) {
				if isComment {
					// Comment.
					if isBlockComment {
						// Check if in end of block comment.
						if curChar == '/' && code[pos-1] == '*' {
							break CutWords
						}
					} else {
						// Check if in start of block comment.
						if curChar == '*' && code[pos-1] == '/' {
							isBlockComment = true
						} else if curChar == '\n' {
							break CutWords
						}
					}
				} else {
					// String.
					if curChar == '\'' || curChar == '"' || curChar == '`' {
						if !isString {
							// Set string tag.
							strTag = curChar
							isString = true
						} else {
							// Check if it is end of string or escaped character.
							if (code[pos-1] == '\\' && code[pos-2] == '\\') || code[pos-1] != '\\' {
								// Check string tag.
								if curChar == strTag {
									// Handle string highlight.
									break CutWords
								}
							}
						}
					}

					if !isString {
						switch {
						case curChar == '/' && (code[pos+1] == '/' || code[pos+1] == '*'):
							isComment = true
						case curChar > 47 && curChar < 58: // Ends with number.
						case curChar == '_' && code[pos-1] != ' ': // Underline: _.
						case (curChar != '.' || curChar == '\n'):
							break CutWords
						}
					}
				}
			}

			if pos == l-1 {
				break CutWords
			}
			pos++
		}

		seg := code[last : pos+1]
	CheckLink:
		switch {
		case isComment:
			isComment = false
			isBlockComment = false
			fmt.Fprintf(buf, `<span class="com">%s</span>`, seg)
		case isString:
			isString = false
			fmt.Fprintf(buf, `<span class="str">%s</span>`, template.HTMLEscapeString(seg))
		case seg == "\t":
			fmt.Fprintf(buf, `%s`, "    ")
		case seg == "{":
			if isFuncDecl {
				isFuncDecl = false
				isFuncBlock = true
			}

			if isFuncBlock {
				r.blockLevel++
			}
			fmt.Fprintf(buf, "%s", seg)
		case seg == "}":
			if isFuncBlock {
				r.blockLevel--
			}
			if r.blockLevel == 0 {
				isFuncBlock = false
				r.recv.name = ""
			}
			fmt.Fprintf(buf, "%s", seg)
		case isFuncDecl:
			if isHasRecv {
				if seg != "(" && seg != " " && seg != "*" {
					if len(r.recv.name) == 0 {
						r.recv.name = seg[:len(seg)-1]
					} else {
						r.recv.tp = seg[:len(seg)-1]
						isHasRecv = false
					}
				}
			} else if len(seg) > 1 && code[pos] == '(' {
				if len(r.recv.name) > 0 {
					fmt.Fprintf(buf, "<span id=\"%s_%s\">%s</span>(", r.recv.tp, seg[:len(seg)-1], seg[:len(seg)-1])
				} else {
					fmt.Fprintf(buf, "<span id=\"%s\">%s</span>(", seg[:len(seg)-1], seg[:len(seg)-1])
				}
				break CheckLink
			}
			fallthrough
		case pos-last > 1:
			// Check if the last word of the paragraphy.
			l := len(seg)
			keyword := seg
			if !com.IsLetter(seg[l-1]) {
				keyword = seg[:l-1]
			} else {
				l++
			}

			// Check keywords.
			switch keyword {
			case "return", "break":
				fmt.Fprintf(buf, `<span class="ret">%s</span>%s`, keyword, seg[l-1:])
				break CheckLink
			case "func":
				isFuncDecl = true
				if code[pos+1] == '(' {
					isHasRecv = true
				}
				fallthrough
			case "package", "import", "range", "for", "if", "else", "type", "struct", "select", "case", "var", "const", "switch", "default", "continue":
				if keyword == "type" {
					isTypeDecl = true
				}
				fmt.Fprintf(buf, `<span class="key">%s</span>%s`, keyword, seg[l-1:])
				break CheckLink
			case "new", "append", "make", "panic", "recover", "len", "cap", "copy", "close", "delete", "defer":
				fmt.Fprintf(buf, `<span class="bui">%s</span>%s`, keyword, seg[l-1:])
				break CheckLink
			}

			if isPredeclared(keyword) {
				fmt.Fprintf(buf, `<span class="boo">%s</span>%s`, keyword, seg[l-1:])
				break CheckLink
			}

			// Check links.
			link, ok := r.findType(seg)
			if ok {
				switch {
				case strings.HasSuffix(link.Path, name) && len(link.Name) > 0: // Current file.
					fmt.Fprintf(buf, `<a class="int" title="%s" href="#%s">%s</a>%s`,
						link.Comment, link.Name, link.Name, seg[l-1:])
				case len(link.Path) > 0 && len(link.Name) > 0:
					if strings.HasPrefix(link.Path, "#") {
						fmt.Fprintf(buf, `<a class="ext" title="%s" href="%s">%s</a>%s`,
							link.Comment, link.Path, link.Name, seg[l-1:])
					} else {
						if strings.Index(link.Path, "#") > -1 {
							fmt.Fprintf(buf, `<a class="ext" title="%s" target="_blank" href="%s">%s</a>%s`,
								link.Comment, link.Path, link.Name, seg[l-1:])
						} else {
							fmt.Fprintf(buf, `<a class="ext" title="%s" target="_blank" href="/%s#%s">%s</a>%s`,
								link.Comment, link.Path, link.Name, link.Name, seg[l-1:])
						}
					}
				}
			} else if seg[len(seg)-1] == ' ' || seg[len(seg)-1] == '\n' {
				if isFuncDecl || isTypeDecl {
					isTypeDecl = false
					fmt.Fprintf(buf, "<span id=\"%s\">%s</span>%s", seg[:len(seg)-1], seg[:len(seg)-1], seg[l-1:])
				} else {
					fmt.Fprintf(buf, "%s", seg)
				}
			} else {
				fmt.Fprintf(buf, "%s", seg)
			}
		default:
			fmt.Fprintf(buf, "%s", seg)
		}

		last = pos + 1
		pos++
		// End of code.
		if pos == l {
			fmt.Fprintf(buf, "%s", code[last:])
			break
		}
	}

	return buf.Bytes()
}

func (r *Render) findType(name string) (*Link, bool) {
	if !com.IsLetter(name[0]) {
		return nil, false
	}

	// We cannot deal with struct field now.
	if name[len(name)-1] == '[' || name[len(name)-1] == ']' ||
		name[len(name)-1] == ' ' || name[len(name)-1] == ':' {
		return nil, false
	}

	// We cannot deal with chain operation.
	if name[0] == '.' {
		return nil, false
	}

	name = name[:len(name)-1]

	// This is for functions and types from imported packages.
	i := strings.Index(name, ".")
	// We cannot deal with struct field or chain operation now.
	if i != strings.LastIndex(name, ".") {
		return nil, false
	}

	if filte := r.FilteList[name]; filte {
		return nil, false
	}

	var left, right string
	if i > -1 {
		left = name[:i+1]
		right = name[i+1:]
	}

	for _, l := range r.Links {
		if i == -1 {
			// Exported types and functions in current package.
			if l.Name == name {
				return l, true
			}
		} else {
			// Functions and types from imported packages.
			if l.Name == left {
				if len(l.Path) > 0 {
					return &Link{Name: name, Path: "/" + l.Path + "#" + right}, true
				} else {
					return &Link{Name: name, Path: "#" + right}, true
				}
			} else if r.recv.name == left[:len(left)-1] {
				// fmt.Println(r.recv.tp + "." + right)
				// fmt.Println(r.findType(r.recv.tp + "." + right))
				return nil, false
			}
		}
	}

	r.FilteList[name] = true
	return nil, false
}

// GetLinks returns objects with its jump link.
func GetLinks(prefix string, imports []string, files map[string]File) []*Link {
	prefix += "?f="
	links := make([]*Link, 0, len(imports)+len(files)*20+10)
	for fname, f := range files {
		// Get all types, functions and import packages.
		for _, t := range f.Types {
			links = append(links, &Link{
				Name:    t.Name,
				Path:    prefix + fname,
				Comment: template.HTMLEscapeString(t.Doc),
			})

			for _, f := range t.Funcs {
				links = append(links, &Link{
					Name:    f.Name,
					Path:    prefix + fname,
					Comment: template.HTMLEscapeString(f.Doc),
				})
			}

			for _, f := range t.IFuncs {
				links = append(links, &Link{
					Name:    f.Name,
					Path:    prefix + fname,
					Comment: template.HTMLEscapeString(f.Doc),
				})
			}
		}

		for _, t := range f.Itypes {
			links = append(links, &Link{
				Name:    t.Name,
				Path:    prefix + fname,
				Comment: template.HTMLEscapeString(t.Doc),
			})

			for _, f := range t.Funcs {
				links = append(links, &Link{
					Name:    f.Name,
					Path:    prefix + fname,
					Comment: template.HTMLEscapeString(f.Doc),
				})
			}

			for _, f := range t.IFuncs {
				links = append(links, &Link{
					Name:    f.Name,
					Path:    prefix + fname,
					Comment: template.HTMLEscapeString(f.Doc),
				})
			}
		}

		for _, f := range f.Funcs {
			links = append(links, &Link{
				Name:    f.Name,
				Path:    prefix + fname,
				Comment: template.HTMLEscapeString(f.Doc),
			})
		}

		for _, f := range f.Ifuncs {
			links = append(links, &Link{
				Name:    f.Name,
				Path:    prefix + fname,
				Comment: template.HTMLEscapeString(f.Doc),
			})
		}
	}

	for _, v := range imports {
		if v != "C" {
			links = append(links, &Link{
				Name: path.Base(v) + ".",
				Path: v,
			})
		}
	}
	return links
}

func isPredeclared(name string) bool {
	_, ok := predeclared[name]
	return ok
}
