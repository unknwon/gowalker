// Copyright 2014 Unknwon
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

package setting

import (
	"time"

	"github.com/Unknwon/com"
	"github.com/Unknwon/macaron"
	"gopkg.in/ini.v1"

	"github.com/Unknwon/gowalker/modules/log"
)

var (
	// Application settings.
	AppVer string

	// Server settings.
	HttpPort string

	// Global setting objects.
	Cfg               *ini.File
	ProdMode          bool
	GithubCredentials string
	FetchTimeout      time.Duration = 60 * time.Second

	DocsJsPath  = "raw/docs/"
	DocsGobPath = "raw/gob/"
)

func init() {
	log.NewLogger(0, "console", `{"level": 0}`)

	sources := []interface{}{"conf/app.ini"}
	if com.IsFile("custom/app.ini") {
		sources = append(sources, "custom/app.ini")
	}

	var err error
	Cfg, err = macaron.SetConfig(sources[0], sources[1:]...)
	if err != nil {
		log.Fatal(4, "Fail to set configuration: %v", err)
	}

	if Cfg.Section("").Key("RUN_MODE").MustString("dev") == "prod" {
		ProdMode = true
		macaron.Env = macaron.PROD
		macaron.ColorLog = false
	}

	HttpPort = Cfg.Section("server").Key("HTTP_PORT").MustString("8080")

	GithubCredentials = "client_id=" + Cfg.Section("github").Key("CLIENT_ID").String() +
		"&client_secret=" + Cfg.Section("github").Key("CLIENT_SECRET").String()
}
