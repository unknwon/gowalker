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
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
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

func FormatCode(w io.Writer, code string, links []*Link) {
	length := len(code) // Length of whole code.
	if length == 0 {
		return
	}

	isString := false       // Indicates if right now is checking string.
	isRawString := false    // Indicates if right now is checking raw string.
	isComment := false      // Indicates if right now is checking comments.
	isBlockComment := false // Indicates if right now is checking block comments.
	last := 0               // Start index of the word.
	pos := 0                // Current index.

	for {
		// Cut words.
	CutWords:
		for {
			if code[pos] < 'A' || code[pos] > 'z' || (code[pos] > 'Z' && code[pos] < 'a') {
				switch {
				case !isComment && code[pos] == '`':
					if !isString {
						isString = true
						isRawString = true
					} else {
						// Handle string highlight.
						break CutWords
					}
				case !isComment && !isRawString && code[pos] == '"' && (pos > 0) && code[pos-1] != '\\':
					if !isString {
						isString = true
					} else {
						// Handle string highlight.
						break CutWords
					}
				case !isString && !isComment && code[pos] == '/' && (pos > 0) && code[pos+1] == '/':
					isComment = true
				case isComment:
					if isBlockComment {
						// End of block comments.
						if code[pos] == '/' && code[pos-1] == '*' {
							break CutWords
						}
					} else {
						switch {
						case code[pos] == '*' && code[pos-1] == '/':
							// Start of block comments.
							isBlockComment = true
						case code[pos] == '\n':
							break CutWords
						}
					}
				case !isString && code[pos] > 47 && code[pos] < 58: // Number
				case !isString && (code[pos] != '.' || code[pos] == '\n'):
					break CutWords
				}
			}

			if pos == length-1 {
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
			fmt.Fprintf(w, `<span class="com">%s</span>`, seg)
		case isString:
			isString = false
			isRawString = false
			fmt.Fprintf(w, `<span class="str">%s</span>`, seg)
		case pos-last > 1:
			// Check if the last word of the paragraphy.
			l := len(seg)
			if pos+1 == length {
				l++
			}

			// Check keywords.
			switch seg[:l-1] {
			case "return":
				fmt.Fprintf(w, `<span class="ret">%s</span>%s`, "return", seg[l-1:])
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
			fmt.Fprintf(w, "%s", code[last:])
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
