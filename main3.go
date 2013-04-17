package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	biFunc string // built-in functions
)

type contentWriter struct {
	indexContent bytes.Buffer
	// Constants
	constMap           map[string]constInfo
	constContent       bytes.Buffer
	singleConstContent bytes.Buffer
	// Variables
	varMap           map[string]varInfo
	varContent       bytes.Buffer
	singleVarContent bytes.Buffer
}

type constInfo struct {
	constType, constValue, comments string
}

type varInfo struct {
	varType, comments string
}

func main() {

	// Initialize built-in functions
	//biFunc = "len|"
	//walkDir("weixin")
	//fmt.Println("DONE")
}

func walkDir(srcDirPath string) {
	//fmt.Println("Handling path...", srcDirPath)
	// Initialize walker
	walker := new(contentWriter)
	walker.constMap = make(map[string]constInfo)
	walker.varMap = make(map[string]varInfo)
	walker.indexContent.WriteString("<h3 id=\"_index\" >Index</h3>\n<ul class=\"unstyled\" >\n")
	walker.indexContent.WriteString("$CONSTANTS\n$VARIABLES")

	// Open root directory
	dir, err := os.Open(srcDirPath)
	if err != nil {
		println(err)
	}
	defer dir.Close()

	// Get file info slice
	fis, err := dir.Readdir(0)
	if err != nil {
		println(err)
	}

	var curPath string
	for _, fi := range fis {
		// Append path
		curPath = srcDirPath + "/" + fi.Name()
		// Check if it is directory or file
		if fi.IsDir() {
			// Directory
			if !(fi.Name()[0] == '.') {
				walkDir(curPath)
			}
		} else {
			// File, only handle .go file
			if filepath.Ext(curPath) == ".go" {
				//fmt.Println("Handling file...", curPath)
				HandleFile(curPath, walker)
			}
		}
	}

	walker.indexContent.WriteString("\n</ul>")
	WriteContentToFile(srcDirPath, walker)
	// TODO:渲染文件
}

func HandleFile(filePath string, walker *contentWriter) {
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
			if len(line) > 2 && line[2] != ' ' {
				comments += " "
			}
			comments += line
			continue
		case len(line) > 3 && line[:4] == "var ": // Variables
			// TODO:类型为匿名struct的变量
			//fmt.Println(line)
			if walker.varContent.Len() == 0 {
				walker.varContent.WriteString("<h3 id=\"_variables\" >Variables</h3>\n")
			}
			// Check if it is singe line or group
			if line[4] == '(' {
				// Group
				if len(comments) > 0 {
					walker.varContent.WriteString("<p>" + strings.Replace(comments, "//", "", -1) + "</p>\n")
				}
				comments = ""
				// Add const content
				walker.varContent.WriteString("<pre class=\"pre-x-scrollable\" >\n var (\n")
				for {
					line, _ = buf.ReadString('\n')
					line = strings.TrimSpace(line)

					if line == ")" {
						break
					}

					ReadVars(pkgName, line, true, walker)
				}
				walker.varContent.WriteString(" )\n</pre>\n")
			}
			continue
		case len(line) > 5 && line[:6] == "const ": // Constants
			if walker.constContent.Len() == 0 {
				walker.constContent.WriteString("<h3 id=\"_constants\" >Constants</h3>\n")
			}
			// Check if it is singe line or group
			if line[6] == '(' {
				// Group
				if len(comments) > 0 {
					walker.constContent.WriteString("<p>" + strings.Replace(comments, "//", "", -1) + "</p>\n")
				}
				comments = ""
				// Add const content
				walker.constContent.WriteString("<pre class=\"pre-x-scrollable\" >\n const (\n")
				for {
					line, _ = buf.ReadString('\n')
					line = strings.TrimSpace(line)

					if line == ")" {
						break
					}

					ReadConsts(pkgName, line, true, walker)
				}
				walker.constContent.WriteString(" )\n</pre>\n")
			} else {
				// Singe line
				if walker.singleConstContent.Len() == 0 {
					walker.singleConstContent.WriteString("<pre class=\"pre-x-scrollable\" >\n const (\n")
				}

				if len(comments) > 0 {
					comments = strings.Replace(comments, "//", "", -1)
					walker.singleConstContent.WriteString(strings.Replace(comments, comments,
						"    <span class=\"com\" > // "+strings.TrimSpace(comments)+"</span>\n", 1))
				}
				comments = ""
				ReadConsts(pkgName, line[6:], false, walker)
			}
			continue
		case len(line) > 6 && line[0:7] == "package": // Package
			pkgName = line[8:] // Get package name
			comments = ""
			continue
		}
	}

	// Close file
	defer f.Close()
}

