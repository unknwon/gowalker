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
	"github.com/Unknwon/log"
	"github.com/Unknwon/macaron"
	"gopkg.in/ini.v1"
)

var (
	// Application settings.
	AppVer   string
	ProdMode bool

	// Server settings.
	HTTPPort     int
	FetchTimeout time.Duration
	DocsJsPath   string
	DocsGobPath  string

	// Global settings.
	Cfg               *ini.File
	GitHubCredentials string
)

func init() {
	log.Prefix = "[Go Walker]"

	sources := []interface{}{"conf/app.ini"}
	if com.IsFile("custom/app.ini") {
		sources = append(sources, "custom/app.ini")
	}

	var err error
	Cfg, err = macaron.SetConfig(sources[0], sources[1:]...)
	if err != nil {
		log.FatalD(4, "Fail to set configuration: %v", err)
	}

	if Cfg.Section("").Key("RUN_MODE").String() == "prod" {
		ProdMode = true
		macaron.Env = macaron.PROD
		macaron.ColorLog = false
	}

	sec := Cfg.Section("server")
	HTTPPort = sec.Key("HTTP_PORT").MustInt(8080)
	FetchTimeout = time.Duration(sec.Key("FETCH_TIMEOUT").MustInt(60)) * time.Second
	DocsJsPath = sec.Key("DOCS_JS_PATH").MustString("raw/docs/")
	DocsGobPath = sec.Key("DOCS_GOB_PATH").MustString("raw/gob/")

	GitHubCredentials = "client_id=" + Cfg.Section("github").Key("CLIENT_ID").String() +
		"&client_secret=" + Cfg.Section("github").Key("CLIENT_SECRET").String()

}
