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

package routers

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/models"
	"github.com/astaxie/beego"
)

// ExamplesRouter serves examples pages.
type ExamplesRouter struct {
	beego.Controller
}

// Get implemented Get method for ExamplesRouter.
func (this *ExamplesRouter) Get() {
	this.Data["IsExamples"] = true
	// Set language version.
	curLang := globalSetting(this.Ctx, this.Input(), this.Data)

	// Set properties.
	this.TplNames = "examples_" + curLang.Lang + ".html"

	// Get argument(s).
	gist := strings.TrimSpace(this.Input().Get("gist"))
	q := strings.TrimSpace(this.Input().Get("q"))

	if len(gist) == 0 || len(q) == 0 {
		this.Data["IsShowExam"] = true
		pkgExams, _ := models.GetAllExams()
		this.Data["ExamNum"] = len(pkgExams)
		this.Data["AllExams"] = pkgExams
		return
	}

	this.Data["ImportPath"] = q

	if !strings.Contains(gist, "gist.github.com") {
		this.Data["IsHasError"] = true
		this.Data["ErrMsg"] = fmt.Sprintf("Gist path[ %s ] is not legal.", gist)
		return
	}

	// Get Gist.
	gist = strings.TrimPrefix(gist, "http://")
	if !strings.HasPrefix(gist, "https://") {
		gist = "https://" + gist
	}
	gist += "/raw"

	req, err := http.NewRequest("GET", gist, nil)
	if err != nil {
		fmt.Println("ExamplesRouter.Get -> New request:", err)
		this.Data["IsHasError"] = true
		this.Data["ErrMsg"] = err.Error()
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/29.0.1541.0 Safari/537.36")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("ExamplesRouter.Get -> Get response:", err)
		this.Data["IsHasError"] = true
		this.Data["ErrMsg"] = err.Error()
		return
	}
	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ExamplesRouter.Get -> Get body:", err)
		this.Data["IsHasError"] = true
		this.Data["ErrMsg"] = err.Error()
		return
	}

	// Parse examples.
	g, err := parseExamples(html, gist, q)
	if err != nil {
		this.Data["IsHasError"] = true
		this.Data["ErrMsg"] = err.Error()
		return
	}

	if len(g.Examples) == 0 {
		this.Data["IsHasError"] = true
		this.Data["ErrMsg"] = "No example has beed found."
		return
	}

	err = saveExamples(g)
	if err != nil {
		this.Data["IsHasError"] = true
		this.Data["ErrMsg"] = err.Error()
		return
	}

	this.Redirect("/"+q, 302)
}

func saveExamples(gist *doc.Gist) error {
	var buf bytes.Buffer
	// Examples.
	for _, e := range gist.Examples {
		buf.WriteString(e.Name)
		buf.WriteString("&E#")
		buf.WriteString(e.Doc)
		buf.WriteString("&E#")
		buf.WriteString(*doc.CodeEncode(&e.Code))
		buf.WriteString("&E#")
		// buf.WriteString(e.Play)
		// buf.WriteString("&E#")
		buf.WriteString(e.Output)
		buf.WriteString("&$#")
	}

	pkgExam := &models.PkgExam{
		Path:     gist.ImportPath,
		Gist:     gist.Gist,
		Examples: buf.String(),
	}

	return models.SavePkgExam(pkgExam)
}

func parseExamples(html []byte, gist, path string) (*doc.Gist, error) {
	gist = strings.TrimPrefix(gist, "https://")
	gist = strings.TrimSuffix(gist, "/raw")
	gist = strings.TrimSuffix(gist, "/")
	g := &doc.Gist{Gist: gist}
	exam := &doc.Example{}

	var status int
	// Status.
	const (
		EMPTY = iota
		NAME
		CODE
		OUTPUT
	)

	for i, v := range strings.Split(string(html), "\n") {
		// Check status.
		switch {
		case len(v) == 0: // Empty line.
			status = EMPTY
		case len(v) >= 2 && v[:2] == "//": // Comment.
			if len(exam.Name) > 0 {
				status = CODE
				break
			}

			status = EMPTY
		case len(v) >= 3 && v[0] == '[' && v[len(v)-1] == ']': // Name
			if len(g.ImportPath) == 0 {
				return nil, errors.New(fmt.Sprintf("Line %d: Expect import path, but found example name[ %s ]", i+1, v))
			}

			// Add example to slice.
			if len(exam.Name) > 0 && len(exam.Code) > 0 {
				g.Examples = append(g.Examples, exam)
				exam = &doc.Example{}
			}
			status = NAME
		case len(g.ImportPath) == 0 && strings.HasPrefix(v, "import_path"):
			index := strings.Index(v, "=")
			g.ImportPath = strings.TrimSpace(v[index+1:])

			if g.ImportPath != path {
				return nil, errors.New(fmt.Sprintf("Line %d: Expect import path[ %s ], but found[ %s ]", i+1, path, g.ImportPath))
			}
		case len(v) >= 7 && strings.Contains(strings.ToLower(v), "output:"):
			if len(exam.Name) == 0 {
				return nil, errors.New(fmt.Sprintf("Line %d: Expect example name, but found output[ %s ]", i+1, v))
			}
			status = OUTPUT
		default:
			if status != OUTPUT {
				status = CODE
			}
		}

		// Get content.
		switch status {
		case EMPTY:
		case NAME:
			exam.Name = v[1 : len(v)-1]
		case CODE:
			exam.Code += v + "\n"
		case OUTPUT:
			if !strings.Contains(strings.ToLower(v), "output:") {
				exam.Output += v + "\n"
			}
		}
	}

	// Add example to slice.
	if len(exam.Name) > 0 && len(exam.Code) > 0 {
		g.Examples = append(g.Examples, exam)
		exam = &doc.Example{}
	}

	return g, nil
}
