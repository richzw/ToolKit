
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

Time Wheel
======

[Hashed and Hierarchical Timing Wheels  ](https://blog.acolyer.org/2015/11/23/hashed-and-hierarchical-timing-wheels/)

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

- grpc [default timeout](https://github.com/grpc-ecosystem/grpc-gateway/blob/6d2b64e3a9edc3d206345280e594703a7d4c5543/runtime/context.go#L39-L41)



