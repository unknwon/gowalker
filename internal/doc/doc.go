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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"go/doc"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/unknwon/com"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"
	// "github.com/davecgh/go-spew/spew"

	"github.com/unknwon/gowalker/internal/db"
	"github.com/unknwon/gowalker/internal/setting"
)

var (
	ErrFetchTimeout = errors.New("Fetch package timeout")
)

// A link describes the (HTML) link information for an identifier.
// The zero value of a link represents "no link".
//
type Link struct {
	Path, Name, Comment string // package path, identifier name, and comments.
}

// FormatCode highlights keywords and adds HTML links to them.
func FormatCode(w io.Writer, code *string, links []*Link) {
	length := len(*code) // Length of whole code.
	if length == 0 {
		return
	}

	*code = strings.Replace(*code, "&#34;", `"`, -1)
	*code = strings.Replace(*code, "&#39;", `'`, -1)
	length = len(*code)

	strTag := uint8(0)      // Indicates what kind of string is chekcing.
	isString := false       // Indicates if right now is checking string.
	isComment := false      // Indicates if right now is checking comments.
	isBlockComment := false // Indicates if right now is checking block comments.
	last := 0               // Start index of the word.
	pos := 0                // Current index.

	for {
		// Cut words.
	CutWords:
		for {
			curChar := (*code)[pos] // Current check character.
			if !com.IsLetter(curChar) {
				if !isComment {
					switch {
					case curChar == '\'' || curChar == '"' || curChar == '`': // String.
						if !isString {
							// Set string tag.
							strTag = curChar
							isString = true
						} else {
							// CHeck if it is end of string or escaped character.
							if ((*code)[pos-1] == '\\' && (*code)[pos-2] == '\\') || (*code)[pos-1] != '\\' {
								// Check string tag.
								if curChar == strTag {
									// Handle string highlight.
									break CutWords
								}
							}
						}
					case !isString && curChar == '/' && ((*code)[pos+1] == '/' || (*code)[pos+1] == '*'):
						isComment = true
					case !isString && curChar > 47 && curChar < 58: // Ends with number.
					case !isString && curChar == '_' && pos > 0 && (*code)[pos-1] != ' ': // Underline: _
					case !isString && (curChar != '.' || curChar == '\n'):
						break CutWords
					}
				} else {
					if isBlockComment {
						// End of block comments.
						if curChar == '/' && (*code)[pos-1] == '*' {
							break CutWords
						}
					} else {
						switch {
						case curChar == '*' && (*code)[pos-1] == '/':
							// Start of block comments.
							isBlockComment = true
						case curChar == '\n':
							break CutWords
						}
					}
				}
			}

			if pos == length-1 {
				break CutWords
			}
			pos++
		}

		seg := (*code)[last : pos+1]
	CheckLink:
		switch {
		case isComment:
			isComment = false
			isBlockComment = false
			fmt.Fprintf(w, `<span class="com">%s</span>`, seg)
		case isString:
			isString = false
			fmt.Fprintf(w, `<span class="str">%s</span>`, template.HTMLEscapeString(seg))
		case seg == "\t":
			fmt.Fprintf(w, `%s`, "    ")
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
				fmt.Fprintf(w, `<span class="ret">%s</span>%s`, keyword, seg[l-1:])
				break CheckLink
			case "func", "range", "for", "if", "else", "type", "struct", "select", "case", "var", "const", "switch", "default", "continue":
				fmt.Fprintf(w, `<span class="key">%s</span>%s`, keyword, seg[l-1:])
				break CheckLink
			case "true", "false", "nil":
				fmt.Fprintf(w, `<span class="boo">%s</span>%s`, keyword, seg[l-1:])
				break CheckLink
			case "new", "append", "make", "panic", "recover", "len", "cap", "copy", "close", "delete", "defer":
				fmt.Fprintf(w, `<span class="bui">%s</span>%s`, keyword, seg[l-1:])
				break CheckLink
			}

			// Check links.
			link, ok := findType(seg[:l-1], links)
			if ok {
				switch {
				case len(link.Path) == 0 && len(link.Name) > 0:
					// Exported types in current package.
					fmt.Fprintf(w, `<a class="int" title="%s" href="#%s">%s</a>%s`,
						link.Comment, link.Name, link.Name, seg[l-1:])
				case len(link.Path) > 0 && len(link.Name) > 0:
					if strings.HasPrefix(link.Path, "#") {
						fmt.Fprintf(w, `<a class="ext" title="%s" href="%s">%s</a>%s`,
							link.Comment, link.Path, link.Name, seg[l-1:])
					} else {
						fmt.Fprintf(w, `<a class="ext" title="%s" target="_blank" href="%s">%s</a>%s`,
							link.Comment, link.Path, link.Name, seg[l-1:])
					}
				}
			} else if seg[len(seg)-1] == ' ' {
				fmt.Fprintf(w, "<span id=\"%s\">%s</span> ", seg[:len(seg)-1], seg[:len(seg)-1])
			} else {
				fmt.Fprintf(w, "%s", seg)
			}
		default:
			fmt.Fprintf(w, "%s", seg)
		}

		last = pos + 1
		pos++
		// End of code.
		if pos == length {
			fmt.Fprintf(w, "%s", (*code)[last:])
			return
		}
	}
}

