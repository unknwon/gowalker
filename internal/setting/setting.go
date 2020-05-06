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

	"github.com/unknwon/com"
	log "gopkg.in/clog.v1"
	"gopkg.in/ini.v1"
	"gopkg.in/macaron.v1"
)

var (
	// Application settings
	AppVer           string
	ProdMode         bool
	DisableRouterLog bool

	// Server settings
	HTTPPort     int
	FetchTimeout time.Duration
	DocsJSPath   string
	DocsGobPath  string

	DigitalOcean struct {
		Spaces struct {
			Enabled   bool
			Endpoint  string
			AccessKey string
			SecretKey string
			Bucket    string
			BucketURL string `ini:"BUCKET_URL"`
		}
	}

	Maintenance struct {
		JSRecycleDays int `ini:"JS_RECYCLE_DAYS"`
	}

	// Global settings
	Cfg    *ini.File
	GitHub struct {
		ClientID     string `ini:"CLIENT_ID"`
		ClientSecret string
	}
	RefreshInterval = 5 * time.Minute
)

func init() {
	log.New(log.CONSOLE, log.ConsoleConfig{})

	sources := []interface{}{"conf/app.ini"}
	if com.IsFile("custom/app.ini") {
		sources = append(sources, "custom/app.ini")
	}

	var err error
	Cfg, err = macaron.SetConfig(sources[0], sources[1:]...)
	if err != nil {
		log.Fatal(2, "Failed to set configuration: %v", err)
	}
	Cfg.NameMapper = ini.AllCapsUnderscore

	if Cfg.Section("").Key("RUN_MODE").String() == "prod" {
		ProdMode = true
		macaron.Env = macaron.PROD
		macaron.ColorLog = false

		log.New(log.CONSOLE, log.ConsoleConfig{
			Level:      log.INFO,
			BufferSize: 100,
		})
	}

	DisableRouterLog = Cfg.Section("").Key("DISABLE_ROUTER_LOG").MustBool()

	sec := Cfg.Section("server")
	HTTPPort = sec.Key("HTTP_PORT").MustInt(8080)
	FetchTimeout = time.Duration(sec.Key("FETCH_TIMEOUT").MustInt(60)) * time.Second
	DocsJSPath = sec.Key("DOCS_JS_PATH").MustString("raw/docs/")
	DocsGobPath = sec.Key("DOCS_GOB_PATH").MustString("raw/gob/")

	if err = Cfg.Section("github").MapTo(&GitHub); err != nil {
		log.Fatal(2, "Failed to map GitHub settings: %v", err)
	} else if err = Cfg.Section("digitalocean.spaces").MapTo(&DigitalOcean.Spaces); err != nil {
		log.Fatal(2, "Failed to map DigitalOcean.Spaces settings: %v", err)
	} else if err = Cfg.Section("maintenance").MapTo(&Maintenance); err != nil {
		log.Fatal(2, "Failed to map Maintenance settings: %v", err)
	}

	sec = Cfg.Section("log.discord")
	if sec.Key("ENABLED").MustBool() {
		log.New(log.DISCORD, log.DiscordConfig{
			Level:      log.ERROR,
			BufferSize: 100,
			URL:        sec.Key("URL").MustString(""),
			Username:   "Go Walker",
		})
	}
}
