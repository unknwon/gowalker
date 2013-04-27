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

package main

import (
	"os"
	"strings"

	"github.com/astaxie/beego"
	"github.com/unknwon/gowalker/controllers"
	"github.com/unknwon/gowalker/models"
)

const (
	VERSION = "0.0.7.0427"
)

func main() {
	beego.Info("Go Walker " + VERSION)

	// Initialization
	beego.Info("Initialize database: gowalker.db")
	os.Mkdir("./data", os.ModePerm)
	if err := models.InitDb(); err != nil {
		beego.Error(err)
	}
	beego.Info("Initialize directory: docs/")
	os.Mkdir("./docs", os.ModePerm)

	// Load VCS
	vcs := strings.Split(beego.AppConfig.String("VCS"), "|")
	// Set static path
	for _, v := range vcs {
		beego.SetStaticPath("/"+v, "docs/"+v)
	}

	// Load languages
	langs := strings.Split(beego.AppConfig.String("language"), "|")

	// Register routers
	beego.Router("/", &controllers.HomeController{})
	// Languages
	for _, v := range langs {
		lang := "/" + v
		beego.Router(lang, &controllers.HomeController{})
		beego.Router(lang+"/search", &controllers.SearchController{})
		beego.Router(lang+"/index", &controllers.IndexController{})
		beego.Router(lang+"/about", &controllers.AboutController{})
	}

	// For all 404 pages
	beego.Router("/:all", &controllers.ErrorController{})

	// Template functions
	beego.Run()
}
