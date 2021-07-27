
Socket
======

epoll implementation based on `golang.org/x/sys/unix`

- SO_REUSEPORT
- Zero copy 
  - replace `recvmmsg` with `recvmsg`
  - https://developer.ibm.com/languages/java/articles/j-zerocopy/
  - https://github.com/newtools/zsocket
  - `ReadFrom`(tcpsock.go) support transparently copying (even via io.Copy) between two TCP connections using zero-copy techniques (splice() and sendfile() on linux). 
- 网卡多队列绑定CPU核心优化
  - 将网卡多队列均匀绑定到CPU多核心上，同时设置SO_INCOMING_CPU属性，将socket的处理与某个CPU核心绑定，同时逻辑线程与某个CPU核心进行亲和性绑定
  - Ref http://www.cnhalo.net/2016/03/13/tcp-accelerate-report
    
- CPU cache
  - 利用cache line，读入cpu的结构体填满一个cache line
    
- 系统调用剥离

Redis lock
=======

`redmutex` is distribute lock based on `Redis`.

based on [doc](https://redis.io/topics/distlock)

Redis `unlink` under the hook
=========

Both `DEL` and `UNLINK` free the key part in blocking mode. And the difference is the way they free the value part.

`DEL` always frees the value part in **blocking** mode. However, if the value is too large, e.g. too many allocations for a large `LIST` or `HASH`, it blocks Redis for a long time. In order to solve the problem, Redis implements the `UNLINK` command, i.e. an 'non-blocking' delete.

In fact, `UNLINK` is **NOT always non-blocking/async**. If the value is small, e.g. the size of `LIST` or `HASH` is less than __64__,
the value will be freed immediately. In this way, `UNLINK` is almost the same as `DEL`, except that it costs a few more function calls than `DEL`. 
However, if the value is large, Redis puts the value into a list, and the value will be freed by another thread i.e. the non-blocking free (`lazyfree.c`). 
In this way, the main thread has to do some synchronization with the background thread (`bio dbAsyncDelete`), and that's also a cost.

`dbAsyncDelete -->  BIO_LAZY_FREE list --> bioProcessBackgroundJobs --> delete key through signal broadcast`
```
bio job type:
  - BIO_CLOSE_FILE    0 /* Deferred close(2) syscall. */
  - BIO_AOF_FSYNC     1 /* Deferred AOF fsync. */
  - BIO_LAZY_FREE     2 /* Deferred objects freeing. */

```

In a conclusion, if __the value is small__, `DEL` is at least, as good as `UNLINK`. If value is very large, e.g. LIST with thousands or millions of items, UNLINK is much better than DEL. You can always safely replace DEL with UNLINK. However, if you find the thread synchronization becomes a problem (multi-threading is always a headache), you can rollback to DEL.

Since Redis 6.0, there's a new configuration: `lazyfree-lazy-user-del`. You can set it to be yes, and Redis will run the `DEL` command as if running a `UNLINK` command.

Kafka fast consumer
======

并行kafka consumer

Time Wheel
======

[Hashed and Hierarchical Timing Wheels](https://blog.acolyer.org/2015/11/23/hashed-and-hierarchical-timing-wheels/)

Delay Queue
======

- memory list (DelayQueue, based on time wheel)
- message queue (rabbitmq, kafka)
- redis 
  - [ref1](https://segmentfault.com/a/1190000022027194)
    - 新增一个JOB，会在ZING:DELAY_QUEUE:JOB_POOL中插入一条数据，记录了业务方消费方。ZING:DELAY_QUEUE:BUCKET也会插入一条记录，记录执行的时间戳
    - 搬运线程会去ZING:DELAY_QUEUE:BUCKET中查找哪些执行时间戳的RunTimeMillis比现在的时间小，将这些记录全部删除；同时会解析出每个任务的Topic是什么，然后将这些任务PUSH到TOPIC对应的列表ZING:DELAY_QUEUE:QUEUE中
    - 每个TOPIC的LIST都会有一个监听线程去批量获取LIST中的待消费数据，获取到的数据全部扔给这个TOPIC的消费线程池
    - 消费线程池执行会去ZING:DELAY_QUEUE:JOB_POOL查找数据结构，返回给回调结构，执行回调方法。
  - [redis lab](https://redislabs.com/ebook/part-2-core-concepts/chapter-6-application-components-in-redis/6-4-task-queues/6-4-2-delayed-tasks/)
  
gRPC internal
======

- gRPC [default timeout](https://github.com/grpc-ecosystem/grpc-gateway/blob/6d2b64e3a9edc3d206345280e594703a7d4c5543/runtime/context.go#L39-L41)

- `gRPC` and `gRPC-gateway` debug with wireshark 

```go
        // grpc-gateway
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(CustomHTTPError),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterGreeterHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	runtime.DefaultContextTimeout = 8 * time.Second // RST_STREAM from grpc gate-way to grpc server with error code cancel
	// 504 gate way timeout to client

	// 408 request timeout for curl  curl  http://localhost:8081/say/benishere  -m 5  with 5 seconds
	// client send fin-ack to grpc-gateway, and grpc-gateway forward RST_STREAM to grpc server
	svr := http.Server{ReadTimeout: 3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Addr:         ":8081",
		Handler:      mux,
	}
	
	// grpc server
        s := grpc.NewServer(grpc.StatsHandler(&serverStats{}),
            grpc.MaxRecvMsgSize(300000), // 429 Too Many Requests, received message larger than max
            grpc.KeepaliveParams(keepalive.ServerParameters{
                MaxConnectionIdle: time.Duration(15) * time.Second, // send GO_AWAY frame
                Timeout:           time.Duration(3) * time.Second,  // timeout of send PING frame
                Time:              time.Duration(5) * time.Second,  // send PING frame
            }), 
        )
```

- keepalive
  - what
    
     gRPC sends http2 pings on the transport to detect if the connection is down.
     If the ping is not acknowledged by the other side within a certain period, the connection will be closed.
  - why
    - fast to detect TCP connection failure. 
    - keep the connection alive, some L4 proxies are configured to kill 'idle' connection
  - how
    - client side
      ```go
      keepalive.ClientParameters{
        Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
        Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
        PermitWithoutStream: true,             // send pings even without active streams
      }
      ```
  
    - server side
      ```go
      keepalive.EnforcementPolicy{
        MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
        PermitWithoutStream: true,            // Allow pings even when there are no active streams
      }
      
      var kasp = keepalive.ServerParameters{
        MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
        MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
        MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
        Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
        Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
      }
      ```

      `Enforcement policy` is a special setting on server side to protect server from malicious or misbehaving clients.

       Server sends `GOAWAY` with `ENHANCE_YOUR_CALM` and close the connection when bad behaviors are detected:

         - Client sends too frequent pings
         - Client sends pings when there's no stream and this is disallowed by server config

  

Golang http 
=========

TODO:

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




