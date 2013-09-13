// Copyright 2013 Unknown
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

package utils

import (
	"net/url"

	"github.com/astaxie/beego/context"
)

var LangTypes []*LangType // Languages are supported.

// LangType represents a language type.
type LangType struct {
	Lang, Name string
}

// SetLangVer sets site language version.
// It returns true when language is changed by URL parameters.
func SetLangVer(ctx *context.Context, input url.Values, data map[interface{}]interface{}) (bool, LangType) {
	isNeedRedir := false

	// 1. Check URL arguments.
	lang := input.Get("lang")

	// 2. Get language information from cookies.
	if len(lang) == 0 {
		ck, err := ctx.Request.Cookie("lang")
		if err == nil {
			lang = ck.Value
		}
	} else {
		isNeedRedir = true
	}

	// Check again in case someone modify by purpose.
	isValid := false
	for _, v := range LangTypes {
		if lang == v.Lang {
			isValid = true
			break
		}
	}
	if !isValid {
		lang = ""
		isNeedRedir = false
	}

	// 3. Get language information from 'Accept-Language'.
	if len(lang) == 0 {
		al := ctx.Request.Header.Get("Accept-Language")
		if len(al) > 2 {
			al = al[:2] // Only compare first two letters.
			for _, v := range LangTypes {
				if al == v.Lang {
					lang = al
					break
				}
			}
		}
	}

	// 4. DefaucurLang language is English.
	if len(lang) == 0 {
		lang = "en-US"
		isNeedRedir = false
	}

	curLang := LangType{
		Lang: lang,
	}

	// Save language information in cookies.
	ctx.SetCookie("lang", curLang.Lang, 1<<31-1, "/")

	restLangs := make([]*LangType, 0, len(LangTypes)-1)
	for _, v := range LangTypes {
		if lang != v.Lang {
			restLangs = append(restLangs, v)
		} else {
			curLang.Name = v.Name
		}
	}

	// Set language properties.
	data["Lang"] = curLang.Lang
	data["CurLang"] = curLang.Name
	data["RestLangs"] = restLangs

	return isNeedRedir, curLang
}
