
no free connections available to host
-------

- 原因： 连接数已经达到了 maxConns = DefaultMaxConnsPerHost = 512(默认值)。连接数达到最大值了

```go
func (c *HostClient) acquireConn() (*clientConn, error) {
    if n == 0 {
        maxConns := c.MaxConns
        if maxConns <= 0 {
            maxConns = DefaultMaxConnsPerHost
        }
        if c.connsCount < maxConns {
            c.connsCount++
            createConn = true
            if !c.connsCleanerRun {
                startCleaner = true
                c.connsCleanerRun = true
            }
        }
    } else {
        n--
        cc = c.conns[n]
        c.conns[n] = nil
        c.conns = c.conns[:n]
    }
    c.connsLock.Unlock()
```

处理链接源码 

```go
func clientDoDeadline(req *Request, resp *Response, deadline time.Time, c clientDoer) error {
   ...

    // Note that the request continues execution on ErrTimeout until
    // client-specific ReadTimeout exceeds. This helps limiting load
    // on slow hosts by MaxConns* concurrent requests.
    //
    // Without this 'hack' the load on slow host could exceed MaxConns*
    // concurrent requests, since timed out requests on client side
    // usually continue execution on the host.

    var mu sync.Mutex
    var timedout bool
        //这个goroutine是用来处理连接以及发送请求的
    gofunc() {
        errDo := c.Do(reqCopy, respCopy)
        mu.Lock()
        {
            if !timedout {
                if resp != nil {
                    respCopy.copyToSkipBody(resp)
                    swapResponseBody(resp, respCopy)
                }
                swapRequestBody(reqCopy, req)
                ch <- errDo
            }
        }
        mu.Unlock()

        ReleaseResponse(respCopy)
        ReleaseRequest(reqCopy)
    }()
        //这块内容是用来处理超时的
    tc := AcquireTimer(timeout)
    var err error
    select {
    case err = <-ch:
    case <-tc.C:
        mu.Lock()
        {
            timedout = true
            err = ErrTimeout
        }
        mu.Unlock()
    }
    ReleaseTimer(tc)

    select {
    case <-ch:
    default:
    }
    errorChPool.Put(chv)

    return err
}
```

```go
// DoTimeout performs the given request and waits for response during
// the given timeout duration.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.
//
// The function doesn't follow redirects. Use Get* for following redirects.
//
// Response is ignored if resp is nil.
//
// ErrTimeout is returned if the response wasn't returned during
// the given timeout.
//
// ErrNoFreeConns is returned if all HostClient.MaxConns connections
// to the host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
//
// Warning: DoTimeout does not terminate the request itself. The request will
// continue in the background and the response will be discarded.
// If requests take too long and the connection pool gets filled up please
// try setting a ReadTimeout.
func (c *HostClient) DoTimeout(req *Request, resp *Response, timeout time.Duration) error {
    return clientDoTimeout(req, resp, timeout, c)
}
```

- Solution：

需要设置 ReadTimeout 字段，达到 ReadTimeout 时间还没有得到返回值，客户端就会把连接断开（


ErrConnectionClosed
--------------

fasthttp default idle timeout is 15 seconds. 如果对方默认keep-live 时间是8s。
需要设置 IdleConnTimeout 小于 8s 

原生http木有问题，是因为两个 协程 在 loop write read,所以对 server 端 FIN 是 及时响应的，也就是client 及时也关闭了链接


Source
-------

- https://www.jianshu.com/p/12f3955c7e1c


