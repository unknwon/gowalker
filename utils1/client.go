// Copyright 2012 Gary Burd
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

// This file implements an http.Client with request timeouts set by command
// line flags. The logic is not perfect, but the code is short.

package utils

import (
	"flag"
	"net"
	"net/http"
	"time"
)

var (
	dialTimeout  = flag.Duration("dial_timeout", 5*time.Second, "Timeout for dialing an HTTP connection.")
	readTimeout  = flag.Duration("read_timeout", 15*time.Second, "Timeoout for reading an HTTP response.")
	writeTimeout = flag.Duration("write_timeout", 5*time.Second, "Timeout writing an HTTP request.")
)

type timeoutConn struct {
	net.Conn
}

func (c *timeoutConn) Read(p []byte) (int, error) {
	return c.Conn.Read(p)
}

func (c *timeoutConn) Write(p []byte) (int, error) {
	// Reset timeouts when writing a request.
	c.Conn.SetWriteDeadline(time.Now().Add(*readTimeout))
	c.Conn.SetWriteDeadline(time.Now().Add(*writeTimeout))
	return c.Conn.Write(p)
}

func timeoutDial(network, addr string) (net.Conn, error) {
	c, err := net.DialTimeout(network, addr, *dialTimeout)
	if err != nil {
		return nil, err
	}
	return &timeoutConn{Conn: c}, nil
}

var HttpTransport = &http.Transport{Dial: timeoutDial}
var HttpClient = &http.Client{Transport: HttpTransport}
