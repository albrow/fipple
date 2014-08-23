package fipple

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
)

type Recorder struct {
	t       *testing.T
	client  *http.Client
	baseURL string
}

func NewRecorder(t *testing.T, baseURL string) *Recorder {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &Recorder{
		t:       t,
		client:  &http.Client{Jar: jar},
		baseURL: baseURL,
	}
}

func (r *Recorder) newResponse(resp *http.Response) *Response {
	return &Response{
		Response: resp,
		recorder: r,
	}
}

func (r *Recorder) NewRequest(method string, path string) *http.Request {
	fullURL := r.baseURL + path
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		r.t.Fatal(err)
	}
	return req
}

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

// Get sends req and records the results into a fipple.Response.
// Note that because an http.Request should have already been created
// with a full, valid url, the baseURL of the Recorder will not be prepended
// to the url for req. You can run methods on the response to check
// the results.
func (r *Recorder) Do(req *http.Request) *Response {
	httpResp, err := r.client.Do(req)
	if err != nil {
		r.t.Fatal(err)
	}
	resp := r.newResponse(httpResp)
	resp.ReadBody()
	return resp
}

// Get sends a GET request to the given path and records the results into
// a fipple.Response. You can run methods on the response
// to check the results.
func (r *Recorder) Get(path string) *Response {
	req := r.newRequest("GET", path)
	return r.Do(req)
}

// Post sends a POST request to the given path using the given data as post
// parameters and records the results into a fipple.Response. You
// can run methods on the response to check the results.
func (r *Recorder) Post(path string, data map[string]string) *Response {
	req := r.newRequestWithData("POST", path, data)
	return r.Do(req)
}

// Put sends a PUT request to the given path using the given data as
// parameters and records the results into a fipple.Response. You
// can run methods on the response to check the results.
func (r *Recorder) Put(path string, data map[string]string) *Response {
	req := r.newRequestWithData("PUT", path, data)
	return r.Do(req)
}

// Delete sends a DELETE request to the given path and records the results
// into a fipple.Response. You can run methods on the response to check the
// results.
func (r *Recorder) Delete(path string) *Response {
	req := r.newRequest("DELETE", path)
	return r.Do(req)
}

func (r *Recorder) GetCookies() []*http.Cookie {
	fullURL, err := url.Parse(r.baseURL)
	if err != nil {
		r.t.Fatal(err)
	}
	return r.client.Jar.Cookies(fullURL)
}
