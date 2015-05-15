// Copyright 2015 Alex Browne.  All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package fipple

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/wsxiaoys/terminal/color"
)

// Response represents the response from an http request and has methods to
// make testing easier.
type Response struct {
	*http.Response
	Body     []byte
	recorder *Recorder
	once     sync.Once
}

// readBody reads r.Response.Body into r.Body. If the content-type is json,
// the body is automatically indented.
func (r *Response) readBody() {
	buf := bytes.NewBuffer(r.Body)
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		body, _ := ioutil.ReadAll(r.Response.Body)
		json.Indent(buf, body, "", "\t")
	} else {
		buf.ReadFrom(r.Response.Body)
	}
	r.Body = buf.Bytes()
}

// ExpectOk causes a test error if response code != 200
func (r *Response) ExpectOk() {
	r.ExpectCode(200)
}

// ExpectCode causes a test error if response code != the given code
func (r *Response) ExpectCode(code int) {
	if r.StatusCode != code {
		r.PrintErrorOnce()
		r.recorder.t.Errorf("Expected response code %d but got: %d", code, r.StatusCode)
	}
}

// ExpectBodyContains causes a test error if the response body does
// not contain the given string.
func (r *Response) ExpectBodyContains(str string) {
	if !strings.Contains(string(r.Body), str) {
		r.PrintErrorOnce()
		r.recorder.t.Errorf("Expected response to contain `%s` but it did not.", str)
	}
}

// PrintError prints some information about the response via t.Errorf. This includes
// a message about the method and path for the sent request, and the entire content
// of the response body.
func (r *Response) PrintError() {
	body := string(r.Body)
	if body == "" {
		r.recorder.t.Errorf("%s request to %s failed. Response was empty.",
			r.Request.Method,
			r.Request.URL.Path)
	} else {
		if Colorize {
			body = r.colorBody()
		}
		r.recorder.t.Errorf("%s request to %s failed. Response was: \n%s",
			r.Request.Method,
			r.Request.URL.Path,
			body)
	}
}

// PrintErrorOnce will only print the response if it has not already been printed.
// Useful in cases where there are multiple Expect* methods called on the same response and
// we don't want to repeatedly print out the response body for each expection failure.
func (r *Response) PrintErrorOnce() {
	r.once.Do(r.PrintError)
}

// colorBody returns a colorized version of the response body.
// By default the color is dark grey-ish.
func (r *Response) colorBody() string {
	return color.Sprintf("@{.}%s", string(r.Body))
}
