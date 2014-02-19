package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/coocood/qbs"
	_ "github.com/mattn/go-sqlite3"
)

type PkgInfo struct {
	Id    int64
	Path  string
	Views int64
}

func main2() {
	qbs.Register("sqlite3", "gowalker.db", "", qbs.NewSqlite3())

	q, _ := qbs.GetQbs()
	defer q.Close()

	pinfo := new(PkgInfo)
	q.Iterate(pinfo, func() error {

		if pinfo.Id > 198 {
			c := &http.Client{}
			req, err := http.NewRequest("GET",
				"http://gowalker.org/add?path="+pinfo.Path+"&views="+fmt.Sprintf("%d", pinfo.Views), nil)
			if err != nil {
				return err
			}
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
			resp, err := c.Do(req)
			defer resp.Body.Close()

			fmt.Println(pinfo.Id, pinfo.Path)
			if pinfo.Id >= 10000 {
				return errors.New("FINISH")
			}
		}
		return nil
	})
}