func findType(name string, links []*Link) (*Link, bool) {
	// This is for functions and types from imported packages.
	i := strings.Index(name, ".")
	var left, right string
	if i > -1 {
		left = name[:i+1]
		right = name[i+1:]
	}

	for _, l := range links {
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
			}
		}
	}
	return nil, false
}

// getLinks returns exported objects with its jump link.
func getLinks(pdoc *Package) []*Link {
	links := make([]*Link, 0, len(pdoc.Types)+len(pdoc.Imports)+len(pdoc.Funcs)+10)
	// Get all types, functions and import packages
	for _, t := range pdoc.Types {
		links = append(links, &Link{
			Name:    t.Name,
			Comment: template.HTMLEscapeString(t.Doc),
		})
	}

	for _, f := range pdoc.Funcs {
		links = append(links, &Link{
			Name:    f.Name,
			Comment: template.HTMLEscapeString(f.Doc),
		})
	}

	for _, t := range pdoc.Types {
		for _, f := range t.Funcs {
			links = append(links, &Link{
				Name:    f.Name,
				Comment: template.HTMLEscapeString(f.Doc),
			})
		}
	}

	for _, v := range pdoc.Imports {
		if v != "C" {
			links = append(links, &Link{
				Name: path.Base(v) + ".",
				Path: v,
			})
		}
	}
	return links
}

func addFunc(f *Func, path, name string, links []*Link) {
	var buf bytes.Buffer
	f.FullName = name
	f.Code = strings.Replace(f.Code, "<pre>", "&lt;pre&gt;", -1) + "}"
	FormatCode(&buf, &f.Code, links)
	f.Code = buf.String()
}

// NOTE: it can be only use for pure functions(not belong to any type), not methods.
func addFuncs(fs []*Func, path string, links []*Link) {
	for _, f := range fs {
		addFunc(f, path, f.Name, links)
	}
}

func renderFuncs(pdoc *Package) {
	links := getLinks(pdoc)

	// Functions.
	addFuncs(pdoc.Funcs, pdoc.ImportPath, links)
	addFuncs(pdoc.Ifuncs, pdoc.ImportPath, links)

	// Types.
	for _, v := range pdoc.Types {
		// Functions.
		for _, m := range v.Funcs {
			addFunc(m, pdoc.ImportPath, m.Name, links)
		}

		// Methods.
		for _, m := range v.Methods {
			addFunc(m, pdoc.ImportPath, v.Name+"_"+m.Name, links)
		}
	}
}

