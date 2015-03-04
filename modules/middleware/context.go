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

package middleware

import (
	"fmt"
	"strings"

	"github.com/Unknwon/log"
	"github.com/Unknwon/macaron"
	"github.com/macaron-contrib/session"

	"github.com/Unknwon/gowalker/modules/base"
	"github.com/Unknwon/gowalker/modules/setting"
)

// Context represents context of a request.
type Context struct {
	*macaron.Context
	Flash *session.Flash
}

// HasError returns true if error occurs in form validation.
func (ctx *Context) HasError() bool {
	hasErr, ok := ctx.Data["HasError"]
	if !ok {
		return false
	}
	ctx.Flash.ErrorMsg = ctx.Data["ErrorMsg"].(string)
	ctx.Data["Flash"] = ctx.Flash
	return hasErr.(bool)
}

// HTML calls Context.HTML and converts template name to string.
func (ctx *Context) HTML(status int, name base.TplName) {
	ctx.Context.HTML(status, string(name))
}

// RenderWithErr used for page has form validation but need to prompt error to users.
func (ctx *Context) RenderWithErr(msg string, tpl base.TplName, form interface{}) {
	if form != nil {
		// auth.AssignForm(form, ctx.Data)
	}
	ctx.Flash.ErrorMsg = msg
	ctx.Data["Flash"] = ctx.Flash
	ctx.HTML(200, tpl)
}

// Handle handles and logs error by given status.
func (ctx *Context) Handle(status int, title string, err error) {
	if err != nil {
		log.ErrorD(4, "%s: %v", title, err)
		if macaron.Env != macaron.PROD {
			ctx.Data["ErrorMsg"] = err
		}
	}

	switch status {
	case 404:
		ctx.Data["Title"] = "Page Not Found"
	case 500:
		ctx.Data["Title"] = "Internal Server Error"
	}
	ctx.HTML(status, base.TplName(fmt.Sprintf("status/%d", status)))
}

// Contexter initializes a classic context for a request.
func Contexter() macaron.Handler {
	return func(c *macaron.Context, f *session.Flash) {
		ctx := &Context{
			Context: c,
			Flash:   f,
		}

		// Compute current URL for real-time change language.
		ctx.Data["Link"] = ctx.Req.URL.Path

		ctx.Data["ProdMode"] = setting.ProdMode
		ctx.Data["SubStr"] = base.SubStr
		ctx.Data["RearSubStr"] = base.RearSubStr
		ctx.Data["HasPrefix"] = strings.HasPrefix

		c.Map(ctx)
	}
}
