package fipple

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
)

// Recorder may be used to record http responses
type Recorder struct {
	t       *testing.T
	client  *http.Client
	baseURL string
}

// File is a fipple representation of a file consisting only of a
// filename and an io.Reader capable of reading the file contents
type File struct {
	Name    string
	Content io.Reader
}

// NewRecorder creates a new recorder with the given baseURL.
// t will be used to print out helpful error messages if any
// assertions fail.
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

// NewRequest creates a new request object with the given http
// method and path. The path will be appended to the baseURL
// for the recorder to create the full URL. You are free to
// add additional parameters or headers to the request before
// sending it. Any errors that occur will be passed to t.Fatal.
func (r *Recorder) NewRequest(method string, path string) *http.Request {
	fullURL := r.baseURL + path
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		r.t.Fatal(err)
	}
	return req
}

// NewRequestWithData can be used to easily send a request with
// form data (encoded as application/x-www-form-urlencoded). The
// path will be appended to the baseURL for the recorder to create
// the full URL. Any errors tha occur will be passed to t.Fatal.
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
// files is a map of key to *fipple.File
func (r *Recorder) NewMultipartRequest(method string, path string, fields map[string]string, files map[string]*File) *http.Request {
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
		fileWriter, err := form.CreateFormFile(fieldname, file.Name)
		if err != nil {
			r.t.Fatal(err)
		}
		if _, err := io.Copy(fileWriter, file.Content); err != nil {
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
	resp.ReadBody()
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