// getExamples returns index of function example if it exists.
func getExamples(pdoc *Package, typeName, name string) (exams []*Example) {
	matchName := name
	if len(typeName) > 0 {
		matchName = typeName + "_" + name
	}

	for i, v := range pdoc.Examples {
		// Already used or doesn't match.
		if v.IsUsed || !strings.HasPrefix(v.Name, matchName) {
			continue
		}

		// Check if it has right prefix.
		index := strings.Index(v.Name, "_")
		// Not found "_", name length shoule be equal.
		if index == -1 && (len(v.Name) != len(name)) {
			continue
		}

		// Found "_", prefix length shoule be equal.
		if index > -1 && len(typeName) == 0 && (index > len(name)) {
			continue
		}

		pdoc.Examples[i].IsUsed = true
		exams = append(exams, v)
	}
	return exams
}

// SaveDocPage saves doc. content to JS file(s),
// it returns max index of JS file(s);
// it returns -1 when error occurs.
func SaveDocPage(docPath string, data []byte) int {
	data = com.Html2JS(data)
	docPath = setting.DocsJSPath + docPath

	buf := new(bytes.Buffer)
	count := 0
	d := string(data)
	l := len(d)
	if l < 80000 {
		buf.WriteString("document.write(\"")
		buf.Write(data)
		buf.WriteString("\")")

		os.MkdirAll(path.Dir(docPath+".js"), os.ModePerm)
		if err := ioutil.WriteFile(docPath+".js", buf.Bytes(), 0655); err != nil {
			log.Error(2, "SaveDocPage %q: %v", docPath, err)
			return -1
		}
	} else {
		// Too large, need to sperate.
		start := 0
		end := start + 40000
		for {
			if end >= l {
				end = l
			} else {
				// Need to break in space.
				for {
					if d[end-3:end] == "/b>" {
						break
					}
					end += 1

					if end >= l {
						break
					}
				}
			}

			buf.WriteString("document.write(\"")
			buf.WriteString(d[start:end])
			buf.WriteString("\")\n")

			p := docPath
			if count != 0 {
				p += fmt.Sprintf("-%d", count)
			}

			os.MkdirAll(path.Dir(p+".js"), os.ModePerm)
			if err := ioutil.WriteFile(p+".js", buf.Bytes(), 0655); err != nil {
				log.Error(2, "SaveDocPage %q: %v", p, err)
				return -1
			}

			if end >= l {
				break
			}

			buf.Reset()
			start = end
			end += 204800
			count++
		}
	}

	return count
}

// SavePkgDoc saves readered readme.md file data.
func SavePkgDoc(docPath string, readmes map[string][]byte) {
	for lang, data := range readmes {
		if len(data) == 0 {
			continue
		}

		if data[0] == '\n' {
			data = data[1:]
		}

		data = com.Html2JS(data)
		localeDocPath := setting.DocsJSPath + docPath + "_RM_" + lang
		os.MkdirAll(path.Dir(localeDocPath), os.ModePerm)

		buf := new(bytes.Buffer)
		buf.WriteString("document.write(\"")
		buf.Write(data)
		buf.WriteString("\")")
		if err := ioutil.WriteFile(localeDocPath+".js", buf.Bytes(), 0655); err != nil {
			log.Error(2, "SavePkgDoc %q: %v", localeDocPath, err)
		}
	}
}

type exportSearchObject struct {
	Title string `json:"title"`
}

