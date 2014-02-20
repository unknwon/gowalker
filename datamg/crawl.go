package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Unknwon/com"
)

func main() {
	log.Println("201402191")
	client := &http.Client{}
	p, err := com.HttpGetBytes(client, "http://godoc.org/-/index", nil)
	if err != nil {
		log.Fatalf("Fail to load page: %v", err)
	}
	content := string(p)
	start := strings.Index(content, "<tbody>") + 7
	end := strings.Index(content, "</tbody>")
	content = content[start:end]

	pkgs := strings.Split(content, "<tr>")[1:]

	skipUntilIndex := 9052
	endWhenIndex := 12000
	for i, name := range pkgs {
		if i < skipUntilIndex {
			continue
		} else if i == endWhenIndex {
			break
		}
		name = strings.TrimSpace(name)[14:]
		end := strings.Index(name, "\">")
		name = name[:end]
		log.Printf("#%d %s", i, name)
		_, err = com.HttpGet(client, "https://gowalker.org/"+name, nil)
		if err != nil {
			log.Fatalf("Fail to load page: %v", err)
		}
		time.Sleep(0 * time.Second)
	}
}
