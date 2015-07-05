// Copyright 2014 Unknwon
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

package base

import (
	"path/filepath"
	"strings"

	"github.com/Unknwon/i18n"
	"github.com/Unknwon/log"
	"gopkg.in/fsnotify.v1"

	"github.com/Unknwon/gowalker/modules/setting"
)

func monitorI18nLocale() {
	log.Info("Monitor i18n locale files enabled")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.FatalD(4, "Fail to init locale watcher: %v", err)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				switch filepath.Ext(event.Name) {
				case ".ini":
					if err := i18n.ReloadLangs(); err != nil {
						log.ErrorD(4, "Fail to relaod locale file reloaded: %v", err)
					}
					log.Debug("Locale file reloaded: %s", strings.TrimPrefix(event.Name, "conf/locale/"))
				}
			}
		}
	}()

	if err := watcher.Add("conf/locale"); err != nil {
		log.FatalD(4, "Fail to start locale watcher: %v", err)
	}
}

func init() {
	if !setting.ProdMode {
		monitorI18nLocale()
	}
}

func SubStr(str string, start, length int) string {
	if len(str) == 0 {
		return ""
	}
	end := start + length
	if len(str) < end {
		return str
	}
	return str[start:end] + "..."
}

func RearSubStr(str string, length int) string {
	if len(str) == 0 {
		return ""
	}
	if len(str) < length {
		return str
	}
	return "..." + str[len(str)-length:]
}
