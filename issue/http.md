Golang http
==========

Http 重用底层TCP链接
---------------

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

Timeout vs Deadline
------------

Go exposes to implement `timeouts`: `Deadlines`.

Exposed by `net.Conn` with the Set [Read|Write]Deadline(time.Time) methods, Deadlines are an __absolute time__ which when reached makes all I/O operations fail with a timeout error.

- Server Timeout

![img.png](client_timeout.png)

```go
srv := &http.Server{
    ReadTimeout: 5 * time.Second,
    WriteTimeout: 10 * time.Second,
}

log.Println(srv.ListenAndServe())
```

  - ReadTimeout covers the time from when the connection is accepted to when the request body is fully read (if you do read the body, otherwise to the end of the headers). It's implemented in net/http by calling SetReadDeadline immediately after Accept.

  - WriteTimeout normally covers the time from the end of the request header read to the end of the response write (a.k.a. the lifetime of the ServeHTTP), by calling SetWriteDeadline

- Client Timeout

![img_1.png](server_timeout.png)

```go
c := &http.Client{
    Transport: &http.Transport{
        Dial: (&net.Dialer{
                Timeout:   30 * time.Second,
                KeepAlive: 30 * time.Second,
        }).Dial,
        MaxIdleConns:          100,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        ForceAttemptHTTP2:     true,
    }
}

```

KeepAlive
-----------

- HTTP KeepAlive

  - 原理
    
    HTTP 的 Keep-Alive，是由应用层（用户态） 实现的，称为 HTTP 长连接
      - 客户端请求的包头中 `Connection: Keep-Alive`
      - web 服务软件一般都会提供 `keepalive_timeout` 参数，用来指定 HTTP 长连接的超时时间
   
  -  线上存在网络问题，会导致 `GRPC HOL blocking`, 于是决定把 GRPC client改写成 HTTP client
    
  -  HTTP Client 存在 EOF 错误，EOF 一般是跟 IO 关闭有关系的, 网络 IO 的Keep-Alive 机制有关系
    
    Client side
    ```go
     c := &http.Client{
        Transport: &http.Transport{
           MaxIdleConnsPerHost: 1,
           DialContext: (&net.Dialer{
                Timeout:   time.Second * 2,
                KeepAlive: time.Second * 60,
           }).DialContext,
           DisableKeepAlives: false,  // Enable keep-alive
           IdleConnTimeout:   90 * time.Second,
          },
        Timeout: time.Second * 2,
     }
    ```
    
  - Doc
    
    ```go
    type Dialer struct {
    ...
     // KeepAlive specifies the interval between keep-alive
     // probes for an active network connection.
     // If zero, keep-alive probes are sent with a default value
     // (currently 15 seconds), if supported by the protocol and operating
     // system. Network protocols or operating systems that do
     // not support keep-alives ignore this field.
     // If negative, keep-alive probes are disabled.
     KeepAlive time.Duration
    ...
    }
    ```

- TCP KeepAlive

  - 原理
    
    该功能是由「内核」实现的，当客户端和服务端长达一定时间没有进行数据交互时，内核为了确保该连接是否还有效，就会发送探测报文，来检测对方是否还在线，然后来决定是否要关闭该连接
    需要通过 socket 接口设置 SO_KEEPALIVE 选项才能够生效，如果没有设置，那么就无法使用 TCP 保活机制
  - 测试
    ```shell
    > redis-cli -h 172.24.213.39 -p 6380
    > tcpdump -i eth0 -n host 172.24.213.39
     # 会看到 client 每隔 15s 会发送空的 ACK 包给 server, 并收到 server 返回的 ACK, 实际上这就是 client 端的 tcp keepalive 在起作用。
     # 然后我们在 server 设置 iptables, 人为制造网络隔离
    > iptables -I INPUT -s 172.24.213.40 -j DROP;iptables -I OUTPUT -d 172.24.213.40 -j DROP;iptables -nvL
    # client 172.24.213.40 每 5s 发送一个 ACK 三次，最后发一个 RST 包销毁连接。当然这个 RST redis-server 肯定也没有接收到。过一会将 server 防火墙删除
    > iptables -D INPUT -s 172.24.213.40 -j DROP;iptables -D OUTPUT -d 172.24.213.40 -j DROP;iptables -nvL
    > ss -a | grep 6380
    ```
  - 参数
    ```shell
    > cat /proc/sys/net/ipv4/tcp_keepalive_time
    7200
    > cat /proc/sys/net/ipv4/tcp_keepalive_probes
    9
    > cat /proc/sys/net/ipv4/tcp_keepalive_intvl
    75
    ```
    
  - Go TCP: 
    
    经过 net: enable TCP keepalives by default 和 net: add KeepAlive field to ListenConfig之后，从 go1.13 开始，默认都会开启 client 端与 server 端的 keepalive, 默认是 15s
    
    `func (ln *TCPListener) accept() (*TCPConn, error) `
    
    `func setKeepAlivePeriod(fd *netFD, d time.Duration) error`



