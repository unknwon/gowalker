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

package base

import (
	"sort"
)

// MapToSortedStrings converts a string map to a alphabet sorted slice without duplication.
func MapToSortedStrings(m map[string]bool) []string {
	strs := make([]string, 0, len(m))
	for s := range m {
		strs = append(strs, s)
	}
	sort.Strings(strs)
	return strs
}
