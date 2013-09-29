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
	"html/template"
	"strings"

	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

var robotTxt string

const robotTpl = `{{$disallow := .Disallow}}{{range .Uas}}User-agent: {{.}}
Disallow: {{$disallow}}

{{end}}User-agent: *
Disallow: /
`

// RobotRouter implemented global settings for all other routers.
type RobotRouter struct {
	beego.Controller
}

// Get implemented Prepare method for RobotRouter.
func (this *RobotRouter) Get() {
	if len(robotTxt) == 0 {
		// Generate "robot.txt".
		t := template.New("robotTpl")
		t.Parse(robotTpl)
		uas := strings.Split(utils.Cfg.MustValue("robot", "uas"), "|")

		data := make(map[string]interface{})
		data["Uas"] = uas
		data["Disallow"] = utils.Cfg.MustValue("robot", "disallow")

		buf := new(bytes.Buffer)
		t.Execute(buf, data)
		robotTxt = buf.String()
	}

	this.Ctx.WriteString(robotTxt)
	return
}
