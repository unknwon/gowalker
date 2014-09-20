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

// Go Walker is a server that generates Go projects API documentation on the fly.
package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/Unknwon/macaron"
	"github.com/macaron-contrib/i18n"
	"github.com/macaron-contrib/session"
	"github.com/macaron-contrib/toolbox"

	"github.com/Unknwon/gowalker/models"
	"github.com/Unknwon/gowalker/modules/base"
	"github.com/Unknwon/gowalker/modules/log"
	"github.com/Unknwon/gowalker/modules/middleware"
	"github.com/Unknwon/gowalker/modules/setting"
	"github.com/Unknwon/gowalker/routers"
)

const (
	APP_VER = "1.1.0.0919"
)

func catchExit() {
	sigTerm := syscall.Signal(15)
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, sigTerm)

	for {
		switch <-sig {
		case os.Interrupt, sigTerm:
			fmt.Println()
			log.Warn("INTERRUPT SIGNAL DETECTED!!!")
			routers.FlushCache()
			log.Warn("[WARN] READY TO EXIT")
			os.Exit(0)
		}
	}
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	setting.AppVer = APP_VER
}

// newMacaron initializes Macaron instance.
func newMacaron() *macaron.Macaron {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("static",
		macaron.StaticOptions{
			SkipLogging: !setting.DisableRouterLog,
		},
	))
	m.Use(macaron.Static("public",
		macaron.StaticOptions{
			Prefix:      "public",
			SkipLogging: !setting.DisableRouterLog,
		},
	))
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:  "templates",
		Funcs:      []template.FuncMap{base.TemplateFuncs},
		IndentJSON: macaron.Env != macaron.PROD,
	}))
	m.Use(i18n.I18n(i18n.Options{
		Langs:    setting.Langs,
		Names:    setting.Names,
		Redirect: true,
	}))
	m.Use(session.Sessioner())
	m.Use(toolbox.Toolboxer(m, toolbox.Options{
		HealthCheckFuncs: []*toolbox.HealthCheckFuncDesc{
			&toolbox.HealthCheckFuncDesc{
				Desc: "Database connection",
				Func: models.Ping,
			},
		},
	}))
	m.Use(middleware.Contexter())
	return m
}

func main() {
	log.Info("Go Walker %s", APP_VER)
	log.Info("Run Mode: %s", strings.Title(macaron.Env))

	go catchExit()

	m := newMacaron()

	// // Register routers.
	// beego.Router("/", &routers.HomeRouter{})
	// beego.Router("/refresh", &routers.RefreshRouter{})
	// beego.Router("/search", &routers.SearchRouter{})
	// beego.Router("/index", &routers.IndexRouter{})
	// // beego.Router("/label", &routers.LabelsRouter{})
	// beego.Router("/function", &routers.FuncsRouter{})
	// beego.Router("/example", &routers.ExamplesRouter{})
	// beego.Router("/about", &routers.AboutRouter{})

	// beego.Router("/api/docs", &routers.ApiRouter{}, "get:Docs")
	// beego.Router("/api/v1/badge", &routers.ApiRouter{}, "get:Badge")
	// beego.Router("/api/v1/search", &routers.ApiRouter{}, "get:Search")
	// beego.Router("/api/v1/refresh", &routers.ApiRouter{}, "get:Refresh")
	// beego.Router("/api/v1/pkginfo", &routers.ApiRouter{}, "get:PkgInfo")

	// // Register template functions.
	// beego.AddFuncMap("isHasEleS", isHasEleS)
	// beego.AddFuncMap("isHasEleE", isHasEleE)
	// beego.AddFuncMap("isNotEmptyS", isNotEmptyS)

	// // "robot.txt"
	// beego.Router("/robots.txt", &routers.RobotRouter{})

	// // For all unknown pages.
	// beego.Router("/:all", &routers.HomeRouter{})

	listenAddr := "0.0.0.0:" + setting.HttpPort
	log.Info("Listen: http://%s", listenAddr)
	if err := http.ListenAndServe(listenAddr, m); err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
}
