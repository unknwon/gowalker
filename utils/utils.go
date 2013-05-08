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
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
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

// A link describes the (HTML) link information for an identifier.
// The zero value of a link represents "no link".
//
type Link struct {
	Path, Name, Comment string // package path, identifier name, and comments.
}

func isLetter(l uint8) bool {
	return (l >= 'A' && l <= 'Z') || (l >= 'a' && l <= 'z')
}

// FormatCode highlights keywords and adds HTML links to them.
func FormatCode(w io.Writer, code *string, links []*Link) {
	length := len(*code) // Length of whole code.
	if length == 0 {
		return
	}

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
			if !isLetter(curChar) {
				if !isComment {
					switch {
					case (curChar == '\'' || curChar == '"' || curChar == '`') && (*code)[pos-1] != '\\': // String.
						if !isString {
							// Set string tag.
							strTag = curChar
							isString = true
						} else {
							// Check string tag.
							if curChar == strTag {
								// Handle string highlight.
								break CutWords
							}
						}
					case !isString && curChar == '/': //&& (pos > 0) && code[pos+1] == '/':
						isComment = true
					case !isString && curChar > 47 && curChar < 58: // Number
					case !isString && curChar == '_' && (*code)[pos-1] != ' ': // _
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
			fmt.Fprintf(w, `<span class="str">%s</span>`, seg)
		case pos-last > 1:
			// Check if the last word of the paragraphy.
			l := len(seg)
			keyword := seg
			if !isLetter(seg[l-1]) {
				keyword = seg[:l-1]
			} else {
				l++
			}

			// Check keywords.
			switch keyword {
			case "return", "break":
				fmt.Fprintf(w, `<span class="ret">%s</span>%s`, keyword, seg[l-1:])
				break CheckLink
			case "func", "range", "for", "if", "else", "type", "struct", "select", "case", "var", "const":
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
					fmt.Fprintf(w, `<a class="ext" title="%s" target="_blank" href="%s">%s</a>%s`,
						link.Comment, link.Path, link.Name, seg[l-1:])
				}
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
			// Exported types in current package.
			if l.Name == name {
				return l, true
			}
		} else {
			// Functions and types from imported packages.
			if l.Name == left {
				return &Link{Name: name, Path: "/" + l.Path + "#" + right}, true
			}
		}
	}
	return nil, false
}

var badSynopsisPrefixes = []string{
	"Autogenerated by Thrift Compiler",
	"Automatically generated ",
	"Auto-generated by ",
	"Copyright ",
	"COPYRIGHT ",
	`THE SOFTWARE IS PROVIDED "AS IS"`,
	"TODO: ",
	"vim:",
}

// Synopsis extracts the first sentence from s. All runs of whitespace are
// replaced by a single space.
func Synopsis(s string) string {

	parts := strings.SplitN(s, "\n\n", 2)
	s = parts[0]

	var buf []byte
	const (
		other = iota
		period
		space
	)
	last := space
Loop:
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch b {
		case ' ', '\t', '\r', '\n':
			switch last {
			case period:
				break Loop
			case other:
				buf = append(buf, ' ')
				last = space
			}
		case '.':
			last = period
			buf = append(buf, b)
		default:
			last = other
			buf = append(buf, b)
		}
	}

	// Ensure that synopsis fits an App Engine datastore text property.
	const m = 400
	if len(buf) > m {
		buf = buf[:m]
		if i := bytes.LastIndex(buf, []byte{' '}); i >= 0 {
			buf = buf[:i]
		}
		buf = append(buf, " ..."...)
	}

	s = string(buf)

	r, n := utf8.DecodeRuneInString(s)
	if n < 0 || unicode.IsPunct(r) || unicode.IsSymbol(r) {
		// ignore Markdown headings, editor settings, Go build constraints, and * in poorly formatted block comments.
		s = ""
	} else {
		for _, prefix := range badSynopsisPrefixes {
			if strings.HasPrefix(s, prefix) {
				s = ""
				break
			}
		}
	}

	return s
}