func ReadConsts(pkgName string, line string, isGroup bool, walker *contentWriter) {
	srcLine := line
	line = ReduceSpaces(line)
	// Check if it has comments
	comments := "?"
	comIndex := strings.IndexAny(line, "/")

	if comIndex > -1 {
		// Check if it is the comment line 
		if comIndex == 0 {
			srcLine = strings.Replace(srcLine, "//", "", -1)
			if isGroup {
				walker.constContent.WriteString(strings.Replace(srcLine, srcLine,
					"    <span class=\"com\" > // "+strings.TrimSpace(srcLine)+"</span>\n", 1))
			}
			return
		}
		comments = strings.TrimSpace(line[comIndex+2:])
		// non-comments line
		line = line[:comIndex-1]
	}

	// Get variable type and value
	varType, varValue := "?", ""
	// Devide line to two parts
	equalIndex := strings.IndexAny(line, "=")
	if equalIndex > -1 {
		varValue = line[equalIndex+2:]
		line = line[:equalIndex-1]
	}
	items := strings.Split(line, " ")
	// Get variable name
	varName := items[0]

	// Check if it has type
	if len(items) == 2 {
		varType = items[1]
	}

	walker.constMap[pkgName+"."+varName] =
		constInfo{varType, varValue, comments}

	// Check if const value has functions (only built-in)
	var isHasFunc bool
	var render string
	if len(varValue) > 0 {
		render, isHasFunc = CheckFunc(varValue)
	}

	srcLine = strings.Replace(srcLine, "//", "", -1)
	if isHasFunc {
		srcLine = strings.Replace(srcLine, varValue, render, 1)
	}

	srcLine = strings.Replace(srcLine, varName, "    <span id=\""+varName+"\" >"+varName+"</span> ", 1)
	if isGroup {
		if comIndex > -1 {
			walker.constContent.WriteString(strings.Replace(srcLine, comments, "<span class=\"com\" > // "+comments+"</span>\n", 1))
		} else {
			walker.constContent.WriteString(srcLine + "\n")
		}
	} else {
		if comIndex > -1 {
			walker.singleConstContent.WriteString(strings.Replace(srcLine, comments, "<span class=\"com\" > // "+comments+"</span>\n", 1))
		} else {
			walker.singleConstContent.WriteString(srcLine + "\n")
		}
	}
}

func ReadVars(pkgName string, line string, isGroup bool, walker *contentWriter) {
	srcLine := line
	line = ReduceSpaces(line)
	// Check if it has comments
	comments := "?"
	comIndex := strings.IndexAny(line, "/")

	if comIndex > -1 {
		// Check if it is the comment line 
		if comIndex == 0 {
			srcLine = strings.Replace(srcLine, "//", "", -1)
			if isGroup {
				walker.varContent.WriteString(strings.Replace(srcLine, srcLine,
					"    <span class=\"com\" > // "+strings.TrimSpace(srcLine)+"</span>\n", 1))
			}
			return
		}
		comments = strings.TrimSpace(line[comIndex+2:])
		// non-comments line
		line = line[:comIndex-1]
	}

	// Get variable type and value
	varType, varValue := "?", ""
	// Devide line to two parts
	equalIndex := strings.IndexAny(line, "=")
	if equalIndex > -1 {
		varValue = line[equalIndex+2:]
		line = line[:equalIndex-1]
	}
	items := strings.Split(line, " ")
	// Get variable name
	varName := items[0]

	// Check if it has type
	if len(items) == 2 {
		varType = items[1]
	}

	walker.varMap[pkgName+"."+varName] =
		varInfo{varType, comments}

	// Check if const value has functions (only built-in)
	var isHasFunc bool
	var render string
	if len(varValue) > 0 {
		render, isHasFunc = CheckFunc(varValue)
	}

	srcLine = strings.Replace(srcLine, "//", "", -1)
	if isHasFunc {
		srcLine = strings.Replace(srcLine, varValue, render, 1)
	}

	srcLine = strings.Replace(srcLine, varName, "    <span id=\""+varName+"\" >"+varName+"</span> ", 1)
	if isGroup {
		if comIndex > -1 {
			walker.varContent.WriteString(strings.Replace(srcLine, comments, "<span class=\"com\" > // "+comments+"</span>\n", 1))
		} else {
			walker.varContent.WriteString(srcLine + "\n")
		}
	} else {
		if comIndex > -1 {
			walker.singleVarContent.WriteString(strings.Replace(srcLine, comments, "<span class=\"com\" > // "+comments+"</span>\n", 1))
		} else {
			walker.singleVarContent.WriteString(srcLine + "\n")
		}
	}
}

