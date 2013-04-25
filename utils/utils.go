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

var urlPattern = regexp.MustCompile(`[a-zA-z]+://[^\s]*`)

// ParseDoc converts two line breaks("\n\n") to <p></p> tags,
// finds code blocks with <pre></pre> tags
func ParseDoc1(doc string) string {
	docBuf := bytes.NewBufferString(doc)
	htmlBuf := new(bytes.Buffer)
	preTag := false // Indicates begin of code block
	inCode := false // Indicates if need to check code block
	indentLevel := 0

	htmlBuf.WriteString("<p>")
	for {
		line, err := docBuf.ReadString('\n')
		fmt.Println(line)
		if err != nil {
			// Unexpected error
			if err != io.EOF {
				return doc
			}

			// Reached end of documentation, if nothing to read then break,
			// otherwise handle the last line.
			if len(line) == 0 {
				break
			}
		}

		switch {
		case len(line) == 0: // Empty line
			continue
		case len(line) > 1 && line[len(line)-2] == ':':
			htmlBuf.WriteString(line + "</p>")
			i := 0
			for {
				if line[i] == '\t' {
					indentLevel = i
				} else {
					break
				}
				i++
			}
			indentLevel++
			preTag = true
			inCode = true
		case line == "\n":
		case preTag && len(line) > indentLevel && line[indentLevel] == '\t':
			htmlBuf.WriteString("<pre>" + line)
			preTag = false
		case inCode && len(line) > indentLevel && line[indentLevel] == '\t':
			htmlBuf.WriteString(line)
		case inCode && len(line) > indentLevel && line[indentLevel] != '\t':
			htmlBuf.WriteString("</pre><p>" + line)
		default:
			htmlBuf.WriteString(line + "</p>")
		}

		// Reached end of documentation
		if err == io.EOF {
			break
		}
	}

	return htmlBuf.String()
}

func ParseDoc(doc string) string {
	tagDoc := new(bytes.Buffer)
	code := false
	contCode := false // Indicates if this paragraph can also be code block
	// Get paragraphs
	paras := strings.Split(doc, "\n\n")
	for i, p := range paras {
		fmt.Println(p)
		links := urlPattern.FindAllString(p, -1)
		// Check links
		for _, s := range links {
			if i := strings.Index(s, "\""); i > -1 {
				s = s[:i]
			}
			p = strings.Replace(p, s, "<a href=\""+s+"\">"+s+"</a>", 1)
		}

		// Check if this paragraph is code example
		if code {
			tagDoc.WriteString("<pre>" + p)
			code = false
			// End of the documentation
			if i == len(paras)-1 {
				tagDoc.WriteString("</pre>")
			}
		} else if contCode {
			if strings.Index(p, ":=") > -1 || isHasKeyword(p) {
				tagDoc.WriteString("\n\n" + p)
				// End of the documentation
				if i == len(paras)-1 {
					tagDoc.WriteString("</pre>")
				}
			} else {
				tagDoc.WriteString("</pre><p>" + p + "</p>")
				contCode = false
				if len(p) > 0 && p[len(p)-1] == ':' {
					code = true
					contCode = true
				}
			}
		} else {
			isHasSymbol := len(p) > 0 && p[len(p)-1] == ':'
			if j := strings.Index(p, "xample:"); j > -1 && !isHasSymbol {
				tagDoc.WriteString("<p>" + p[:j+7] + "</p><pre>" + p[j+7:])
				// End of the documentation
				if i == len(paras)-1 {
					tagDoc.WriteString("</pre>")
				}
				contCode = true
			} else {
				tagDoc.WriteString("<p>" + p + "</p>")
				if isHasSymbol {
					code = true
					contCode = true
				}
			}
		}
	}

	return tagDoc.String()
}

var keywords = []string{"func ", "import ", "log.", "http."}

func isHasKeyword(str string) bool {
	isCode := false
	for _, k := range keywords {
		if strings.Index(str, k) > -1 {
			isCode = true
			break
		}
	}
	return isCode
}
