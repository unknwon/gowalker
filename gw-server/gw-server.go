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

// Go Walker is a web server that generates Go projects API documentation with source code on the fly.
package main

import (
	"os"
	"runtime"

	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/gw-server/routers"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

const (
	APP_VER = "0.8.6.0731"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Set App version and log level.
	routers.AppVer = "v" + APP_VER

	if beego.AppConfig.String("runmode") == "pro" {
		beego.SetLevel(beego.LevelInfo)

		beego.Info("Go Walker", APP_VER)

		os.Mkdir("../log", os.ModePerm)
		fw := beego.NewFileWriter("../log/log", true)
		err := fw.StartLogger()
		if err != nil {
			beego.Critical("NewFileWriter ->", err)
		}
	}

	doc.SetGithubCredentials(utils.Cfg.MustValue("github", "client_id"),
		utils.Cfg.MustValue("github", "client_secret"))
}

func main() {
	beego.AppName = "Go Walker Server"
	beego.Info("Go Walker Server", APP_VER)

	// Register routers.
	beego.Router("/", &routers.HomeRouter{})
	beego.Router("/refresh", &routers.RefreshRouter{})
	beego.Router("/search", &routers.SearchRouter{})
	beego.Router("/index", &routers.IndexRouter{})
	beego.Router("/labels", &routers.LabelsRouter{})
	beego.Router("/examples", &routers.ExamplesRouter{})
	beego.Router("/funcs", &routers.FuncsRouter{})
	beego.Router("/about", &routers.AboutRouter{})

	// Register template functions.

	// For all unknown pages.
	beego.Router("/:all", &routers.HomeRouter{})
	beego.Run()
}
