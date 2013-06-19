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
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/Unknwon/gowalker/routers"
	"github.com/Unknwon/gowalker/utils"
	"github.com/astaxie/beego"
)

const (
	VERSION = "0.3.3.0619" // Application version.
)

var (
	logTicker   *time.Ticker
	logFileName string
	logFile     *os.File
)

func init() {
	// Try to have highest performance.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Set application log level.
	if beego.AppConfig.String("runmode") == "pro" {
		beego.SetLevel(beego.LevelInfo)
	}

	// ----- Initialize log file -----
	os.Mkdir("./log", os.ModePerm)

	// Start log ticker.
	logTicker = time.NewTicker(time.Minute)
	go logTickerCheck(logTicker.C)

	beego.Info("Go Walker", VERSION)
	setLogger()
}

// logTickerCheck checks for log files.
// Because we need to record log in different files for different time period.
func logTickerCheck(logChan <-chan time.Time) {
	for {
		<-logChan
		setLogger()
	}
}

// setLogger sets corresponding log file for beego.Logger.
func setLogger() {
	// Compute log file name as format '<year>-<month>-<day>.txt', eg.'2013-06-19.txt'.
	logName := fmt.Sprintf("./log/%04d-%02d-%02d.txt",
		time.Now().Year(), time.Now().Month(), time.Now().Day())
	// Check if need to create new log file.
	if logName == logFileName {
		return
	}

	logFileName = logName
	// Open or create log file.
	var err error
	if utils.IsExist(logName) {
		logFile, err = os.OpenFile(logName, os.O_RDWR|os.O_APPEND, 0644)
	} else {
		logFile, err = os.Create(logName)
	}
	if err != nil {
		beego.Critical("Failed to init log file ->", err)
		return
	}

	beego.SetLogger(log.New(logFile, "", log.Ldate|log.Ltime))
}

func main() {
	beego.AppName = "Go Walker"
	beego.Info("Go Walker", VERSION)

	// Register routers.
	beego.Router("/", &routers.HomeRouter{})

	beego.Run()
}
