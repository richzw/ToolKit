
- [gRPC flow control](https://www.youtube.com/watch?v=EEEGBwEA8yA)
  - TCP congestion control
    - Common Algorithm
      - Reno
      - BIC
      - CUBIC
    - General Strategy
      - Incr sending rate if ACK
      - Decr if ACK missed
  - TCP flow control
    - Stop reading kernel buffer
    - Receiver drop further packets
    - Receiver being protected
    - Trigger Impact
      - Reduce throughput
      - Degraded multiplexing
  - gRPC flow control
    - Algorithm similar to Token Bucket (WINDOW_UPDATE Frame)
    - Features
      - High performance
      - Fine grained throttling - Stream (RPC)/ Connection
      - Frame Priority
  - Window Size
    - Solution: BDP estimator - Bandwidth Delay Product: the amount of data that can be in transit in the network
    - Goal: intelligently avoid triggering flow control
    - Measure BDP through PING frame and PID controller
    - Set the _init window size_ to BDP
  - Challenge
    - Fairness between RPCs  - HTTP2 flow control supports multiplexing
    - Throttle based on performance - gRPC has BDP estimator
    - Flow control From End to End - gRPC has built-in Flow control

- [Best Practice](https://www.youtube.com/watch?v=Z_yD7YPL2oE)

  - API Design - Idempotency   
    - Request should include timestamp/guid to make idempotency
  - API Design - Performance
    - Request: can imply unbounded work -- set limits
    - Response: pagination
    - Avoid long-running operation  - 
  - API Design - Defaults
    - Unset enums default to zero value, perfer UNKNOWN/UNSECIFIED as the default
    - Backward compatibility
  - API Design - Errors
    - Do not include in response payload in most cases
    - Avoid batching multiple, independent operations
  - Error Handling - Don't Panic!
    - Do not blindly return errors from libs or ther services
  - Deadlines - Propagation
    - `Context.WithDeadline(.. Time percificed)` or `WithTimeout(…  -` 
  - Rate Limiting
    - Local rate limits
      - `Grpc.InTapHandler(rateLimiter`)`
      - `Golang.org/x/time/rate`    -- rate.NewLimiter(…)
  - Retries.
    - Officail grpc plan to do that
      - Configured via server config
      - Supports
        - Sequential retries with backoff
        - Concurrent hedged request
    - Until then: use a client wrapper or interceptor
      - Accept a content and use its deadline
  - Memory Management
    - Grpc does not limit goroutines
      - Option1: set listener limits and concurrent stream limits
        - `Listener = netutil.LimitListener(listener, connectionLimit)`
        - `Grpc.NewServer(grpc.MaxConcurrentSteams(streamsLimit))`
      - Option2: use TapHandler to error when too many rpcs are in flight
      - Option3: use health report and load balance to redirect traffic
    - Large request can OOM
      - Set a max request payload size
         `Grpc.NewServer(grpc.MaxRecvMsgSize(4096/*bytes*/))`
  - Always re-use stubs and channels when possible.
  - Use _keepalive pings_ to keep HTTP/2 connections alive during periods of inactivity to allow initial RPCs to be made quickly without a delay (i.e. C++ channel arg GRPC_ARG_KEEPALIVE_TIME_MS).
  - Use _streaming RPCs_ when handling a long-lived logical flow of data from the client-to-server, server-to-client, or in both directions. Streams can avoid continuous RPC initiation, which includes connection load balancing at the client-side, starting a new HTTP/2 request at the transport layer, and invoking a user-defined method handler on the server side.
  - Each gRPC channel uses 0 or more HTTP/2 connections and each connection usually has a limit on the number of concurrent streams. When the number of active RPCs on the connection reaches this limit, additional RPCs are queued in the client and must wait for active RPCs to finish before they are sent. Applications with high load or long-lived streaming RPCs might see performance issues because of this queueing. There are two possible solutions:
    - Create a **separate channel** for each area of high load in the application.
    - Use a **pool of gRPC channels** to distribute RPCs over multiple connections (channels must have different channel args to prevent re-use so define a use-specific channel arg such as channel number).
      
      _Side note_: The gRPC team has plans to add a feature to fix these performance issues [grpc/grpc#21386](https://github.com/grpc/grpc/issues/21386)