func ReduceSpaces(srcStr string) string {
	var buf bytes.Buffer
	for i, _ := range srcStr {
		if srcStr[i] == ' ' {
			if i == 0 || srcStr[i-1] != ' ' {
				buf.WriteByte(srcStr[i])
			}
		} else {
			buf.WriteByte(srcStr[i])
		}
	}
	return buf.String()
}

func WriteContentToFile(filePath string, walker *contentWriter) {
	// Read head and bottom
	f, _ := os.Open("head.txt")
	fi, _ := f.Stat()
	head := make([]byte, fi.Size())
	f.Read(head)
	f.Close()
	f, _ = os.Open("bottom.txt")
	fi, _ = f.Stat()
	bottom := make([]byte, fi.Size())
	f.Read(bottom)
	f.Close()
	// Write to file
	f, _ = os.Create(filePath + "/index.html")
	prefix := strings.Repeat("../", len(strings.Split(filePath, "/")))
	headStr := strings.Replace(string(head), "$PREFIX", prefix, -1)
	headStr = strings.Replace(headStr, "$PKGNAME", filePath, 1)
	f.WriteString(headStr)

	// Index
	indexStr := walker.indexContent.String()
	// Check constants
	var indexTitle string
	if len(walker.constMap) > 0 {
		indexTitle = "<li><a href=\"#_constants\" >Constants</a></li>"
	}
	indexStr = strings.Replace(indexStr, "$CONSTANTS", indexTitle, 1)
	// Check variables
	indexTitle = ""
	if len(walker.varMap) > 0 {
		indexTitle = "<li><a href=\"#_variables\" >Variables</a></li>"
	}
	indexStr = strings.Replace(indexStr, "$VARIABLES", indexTitle, 1)

	f.WriteString(indexStr)

	// Write constants
	if len(walker.constMap) > 0 {
		f.Write(walker.constContent.Bytes())
		if walker.singleConstContent.Len() > 0 {
			walker.singleConstContent.WriteString(" )\n</pre>\n")
			f.Write(walker.singleConstContent.Bytes())
		}
	}
	// Write variables
	if len(walker.varMap) > 0 {
		f.Write(walker.varContent.Bytes())
		if walker.singleVarContent.Len() > 0 {
			walker.singleVarContent.WriteString(" )\n</pre>\n")
			f.Write(walker.singleVarContent.Bytes())
		}
	}

	f.Write(bottom)
	f.Close()
}

func CheckFunc(varValue string) (string, bool) {
	var isHasFunc bool
	oriValue := varValue
	varValue = strings.Replace(varValue, " (", "", -1)
	funcParts := strings.Split(varValue, "(")

	num := len(funcParts)
	if num > 1 {
		for i, v := range funcParts {
			if i < num-1 {
				isHasFunc = true
				items := strings.Split(v, " ")
				// Check if from other packages
				tags := strings.Split(items[len(items)-1], ".")
				if len(tags) > 1 {
					// From other packages
					fmt.Print("")
				} else {
					// Check if it is built-in function
					if strings.Index(biFunc, tags[0]+"|") > -1 {
						varValue = strings.Replace(oriValue, tags[0],
							"<a href=\"http://golang.org/pkg/builtin/#"+tags[0]+"\" >"+tags[0]+"</a>", -1)
					} else {
						varValue = strings.Replace(oriValue, tags[0],
							"<a href=\"#"+tags[0]+"\" >"+tags[0]+"</a>", -1)
					}
					// TODO:多个同名函数时会浪费资源多次替换
				}
			}
		}
	}
	return varValue, isHasFunc
}