// renderDoc renders and saves the documentation file,
// and returns the new JSFile object corresponding to this generation.
func renderDoc(render macaron.Render, pdoc *Package, docPath string) (*db.JSFile, error) {
	data := make(map[string]interface{})
	data["PkgFullIntro"] = pdoc.Doc
	data["IsGoRepo"] = pdoc.IsGoRepo

	exports := make([]exportSearchObject, 0, 10)

	var buf bytes.Buffer
	links := make([]*Link, 0, len(pdoc.Types)+len(pdoc.Imports)+len(pdoc.TestImports)+
		len(pdoc.Funcs)+10)
	// Get all types, functions and import packages
	for _, t := range pdoc.Types {
		links = append(links, &Link{
			Name:    t.Name,
			Comment: template.HTMLEscapeString(t.Doc),
		})
		exports = append(exports, exportSearchObject{t.Name})
	}

	for _, f := range pdoc.Funcs {
		f.Code = template.HTMLEscapeString(f.Code)
		links = append(links, &Link{
			Name:    f.Name,
			Comment: template.HTMLEscapeString(f.Doc),
		})
		exports = append(exports, exportSearchObject{f.Name})
	}

	for _, t := range pdoc.Types {
		for _, f := range t.Funcs {
			links = append(links, &Link{
				Name:    f.Name,
				Comment: template.HTMLEscapeString(f.Doc),
			})
			exports = append(exports, exportSearchObject{f.Name})
		}

		for _, m := range t.Methods {
			exports = append(exports, exportSearchObject{t.Name + "." + m.Name})
		}
	}

	// Ignore C.
	for _, v := range append(pdoc.Imports, pdoc.TestImports...) {
		if v != "C" {
			links = append(links, &Link{
				Name: path.Base(v) + ".",
				Path: v,
			})
		}
	}

	// Set exported objects type-ahead.
	if len(exports) > 0 {
		pdoc.IsHasExport = true
		data["IsHasExports"] = true
		exportDataSrc, _ := json.Marshal(exports)
		data["ExportDataSrc"] = "<script>var exportDataSrc = " + string(exportDataSrc) + ";</script>"
	}

	pdoc.IsHasConst = len(pdoc.Consts) > 0
	pdoc.IsHasVar = len(pdoc.Vars) > 0
	if len(pdoc.Examples) > 0 {
		pdoc.IsHasExample = true
		data["IsHasExample"] = pdoc.IsHasExample
		data["Examples"] = pdoc.Examples
	}

	// Constants.
	data["IsHasConst"] = pdoc.IsHasConst
	data["Consts"] = pdoc.Consts
	for i, v := range pdoc.Consts {
		if len(v.Doc) > 0 {
			buf.Reset()
			doc.ToHTML(&buf, v.Doc, nil)
			v.Doc = buf.String()
		}
		buf.Reset()
		v.Decl = template.HTMLEscapeString(v.Decl)
		v.Decl = strings.Replace(v.Decl, "&#34;", "\"", -1)
		FormatCode(&buf, &v.Decl, links)
		v.FmtDecl = buf.String()
		pdoc.Consts[i] = v
	}

	// Variables.
	data["IsHasVar"] = pdoc.IsHasVar
	data["Vars"] = pdoc.Vars
	for i, v := range pdoc.Vars {
		if len(v.Doc) > 0 {
			buf.Reset()
			doc.ToHTML(&buf, v.Doc, nil)
			v.Doc = buf.String()
		}
		buf.Reset()
		FormatCode(&buf, &v.Decl, links)
		v.FmtDecl = buf.String()
		pdoc.Vars[i] = v
	}

	// Files.
	if len(pdoc.Files) > 0 {
		pdoc.IsHasFile = true
		data["IsHasFiles"] = pdoc.IsHasFile
		data["Files"] = pdoc.Files

		var query string
		if i := strings.Index(pdoc.Files[0].BrowseUrl, "?"); i > -1 {
			query = pdoc.Files[0].BrowseUrl[i:]
		}

		viewFilePath := path.Dir(pdoc.Files[0].BrowseUrl) + "/" + query
		// GitHub URL change.
		if strings.HasPrefix(viewFilePath, "github.com") {
			viewFilePath = strings.Replace(viewFilePath, "blob/", "tree/", 1)
		}
		data["ViewFilePath"] = viewFilePath
	}

	var err error
	renderFuncs(pdoc)

	data["Funcs"] = pdoc.Funcs
	for i, f := range pdoc.Funcs {
		if len(f.Doc) > 0 {
			buf.Reset()
			doc.ToHTML(&buf, f.Doc, nil)
			f.Doc = buf.String()
		}
		buf.Reset()
		FormatCode(&buf, &f.Decl, links)
		f.FmtDecl = buf.String() + " {"
		if exs := getExamples(pdoc, "", f.Name); len(exs) > 0 {
			f.Examples = exs
		}
		pdoc.Funcs[i] = f
	}

	data["Types"] = pdoc.Types
	for i, t := range pdoc.Types {
		for j, v := range t.Consts {
			if len(v.Doc) > 0 {
				buf.Reset()
				doc.ToHTML(&buf, v.Doc, nil)
				v.Doc = buf.String()
			}
			buf.Reset()
			v.Decl = template.HTMLEscapeString(v.Decl)
			v.Decl = strings.Replace(v.Decl, "&#34;", "\"", -1)
			FormatCode(&buf, &v.Decl, links)
			v.FmtDecl = buf.String()
			t.Consts[j] = v
		}
		for j, v := range t.Vars {
			if len(v.Doc) > 0 {
				buf.Reset()
				doc.ToHTML(&buf, v.Doc, nil)
				v.Doc = buf.String()
			}
			buf.Reset()
			FormatCode(&buf, &v.Decl, links)
			v.FmtDecl = buf.String()
			t.Vars[j] = v
		}

		for j, f := range t.Funcs {
			if len(f.Doc) > 0 {
				buf.Reset()
				doc.ToHTML(&buf, f.Doc, nil)
				f.Doc = buf.String()
			}
			buf.Reset()
			FormatCode(&buf, &f.Decl, links)
			f.FmtDecl = buf.String() + " {"
			if exs := getExamples(pdoc, "", f.Name); len(exs) > 0 {
				f.Examples = exs
			}
			t.Funcs[j] = f
		}
		for j, m := range t.Methods {
			if len(m.Doc) > 0 {
				buf.Reset()
				doc.ToHTML(&buf, m.Doc, nil)
				m.Doc = buf.String()
			}
			buf.Reset()
			FormatCode(&buf, &m.Decl, links)
			m.FmtDecl = buf.String() + " {"
			if exs := getExamples(pdoc, t.Name, m.Name); len(exs) > 0 {
				m.Examples = exs
			}
			t.Methods[j] = m
		}
		if len(t.Doc) > 0 {
			buf.Reset()
			doc.ToHTML(&buf, t.Doc, nil)
			t.Doc = buf.String()
		}
		buf.Reset()
		FormatCode(&buf, &t.Decl, links)
		t.FmtDecl = buf.String()
		if exs := getExamples(pdoc, "", t.Name); len(exs) > 0 {
			t.Examples = exs
		}
		pdoc.Types[i] = t
	}

	// Examples.
	links = append(links, &Link{
		Name: path.Base(pdoc.ImportPath) + ".",
	})

	for _, e := range pdoc.Examples {
		buf.Reset()
		FormatCode(&buf, &e.Code, links)
		e.Code = buf.String()
	}

	data["ProjectPath"] = pdoc.ProjectPath
	data["ImportPath"] = pdoc.ImportPath

	// GitHub redirects non-HTTPS link and Safari loses "#XXX".
	if strings.HasPrefix(pdoc.ProjectPath, "github") {
		data["Secure"] = "s"
	}

	result, err := render.HTMLBytes("docs/tpl", data)
	if err != nil {
		return nil, fmt.Errorf("rendering HTML: %v", err)
	}

	numExtraFiles := SaveDocPage(docPath, result)
	if numExtraFiles == -1 {
		return nil, errors.New("save JS file wasn't successful")
	}
	SavePkgDoc(pdoc.ImportPath, pdoc.Readme)

	data["UtcTime"] = time.Unix(pdoc.Created, 0).UTC()
	return &db.JSFile{
		Etag:          pdoc.Etag,
		Status:        db.JSFileStatusGenerated,
		NumExtraFiles: numExtraFiles,
	}, nil
}

