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
	"github.com/astaxie/beego"
	"github.com/unknwon/gowalker/controllers"
	"github.com/unknwon/gowalker/models"
)

const (
	VERSION = "0.0.1"
)

func main() {
	beego.Info("Go Walker " + VERSION)

	beego.Info("Initialize database")
	if err := models.InitDb(); err != nil {
		beego.Error(err)
	}

	// Set static path
	beego.SetStaticPath("/github.com", "docs/github.com")
	beego.SetStaticPath("/code.google.com", "docs/code.google.com")

	// Register routers
	beego.Router("/", &controllers.HomeController{})
	beego.Router("/search", &controllers.SearchController{})
	beego.Router("/-/index", &controllers.IndexController{})
	beego.Router("/-/about", &controllers.AboutController{})

	// For 404 pages
	beego.Router("/:all", &controllers.ErrorController{})
	beego.Run()
}
