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
	"os"

	"github.com/Unknwon/com"
	"github.com/Unknwon/goconfig"
	"github.com/Unknwon/macaron"

	"github.com/Unknwon/gowalker/modules/log"
)

var (
	// App settings.
	AppVer string

	// Server settings.
	HttpPort string

	// Server settings.
	DisableRouterLog bool

	// Global setting objects.
	Cfg               *goconfig.ConfigFile
	ProdMode          bool
	GithubCredentials string

	// I18n settings.
	Langs, Names []string
)

func init() {
	log.NewLogger(0, "console", `{"level": 0}`)

	var err error
	Cfg, err = goconfig.LoadConfigFile("conf/app.ini")
	if err != nil {
		log.Fatal(4, "Fail to parse 'conf/app.ini': %v", err)
	}
	if com.IsFile("custom/app.ini") {
		if err = Cfg.AppendFiles("custom/app.ini"); err != nil {
			log.Fatal(4, "Fail to load 'custom/app.ini': %v", err)
		}
	}

	if Cfg.MustValue("", "RUN_MODE", "dev") == "prod" {
		macaron.Env = macaron.PROD
		ProdMode = true
	}

	HttpPort = Cfg.MustValue("server", "HTTP_PORT", "8080")

	DisableRouterLog = Cfg.MustBool("server", "DISABLE_ROUTER_LOG")

	GithubCredentials = "client_id=" + Cfg.MustValue("github", "CLIENT_ID") +
		"&client_secret=" + Cfg.MustValue("github", "CLIENT_SECRET")

	Langs = Cfg.MustValueArray("i18n", "LANGS", ",")
	Names = Cfg.MustValueArray("i18n", "NAMES", ",")
}

// SaveConfig saves configuration file.
func SaveConfig() error {
	os.MkdirAll("custom", os.ModePerm)
	return goconfig.SaveConfigFile(Cfg, "custom/app.ini")
}
