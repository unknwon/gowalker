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

// Go Walker Admin is the deamon process for Go Walker Server.
package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/astaxie/beego"
)

const (
	APP_VER = "0.0.1.0822"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Set App version and log level.
	// routers.AppVer = "v" + APP_VER

	if beego.AppConfig.String("runmode") == "pro" {
		beego.SetLevel(beego.LevelInfo)
		beego.Info("Product mode enabled")
		beego.Info("Go Walker Admin", APP_VER)

		os.Mkdir("../log", os.ModePerm)
		fw := beego.NewFileWriter("../log/admin", true)
		err := fw.StartLogger()
		if err != nil {
			beego.Critical("NewFileWriter ->", err)
		}
	}
}

func main() {
	beego.AppName = "Go Walker Admin"
	beego.Info("Go Walker Admin", APP_VER)

	// Register routers.

	// Register template functions.

	// For all unknown pages.

	// Deamon.
	go deamon()

	beego.Run()
}

func deamon() {
	beego.Info("Start deamon process")
	for {
		runServer()
	}
}

func runServer() {
	buf := new(bytes.Buffer)

	gwsPath := "../gw-server/gw-server"
	if runtime.GOOS == "windows" {
		gwsPath += ".exe"
	}
	gws := exec.Command(gwsPath)
	stdout, err := gws.StdoutPipe()
	if err != nil {
		beego.Error(err)
	}

	stderr, err := gws.StderrPipe()
	if err != nil {
		beego.Error(err)
	}

	err = gws.Start()
	if err != nil {
		beego.Error(err)
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(buf, stderr)
	gws.Wait()

	beego.Error(buf.String())
}
