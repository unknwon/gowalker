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
	"net/http"
	"runtime"
	"strings"

	"github.com/Unknwon/log"
	"github.com/Unknwon/macaron"
	"github.com/macaron-contrib/i18n"
	"github.com/macaron-contrib/pongo2"
	"github.com/macaron-contrib/session"

	"github.com/Unknwon/gowalker/modules/middleware"
	"github.com/Unknwon/gowalker/modules/setting"
	"github.com/Unknwon/gowalker/routers"
	"github.com/Unknwon/gowalker/routers/apiv1"
)

const APP_VER = "1.4.3.0325"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	setting.AppVer = APP_VER
}

// newMacaron initializes Macaron instance.
func newMacaron() *macaron.Macaron {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public",
		macaron.StaticOptions{
			SkipLogging: setting.ProdMode,
		},
	))
	m.Use(macaron.Static("raw",
		macaron.StaticOptions{
			Prefix:      "raw",
			SkipLogging: setting.ProdMode,
		}))
	m.Use(pongo2.Pongoer(pongo2.Options{
		IndentJSON: !setting.ProdMode,
	}))
	m.Use(i18n.I18n())
	m.Use(session.Sessioner())
	m.Use(middleware.Contexter())
	return m
}

func main() {
	log.Info("Go Walker %s", APP_VER)
	log.Info("Run Mode: %s", strings.Title(macaron.Env))

	m := newMacaron()
	m.Get("/", routers.Home)
	m.Get("/search", routers.Search)
	m.Get("/search/json", routers.SearchJSON)

	m.Group("/api", func() {
		m.Group("/v1", func() {
			m.Get("/badge", apiv1.Badge)
		})
	})

	m.Get("/robots.txt", func() string {
		return `User-agent: *
Disallow: /search`
	})
	m.Get("/*", routers.Docs)

	listenAddr := fmt.Sprintf("0.0.0.0:%d", setting.HTTPPort)
	log.Info("Listen: http://%s", listenAddr)
	if err := http.ListenAndServe(listenAddr, m); err != nil {
		log.FatalD(4, "Fail to start server: %v", err)
	}
}
