// Copyright 2015 Unknwon
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

package db

import (
	"fmt"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/robfig/cron"
	log "gopkg.in/clog.v1"
	"xorm.io/core"

	"github.com/unknwon/gowalker/internal/setting"
)

var x *xorm.Engine

func init() {
	sec := setting.Cfg.Section("database")
	var err error
	x, err = xorm.NewEngine("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		sec.Key("USER").String(),
		sec.Key("PASSWD").String(),
		sec.Key("HOST").String(),
		sec.Key("NAME").String()))
	if err != nil {
		log.Fatal(2, "Failed to init new engine: %v", err)
	}
	x.SetMapper(core.GonicMapper{})

	if err = x.Sync(new(PkgInfo), new(PkgRef), new(JSFile)); err != nil {
		log.Fatal(2, "Failed to sync database: %v", err)
	}

	numTotalPackages, _ = x.Count(new(PkgInfo))
	c := cron.New()
	if err = c.AddFunc("@every 1m", RefreshNumTotalPackages); err != nil {
		log.Fatal(2, "Failed to add func: %v", err)
	} else if err = c.AddFunc("@every 1m", DistributeJSFiles); err != nil {
		log.Fatal(2, "Failed to add func: %v", err)
	} else if err = c.AddFunc("@every 5m", RecycleJSFiles); err != nil {
		log.Fatal(2, "Failed to add func: %v", err)
	}
	c.Start()

	time.AfterFunc(5*time.Second, DistributeJSFiles)
	time.AfterFunc(10*time.Second, RecycleJSFiles)
}

// NOTE: Must be operated atomically
var numTotalPackages int64

func NumTotalPackages() int64 {
	return atomic.LoadInt64(&numTotalPackages)
}
