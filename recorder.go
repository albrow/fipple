// Copyright 2015 Alex Browne.  All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

// Package fipple is a lightweight utility that makes it easy to write
// integration tests for REST APIs.
package fipple

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

// Recorder can be used to send http requests and record the responses.
type Recorder struct {
	t       *testing.T
	client  *http.Client
	baseURL string
	server  *httptest.Server
	// Colorize is used to determine whether or not to colorize the errors when
	// printing to the console using t.Error. The default is true.
	Colorize bool
}

// NewRecorder returns a recorder that sends requests through the given handler.
// The recorder will report any errors using t.Error or t.Fatal.
func NewRecorder(t *testing.T, handler http.Handler) *Recorder {
	server := httptest.NewServer(handler)
	return &Recorder{
		t:        t,
		client:   newTestClient(t),
		baseURL:  server.URL,
		server:   server,
		Colorize: true,
	}
}

// NewURLRecorder creates a new recorder with the given baseURL. The recorder
// will report any errors using t.Error or t.Fatal.
func NewURLRecorder(t *testing.T, baseURL string) *Recorder {
	return &Recorder{
		t:        t,
		client:   newTestClient(t),
		baseURL:  baseURL,
		Colorize: true,
	}
}

// Close closes the recorder. You must call Close when you are done using a
// recorder.
func (r *Recorder) Close() {
	if r.server != nil {
		r.server.Close()
	}
}

// newResponse creates and returns a *fipple.Response, which is a lightweight
// wrapper around an *http.Response.
func (r *Recorder) newResponse(resp *http.Response) *Response {
	return &Response{
		Response: resp,
		recorder: r,
	}
}

// newTestClient returns an *http.Client with a cookiejar which can be used to
// store and retrieve cookies.
func newTestClient(t *testing.T) *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar}
}

// NewRequest creates a new request object with the given http method and path.
// The path will be appended to the baseURL for the recorder to create the full
// URL. You are free to add additional parameters or headers to the request
// before sending it. Any errors that occur will be passed to t.Fatal.
func (r *Recorder) NewRequest(method string, path string) *http.Request {
	fullURL := r.baseURL + path
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		r.t.Fatal(err)
	}
	return req
}

// NewRequestWithData can be used to easily send a request with form data
// (encoded as application/x-www-form-urlencoded). The path will be appended to
// the baseURL for the recorder to create the full URL. The Content-Type header
// will automatically be added. Any errors tha occur will be passed to t.Fatal.
func (r *Recorder) NewRequestWithData(method string, path string, data map[string]string) *http.Request {
	fullURL := r.baseURL + path
	v := url.Values{}
	for key, value := range data {
		v.Add(key, value)
	}
	req, err := http.NewRequest(method, fullURL, strings.NewReader(v.Encode()))
	if err != nil {
		r.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// NewMultipartRequest can be used to easily create (and later send)
// a request with form data and/or files (encoded as multipart/form-data).
// fields is a key-value map of basic string fields for the form data, and
// files is a map of key to *os.File. The Content-Type header will
// automatically be added. Any errors tha occur will be passed to t.Fatal.
func (r *Recorder) NewMultipartRequest(method string, path string, fields map[string]string, files map[string]*os.File) *http.Request {
	fullURL := r.baseURL + path

	// First, create a new multipart form writer.
	body := bytes.NewBuffer([]byte{})
	form := multipart.NewWriter(body)

	// Add the key-value field params to the form
	for fieldname, value := range fields {
		if err := form.WriteField(fieldname, value); err != nil {
			r.t.Fatal(err)
		}
	}

	// Add the files to the form
	for fieldname, file := range files {
		fileWriter, err := form.CreateFormFile(fieldname, file.Name())
		if err != nil {
			r.t.Fatal(err)
		}
		if _, err := io.Copy(fileWriter, file); err != nil {
			r.t.Fatal(err)
		}
	}

	// Close the form to finish writing
	if err := form.Close(); err != nil {
		r.t.Fatal(err)
	}

	// Create and return the request object
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		r.t.Fatal(err)
	}
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+form.Boundary())
	return req
}

// NewJSONRequest creates and returns a JSON request with the given
// method and path (which is appended to the baseURL. data can be any
// data structure but cannot include functions or recursiveness. NewJSONRequest
// will convert data into json using json.Marshall. The Content-Type header
// will automatically be added. Any errors tha occur will be passed to t.Fatal.
func (r *Recorder) NewJSONRequest(method string, path string, data interface{}) *http.Request {
	// Create and write to the body
	body := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(body)
	if err := encoder.Encode(data); err != nil {
		r.t.Fatal(err)
	}

	// Create and return the request object
	fullURL := r.baseURL + path
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		r.t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	return req
}

// Do sends req and records the results into a fipple.Response.
// Note that because an http.Request should have already been created
// with a full, valid url, the baseURL of the Recorder will not be prepended
// to the url for req. You can run methods on the response to check
// the results. Any errors that occur will be passed to t.Fatal
func (r *Recorder) Do(req *http.Request) *Response {
	httpResp, err := r.client.Do(req)
	if err != nil {
		r.t.Fatal(err)
	}
	resp := r.newResponse(httpResp)
	resp.readBody()
	return resp
}

// Get sends a GET request to the given path and records the results into
// a fipple.Response. path will be appended to the baseURL for the recorder
// to create the full URL. You can run methods on the response to check the
// results. Any errors that occur will be passed to t.Fatal
func (r *Recorder) Get(path string) *Response {
	req := r.NewRequest("GET", path)
	return r.Do(req)
}

// Post sends a POST request to the given path using the given data as post
// parameters and records the results into a fipple.Response. path will be
// appended to the baseURL for the recorder to create the full URL. You
// can run methods on the response to check the results. Any errors that occur
// will be passed to t.Fatal
func (r *Recorder) Post(path string, data map[string]string) *Response {
	req := r.NewRequestWithData("POST", path, data)
	return r.Do(req)
}

// Put sends a PUT request to the given path using the given data as
// parameters and records the results into a fipple.Response. path
// will be appended to the baseURL for the recorder to create the
// full URL. You can run methods on the response to check the results.
// Any errors that occur will be passed to t.Fatal
func (r *Recorder) Put(path string, data map[string]string) *Response {
	req := r.NewRequestWithData("PUT", path, data)
	return r.Do(req)
}

// Delete sends a DELETE request to the given path and records the results
// into a fipple.Response. path will be appended to the baseURL for the recorder
// to create the full URL. You can run methods on the response to check the
// results. Any errors that occur will be passed to t.Fatal
func (r *Recorder) Delete(path string) *Response {
	req := r.NewRequest("DELETE", path)
	return r.Do(req)
}

// GetCookies returns the raw cookies that have been set as a result
// of any requests recorded by a Recorder. Any errors that occur will be
// passed to t.Fatal
func (r *Recorder) GetCookies() []*http.Cookie {
	fullURL, err := url.Parse(r.baseURL)
	if err != nil {
		r.t.Fatal(err)
	}
	return r.client.Jar.Cookies(fullURL)
}
