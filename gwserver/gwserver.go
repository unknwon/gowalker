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

// Go Walker Server generates Go projects API documentation and Hacker View on the fly.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Unknwon/com"
	"github.com/Unknwon/gowalker/doc"
	"github.com/Unknwon/gowalker/gwserver/routers"
	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/utils"
	"github.com/Unknwon/hv"
	"github.com/astaxie/beego"
	"github.com/beego/i18n"
)

const (
	APP_VER = "1.0.0.0929"
)

// We have to call a initialize function manully
// because we use `bee bale` to pack static resources
// and we cannot make sure that which init() execute first.
func initialize() {
	// Load configuration, set app version and log level.
	err := utils.LoadConfig("conf/app.ini")
	if err != nil {
		panic("Fail to load configuration file: " + err.Error())
	}
	err = i18n.SetMessage("conf/message.ini")
	if err != nil {
		panic("Fail to set message file: " + err.Error())
	}

	// Initialize data.
	models.InitDb()
	routers.InitRouter()

	// Trim 4th part.
	routers.AppVer = strings.Join(strings.Split(APP_VER, ".")[:3], ".")

	beego.AppName = utils.Cfg.MustValue("beego", "app_name")
	beego.RunMode = utils.Cfg.MustValue("beego", "run_mode")
	beego.HttpPort = utils.Cfg.MustInt("beego", "http_port_"+beego.RunMode)

	routers.IsBeta = utils.Cfg.MustBool("server", "beta")
	routers.IsProMode = beego.RunMode == "pro"
	if routers.IsProMode {
		beego.SetLevel(beego.LevelInfo)
		beego.Info("Product mode enabled")
		beego.Info(beego.AppName, APP_VER)

		os.Mkdir("../log", os.ModePerm)
		beego.BeeLogger.SetLogger("file", "../log/server")
	}

	doc.SetGithubCredentials(utils.Cfg.MustValue("github", "client_id"),
		utils.Cfg.MustValue("github", "client_secret"))
}

func catchExit() {
	sigTerm := syscall.Signal(15)
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, sigTerm)

	for {
		switch <-sig {
		case os.Interrupt, sigTerm:
			fmt.Println()
			com.ColorLog("[WARN] INTERRUPT SIGNAL DETECTED!!!\n")
			routers.FlushCache()
			com.ColorLog("[WARN] READY TO EXIT\n")
			os.Exit(0)
		}
	}
}

func main() {
	initialize()
	go catchExit()

	beego.Info(beego.AppName, APP_VER)

	// Register routers.
	beego.Router("/", &routers.HomeRouter{})
	// beego.Router("/refresh", &routers.RefreshRouter{})
	beego.Router("/search", &routers.SearchRouter{})
	// beego.Router("/index", &routers.IndexRouter{})
	// beego.Router("/label", &routers.LabelsRouter{})
	// beego.Router("/function", &routers.FuncsRouter{})
	// beego.Router("/example", &routers.ExamplesRouter{})
	// beego.Router("/about", &routers.AboutRouter{})

	// Register template functions.
	beego.AddFuncMap("i18n", i18n.Tr)
	beego.AddFuncMap("isHasEleS", isHasEleS)
	beego.AddFuncMap("isHasEleE", isHasEleE)
	beego.AddFuncMap("isNotEmptyS", isNotEmptyS)

	// "robot.txt"
	beego.Router("/robots.txt", &routers.RobotRouter{})

	// For all unknown pages.
	beego.Router("/:all", &routers.HomeRouter{})
	beego.Run()
}

func isHasEleS(s []string) bool {
	if len(s) == 1 && len(s[0]) == 0 {
		return false
	}
	return len(s) > 0
}

func isHasEleE(s []*hv.Example) bool {
	return len(s) > 0
}

func isNotEmptyS(s string) bool {
	return len(s) > 0
}
