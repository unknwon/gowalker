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
	"html/template"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/Unknwon/com"
	"github.com/Unknwon/goconfig"
)

var Cfg *goconfig.ConfigFile

// LoadConfig loads configuration file.
func LoadConfig(cfgPath string) (err error) {
	if !com.IsExist(cfgPath) {
		os.Create(cfgPath)
	}

	Cfg, err = goconfig.LoadConfigFile(cfgPath)
	return err
}

// SaveConfig saves configuration file.
func SaveConfig() error {
	return goconfig.SaveConfigFile(Cfg, "conf/app.ini")
}

var readmePat = regexp.MustCompile(`^[Rr][Ee][Aa][Dd][Mm][Ee](?:$|\.)`)

func IsDocFile(n string) bool {
	if strings.HasSuffix(n, ".go") && n[0] != '_' && n[0] != '.' {
		return true
	}
	return strings.HasPrefix(strings.ToLower(n), "readme")
}

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
					case !isString && curChar == '_' && (*code)[pos-1] != ' ': // Underline: _.
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
