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
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/Unknwon/gowalker/controllers"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

const (
	VERSION = "0.1.6.0506" // Application version.
)

func init() {
	// Try to have highest performance.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Set application log level.
	beego.SetLevel(beego.LevelTrace)

	// Initialize log file.
	os.Mkdir("./log", os.ModePerm)
	// Compute log file name as format '<year>-<month>-<day>.txt', eg.'2013-5-6.txt'.
	logName := fmt.Sprintf("./log/%d-%d-%d.txt", time.Now().Year(), time.Now().Month(), time.Now().Day())
	// Open or create log file.
	var fl *os.File
	var err error
	if utils.IsExist(logName) {
		fl, err = os.OpenFile(logName, os.O_RDWR|os.O_APPEND, 0644)
	} else {
		fl, err = os.Create(logName)
	}
	if err != nil {
		beego.Trace("Failed to init log file ->", err)
		return
	}
	beego.Info("Go Walker", VERSION)
	beego.SetLogger(log.New(fl, "", log.Ldate|log.Ltime))
}

func main() {
	beego.AppName = "Go Walker"
	beego.Info("Go Walker", VERSION)

	// Register routers.
	beego.Router("/", &controllers.HomeController{})
	beego.Router("/index", &controllers.IndexController{})
	beego.Router("/about", &controllers.AboutController{})
	beego.Router("/search", &controllers.SearchController{})
	beego.Router("/refresh", &controllers.RefreshController{})

	// For all unknown pages.
	beego.Router("/:all", &controllers.HomeController{})
	beego.Run()
}
