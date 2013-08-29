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

// Go Walker Local is the local version of Go Walker as an alternative of godoc.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/Unknwon/com"
	"github.com/Unknwon/gowalker/utils"
)

const (
	APP_VER = "0.0.2.0824"
)

var (
	srcPath  string
	httpPort = 8082
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	com.ColorLog("[INFO] Go Walker Local v%s.\n", APP_VER)

	var err error
	srcPath, err = utils.GetAppPath("github.com/Unknwon/gowalker/gw-local", "conf")
	if err != nil {
		com.ColorLog("[ERRO] Cannot assign 'srcPath'[ %s ]\n", err)
		return
	}

	com.ColorLog("[INFO] File server( %s )\n", srcPath)

	// Get 'args'.
	args := os.Args[1:]
	if len(args) > 0 {
		hp, err := strconv.Atoi(args[0])
		if err == nil {
			httpPort = hp
		}
	}

	http.HandleFunc("/", serverHome)
	http.HandleFunc("/index", serverIndex)
	http.HandleFunc("/funcs", serverFuncs)
	http.HandleFunc("/about", serverAbout)

	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(srcPath+"static/"))))

	http.ListenAndServe(":"+fmt.Sprint(httpPort), nil)
}

func serverHome(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Home page")
}

func serverIndex(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Index")
}

func serverFuncs(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Funcs")
}

func serverAbout(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "About")
}
