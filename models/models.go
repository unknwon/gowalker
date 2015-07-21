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

package models

import (
	"fmt"

	"github.com/Unknwon/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/robfig/cron"

	"github.com/Unknwon/gowalker/modules/setting"
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
		log.FatalD(4, "Fail to init new engine: %v", err)
	}
	x.SetLogger(nil)
	x.SetMapper(core.GonicMapper{})

	if err = x.Sync(new(PkgInfo), new(PkgRef)); err != nil {
		log.FatalD(4, "Fail to sync database: %v", err)
	}

	numOfPackages, _ = x.Count(new(PkgInfo))
	c := cron.New()
	c.AddFunc("@every 5m", func() {
		numOfPackages, _ = x.Count(new(PkgInfo))
	})
	c.Start()
}

var numOfPackages int64

func NumOfPackages() int64 {
	return numOfPackages
}
