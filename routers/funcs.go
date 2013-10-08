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

package routers

import (
	"strings"

	"github.com/Unknwon/gowalker/models"
)

// FuncsRouter serves AJAX function code API page.
type FuncsRouter struct {
	baseRouter
}

// Get implemented Get method for FuncsRouter.
func (this *FuncsRouter) Get() {
	this.Data["IsFuncs"] = true
	this.TplNames = "funcs.html"

	// Get argument(s).
	q := strings.TrimSpace(this.Input().Get("q"))

	// Set properties.
	this.Data["Keyword"] = q

	if len(q) > 0 {
		// Function search.
		pfuncs := models.SearchFunc(q)

		// Show results after searched.
		if len(pfuncs) > 0 {
			this.Data["IsFindFunc"] = true
			this.Data["Results"] = pfuncs
		}
	} else {
		this.Data["IsShowIntro"] = true
	}
}