type requestType int

const (
	RequestTypeHuman requestType = iota
	RequestTypeRefresh
)

// CheckPackage checks package by import path.
func CheckPackage(importPath string, render macaron.Render, rt requestType) (*db.PkgInfo, error) {
	// Trim prefix of standard library
	importPath = strings.TrimPrefix(importPath, "github.com/golang/go/tree/master/src")

	pinfo, err := db.GetPkgInfo(importPath)
	if rt != RequestTypeRefresh {
		if err == nil {
			gobPath := setting.DocsGobPath + importPath + ".gob"
			if !setting.ProdMode && com.IsFile(gobPath) {
				pdoc := new(Package)
				fr, err := os.Open(gobPath)
				if err != nil {
					return nil, fmt.Errorf("read gob: %v", err)
				} else if err = gob.NewDecoder(fr).Decode(pdoc); err != nil {
					fr.Close()
					return nil, fmt.Errorf("decode gob: %v", err)
				}
				fr.Close()

				_, err = renderDoc(render, pdoc, importPath)
				if err != nil {
					return nil, fmt.Errorf("render cached doc: %v", err)
				}
			}

			pinfo.Views++
			pinfo.LastViewed = time.Now().Unix()
			if err = db.SavePkgInfo(pinfo, false); err != nil {
				return nil, fmt.Errorf("update views: %v", err)
			}
			return pinfo, nil
		}
	}

	// Just in case, should never happen
	if err == db.ErrEmptyPackagePath {
		return nil, err
	}

	var etag string
	if err != db.ErrPackageVersionTooOld && pinfo != nil {
		etag = pinfo.Etag
	}

	// Fetch package from VCS
	c := make(chan crawlResult, 1)
	go func() {
		pdoc, err := crawlDoc(importPath, etag)
		c <- crawlResult{pdoc, err}
	}()

	var pdoc *Package
	err = nil // Reset
	select {
	case cr := <-c:
		if cr.err == nil {
			pdoc = cr.pdoc
		} else {
			err = cr.err
		}
	case <-time.After(setting.FetchTimeout):
		err = ErrFetchTimeout
	}

	if err != nil {
		if err == ErrPackageNotModified {
			log.Trace("Package has not been modified: %s", pinfo.ImportPath)
			// Update time so cannot refresh too often
			pinfo.Created = time.Now().UTC().Unix()
			return pinfo, db.SavePkgInfo(pinfo, false)
		} else if err == ErrInvalidRemotePath {
			return nil, ErrInvalidRemotePath // Allow caller to make redirect to search.
		}
		return nil, fmt.Errorf("check package: %v", err)
	}

	if !setting.ProdMode {
		gobPath := setting.DocsGobPath + importPath + ".gob"
		os.MkdirAll(path.Dir(gobPath), os.ModePerm)
		fw, err := os.Create(gobPath)
		if err != nil {
			return nil, fmt.Errorf("create gob: %v", err)
		}
		defer fw.Close()
		if err = gob.NewEncoder(fw).Encode(pdoc); err != nil {
			return nil, fmt.Errorf("encode gob: %v", err)
		}
	}

	log.Trace("Walked package %q, Goroutine #%d", pdoc.ImportPath, runtime.NumGoroutine())

	jsFile, err := renderDoc(render, pdoc, importPath)
	if err != nil {
		return nil, fmt.Errorf("render doc: %v", err)
	}

	if pinfo != nil {
		pdoc.ID = pinfo.ID
		pdoc.RefIDs = pinfo.RefIDs
		pdoc.RefNum = pinfo.RefNum
	}

	pdoc.Created = time.Now().UTC().Unix()
	pdoc.LastViewed = time.Now().Unix()
	if err = db.SavePkgInfo(pdoc.PkgInfo, true); err != nil {
		return nil, fmt.Errorf("SavePkgInfo[%s]: %v", importPath, err)
	}

	jsFile.PkgID = pdoc.PkgInfo.ID
	if err = db.SaveJSFile(jsFile); err != nil {
		return nil, fmt.Errorf("SaveJSFile[%s]: %v", importPath, err)
	}
	pdoc.JSFile = jsFile

	return pdoc.PkgInfo, nil
}
