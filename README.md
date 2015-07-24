Fipple
------

[![GoDoc](https://godoc.org/github.com/albrow/fipple?status.svg)](https://godoc.org/github.com/albrow/fipple)

A lightweight utility written in go that makes it easy to write integration tests for REST APIs.

### Installation

Install like you would any other go package

``` bash
go get github.com/albrow/fipple
```

Then import the package in your source code

``` go
import "github.com/albrow/fipple"
```

### Setting Up

Fipple is designed for integration testing. It helps you test the endpoints of
REST APIs by sending requests to your server, recording the results, and
providing convenient methods for checking the response.

Start with the `Recorder` object, which you can use to create and send http
requests and record the response. The `Recorder` will use a `*testing.T` to
report errors in a very readable format.

The second argument to `fipple.NewRecorder` is an `http.Handler`. The recorder
will route all requests through the the given handler. Here's an example of how
to use fipple with httptest and the
[martini web framework](https://github.com/go-martini/martini).

```go
func TestUsersCreate(t *testing.T) {
	// Create a martini instance and add a single route handler. Typically you
	// would want to wrap this into a function that declares all your routes.
	m := martini.Classic()
	m.Get("/", func() string {
		return "Hello world!"
	})
	// Create a new recorder using the martini instance. Requests sent using the
	// recorder will be sent through the martini instance.
	rec := fipple.NewRecorder(t, m)
}
```

We used martini in the above example, but other popular web frameworks such as
[gin](https://github.com/gin-gonic/gin) or
[negroni](https://github.com/codegangsta/negroni) would work just as well.
You can create a recorder with anything that implements
[`http.Handler`](http://golang.org/pkg/net/http/#Handler).

If you're using a web framework that does not expose an `http.Handler`, don't
fret! You can still use fipple by starting a server in a separate process and
then using `fipple.NewURLRecorder` to create the recorder. So, for example if
you started your server in a separate process listening on port 4000, here's how
you would create a recorder:

```go
func TestUsersCreate(t *testing.T) {
	// Create a recorder that points to our already running server on port 4000.
	rec := fipple.NewURLRecorder(t, "http://localhost:4000")
}
```

### Example Usage

In this example, we're writing an integration test for creating users. We'll use
the `Recorder` object to construct and send a POST request with some parameters
for creating a user. The response will be captured in a `Response` object, a
lightweight wrapper around an `http.Response`, which we'll call `res`. We'll use
the `Post` method for this. It takes two arguments: the path (which is appended
to the base url for the `Recorder`) and the data you want to send as a map of
string to string. Here, we'll do a POST request to "/users" with
an email and password.

```go
res := rec.Post("/users", map[string]string{
	"email": "foo@example.com",
	"name": "Mr. Foo Bar",
	"password": "password123",
})
```

Note that there's no need to write error handling yourself. Since `rec` has a
`*testing.T`, it'll report errors automatically with `t.Error`.

We can then use the `Response` object to make sure the response from our server
was correct. Typically this means using the `ExpectOk` (or `ExpectCode` if you
expect a code other than 200) and `ExpectBodyContains` methods.
`ExpectBodyContains` will check the body of the response for the provided
string. If the body does not contain that string, it reports an error via
`t.Error`.

In our example, we expect the response code to be 200 (i.e. Ok) and the body of
the response to contain the user we just created, encoded as JSON. Here's how we
could do that:

```go
res.ExpectOk()
res.ExpectBodyContains(`"user": `)
res.ExpectBodyContains(`"email": "foo@example.com"`)
res.ExpectBodyContains(`"name": "Mr. Foo Bar"`)
```

If any of the expectations failed, fipple will print out a nice summary of the
request, including the actual body of the response, and a list of what went
wrong. Here's an example output for a failed test:

![Example Failed Test Output](http://oi59.tinypic.com/rj37kk.jpg)

If the response contains JSON, it will automatically be formatted for you. You
can also turn the colorization off via the `Colorize` option. (With colorization
off, the body of the response will be the same color as all other text, instead
of dark grey-ish).

### Table-Driven Tests

Fipple works great for table-driven tests. Let's say that you wanted to test
validations for the create users endpoint. For inputs that are incorrect, you
expect a specific validation error in the response indicating what was wrong.
Here's an example:

```go
validationTests := []struct {
	data            map[string]string
	expectedMessage string
}{
	{
		data: map[string]string{
			"email": "foo2@example.com",
		},
		expectedMessage: "name is required",
	},
	{
		data: map[string]string{
			"email": "not_valid_email",
		},
		expectedMessage: "email is invalid",
	},
	{
		data: map[string]string{
			"name": "Foo Bar Jr.",
		},
		expectedMessage: "email is required",
	},
}
for _, test := range validationTests {
	res := rec.Post("users", test.data)
	res.ExpectCode(422)
	res.ExpectBodyContains(test.expectedMessage)
}
```

### More Complicated Requests

Right out of the box, the `Recorder` object supports the GET, POST, PUT, and
DELETE methods. If you need to test a different type of method, you can do so by
constructing your own request and then using the `Do` method of the `Recorder`.
Here's an example:

```go
req := rec.NewRequest("BREW", "coffees/french_roast")
res := rec.Do(req)
res.ExpectCode(418)
res.ExpectBodyContains("I am a teapot!")
```

Since NewRequest returns a vanilla `*http.Request` object, you can also use it
to add custom headers to the request before passing the request to `Do`. Here's
an example of adding a JWT token to the request before sending it.

```go
req := rec.NewRequest("DELETE", "users/" + userId)
req.Header.Add("AUTHORIZATION", "Bearer " + authToken)
res := rec.Do(req)
res.ExpectOk()
```

Finally, fipple also supports multipart requests (with file uploads!) via the
`NewMultipartRequest` method.

```go
userFields := map[string]string{
	"email":    "bar@example.com",
	"name":     "Ms. Foo Bar",
	"password": "password123",
}
profileImageFile, err := os.Open("images/profile.png")
if err != nil {
	t.Fatal(err)
}
userFiles := map[string]*os.File{
	"profilePicture": profileImageFile,
}
req := rec.NewMultipartRequest("POST", "users", userFields, userFiles)
res := rec.Do(req)
res.ExpectOk()
```

[Full documentation](http://godoc.org/github.com/albrow/fipple) is available on
godoc.org.


License
-------

Fipple is licensed under the MIT License. See the LICENSE file for more
information.
