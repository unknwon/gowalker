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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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
	curLang := setLangVer(this.Ctx, this.Input(), this.Data)

	// Set properties
	this.TplNames = "examples_" + curLang.Lang + ".html"

	// Get query field.
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
	// Get Gist.
	if !strings.HasPrefix(gist, "https://") {
		gist = "https://" + gist
	}

	req, err := http.NewRequest("GET", gist, nil)
	if err != nil {
		fmt.Println("ExamplesRouter.Get -> New request:", err)
		this.Data["ErrMsg"] = err.Error()
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/29.0.1541.0 Safari/537.36")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("ExamplesRouter.Get -> Get response:", err)
		this.Data["ErrMsg"] = err.Error()
		return
	}
	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ExamplesRouter.Get -> Get body:", err)
		this.Data["ErrMsg"] = err.Error()
		return
	}

	// Parse examples.
	err = parseExamples(html, q)
	if err != nil {
		this.Data["ErrMsg"] = err.Error()
		return
	}
}

func parseExamples(html []byte, path string) error {
	return errors.New("Unrecognized Gist.")
}
