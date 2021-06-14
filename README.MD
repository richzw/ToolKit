
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

Kafka fast consumer
======

并行kafka consumer

TimeWheel
======

[Hashed and Hierarchical Timing Wheels](https://blog.acolyer.org/2015/11/23/hashed-and-hierarchical-timing-wheels/)


gRPC internal
======

- grpc [default timeout](https://github.com/grpc-ecosystem/grpc-gateway/blob/6d2b64e3a9edc3d206345280e594703a7d4c5543/runtime/context.go#L39-L41)



