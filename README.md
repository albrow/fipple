Fipple
------

[![GoDoc](https://godoc.org/github.com/albrow/fipple?status.svg)](https://godoc.org/github.com/albrow/fipple)

A testing utility for go which lets you easily record and test http responses from any url.

### Installation

Install like you would any other go package

``` bash
go get github.com/albrow/fipple
```

Then import the package in your source code

``` go
import "github.com/albrow/fipple"
```

### Example Usage

Fipple is designed for integration testing. In a typical setup, you'll have your server
running in one process (perhaps in a test environment) and listening on some port on localhost.
Then you'd run `go test ./...` in a separate process. You could also use fipple to test
3rd-party REST APIs to make sure they work the way you expect and don't change unexpectedly.

Start with the `Recorder` object, which you can use to create and send http requests and record
the response. The `Recorder` will use a `*testing.T` to report errors in a very readable format.
The second argument is the base url which will be prepended to the path in any requests you make
with the recorder.

```go
func TestUsersCreate(t *testing.T) {
	rec := fipple.NewRecorder(t, "http://localhost:4000/")
}
```

In this example, we're writing an integration test for creating users. We'll use the `Recorder`
object to construct and send a POST request with some parameters for creating a user. The response
will be captured in a `Response` object, a lightweight wrapper around an `http.Response`, which
we'll call `res`. We'll use the `Post` method for this. It takes two arguments: the path (which
is appended to the base url for the `Recorder`) and the data you want to send as a map of string
to string. Here, we'll do a POST request to "localhost:4000/users" with an email and password.

```go
res := rec.Post("users", map[string]string{
	"email": "foo@example.com",
	"name": "Mr. Foo Bar",
	"password": "password123",
})
```

Note that there's no need to write error handling yourself. Since `rec` has a `*testing.T`,
it'll report errors automatically with `t.Error`.

We can then use the `Response` object to make sure the response from our server was correct.
Typically this means using the `AssertOk` (or `AssertCode` if you expect a code other than 200)
and `AssertBodyContains` methods. `AssertBodyContains` will check the body of the response for
the provided string. If the body does not contain that string, it reports an error via `t.Error`.

In our example, we expect the response code to be 200 (i.e. Ok) and the body of the response to
contain the user we just created, encoded as JSON. Here's how we could do that:

```go
res.AssertOk()
res.AssertBodyContains(`"user": `)
res.AssertBodyContains(`"email": "foo@example.com"`)
res.AssertBodyContains(`"name": "Mr. Foo Bar"`)
```

If any of the assertions failed, fipple will print out a nice summary of the request, including
the actual body of the response, and a list of what went wrong. Here's an example output for a
failed test:

![Example Failed Test Output](http://oi59.tinypic.com/rj37kk.jpg)

### Table-Driven Tests

Fipple works great for table-driven tests. Let's say that you wanted to test validations for
the create users endpoint. For inputs that are incorrect, you expect a specific validation
error in the response indicating what was wrong. Here's an example:

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
	res.AssertCode(422)
	res.AssertBodyContains(test.expectedMessage)
}
```

### More Complicated Requests

Right out of the box, the `Recorder` object supports the GET, POST, PUT, and DELETE methods.
If you need to test a different type of method, you can do so by constructing your own request
and then using the `Do` method of the `Recorder`. Here's an example:

```go
req := rec.NewRequest("BREW", "coffees/french_roast")
res := rec.Do(req)
res.AssertCode(418)
res.AssertBodyContains("I am a teapot!")
```

Since NewRequest returns a vanilla `*http.Request` object, you can also use it to add custom
headers to the request before passing the request to `Do`. Here's an example of adding a JWT
token to the request before sending it.

```go
req := rec.NewRequest("DELETE", "users/" + userId)
req.Header.Add("AUTHORIZATION", "Bearer " + authToken)
res := rec.Do(req)
res.AssertOk()
```

Finally, fipple also supports multipart requests (with file uploads!) via the `NewMultipartRequest`
method.

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
userFiles := map[string]*fipple.File{
	"profilePicture": profileImageFile,
}
req := rec.NewMultipartRequest("POST", "users", userFields, userFiles)
res := rec.Do(req)
res.AssertOk()
```

[Full documentation](http://godoc.org/github.com/albrow/fipple) is available on godoc.org.
