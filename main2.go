package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	LineBreak = "\n"
	VarMap    map[string]TypeStruct
)

type TypeStruct struct {
	Addr, Comment, VarType, VarValue string
}

func main() {
	VarMap = make(map[string]TypeStruct)
	walkDir("weixin")
	fmt.Println(VarMap)
}

func walkDir(srcDirPath string) {
	// Open root directory
	dir, err := os.Open("weixin")
	if err != nil {
		println(err)
	}
	defer dir.Close()

	// Get file info slice
	fis, err := dir.Readdir(0)
	if err != nil {
		println(err)
	}

	for _, fi := range fis {
		// Append path
		curPath := srcDirPath + "/" + fi.Name()
		// Check if it is directory or file
		if fi.IsDir() {
			// Directory
			println("Handling path...", curPath)
			walkDir(curPath)
		} else {
			// File, only handle .go file
			if filepath.Ext(curPath) == ".go" {
				println("Handling file...", curPath)
				HandleFile(curPath)
			}
		}
	}
}

func HandleFile(filePath string) {
	// Read file by filename
	var f *os.File
	var err error
	if f, err = os.Open(filePath); err != nil {
		println(err)
	}

	// Create buffer reader
	buf := bufio.NewReader(f)

	var pkgName string
	var comments string
	// Parse line-by-line
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		//println(line)

		if err != nil {
			// Unexpected error
			if err != io.EOF {
				println(err)
				break
			}

			// Reached end of file, if nothing to read then break,
			// otherwise handle the last line.
			if len(line) == 0 {
				break
			}
		}

		// switch written for readability (not performance)
		switch {
		case len(line) < 2: // Empty line
			continue
		case line[0:2] == "//": // Comments
			// Append comments
			if len(comments) == 0 {
				comments = line
			} else {
				comments += LineBreak + line
			}
			continue
		case len(line) > 5 && line[0:6] == "func (": // Method
			// Skip import for now
			continue
		case len(line) > 3 && line[0:4] == "func": // Function
			// Skip import for now
			pkgName += pkgName
			continue
		case line[0:3] == "var": // Variable
			// Check if it uses group
			if line[4] == '(' {
				for {
					line, _ = buf.ReadString('\n')
					line = strings.TrimSpace(line)

					if line == ")" {
						break
					}

					ReadVars(pkgName, line)
				}

			} else {
				ReadVars(pkgName, line[4:])
			}
			continue
		case line[0:5] == "const": // Constants
			// Check if it uses group
			if line[6] == '(' {
				for {
					line, _ = buf.ReadString('\n')
					line = strings.TrimSpace(line)

					if line == ")" {
						break
					}

					ReadConsts(pkgName, line)
				}

			} else {
				ReadConsts(pkgName, line[6:])
			}
			continue
		case line[0:6] == "import" && line[:1] == "\"": // Import
			// Skip import for now
			continue
		case line[0:7] == "package": // Package
			pkgName = line[8:12] // Get package name
			println("In package: ", pkgName)
			continue
		}
	}

	defer f.Close()
}

func ReadConsts(pkgName string, line string) {
	// Check if it has comments
	comments := "?"
	comIndex := strings.IndexAny(line, "/")
	var items []string
	if comIndex > -1 {
		items = strings.Split(line[:comIndex-1], " ")
		comments = strings.TrimSpace(line[comIndex+2:])
	} else {
		items = strings.Split(line, " ")
	}

	// Get variable name
	varName := items[0]

	// Get variable type and value
	varType, varValue := "?", ""
	// Check if it has type
	if len(items) == 4 {
		// It has type
		varType = items[1]
		varValue = items[3]
	} else {
		varValue = items[2]
	}

	VarMap[pkgName+"."+varName] =
		TypeStruct{pkgName + "#" + varName, comments, varType, varValue}
}

func ReadVars(pkgName string, line string) {
	// Check if it has comments
	comments := "?"
	comIndex := strings.IndexAny(line, "/")
	var items []string
	if comIndex > -1 {
		items = strings.Split(line[:comIndex-1], " ")
		comments = strings.TrimSpace(line[comIndex+2:])
	} else {
		items = strings.Split(line, " ")
	}

	// Get variable name
	varName := items[0]

	// Get variable type and value
	varType, varValue := "?", ""
	// Check if it has type
	if len(items) == 4 {
		// It has type
		varType = items[1]
		varValue = items[3]
	} else {
		varValue = items[2]
	}

	VarMap[pkgName+"."+varName] =
		TypeStruct{pkgName + "#" + varName, comments, varType, varValue}
}
