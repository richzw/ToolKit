
- Http 重用底层TCP链接
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
  - 测试代码
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
- [Timeout vs Deadline](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)
  - Go exposes to implement `timeouts`: `Deadlines`.
  - Exposed by `net.Conn` with the Set [Read|Write]Deadline(time.Time) methods, Deadlines are an __absolute time__ which when reached makes all I/O operations fail with a timeout error.
  - keep in mind that all timeouts are implemented in terms of Deadlines, so they do NOT reset every time data is sent or received.
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
    - [no way of manipulating timeouts in Handler](https://github.com/golang/go/issues/16100)
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
    - Cancel and Context
      - net/http offers two ways to cancel a client request: Request.Cancel and, new in 1.7, Context.
    ```go
    func main() {
    	c := make(chan struct{})
    	timer := time.AfterFunc(5*time.Second, func() {
    		close(c)
    	})
    
            // Serve 256 bytes every second.
    	req, err := http.NewRequest("GET", "http://httpbin.org/range/2048?duration=8&chunk_size=256", nil)
    	if err != nil {
    		log.Fatal(err)
    	}
    	req.Cancel = c
    
    	log.Println("Sending request...")
    	resp, err := http.DefaultClient.Do(req)
    	if err != nil {
    		log.Fatal(err)
    	}
    	defer resp.Body.Close()
    
    	log.Println("Reading body...")
    	for {
    		timer.Reset(2 * time.Second)
                    // Try instead: timer.Reset(50 * time.Millisecond)
    		_, err = io.CopyN(ioutil.Discard, resp.Body, 256)
    		if err == io.EOF {
    			break
    		} else if err != nil {
    			log.Fatal(err)
    		}
    	}
    }
    ```
    ```go
    ctx, cancel := context.WithCancel(context.TODO())
    timer := time.AfterFunc(5*time.Second, func() {
    	cancel()
    })
    
    req, err := http.NewRequest("GET", "http://httpbin.org/range/2048?duration=8&chunk_size=256", nil)
    if err != nil {
    	log.Fatal(err)
    }
    req = req.WithContext(ctx)
    ```
[KeepAlive]()
  - **HTTP KeepAlive**
  - 原理
    - HTTP 的 Keep-Alive，是由应用层（用户态） 实现的，称为 HTTP 长连接. HTTP 是基于 TCP 传输协议实现的
      - 客户端请求的包头中 `Connection: Keep-Alive` 从 HTTP 1.1 开始， 就默认是开启了 Keep-Alive
      - web 服务软件一般都会提供 `keepalive_timeout` 参数，用来指定 HTTP 长连接的超时时间
  -  HTTP Client 存在 EOF 错误，EOF 一般是跟 IO 关闭有关系的, 网络 IO 的Keep-Alive 机制有关系
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
  - **TCP KeepAlive**
  - 原理
    - 该功能是由「内核」实现的，当客户端和服务端长达一定时间没有进行数据交互时，内核为了确保该连接是否还有效，就会发送探测报文，来检测对方是否还在线，然后来决定是否要关闭该连接
      - 如果对端程序是正常工作的。当 TCP 保活的探测报文发送给对端, 对端会正常响应，这样 TCP 保活时间会被重置，等待下一个 TCP 保活时间的到来。
      - 如果对端主机崩溃，或对端由于其他原因导致报文不可达。当 TCP 保活的探测报文发送给对端后，石沉大海，没有响应，连续几次，达到保活探测次数后，TCP 会报告该 TCP 连接已经死亡。
    - 需要通过 socket 接口设置 SO_KEEPALIVE 选项才能够生效，如果没有设置，那么就无法使用 TCP 保活机制
  - 流程
    - • 空闲时间检测：TCP 连接进入空闲状态超过 tcp_keepalive_time 后，开始发送探测包。
    - • 发送探测包：每隔 tcp_keepalive_intvl 秒发送一次探测包。
    - • 响应与失败处理：
      - • 如果对端响应，则认为连接正常；
      - • 如果连续发送 tcp_keepalive_probes 次探测包无响应，则认为连接已断开，关闭连接。
  - 如何确认 TCP Keepalive 是否生效
    - ss -4no state established | grep 2379
    - tcpdump -i eth0 tcp port <port> and tcp[tcpflags] == tcp-ack
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
    - 经过 net: enable TCP keepalives by default 和 net: add KeepAlive field to ListenConfig之后，从 go1.13 开始，默认都会开启 client 端与 server 端的 keepalive, 默认是 15s
    `func (ln *TCPListener) accept() (*TCPConn, error) `
    `func setKeepAlivePeriod(fd *netFD, d time.Duration) error`
- [`i/o timeout` caused by incorrect `setTimeout/setDeadline`](https://mp.weixin.qq.com/s/OI1TXa3JeSdMJV4aM19ZJw)
  - Source
    ```go
    tr = &http.Transport{
        MaxIdleConns: 100,
        Dial: func(netw, addr string) (net.Conn, error) {
            conn, err := net.DialTimeout(netw, addr, time.Second*2) //设置建立连接超时
            if err != nil {
                return nil, err
            }
            err = conn.SetDeadline(time.Now().Add(time.Second * 3)) //设置发送接受数据超时
            if err != nil {
                return nil, err
            }
            return conn, nil
        },
    }
    ```
  - 现象
    - golang服务在发起http调用时，虽然`http.Transport`设置了3s超时，会偶发出现i/o timeout的报错
  - 分析 
    - 抓包发现， 从刚开始三次握手，到最后出现超时报错 i/o timeout。 间隔3s。原因是客户端发送了一个一次Reset请求导致的。 就是客户端3s超时主动断开链接的
  - 查看 
    - SetDeadline是对于链接级别 SetDeadline sets the read and write deadlines associated with the connection.
  - 原理
    - HTTP协议从1.1之后就默认使用`长连接`。golang标准库里也兼容这种实现。通过建立一个连接池，针对每个域名建立一个TCP长连接
     ![img.png](http_connection.png)
    - 一个域名会建立一个连接，一个连接对应一个读goroutine和一个写goroutine。正因为是同一个域名，所以最后才会泄漏3个goroutine，如果不同域名的话，那就会泄漏 1+2*N 个协程，N就是域名数。
  - 正确的姿势 - **超时设置在http**
    ```go
        tr = &http.Transport{
            MaxIdleConns: 100,
        }
        client := &http.Client{
            Transport: tr,
            Timeout: 3*time.Second,  // Timeout specifies a time limit for requests made by this Client.
        }
    ```
    ![img.png](netpoll.png)
    不要在 http.Transport中设置超时，那是连接的超时，不是请求的超时。否则可能会出现莫名 io timeout报错。

- [如何正确设置保活](https://mp.weixin.qq.com/s/EmawKOftz0OAnMd2ydcOgQ)
  - 前情
    - 由于线上存在网络问题，会导致 GRPC HOL blocking, 于是决定把 GRPC client改写成 HTTP client
    - 搜索日志里面会发现有极少数的 EOF 错误
    - EOF 这个东西一般是跟 IO 关闭有关系的，就是 server 和 client 的 Keep-Alive 机制的问题
  - Debug
    - HTTP Client
      ```go
       c := &http.Client{
        Transport: &http.Transport{
         MaxIdleConnsPerHost: 1,
         DialContext: (&net.Dialer{
          Timeout:   time.Second * 2,
          KeepAlive: time.Second * 60,
         }).DialContext,
         DisableKeepAlives: false,
         IdleConnTimeout:   90 * time.Second,
        },
        Timeout: time.Second * 2,
       }
      ```
      官方文档介绍是一个用于TCP Keep-Alive的probe指针，间隔一定的时间发送心跳包。每间隔60S进行一次Keep-Alive
    - HTTP Server
      ```go
      s := http.Server{
        Addr:        ":8080",
        Handler:     http.HandlerFunc(Index),
        ReadTimeout: 10 * time.Second,
        // IdleTimeout: 10 * time.Second,
       }
       s.SetKeepAlivesEnabled(true)
       s.ListenAndServe()
      ```
      Server的 KeepAlive 主要是通过 IdleTimeout 来进行控制的，IdleTimeout 如果为空则使用 ReadTimeout
    - client 侧的 Keep-Alive 是60s，但是 server 侧的时间是间隔10s就去关掉空闲的连接。所以这里很容易就认为是：client 侧的 Keep-Alive 心跳间隔时间太长了，server 侧提前关闭了连接。修改参数，重新上线，持续观察一段时间发现还是有 EOF 错误
    - Mock EOF
      ```go
      func test(w http.ResponseWriter, r *http.Request) {
       log.Println("receive request from:", r.RemoteAddr, r.Header)
       if count%2 == 1 {
        conn, _, err := w.(http.Hijacker).Hijack()
        if err != nil {
         return
        }
      
        conn.Close()
        count++
        return
       }
       w.Write([]byte("ok"))
       count++
      }
      
      func main() {
       s := http.Server{
        Addr:        ":8080",
        Handler:     http.HandlerFunc(test),
        ReadTimeout: 10 * time.Second,
       }
       // s.SetKeepAlivesEnabled(false)
       s.ListenAndServe()
      }
      ```
      在尝试复现 EOF 错误的时候，看到有 Hijack 这种东西，还是挺好用的。可以看到直接在 server 侧关掉连接, client 侧感知不到连接关闭确实是会有 EOF 错误发生的。
    - Mock Keep-Alive
      ```go
      func sendRequest(c *http.Client) {
       req, err := http.NewRequest("POST", "http://localhost:8080", nil)
       if err != nil {
        panic(err)
       }
       resp, err := c.Do(req)
       if err != nil {
        panic(err)
       }
       defer resp.Body.Close()
      
       buf := &bytes.Buffer{}
       buf.ReadFrom(resp.Body)
      
      }
      
      func main() {
       c := &http.Client{
        Transport: &http.Transport{
         MaxIdleConnsPerHost: 1,
         DialContext: (&net.Dialer{
          Timeout:   time.Second * 2,
          KeepAlive: time.Second,
         }).DialContext,
         DisableKeepAlives: false,
         IdleConnTimeout:   90 * time.Second,
        },
        Timeout: time.Second * 2,
       }
       // c := &http.Client{}
       sendRequest(c)
       time.Sleep(time.Second * 3)
       sendRequest(c)
      
      }
      ```
      在本地开始尝试复现 Keep-Alive 的问题，client 侧使用 KeepAlive: time.Second, 每间隔一秒钟的 keep-alive, server 侧同样使用两秒 IdleTimeout: time.Second。
    - Packet Capture
      - Client 使用的基于TCP层面的Keep-alive协议，针对的是整条TCP连接
      - Server 侧明显是基于应用层协议做的判断
    - Source Code
      - Client
       ```go
       func (d *Dialer) DialContext(ctx context.Context, network, address string) (Conn, error) {
       ...
        if tc, ok := c.(*TCPConn); ok && d.KeepAlive >= 0 {
         setKeepAlive(tc.fd, true)
         ka := d.KeepAlive
         if d.KeepAlive == 0 {
          ka = defaultTCPKeepAlive
         }
         setKeepAlivePeriod(tc.fd, ka)
         testHookSetKeepAlive(ka)
        }
       ...
       }
       
       func setKeepAlivePeriod(fd *netFD, d time.Duration) error {
        // The kernel expects seconds so round to next highest second.
        secs := int(roundDurationUp(d, time.Second))
        if err := fd.pfd.SetsockoptInt(syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs); err != nil {
         return wrapSyscallError("setsockopt", err)
        }
        err := fd.pfd.SetsockoptInt(syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, secs)
        runtime.KeepAlive(fd)
        return wrapSyscallError("setsockopt", err)
       }
       ```
      最后调用的是 SetsockoptInt，这个函数就不在这具体的展开了，本质上来讲 Client 侧是在TCP 4层让 OS 来帮忙进行的 Keep-Alive。
    - Server
      ```go
      func (c *conn) serve(ctx context.Context) {
      ...
          defer func() {
              if !c.hijacked() {
         c.close()
         c.setState(c.rwc, StateClosed, runHooks)
        }
          }
          
          for {
              w, err := c.readRequest(ctx)
              ...
              serverHandler{c.server}.ServeHTTP(w, w.req)
              ...
        if d := c.server.idleTimeout(); d != 0 {
         c.rwc.SetReadDeadline(time.Now().Add(d))
         if _, err := c.bufr.Peek(4); err != nil {
          return
         }
        }
          }
      ...
      }
      ```
      defer 就是关闭连接用的，当函数退出的时候server会关闭连接。 for循坏是处理连接请求用的，可以看出来HTTP server本身其实是不支持处理多个请求的，并没有实现HTTP 1.1协议中的Pipeline。
      然后再看keep-alive的操作，先设置ReadDeadline，然后调用c.bufr.Peek这里的调用流程比较长，其实最后会落到conn.Read,本质上是一个阻塞操作。然后开始等待bufr里面的数据，如果client在这个时间段没有发送数据过来，则会退出for循环然后关闭连接
  - conclusion
    - 在上述的场景下想要reuse一个conn主要还是取决于server 侧的idleTimeout。如果没收到client发送的请求是会主动发送fin包进行close的。
  - fix
    - **Retry**
      其实解决方案有很多种，在这里线上采用的是客户端进行重试。这里引申一下，像上面这种错误，如果是GET,HEAD等一些幂等操作的话，client代码库会自动进行重试。我们线上使用的是POST, 所以直接在业务侧进行重试
    - **Increase IdleTimeout**
      另外一个解决方案就是增加server的IdleTimeout，但是这样一来会消耗更多的server资源。
- [Dive into go http timeout](https://adam-p.ca/blog/2022/01/golang-http-server-timeouts/)
- [网络超时]
  - 超时问题
    - 一个是调用 redis 导致程序夯住300s左右，另一个是上游请求频繁超时（120ms~200ms之间），两个问题都发生在同一个服务上
  - Debug
    - 首先排除是不是抖动，通过跟 redis 同学反复确认，他们的耗时处理都是在1ms以下，没发现长时间的耗时，又跟运维确认是否有网络抖动，运维反馈说机房10G网卡确实会出现偶尔的丢包现象，正在换100G网卡，但偶尔丢包会有重试，不会导致耗时这么长，所以排除了丢包问题(
    - 加了更详细的 debug 日志，最后发现是 read的时候阻塞了，并最终收到了 RST 包
    - 各种猜测：丢包、内核参数问题、Gateway 升级等等，同时也发现了更多的现象，比如偶尔出现几台机器同时耗时长的问题，其他部门的同学也反馈有300s耗时问题，同样是调用 redis，种种猜测之后还是决定抓包来看，由于复现的概率比较低，也没有规律，同时机器也比较多，抓包成本比较大，所以抓包耗费了很长时间，直到线上复现了超时问题。
    - redis proxy 机器数比较少相对好抓，client 端机器比较多，所以只抓了个别机器
      - 我们发现 client 端一直在超时重试（Retransmission，这里遵循以TCP_RTO_MIN=HZ/5 = 200ms 开始的退避算法，200ms、400ms，800...），一直到最后服务端返回了 RST，这个正好印证了第二步日志中报的错误"connection reset by peer"
      - server端一直没有返回 ack，但在09:18:12的时候收到了server端发送的 keep-alive 包，也叫 tcp 探活包，时间刚好是从上一次请求后的75s（redis 确认设置了 keep-alive 时间是75s）
      - 从整个流量来看，好像 redis 端收不到包，但 client 可以收到 redis 的包。同一时间redis也同步开启了抓包，但并没收到我们发送的PSH
      - 调用方式采用的VIP，VIP指向 Gateway，但 Gateway 都是公用的，流量很大，很难抓包，所以我们通过排除法，选择直连真实ip，绕过 Gateway，通过两天观察，类似的问题消失了，所以基本就定位到是 Gateway 问题
    - Gateway fullnat模式: fullnat相对nat模式有一个重要的区别就是session表从一个变成了两个（session表用作关联客户端和服务端的ip:port），一个IN一个OUT，如果IN的表丢失了，而OUT表正常，就有可能出现上述抓包的问题流量——client 能收到 server 包，server 收不到 client 包。
  - More
    - 如果加重试造成雪崩怎么办？
      - 有一种观点是不要在下游加重试，统一从最上游加一次重试就行了，这种观点无非是担心中间链路有突发流量，会导致下游雪崩，比如网络抖动导致大规模重试。
      - 解决雪崩的问题不应该是通过减少重试次数来缓解，而是通过限流和熔断，对限流来说，越上层限流越好，而对于重试，是越下层重试越好
    - 那到底该设置多少超时时间和重试次数呢？
      - 其实这两个数是有互相配合的，举个例子大家就明白了，比如我调用下游的成功率要达到五个9，调用下游的耗时98分位是20ms，那如果我设置超时时间20ms就意味着有2%的概率失败，那加一次重试可以降低到0.04%，可用性达到99.96%，不满足需求，那就再加一次重试，可用性达到99.9992%，但如果超时次数太多对上下游压力也会比较大，所以最好不要超过3次，如果重试次数确定了
      - 最理想的方案还是依据SLA，来确定超时时间和重试次数。
- [GET 请求可以发送 body?]()
  - RFC
    - 在 GET 类型的请求里使用 body 是一个没有定义的语义。如果在 GET 请求的 body 里传递参数可能会被某些实现方拒绝该请求。



