Golang http
=========

Http 重用底层TCP链接

官方释疑

```go
    // Body represents the response body.
    //
    // The response body is streamed on demand as the Body field
    // is read. If the network connection fails or the server
    // terminates the response, Body.Read calls return an error.
    //
    // The http Client and Transport guarantee that Body is always
    // non-nil, even on responses without a body or responses with
    // a zero-length body. It is the caller's responsibility to
    // close Body. The default HTTP client's Transport may not
    // reuse HTTP/1.x "keep-alive" TCP connections if the Body is
    // not read to completion and closed.
    //
    // The Body is automatically dechunked if the server replied
    // with a "chunked" Transfer-Encoding.
    //
    // As of Go 1.12, the Body will also implement io.Writer
    // on a successful "101 Switching Protocols" response,
    // as used by WebSockets and HTTP/2's "h2c" mode.
    Body io.ReadCloser`
```

- 对于Body == nil的处理

```go
	if resp.Body == nil {
		// The documentation on the Body field says “The http Client and Transport
		// guarantee that Body is always non-nil, even on responses without a body
		// or responses with a zero-length body.” Unfortunately, we didn't document
		// that same constraint for arbitrary RoundTripper implementations, and
		// RoundTripper implementations in the wild (mostly in tests) assume that
		// they can use a nil Body to mean an empty one (similar to Request.Body).
		// (See https://golang.org/issue/38095.)
		//
		// If the ContentLength allows the Body to be empty, fill in an empty one
		// here to ensure that it is non-nil.
		if resp.ContentLength > 0 && req.Method != "HEAD" {
			return nil, didTimeout, fmt.Errorf("http: RoundTripper implementation (%T) returned a *Response with content length %d but a nil Body", rt, resp.ContentLength)
		}
		resp.Body = ioutil.NopCloser(strings.NewReader(""))
	}
```

- 对于不处理的statusCode

```go
    resp, err := http.Get("http://www.example.com")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusOK {
        var apiRet APIRet
        decoder := json.NewDecoder(resp.Body)
        err := decoder.Decode(&apiRet)
        // ...
    }else{
    	// 使用io.Copy
        io.Copy(ioutil.Discard, resp.Body)
        // ...
    }
```

测试代码

```go
    func main() {
        count := 100
        for i := 0; i < count; i++ {
            resp, err := http.Get("https://www.oschina.net")
            if err != nil {
                panic(err)
            }
    
            //io.Copy(ioutil.Discard, resp.Body)
            resp.Body.Close()
        }
    }
```