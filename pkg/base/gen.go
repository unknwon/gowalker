// Copyright 2014 The Go Authors. All rights reserved.
// Copyright 2015 Unknwon
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"go/format"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/unknwon/com"
)

const (
	goRepoPath = 1 << iota
	goSubrepoPath
	gaeRepoPath
	packagePath
)

var tmpl = template.Must(template.New("").Parse(`// Created by go generate; DO NOT EDIT
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

const (
    goRepoPath = {{.goRepoPath}}
    goSubrepoPath = {{.goSubrepoPath}}
    gaeRepoPath = {{.gaeRepoPath}}
    packagePath = {{.packagePath}}
)

var pathFlags = map[string]int{
{{range $k, $v := .pathFlags}}{{printf "%q" $k}}: {{$v}},
{{end}} }

func PathFlag(path string) int {
	return pathFlags[path]
}

func NumOfPathFlags() int {
	return len(pathFlags)
}

var paths []string

func init() {
	paths = make([]string, 0, len(pathFlags))
	for k := range pathFlags {
		paths = append(paths, k)
	}
}

func Paths() []string {
	return paths
}

var validTLDs = map[string]bool{
{{range  $v := .validTLDs}}{{printf "%q" $v}}: true,
{{end}} }
`))

var output = flag.String("output", "data.go", "file name to write")

type gitTree struct {
	Tree []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	} `json:"tree"`
}

func getGitTree(url string) *gitTree {
	p, err := com.HttpGetBytes(&http.Client{}, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	var t gitTree
	if err = json.Unmarshal(p, &t); err != nil {
		log.Fatal(err)
	}
	return &t
}

// getPathFlags builds map of standard/core repository path flags.
func getPathFlags() map[string]int {
	cmd := exec.Command("go", "list", "std")
	p, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	pathFlags := map[string]int{
		"builtin": packagePath | goRepoPath,
		"C":       packagePath,
	}
	for _, path := range strings.Fields(string(p)) {
		if strings.HasPrefix(path, "cmd/") || strings.HasPrefix(path, "vendor/") {
			continue
		}
		pathFlags[path] |= packagePath | goRepoPath
		for {
			i := strings.LastIndex(path, "/")
			if i < 0 {
				break
			}
			path = path[:i]
			pathFlags[path] |= goRepoPath
		}
	}

	// Get GAE repository path flags.
	t := getGitTree("https://api.github.com/repos/golang/appengine/git/trees/master?recursive=1")
	pathFlags["appengine"] |= packagePath | gaeRepoPath
	for _, blob := range t.Tree {
		if blob.Type != "tree" {
			continue
		}
		pathFlags["appengine/"+blob.Path] |= packagePath | gaeRepoPath
	}
	return pathFlags
}

// getValidTLDs gets and returns list of valid TLDs.
func getValidTLDs() (validTLDs []string) {
	p, err := com.HttpGetBytes(&http.Client{}, "http://data.iana.org/TLD/tlds-alpha-by-domain.txt", nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range strings.Split(string(p), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		validTLDs = append(validTLDs, "."+strings.ToLower(line))
	}
	return validTLDs
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("gen: ")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("usage: decgen [--output filename]")
	}

	// Generate output.
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]interface{}{
		"output":        *output,
		"goRepoPath":    goRepoPath,
		"goSubrepoPath": goSubrepoPath,
		"gaeRepoPath":   gaeRepoPath,
		"packagePath":   packagePath,
		"pathFlags":     getPathFlags(),
		"validTLDs":     getValidTLDs(),
	})
	if err != nil {
		log.Fatal("template error:", err)
	}
	source, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal("source format error:", err)
	}
	fd, err := os.Create(*output)
	_, err = fd.Write(source)
	if err != nil {
		log.Fatal(err)
	}
}
