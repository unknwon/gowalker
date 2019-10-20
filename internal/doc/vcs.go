// Copyright 2012 Gary Burd
// Copyright 2013 Unknwon
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

package doc

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/unknwon/com"

	"github.com/unknwon/gowalker/internal/base"
	"github.com/unknwon/gowalker/internal/db"
)

// TODO: specify with command line flag
const repoRoot = "/tmp/gw"

func init() {
	os.RemoveAll(repoRoot)
}

var urlTemplates = []struct {
	re       *regexp.Regexp
	template string
	lineFmt  string
}{
	{
		regexp.MustCompile(`^git\.gitorious\.org/(?P<repo>[^/]+/[^/]+)$`),
		"https://gitorious.org/{repo}/blobs/{tag}/{dir}{0}",
		"#line%d",
	},
	{
		regexp.MustCompile(`^camlistore\.org/r/p/(?P<repo>[^/]+)$`),
		"http://camlistore.org/code/?p={repo}.git;hb={tag};f={dir}{0}",
		"#l%d",
	},
}

// lookupURLTemplate finds an expand() template, match map and line number
// format for well known repositories.
func lookupURLTemplate(repo, dir, tag string) (string, map[string]string, string) {
	if strings.HasPrefix(dir, "/") {
		dir = dir[1:] + "/"
	}
	for _, t := range urlTemplates {
		if m := t.re.FindStringSubmatch(repo); m != nil {
			match := map[string]string{
				"dir": dir,
				"tag": tag,
			}
			for i, name := range t.re.SubexpNames() {
				if name != "" {
					match[name] = m[i]
				}
			}
			return t.template, match, t.lineFmt
		}
	}
	return "", nil, ""
}

type vcsCmd struct {
	schemes  []string
	download func([]string, string, string) (string, string, error)
}

var vcsCmds = map[string]*vcsCmd{
	"git": &vcsCmd{
		schemes:  []string{"http", "https", "git"},
		download: downloadGit,
	},
}

var lsremoteRe = regexp.MustCompile(`(?m)^([0-9a-f]{40})\s+refs/(?:tags|heads)/(.+)$`)

func downloadGit(schemes []string, repo, savedEtag string) (string, string, error) {
	var p []byte
	var scheme string
	for i := range schemes {
		cmd := exec.Command("git", "ls-remote", "--heads", "--tags", schemes[i]+"://"+repo+".git")
		log.Println(strings.Join(cmd.Args, " "))
		var err error
		p, err = cmd.Output()
		if err == nil {
			scheme = schemes[i]
			break
		}
	}

	if scheme == "" {
		return "", "", com.NotFoundError{"VCS not found"}
	}

	tags := make(map[string]string)
	for _, m := range lsremoteRe.FindAllSubmatch(p, -1) {
		tags[string(m[2])] = string(m[1])
	}

	tag, commit, err := bestTag(tags, "master")
	if err != nil {
		return "", "", err
	}

	etag := scheme + "-" + commit

	if etag == savedEtag {
		return "", "", ErrPackageNotModified
	}

	dir := path.Join(repoRoot, repo+".git")
	p, err = ioutil.ReadFile(path.Join(dir, ".git/HEAD"))
	switch {
	case err != nil:
		if err := os.MkdirAll(dir, 0777); err != nil {
			return "", "", err
		}
		cmd := exec.Command("git", "clone", scheme+"://"+repo, dir)
		log.Println(strings.Join(cmd.Args, " "))
		if err := cmd.Run(); err != nil {
			return "", "", err
		}
	case string(bytes.TrimRight(p, "\n")) == commit:
		return tag, etag, nil
	default:
		cmd := exec.Command("git", "fetch")
		log.Println(strings.Join(cmd.Args, " "))
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return "", "", err
		}
	}

	cmd := exec.Command("git", "checkout", "--detach", "--force", commit)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return "", "", err
	}

	return tag, etag, nil
}

var (
	vcsPattern       = regexp.MustCompile(`^(?P<repo>(?:[a-z0-9.\-]+\.)+[a-z0-9.\-]+(?::[0-9]+)?/[A-Za-z0-9_.\-/]*?)\.(?P<vcs>bzr|git|hg|svn)(?P<dir>/[A-Za-z0-9_.\-/]*)?$`)
	gopkgPathPattern = regexp.MustCompile(`^/(?:([a-zA-Z0-9][-a-zA-Z0-9]+)/)?([a-zA-Z][-.a-zA-Z0-9]*)\.((?:v0|v[1-9][0-9]*)(?:\.0|\.[1-9][0-9]*){0,2})(?:\.git)?((?:/[a-zA-Z0-9][-.a-zA-Z0-9]*)*)$`)
)

