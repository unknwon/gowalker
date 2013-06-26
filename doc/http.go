// Copyright 2011 Gary Burd
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

package doc

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var userAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/29.0.1541.0 Safari/537.36"

var (
	dialTimeout  = flag.Duration("dial_timeout", 10*time.Second, "Timeout for dialing an HTTP connection.")
	readTimeout  = flag.Duration("read_timeout", 10*time.Second, "Timeoout for reading an HTTP response.")
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

var (
	httpTransport = &http.Transport{Dial: timeoutDial}
	httpClient    = &http.Client{Transport: httpTransport}
)

// httpGet gets the specified resource. ErrNotFound is returned if the server
// responds with status 404.
func httpGetBytes(client *http.Client, url string, header http.Header) ([]byte, error) {
	rc, err := httpGet(client, url, header)
	if err != nil {
		return nil, err
	}
	p, err := ioutil.ReadAll(rc)
	rc.Close()
	return p, err
}

// httpGet gets the specified resource. ErrNotFound is returned if the
// server responds with status 404.
func httpGet(client *http.Client, url string, header http.Header) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	for k, vs := range header {
		req.Header[k] = vs
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &RemoteError{req.URL.Host, err}
	}
	if resp.StatusCode == 200 {
		return resp.Body, nil
	}
	resp.Body.Close()
	if resp.StatusCode == 404 { // 403 can be rate limit error.  || resp.StatusCode == 403 {
		err = NotFoundError{"Resource not found: " + url}
	} else {
		err = &RemoteError{req.URL.Host, fmt.Errorf("get %s -> %d", url, resp.StatusCode)}
	}
	return nil, err
}

// fetchFiles fetches the source files specified by the rawURL field in parallel.
func fetchFiles(client *http.Client, files []*source, header http.Header) error {
	ch := make(chan error, len(files))
	for i := range files {
		go func(i int) {
			req, err := http.NewRequest("GET", files[i].rawURL, nil)
			if err != nil {
				ch <- err
				return
			}
			req.Header.Set("User-Agent", userAgent)
			for k, vs := range header {
				req.Header[k] = vs
			}
			resp, err := client.Do(req)
			if err != nil {
				ch <- &RemoteError{req.URL.Host, err}
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				ch <- &RemoteError{req.URL.Host, fmt.Errorf("get %s -> %d", req.URL, resp.StatusCode)}
				return
			}
			files[i].data, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- &RemoteError{req.URL.Host, err}
				return
			}
			ch <- nil
		}(i)
	}
	for _ = range files {
		if err := <-ch; err != nil {
			return err
		}
	}
	return nil
}

func httpGetJSON(client *http.Client, url string, v interface{}) error {
	rc, err := httpGet(client, url, nil)
	if err != nil {
		return err
	}
	defer rc.Close()
	err = json.NewDecoder(rc).Decode(v)
	if _, ok := err.(*json.SyntaxError); ok {
		err = NotFoundError{"JSON syntax error at " + url}
	}
	return err
}
