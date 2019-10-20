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

package context

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-macaron/session"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"

	"github.com/unknwon/gowalker/internal/base"
	"github.com/unknwon/gowalker/internal/setting"
)

// Context represents context of a request.
type Context struct {
	*macaron.Context
	Flash *session.Flash
}

// Title sets "Title" field in template data.
func (c *Context) Title(locale string) {
	c.Data["Title"] = c.Tr(locale)
}

// PageIs sets "PageIsxxx" field in template data.
func (c *Context) PageIs(name string) {
	c.Data["PageIs"+name] = true
}

// HasError returns true if error occurs in form validation.
func (c *Context) HasError() bool {
	hasErr, ok := c.Data["HasError"]
	if !ok {
		return false
	}
	c.Flash.ErrorMsg = c.Data["ErrorMsg"].(string)
	c.Data["Flash"] = c.Flash
	return hasErr.(bool)
}

// HTML calls Context.HTML and converts template name to string.
func (c *Context) HTML(status int, name string) {
	c.Context.HTML(status, name)
}

// Success responses template with status http.StatusOK.
func (c *Context) Success(name string) {
	c.HTML(http.StatusOK, name)
}

// RenderWithErr used for page has form validation but need to prompt error to users.
func (c *Context) RenderWithErr(msg string, tpl string, form interface{}) {
	if form != nil {
		// auth.AssignForm(form, c.Data)
	}
	c.Flash.ErrorMsg = msg
	c.Data["Flash"] = c.Flash
	c.Success(tpl)
}

// Handle handles and logs error by given status.
func (c *Context) Handle(status int, title string, err error) {
	if err != nil {
		log.Error(2, "%s: %v", title, err)
		if macaron.Env != macaron.PROD {
			c.Data["ErrorMsg"] = err
		}
	}

	switch status {
	case 404:
		c.Data["Title"] = "Page Not Found"
	case 500:
		c.Data["Title"] = "Internal Server Error"
	}
	c.HTML(status, fmt.Sprintf("status/%d", status))
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

		ctx.Data["AppVer"] = setting.AppVer
		ctx.Data["ProdMode"] = setting.ProdMode
		ctx.Data["SubStr"] = base.SubStr
		ctx.Data["RearSubStr"] = base.RearSubStr
		ctx.Data["HasPrefix"] = strings.HasPrefix
		ctx.Data["int64"] = base.Int64
		ctx.Data["Year"] = time.Now().Year()

		c.Map(ctx)
	}
}
