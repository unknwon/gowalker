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

// Go Walker is a web server for Go project source code analysis.

package main

import (
	"strings"

	"github.com/Unknwon/gowalker/controllers"
	"github.com/astaxie/beego"
)

const (
	VERSION = "0.0.8.0428"
)

func main() {
	beego.Info("Go Walker", VERSION)

	// Load languages.
	langs := strings.Split(beego.AppConfig.String("language"), "|")
	names := strings.Split(beego.AppConfig.String("langNames"), "|")
	controllers.InitLangs(langs, names)

	// Register routers.
	beego.Router("/", &controllers.HomeController{})
	beego.Router("/index", &controllers.IndexController{})
	beego.Router("/about", &controllers.AboutController{})
	beego.Router("/search", &controllers.SearchController{})

	// For all unknown pages.
	beego.Router("/:all", &controllers.HomeController{})
	beego.Run()
}
