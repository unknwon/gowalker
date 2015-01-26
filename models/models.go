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

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"

	"github.com/Unknwon/gowalker/modules/log"
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
		log.Fatal(4, "Fail to init new engine: %v", err)
	} else if err = x.Sync(new(PkgInfo)); err != nil {
		log.Fatal(4, "Fail to sync database: %v", err)
	}
}