func getVCSDoc(match map[string]string, etagSaved string) (*Package, error) {
	if strings.HasPrefix(match["importPath"], "golang.org/x/") {
		match["owner"] = "golang"
		match["repo"] = path.Dir(strings.TrimPrefix(match["importPath"], "golang.org/x/"))
		return getGitHubDoc(match, etagSaved)
	} else if strings.HasPrefix(match["importPath"], "gopkg.in/") {
		m := gopkgPathPattern.FindStringSubmatch(strings.TrimPrefix(match["importPath"], "gopkg.in"))
		if m == nil {
			return nil, fmt.Errorf("unsupported gopkg.in import path: %s", match["importPath"])
		}
		user := m[1]
		repo := m[2]
		if len(user) == 0 {
			user = "go-" + repo
		}
		match["owner"] = user
		match["repo"] = repo
		match["tag"] = m[3]
		return getGitHubDoc(match, etagSaved)
	}

	cmd := vcsCmds[match["vcs"]]
	if cmd == nil {
		return nil, com.NotFoundError{com.Expand("VCS not supported: {vcs}", match)}
	}

	scheme := match["scheme"]
	if scheme == "" {
		i := strings.Index(etagSaved, "-")
		if i > 0 {
			scheme = etagSaved[:i]
		}
	}

	schemes := cmd.schemes
	if scheme != "" {
		for i := range cmd.schemes {
			if cmd.schemes[i] == scheme {
				schemes = cmd.schemes[i : i+1]
				break
			}
		}
	}

	// Download and checkout.

	tag, _, err := cmd.download(schemes, match["repo"], etagSaved)
	if err != nil {
		return nil, err
	}

	// Find source location.

	urlTemplate, urlMatch, lineFmt := lookupURLTemplate(match["repo"], match["dir"], tag)

	// Slurp source files.

	d := path.Join(repoRoot, com.Expand("{repo}.{vcs}", match), match["dir"])
	f, err := os.Open(d)
	if err != nil {
		if os.IsNotExist(err) {
			err = com.NotFoundError{err.Error()}
		}
		return nil, err
	}
	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	// Get source file data.
	var files []com.RawFile
	for _, fi := range fis {
		if fi.IsDir() || !base.IsDocFile(fi.Name()) {
			continue
		}
		b, err := ioutil.ReadFile(path.Join(d, fi.Name()))
		if err != nil {
			return nil, err
		}
		files = append(files, &Source{
			SrcName:   fi.Name(),
			BrowseUrl: com.Expand(urlTemplate, urlMatch, fi.Name()),
			SrcData:   b,
		})
	}

	// Start generating data.
	w := &Walker{
		LineFmt: lineFmt,
		Pdoc: &Package{
			PkgInfo: &db.PkgInfo{
				ImportPath: match["importPath"],
			},
		},
	}

	srcs := make([]*Source, 0, len(files))
	for _, f := range files {
		s, _ := f.(*Source)
		srcs = append(srcs, s)
	}

	return w.Build(&WalkRes{
		WalkDepth: WD_All,
		WalkType:  WT_Memory,
		WalkMode:  WM_All,
		Srcs:      srcs,
	})
}

var defaultTags = map[string]string{"git": "master", "hg": "default", "svn": "trunk"}

func bestTag(tags map[string]string, defaultTag string) (string, string, error) {
	if commit, ok := tags["go1"]; ok {
		return "go1", commit, nil
	}
	if commit, ok := tags[defaultTag]; ok {
		return defaultTag, commit, nil
	}
	return "", "", com.NotFoundError{"Tag or branch not found."}
}

// checkDir checks if directory has been appended to slice.
func checkDir(dir string, dirs []string) bool {
	for _, d := range dirs {
		if dir == d {
			return true
		}
	}
	return false
}

// Only support .zip.
func getRepoByArchive(match map[string]string, downloadPath string) (bool, string, []com.RawFile, []string, error) {
	stdout, _, err := com.ExecCmd("curl", downloadPath)
	if err != nil {
		return false, "", nil, nil, err
	}
	p := []byte(stdout)

	r, err := zip.NewReader(bytes.NewReader(p), int64(len(p)))
	if err != nil {
		return false, "", nil, nil, errors.New(downloadPath + " -> new zip: " + err.Error())
	}

	if len(r.File) == 0 {
		return false, "", nil, nil, nil
	}

	nameLen := strings.Index(r.File[0].Name, "/")
	dirPrefix := match["dir"]
	if len(dirPrefix) != 0 {
		dirPrefix = dirPrefix[1:] + "/"
	}
	preLen := len(dirPrefix)
	isGoPro := false

	// for k, v := range match {
	// 	println(k, v)
	// }
	comment := r.Comment

	files := make([]com.RawFile, 0, 5)
	dirs := make([]string, 0, 5)
	for _, f := range r.File {
		fileName := f.Name[nameLen+1:]
		// Skip directories and files in wrong directories, get them later.
		if strings.HasSuffix(fileName, "/") || !strings.HasPrefix(fileName, dirPrefix) {
			continue
		}
		//fmt.Println(fileName)

		// Get files and check if directories have acceptable files.
		if d, fn := path.Split(fileName); base.IsDocFile(fn) &&
			base.FilterDirName(d) {
			// Check if it's a Go file.
			if !isGoPro && strings.HasSuffix(fn, ".go") {
				isGoPro = true
			}

			// Check if file is in the directory that is corresponding to import path.
			if d == dirPrefix {
				// Yes.
				if !isGoPro && strings.HasSuffix(fn, ".go") {
					isGoPro = true
				}
				// Get file from archive.
				rc, err := f.Open()
				if err != nil {
					return isGoPro, comment, files, dirs,
						errors.New(downloadPath + " -> open file: " + err.Error())
				}

				p := make([]byte, f.FileInfo().Size())
				rc.Read(p)
				if err != nil {
					return isGoPro, comment, files, dirs,
						errors.New(downloadPath + " -> read file: " + err.Error())
				}
				//fmt.Println(com.Expand(match["browserUrlTpl"], match, fn))
				files = append(files, &Source{
					SrcName:   fn,
					BrowseUrl: com.Expand(match["browserUrlTpl"], match, fn),
					RawSrcUrl: com.Expand(match["rawSrcUrlTpl"], match, fileName[preLen:]),
					SrcData:   p,
				})

			} else {
				sd, _ := path.Split(d[preLen:])
				sd = strings.TrimSuffix(sd, "/")
				if !checkDir(sd, dirs) {
					dirs = append(dirs, sd)
				}
			}
		}
	}
	return isGoPro, comment, files, dirs, nil
}
