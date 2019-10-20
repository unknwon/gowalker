// Copyright 2018 Unknwon
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

package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/unknwon/gowalker/internal/db"
)

var (
	totalJSFilesGaugeFunc = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "gowalker",
		Subsystem: "js_file",
		Name:      "total",
		Help:      "Number of total JS files",
	}, func() float64 {
		return float64(db.NumTotalJSFiles())
	})
	generatedJSFilesGaugeFunc = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "gowalker",
		Subsystem: "js_file",
		Name:      "generated",
		Help:      "Number of generated JS files",
	}, func() float64 {
		return float64(db.NumGeneratedJSFiles())
	})
	distributedJSFilesGaugeFunc = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "gowalker",
		Subsystem: "js_file",
		Name:      "distributed",
		Help:      "Number of distributed JS files",
	}, func() float64 {
		return float64(db.NumDistributedJSFiles())
	})
	recycledJSFilesGaugeFunc = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "gowalker",
		Subsystem: "js_file",
		Name:      "recycled",
		Help:      "Number of recycled JS files",
	}, func() float64 {
		return float64(db.NumRecycledJSFiles())
	})
)

func init() {
	prometheus.MustRegister(
		totalJSFilesGaugeFunc,
		generatedJSFilesGaugeFunc,
		distributedJSFilesGaugeFunc,
		recycledJSFilesGaugeFunc,
	)
}
