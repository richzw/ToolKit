
- 服务发现
    - 服务注册是针对服务端的，服务启动后需要注册，分为几个部分：
        - 启动注册
        - 定时续期
        - 退出撤销
    - 服务发现是针对调用端的，一般分为两类问题：
        - 存量获取
        - 增量侦听
        - 还有一个常见的工程问题是： 应对服务发现故障

- [Fail at Scale, Reliability in the face of rapid change](https://mp.weixin.qq.com/s/BNOr5e92atc2RZstv_afwQ) 
  - 如果大规模请求都慢了，会引起 Go GC 压力增加，最终导致服务不可用。Facebook 采了两种方法解决
    - Controlled Delay: 算法根据不同负载处理排队的请求，解决入队速率与处理请求速率不匹配问题
      ```go
      onNewRequest(req, queue):
      
        if (queue.lastEmptyTime() < (now - N seconds)) {
           timeout = M ms
        } else {
           timeout = N seconds;
        }
        queue.enqueue(req, timeout)
      ```
      如果过去 N 秒内队列不为空，说明处理不过来了，那么超时时间设短一些 M 毫秒，否则超时时间可以设长一些。Facebook 线上 M 设置为 5ms, N 是 100ms
    - Adaptive LIFO: 正常队列是 First In First Out 的，但是当业务请理慢，请求堆积时，超时的请求，用户可能己经重试了，还不如处理后入队的请求
      ![img.png](micro_service_lifo.png)
    - Concurreny Control: 并发控制, 论文描述的其实就是 circuit breaker, 如果 inflight 请求过多，或是错误过多，会触发 Client 熔断

- [Go 官方限流器](https://mp.weixin.qq.com/s/qAKQm9CVNxk-ltUxHql1nw)

  [常见的限流算法](https://mp.weixin.qq.com/s?__biz=MzUzNTY5MzU2MA==&mid=2247486937&idx=1&sn=d4ea6ebb38c52e8004e73f235bde9848&scene=21#wechat_redirect) 有固定窗口、滑动窗口、漏桶、令牌桶
  - 计数器是一种比较简单粗暴的限流算法，其思想是在固定时间窗口内对请求进行计数，与阀值进行比较判断是否需要限流，一旦到了时间临界点，将计数器清零
    - 计数器算法存在“时间临界点”缺陷, 计数器算法实现限流的问题是没有办法应对突发流量
    ```go
    type LimitRate struct {
       rate  int           //阀值
       begin time.Time     //计数开始时间
       cycle time.Duration //计数周期
       count int           //收到的请求数
       lock  sync.Mutex    //锁
    }
    
    func (limit *LimitRate) Allow() bool {
       limit.lock.Lock()
       defer limit.lock.Unlock()
    
       // 判断收到请求数是否达到阀值
       if limit.count == limit.rate-1 {
          now := time.Now()
          // 达到阀值后，判断是否是请求周期内
          if now.Sub(limit.begin) >= limit.cycle {
             limit.Reset(now)
             return true
          }
          return false
       } else {
          limit.count++
          return true
       }
    }
    
    func (limit *LimitRate) Set(rate int, cycle time.Duration) {
       limit.rate = rate
       limit.begin = time.Now()
       limit.cycle = cycle
       limit.count = 0
    }
    
    func (limit *LimitRate) Reset(begin time.Time) {
       limit.begin = begin
       limit.count = 0
    }
    ```
  - 滑动窗口算法将一个大的时间窗口分成多个小窗口，每次大窗口向后滑动一个小窗口，并保证大的窗口内流量不会超出最大值，这种实现比固定窗口的流量曲线更加平滑。
    - 滑动窗口算法是固定窗口的一种改进，但从根本上并没有真正解决固定窗口算法的临界突发流量问题
    - kratos框架里circuit breaker用循环列表保存timeSlot对象的实现，他们这个实现的好处是不用频繁的创建和销毁timeslot对象
     ```go
     type timeSlot struct {
      timestamp time.Time // 这个timeSlot的时间起点
      count     int       // 落在这个timeSlot内的请求数
     }
     
     // 统计整个时间窗口中已经发生的请求次数
     func countReq(win []*timeSlot) int {
      var count int
      for _, ts := range win {
       count += ts.count
      }
      return count
     }
     
     type SlidingWindowLimiter struct {
      mu           sync.Mutex    // 互斥锁保护其他字段
      SlotDuration time.Duration // time slot的长度
      WinDuration  time.Duration // sliding window的长度
      numSlots     int           // window内最多有多少个slot
      windows      []*timeSlot
      maxReq       int // 大窗口时间内允许的最大请求数
     }
     
     func NewSliding(slotDuration time.Duration, winDuration time.Duration, maxReq int) *SlidingWindowLimiter {
      return &SlidingWindowLimiter{
       SlotDuration: slotDuration,
       WinDuration:  winDuration,
       numSlots:     int(winDuration / slotDuration),
       maxReq:       maxReq,
      }
     }
     
     
     func (l *SlidingWindowLimiter) validate() bool {
      l.mu.Lock()
      defer l.mu.Unlock()
     
     
      now := time.Now()
      // 已经过期的time slot移出时间窗
      timeoutOffset := -1
      for i, ts := range l.windows {
       if ts.timestamp.Add(l.WinDuration).After(now) {
        break
       }
       timeoutOffset = i
      }
      if timeoutOffset > -1 {
       l.windows = l.windows[timeoutOffset+1:]
      }
     
      // 判断请求是否超限
      var result bool
      if countReq(l.windows) < l.maxReq {
       result = true
      }
     
      // 记录这次的请求数
      var lastSlot *timeSlot
      if len(l.windows) > 0 {
       lastSlot = l.windows[len(l.windows)-1]
       if lastSlot.timestamp.Add(l.SlotDuration).Before(now) {
        // 如果当前时间已经超过这个时间插槽的跨度，那么新建一个时间插槽
        lastSlot = &timeSlot{timestamp: now, count: 1}
        l.windows = append(l.windows, lastSlot)
       } else {
        lastSlot.count++
       }
      } else {
       lastSlot = &timeSlot{timestamp: now, count: 1}
       l.windows = append(l.windows, lastSlot)
      }
     
     
      return result
     }
     ```
  - 漏桶算法是首先想象有一个木桶，桶的容量是固定的。当有请求到来时先放到木桶中，处理请求的worker以固定的速度从木桶中取出请求进行相应。如果木桶已经满了，直接返回请求频率超限的错误码或者页面。
    - [漏桶算法](https://github.com/kevinyan815/gocookbook/issues/28) 是流量最均匀的限流实现方式，一般用于流量“整形”。例如保护数据库的限流，先把对数据库的访问加入到木桶中，worker再以db能够承受的qps从木桶中取出请求，去访问数据库。
    - 木桶流入请求的速率是不固定的，但是流出的速率是恒定的。这样的话能保护系统资源不被打满，但是面对突发流量时会有大量请求失败，不适合电商抢购和微博出现热点事件等场景的限流。
     ```go
     type LeakyBucket struct {
        rate       float64    // 每秒固定流出速率
        capacity   float64    // 桶的容量
        water      float64    // 当前桶中请求量
        lastLeakMs int64      // 桶上次漏水微秒数
        lock       sync.Mutex // 锁
     }
     
     func (leaky *LeakyBucket) Allow() bool {
        leaky.lock.Lock()
        defer leaky.lock.Unlock()
     
        now := time.Now().UnixNano() / 1e6
        // 计算剩余水量,两次执行时间中需要漏掉的水
        leakyWater := leaky.water - (float64(now-leaky.lastLeakMs) * leaky.rate / 1000)
        leaky.water = math.Max(0, leakyWater)
        leaky.lastLeakMs = now
        if leaky.water+1 <= leaky.capacity {
           leaky.water++
           return true
        } else {
           return false
        }
     }
     
     func (leaky *LeakyBucket) Set(rate, capacity float64) {
        leaky.rate = rate
        leaky.capacity = capacity
        leaky.water = 0
        leaky.lastLeakMs = time.Now().UnixNano() / 1e6
     }
     ```
  - 令牌桶是反向的"漏桶"，它是以恒定的速度往木桶里加入令牌，木桶满了则不再加入令牌。服务收到请求时尝试从木桶中取出一个令牌，如果能够得到令牌则继续执行后续的业务逻辑。如果没有得到令牌，直接返回访问频率超限的错误码或页面等，不继续执行后续的业务逻辑
    - 适合电商抢购或者微博出现热点事件这种场景，因为在限流的同时可以应对一定的突发流量。如果采用漏桶那样的均匀速度处理请求的算法，在发生热点时间的时候，会造成大量的用户无法访问，对用户体验的损害比较大。
     ```go
     type TokenBucket struct {
        rate         int64 //固定的token放入速率, r/s
        capacity     int64 //桶的容量
        tokens       int64 //桶中当前token数量
        lastTokenSec int64 //上次向桶中放令牌的时间的时间戳，单位为秒
     
        lock sync.Mutex
     }
     
     func (bucket *TokenBucket) Take() bool {
        bucket.lock.Lock()
        defer bucket.lock.Unlock()
     
        now := time.Now().Unix()
        bucket.tokens = bucket.tokens + (now-bucket.lastTokenSec)*bucket.rate // 先添加令牌
        if bucket.tokens > bucket.capacity {
           bucket.tokens = bucket.capacity
        }
        bucket.lastTokenSec = now
        if bucket.tokens > 0 {
           // 还有令牌，领取令牌
           bucket.tokens--
           return true
        } else {
           // 没有令牌,则拒绝
           return false
        }
     }
     
     func (bucket *TokenBucket) Init(rate, cap int64) {
        bucket.rate = rate
        bucket.capacity = cap
        bucket.tokens = 0
        bucket.lastTokenSec = time.Now().Unix()
     }
     ```
  - ![img.png](micro_service_ratelimit_gw.png)
  - `golang.org/x/time/rate`。该限流器也是基于 Token Bucket(令牌桶) 实现的。
  - `uber-go/ratelimit`也是一个很好的选择，与Golang官方限流器不同的是Uber的限流器是通过漏桶算法实现的
  - [Ref](https://mp.weixin.qq.com/s/dJ3hiuA-8BdNF_ENL-WIUg)
  ```go
    type Limiter struct {
    mu     sync.Mutex
    limit  Limit
    burst  int // 令牌桶的大小
    tokens float64
    last time.Time // 上次更新tokens的时间
    lastEvent time.Time // 上次发生限速器事件的时间（通过或者限制都是限速器事件）
    }
  ```
  其主要字段的作用是：

  - limit：limit字段表示往桶里放Token的速率，它的类型是Limit，是int64的类型别名。设置limit时既可以用数字指定每秒向桶中放多少个Token，也可以指定向桶中放Token的时间间隔，其实指定了每秒放Token的个数后就能计算出放每个Token的时间间隔了。
  - burst: 令牌桶的大小。
  - tokens: 桶中的令牌。
  - last: 上次往桶中放 Token 的时间。
  - lastEvent：上次发生限速器事件的时间（通过或者限制都是限速器事件）

  可以看到在 timer/rate 的限流器实现中，并没有单独维护一个 Timer 和队列去真的每隔一段时间向桶中放令牌，而是仅仅通过计数的方式表示桶中剩余的令牌。每次消费取 Token 之前会先根据上次更新令牌数的时间差更新桶中Token数。

- [如何获取客户端真实 IP](https://mp.weixin.qq.com/s/C-Xf6haLrOWkmBm2lRTdEQ)
  - background
    - 通常我们可以通过 HTTP 协议 Request Headers 中 X-Forwarded-For 头来获取真实 IP。然而通过 X-Forwarded-For 头获取真实 IP 的方式真的可靠么？
  - 方案
    ```go
    // ClientIP 方法可以获取到请求客户端的IP
    func (c *Context) ClientIP() string {
       // 1. ForwardedByClientIP 默认为 true，此处会优先取 X-Forwarded-For 值，
       // 如果 X-Forwarded-For 为空，则会再尝试取 X-Real-Ip
       if c.engine.ForwardedByClientIP {
          clientIP := c.requestHeader("X-Forwarded-For")
          clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
          if clientIP == "" {
             clientIP = strings.TrimSpace(c.requestHeader("X-Real-Ip"))
          }
          if clientIP != "" {
             return clientIP
          }
       }
       // 2. 如果我们手动配置 ForwardedByClientIP 为 false 且 X-Appengine-Remote-Addr 不为空，则取 X-Appengine-Remote-Addr 作为客户端IP
       if c.engine.AppEngine {
          if addr := c.requestHeader("X-Appengine-Remote-Addr"); addr != "" {
             return addr
          }
       }
       // 3. 最终才考虑取对端 IP 兜底
       if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
          return ip
    }
       return ""
    }
    ```
    Gin 框架到 1.7.2 后突然发现一个 『Bug』，升级后服务端无法获正确的客户端 IP，取而代之的是 Kubernetes 集群中 Nginx Ingress IP
    https://github.com/gin-gonic/gin/issues/2697
    
  - X Forward For 伪造？
    - 客户端是否能伪造 IP，取决于边缘节点（Edge Node）是如何处理 X-Forwarded-For 字段。客户端直接连接的首个 Proxy 节点都叫做边缘节点（Edge Node），无论是网关、CDN、LB 等，只要这一层是直接接入客户端访问的，那么它就是一个边缘节点。
    - 不重写 X-Forwarded-For 的边缘节点 边缘节点如果是透传 HTTP 的 X-Forwarded-For 头，那么它就是不安全的，客户端可以在 HTTP 请求中伪造 X-Forwarded-For 值，且这个值会被向后透传。
    - 重写 X-Forwarded-For 的边缘节点 边缘节点如果重写 $remote_addr 到 X-Forwarded-For ，那么这就是安全的。
  - 继续尝试通过 X-Forwarded-For 获取客户端真实 IP
    - 业务中需配置基础设施所有前置代理到 TrustedProxies 中，包含 CDN 地址池、WAF 地址池、Kunernetest Nginx Ingress 地址池，这种方案基本无法落地：
    - 配置太过复杂，一旦获取 IP 不准，很难排查。
    - 导致业务配置和基础设施耦合，基础设施如果对 CDN、WAF、Ingress 做变动，业务代码必须同步变更。
    - 部分可信代理 IP 根本没法配置，比如 CDN 地址池
  - 尝试通过自定义 Header 获取客户端真实 IP
    - 基础设施团队提供自定义 Header 来获取客户端真实 IP，如 X-Client-Real-IP 或 X-Appengine-Remote-Addr 。这种方案需要基础设施团队在云厂商 CDN 或 WAF 终端上做好相应的配置。这种方案：
    - 配置简单可靠，维护成本低，仅需在 CDN、WAF 终端配置自定义 Header 即可。
    - 如果使用 X-Appengine-Remote-Addr，对于使用 Google Cloud 的 App Engine 的服务不需做任何修改。对于使用的国内云厂商的服务，则需要显式的配置 engine. AppEngine = true，然后继续通过 ctx.ClientIP() 方法即可。 
    - 如果使用其他自定义 Header，如 X-Client-Real-IP 来获取客户端真实 IP，建议可以考虑自行封装 ClientIP(*gin.Context) string 函数，从 X-Client-Real-IP 中获取客户端 IP。
- [一些常见的源 IP 保持方法](https://mp.weixin.qq.com/s/-ntUY5leJva3tWCqTb_ecw)
  - 七层协议的源 IP 保持 - X-Forwarded-For
  - 四层协议的源 IP 保持
    - DNAT
      - IPVS/iptables都支持 DNAT，客户端通过 VIP 访问 LB，请求报文到达 LB 时，LB 根据连接调度算法选择一个后端 Server，将报文的目标地址 VIP 改写成选定 Server 的地址，报文的目标端口改写成选定 Server 的相应端口，最后将修改后的报文发送给选出的 Server。由于 LB 在转发报文时，没有修改报文的源 IP，所以，后端 Server 可以看到客户端的源 IP
    - Transparent Proxy
      - Nginx/Haproxy 支持透明代理(Transparent Proxy)。当开启该配置时，LB 与后端服务建立连接时，会将 socket 的源 IP 绑定为客户端的 IP 地址，这里依赖内核TPROXY以及 socket 的 IP_TRANSPARENT 选项。
    - TOA
      - TOA(TCP Option Address)是基于四层协议（TCP）获取真实源 IP 的方法，本质是将源 IP 地址插入 TCP 协议的 Options 字段。这需要内核安装对应的TOA内核模块。
    - Proxy Protocol
      - Proxy Protocol是 Haproxy 实现的一个四层源地址保留方案。它的原理特别简单，Proxy 在与后端 Server 建立 TCP 连接后，在发送实际应用数据之前，首先发送一个Proxy Protocol协议头(包括客户端源 IP/端口、目标IP/端口等信息)。这样，后端 server 通过解析协议头获取真实的客户端源 IP 地址。
- [CDC](https://mp.weixin.qq.com/s/kzRf6zMDfElrN-VKXfo_Cw)
  - Change Data Capture[1] 简称 CDC, 用于异构数据同步，将 database 的数据同步到第三方，这里的 DB 可以是 MySQL, PG, Mongo 等等一切数据源，英文技术圈称之为 Single Source OF True (SSOT), 目标称为 Derived Data Systems。
    ![img.png](micro_service_cdc.png)
    - alibaba canal
    - go-mysql
  - CDC 旁路模式做同步，主要有三种旁路原理：
    - 触发器 trigger 将更新通过触发器写到专用表 table, 缺点非常多，性能不好，不能异构写 
    - 增加 time 字段，数据惰性删除标记 NULL，CDC 服务定期扫该表。缺点实时无法保证，而且对业务有侵入，要求业务适配 
    - 解析 log, 当前主流实现，Mongodb 读取 oplog, MySQL 读取 binlog

    red hat 出品的 Debezium[6] 开源软件，数据源支持 MySQL,MongoDB,PostgreSQL,SQL Server,Oracle,Db2,Cassandra, 通过 kafka connect 获得增量数据写到 kafka, 供后续业务订阅与消费
  - 事务
    - 基于 binlog 解析实现的 CDC 要求 MySQL 开启 row 模式，并且建义 full image 记录全量字段，这点保证了数据更新的 order 顺序，但是没有保证事务的原子性
    - 解决方案，ringbuffer 等待接收事务所有数据，然后一起提交到消费端
  - 那么如何同步全部基量数据呢
    - Full State Capturue 简称 [FSC](https://netflixtechblog.com/dblog-a-generic-change-data-capture-framework-69351fb9099b)
    - 不做全量备份，开通时间窗口，在窗口期内通过主键 pk 按 chunk (比如 10 条记录一批)扫表，混合扫表和 binlog 解析出来的数据，去掉冲突的部份，不断重复这个过程，直接全部扫完
  - 高可用
    - 可以定期将消费的 binlog pos 同步到 zk/etcd 等外部存储，多个 CDC 服务竞争任务
  - Chore
    - CQRS Command and Query Responsibility Segregation命令查询分离走进视野， [CQRS](https://docs.microsoft.com/en-us/azure/architecture/patterns/cqrs), Event Sourcing

- [负载均衡](https://mp.weixin.qq.com/s/RFjTDNsGiqkMNnkjVus5Zw)
  - 负载均衡的主要作用
    - 高并发：负载均衡通过算法调整负载，尽力均匀的分配应用集群中各节点的工作量，以此提高应用集群的并发处理能力（吞吐量）。
    - 伸缩性：添加或减少服务器数量，然后由负载均衡进行分发控制。这使得应用集群具备伸缩性。
    - 高可用：负载均衡器可以监控候选服务器，当服务器不可用时，自动跳过，将请求分发给可用的服务器。这使得应用集群具备高可用的特性。
    - 安全防护：
  - 软件负载均衡的 主流产品 有：Nginx、HAProxy、LVS。
    - LVS 可以作为四层负载均衡器。其负载均衡的性能要优于 Nginx。
    - HAProxy 可以作为 HTTP 和 TCP 负载均衡器。
    - Nginx、HAProxy 可以作为四层或七层负载均衡器
  - 软件负载均衡从通信层面来看，又可以分为四层和七层负载均衡。
    - 七层负载均衡：就是可以根据访问用户的 HTTP 请求头、URL 信息将请求转发到特定的主机。
      - DNS 重定向
      - HTTP 重定向
      - 反向代理
    - 四层负载均衡：基于 IP 地址和端口进行请求的转发。
      - 修改 IP 地址
      - 修改 MAC 地址
    - DNS 负载均衡
      - 工作原理就是：基于 DNS 查询缓存，按照负载情况返回不同服务器的 IP 地址。
      - 优点：负载均衡工作，交给 DNS 服务器处理，省掉了负载均衡服务器维护的麻烦
      - 缺点：
        - 可用性差 DNS 解析是多级解析，新增/修改 DNS 后，解析时间较长；解析过程中，用户访问网站将失败
        - 扩展性低：DNS 负载均衡的控制权在域名商那里，无法对其做更多的改善和扩展；
        - 维护性差：也不能反映服务器的当前运行状态；支持的算法少；
    - HTTP 负载均衡
    - 反向代理负载均衡
    - IP 负载均衡是在网络层通过修改请求目的地址进行负载均衡
    - 数据链路层负载均衡是指在通信协议的数据链路层修改 mac 地址进行负载均衡
      - 在 Linux 平台上最好的链路层负载均衡开源产品是 LVS (Linux Virtual Server)。LVS 是基于 Linux 内核中 netfilter 框架实现的负载均衡系统。netfilter 是内核态的 Linux 防火墙机制，可以在数据包流经过程中，根据规则设置若干个关卡（hook 函数）来执行相关的操作
      - 进入 PREROUTING 后经过路由查找，确定访问的目的 VIP 是本机 IP 地址，所以数据包进入到 INPUT 链上
      - IPVS 是工作在 INPUT 链上，会根据访问的 vip+port 判断请求是否 IPVS 服务，如果是则调用注册的 IPVS HOOK 函数，进行 IPVS 相关主流程，强行修改数据包的相关数据，并将数据包发往 POSTROUTING 链上。
      - POSTROUTING 上收到数据包后，根据目标 IP 地址（后端服务器），通过路由选路，将数据包最终发往后端的服务器上。
      
- [微服务的超时传递](https://mp.weixin.qq.com/s/vZlFS7lkKplil5cyHirSmA)
  ![img.png](micro_service_timeout.png)
  当进行context传递的时候，比如上图中请求Redis，那么通过如下方式获取剩余时间，然后对比Redis设置的超时时间取较小的时间
    ```go
    dl, ok := ctx.Deadline()
    timeout := time.Now().Add(time.Second * 3)
    if ok := dl.Before(timeout); ok {
        timeout = dl
    }
    ```
  服务间超时传递主要是指 RPC 调用时候的超时传递，对于 gRPC 来说并不需要要我们做额外的处理，gRPC 本身就支持超时传递，原理和上面差不多，是通过 metadata 进行传递，最终会被转化为 grpc-timeout 的值，如下代码所示 grpc-go/internal/transport/handler_server.go:79
    ```go
    if v := r.Header.Get("grpc-timeout"); v != "" {
      to, err := decodeTimeout(v)
      if err != nil {
       return nil, status.Errorf(codes.Internal, "malformed time-out: %v", err)
      }
      st.timeoutSet = true
      st.timeout = to
    }
    ```
- HTTP 鉴权
  - 鉴权的本质：用户 (user / service) 是否有以及如何获得权限 (Authority) 去操作 (Operate) 哪些资源 (Resource)
    - 认证（Authentication）：系统如何正确分辨出操作用户的真实身份？认证方式有很多种，后面会讲解
    - 授权（Authorization）：系统如何控制一个用户该看到哪些数据、能操作哪些功能？授权与认证是硬币的两面
    - 凭证（Credential）：系统如何保证它与用户之间的承诺是双方当时真实意图的体现，是准确、完整且不可抵赖的？我们一般把凭证代称为 TOKEN
    - 保密（Confidentiality）：系统如何保证敏感数据无法被包括系统管理员在内的内外部人员所窃取、滥用？这里要求我们不能存储明文到 DB 中，不能将密码写到 http url 中，同时要求 id 服务仅有少部分人能够访问，并且有审计
    - 传输（Transport Security）：系统如何保证通过网络传输的信息无法被第三方窃听、篡改和冒充？内网无所谓了，外网一般我们都用 https
    - 验证（Verification）：系统如何确保提交到每项服务中的数据是合乎规则的，不会对系统稳定性、数据一致性、正确性产生风险
  - 验证方式
    - 需要将凭证 Credential 放到 header Authorization 里，凭证可是是用户名密码，也可以是自定义生成的 TOKEN
      - Basic/Digest
      - HMAC 
      - Oauth2
- [缓存更新的套路](https://coolshell.cn/articles/17416.html)
  - Cache aside - Why does Facebook use delete to remove the key-value pair in Memcached instead of updating the Memcached during write request to the backend?
    - 失效：应用程序先从cache取数据，没有得到，则从数据库中取数据，成功后，放到缓存中。
    - 命中：应用程序从cache中取数据，取到后返回。
    - 更新：先把数据存到数据库中，成功后，再让缓存失效。
    - 是不是Cache Aside这个就不会有并发问题了？
      - 不是的，比如，一个是读操作，但是没有命中缓存，然后就到数据库中取数据，此时来了一个写操作，写完数据库后，让缓存失效，然后，之前的那个读操作再把老的数据放进去，所以，会造成脏数据。
      - 但，这个case理论上会出现，不过，实际上出现的概率可能非常低，因为这个条件需要发生在读缓存时缓存失效，而且并发着有一个写操作。而实际上数据库的写操作会比读操作慢得多，而且还要锁表，而读操作必需在写操作前进入数据库操作，而又要晚于写操作更新缓存，所有的这些条件都具备的概率基本并不大。
      - 最好还是为缓存设置上过期时间
  - Read through
    - Read Through 套路就是在查询操作中更新缓存，也就是说，当缓存失效的时候（过期或LRU换出），Cache Aside是由调用方负责把数据加载入缓存，而Read Through则用缓存服务自己来加载，从而对应用方是透明的。
  - Write through
    - Write Through 套路和Read Through相仿，不过是在更新数据时发生。当有数据更新的时候，如果没有命中缓存，直接更新数据库，然后返回。如果命中了缓存，则更新缓存，然后再由Cache自己更新数据库（这是一个同步操作）
  - Write behind caching
    - Write Behind 又叫 Write Back。一些了解Linux操作系统内核的同学对write back应该非常熟悉，这不就是Linux文件系统的Page Cache的算法
    - 在更新数据的时候，只更新缓存，不更新数据库，而我们的缓存会异步地批量更新数据库。
    - 其带来的问题是，数据不是强一致性的，而且可能会丢失
    ![img_1.png](micro_service_write_back.png)

  - 缓存系统更新过程之所以会出现数据不一致问题，有内外两大原因：
    - 内部原因：Redis和MySQL的更新不是天然的原子操作，非事务性的组合拳。
    - 外部原因：实际中的读写请求是并发且无序的，可预测性很差，完全不可控。
  - 先淘汰缓存&再更新主存
    - 为了防止其他线程读到缓存中的旧数据，干脆淘汰掉，然后把数据更新到主存储，后续的请求再次读取时触发Cache Miss，从而读取MySQL再将新数据更新到Redis。
    - 明显的漏洞：在淘汰Redis的数据完成后，更新MySQL完成之前，这个时间段内如果有新的读请求过来，发现Cache Miss了，就会把旧数据重新写到Redis中，再次造成不一致，并且毫无察觉后续读的都是旧数据。
  - 先更新主存&再淘汰缓存
    - 先更新MySQL，成功之后淘汰缓存，后续读取请求时触发Cache Miss再将新数据回写Redis
    - 问题：
      ![img.png](micro_service_cache.png)
      - 事件A：更新MySQL前出现一个读请求，且缓存中无数据出现cache miss
      - 事件B：T3时刻回写Redis的操作才完成，在此之前T2时刻清除了缓存
      - 从实际考虑这种综合事件发生的概率非常低，因为写操作远慢于读操作。
  - 延时双删策略
    ![img.png](micro_service_cache1.png)
  - 异步更新缓存
    - 把MySQL的更新操作完成后不直接操作Redis，而是把这个操作命令(消息)扔到一个中间层，然后由Redis自己来消费更新数据，这是一种解耦的异步方案
- [微服务熔断器](https://mp.weixin.qq.com/s/ik_Dnjzmdm_tDbWBbJ_LDg)
  - 初始状态下熔断器是关闭的，当故障超过配置的阈值，则会打开
  - 熔断器是打开状态，在经过一段熔断时间后，部分会打开
  - 如果熔断器是半开的，它可以
    - 再次打开，如果允许通过的请求也失败了
    - 关闭，如果允许通过的请求成功响应
  ![img.png](micro_service_circuit_break.png)
  - Hystrix 
  ![img.png](micro_service_hytrix.png)
- [Raft 分布式共识算法讲义](https://mp.weixin.qq.com/s/JOzK15Y85FFwIuMwJifIAw)
  - [Source](https://www.youtube.com/watch?v=vYp4LYbnnW8)
- Raft vs Paxos
  - Multi-Paxos是自下而上（bottom-up）设计的弱Leader，而Raft是自上而下（top-down）设计的强Leader
    - Raft算法为了设计出易于理解的算法，采用了问题分解和简化状态空间的方法。问题分解体现在它将Leader选举、日志复制和成员变更拆解成子问题，分而治之
    - 为了简化状态空间，Raft增加了更多约束，例如不允许日志出现空洞、限制了日志之间可能变得不一致的方式等，这样得到的系统会更易于推理和校验
  - 共同点
    - 它们都是通过Replicated log来实现
    - 任何时候服务器都可能处于以下三种状态之一：
      - Leader，负责使用 AppendEntries RPC 将写请求添加到replicated log中。
      - Follower，回应Leader发送的AppendEntries RPC。
      - Candidate，Candidate会试图使用 RequestVotes RPC 成为Leader
    - 服务器都会存储一个递增的数字叫做Term（任期号），每条RPC消息都会包含Term。当服务器接收到RPC时，它会检查RPC里包含的Term与自身Term的大小
  - 差异点
    - Leader选举
      - Paxos允许任何服务器成为Leader，然后再从其他服务器补齐缺少的日志，而Raft只允许拥有最新日志的服务器成为Leader
    - Leader是否会修改日志的term
      - 在Raft 中，日志条目的term不会被未来的leader所修改。而在Paxos协议下，这是可能的
    - 日志的连续性
      - Raft 的日志是严格顺序写入的，而Paxos通常允许无序写入，但需要额外的协议来填充可能因此发生的日志间隙。
    - 日志条目提交的条件
      - 在Raft协议里，即使日志条目被多数派的节点所接受（但未提交），也有可能被回滚掉
      - 而在Paxos系统里，只要日志条目被复制到了多数派服务器上，Paxos就可以安全地提交日志条目。
    - Raft增加了一些约束，例如只允许日志从Leader流向其他服务器，当Follower数据与Leader不一致时，Leader会强制Follower的日志复制自己的日志，用Leader的日志条目来覆盖Follower日志中的冲突条目。这使得Raft成为一个强Leader的算法。
- [eBPF 代替 iptables 优化服务网格数据面性能](https://mp.weixin.qq.com/s/U6-wcBsBC-Khffb7kTBtjA)
  - iptables 实现流量劫持
    - 一个 Pod 的创建过程，sidecar injector 会向 Pod 中注入两个容器，istio-init 和 istio-proxy
    - istio-init 是一个 init container，负责创建流量劫持相关的 iptables 规则，在创建完成后会退出
    - istio-proxy 中运行着 envoy，负责代理 Pod 的网络流量，iptables 会将请求劫持到 istio-proxy 处理
    ![img.png](micro_service_network_iptable.png)
  - eBPF
    - eBPF(extended Berkeley Packet Filter) 是一种可以在 Linux 内核中运行用户编写的程序，而不需要修改内核代码或加载内核模块的技术
    - Inbound
      - eBPF 程序会劫持 bind 系统调用并修改地址，例如应用程序 bind 0.0.0.0:80 会被修改为 127.0.0.1:80，应用程序还有可能 bind ipv6 的地址，所以这里有两个 eBPF 程序分别处理 ipv4 和 ipv6 的 bind。
      - 和 iptables 不同，iptables 可以针对每个 netns 单独设置规则，eBPF 程序 attach 到指定 hook 点后，会对整个系统都生效，例如 attach 到 bind 系统调用后，所有 Pod 内以及节点上进程调用 bind 都会触发 eBPF 程序，我们需要区分哪些调用是来自需要由 eBPF 完成流量劫持的 Pod。
      ![img.png](micro_service_ebpf_inbound.png)
    - Outbound
      - TCP
        - _coonect4 通过劫持 connect 系统调用将目的地址修改为127.0.0.1:15001，也就是 envoy 的 VirtualOutboundListerer，同时将连接的原始目的地址保存在 sk_storage_map
        - 在 TCP 连接建立完成后，sockops 会读取 sk_storage_map 中的数据，并以四元组（源IP、目的IP、源端口、目的端口）为 key 将原始目的地址保存在 origin_dst_map
        - _getsockopt通过劫持 getsockopt 系统调用，读取 origin_dst_map 中的数据将原始目的地址返回给 envoy
        ![img.png](micro_service_ebpf_tcp.png)
      - UDP
        - _connect4 和 _sendmsg4 都是负责修改 UDP 的目的地址为 127.0.0.1:15053 并保存原始的目的地址到 sk_storage_map，因为 Linux 提供两种发送 UDP 数据的方式
        - 先调用 connect 再调用 send，这种情况由 _connect4 处理
        - 直接调用 sendto，这种情况由 _sendmsg4 处理
        - recvmsg4 通过读取 sk_storage_map 将回包的源地址改为原始的目的地址，这是因为有些应用程序，例如 nslookup 会校验回包的源地址。
        ![img.png](micro_service_ebpf_udp.png)
- [定时任务系统](https://mp.weixin.qq.com/s/poRF-xFgBHT-AYpBMnQCoA)
  - 操作系统的时间系统
    - PC 机中有两个时钟源，他们分别叫做 RTC（Real Time Clock，实时时钟） 和 OS（操作系统）时钟。
      - RTC（Real Time Clock，实时时钟）也叫做 CMOS 时钟，它是 PC 主机板上的一块芯片（或者叫做时钟电路），它靠电池供电，即使系统断电也可以维持日期和时间. 也被称为硬件时钟
      - OS 时钟产生于 PC 主板上的定时/计数芯片（8253/8254），由操作系统控制这个芯片的工作，OS 时钟的基本单位就是该芯片的计数周期。OS 时钟并不是本质意义上的时钟，它更应该被称为一个计数器. OS 时钟只在开机时才有效，而且完全由操作系统控制，所以也被称为软时钟或系统时钟
    - 时钟中断
      - Linux 的 OS 时钟的物理产生原因是可编程定时/计数器产生的输出脉冲，这个脉冲送入 CPU，就可以引发一个中断请求信号，我们就把它叫做时钟中断
      - 时钟中断执行的频率很高：100 次/秒（Linux 设计者将一个时钟滴答（tick）定义为 10ms），时钟中断的主要工作是处理和时间有关的所有信息、决定是否执行调度程序。
      - 在多处理器系统下，时钟中断实际上是分成两个部分：
        - 局时钟中断，系统中每个 tick 只发生一次。对应的中断处理程序用于更新系统时间和统计系统负载。
        - 本地时钟中断，系统中每个 tick 在每个 CPU 上发生一次。对应的中断处理程序用于统计对应 CPU 和运行于该CPU上的进程的时间，以及触发对应 CPU 上的定时器
    - 时钟中断的应用
      - 更新系统时间：在 Linux 内核中，全局变量 jiffies_64 用于记录系统启动以来所经历的 tick 数
        - 而时钟中断存在丢失的可能。内核中的某些临界区是不能被中断的，所以进入临界区前需要屏蔽中断。
        - 当中断屏蔽取消的时候，硬件只能告诉内核是否曾经发生了时钟中断、却不知道已经发生过多少次。
        - 内核中的全局变量 xtime 用于记录当前时间（自 1970-01-01 起经历的秒数、本秒中经历的纳秒数）。xtime 的初始值就是内核启动时从 RTC 读出的。
        - 在时钟中断处理程序更新 jiffies_64 的值后，便更新 xtime 的值。通过比较 jiffies_64 的当前值与上一次的值（上面说到，差值可能大于 1），决定 xtime 应该更新多少。
        - 系统调用 gettimeofday（对应库函数 time 和 gettimeofday）就是用来读 xtime 变量的，从而让用户程序获取系统时间。
      - 实现定时器：既然已知每个 tick 是 10ms，用 tick 来做定时任务统计再好不过。无论是内核还是应用系统其实都有大量的定时任务需求，这些定时任务类型不一，但是都是依赖于 tick。
    - 操作系统实现定时任务的方式有哪些呢
      - 维护一个带过期时间的任务链表
        - 在一个全局链表中维护一个定时任务链。每次 tick 中断来临，遍历该链表找到 expire 到期的任务
      - 时间轮（Timing-Wheel）算法
        - 时间轮很容易理解，上图有 n 个 bucket，每一个 bucket 表示一秒，当前 bucket 表示当前这一秒到来之后要触发的事件。 每个 bucket 会对应着一个链表，链表中存储的就是当前时刻到来要处理的事件。
        - 扩展算法，即 Hierarchy 时间轮
      - 维护一个基于小根堆算法的定时任务
  - 开箱即用的定时器
    - cron/crontab
    - Netty 封装的时间轮：HashedWheelTimer - 采用了单轮+round 的模式
    - Kafka 中的时间轮：TimingWheel - 多轮的模式
- [Redis分布式锁使用不当](https://mp.weixin.qq.com/s?__biz=MzA5OTAyNzQ2OA==&mid=2649742999&idx=1&sn=dd34f34d916297023c1ab338b2d0d989&chksm=8893e1f4bfe468e27e49c2204b77218d56e34c417820581849ca1d5083c85977c9b6e04ab671&scene=21#wechat_redirect)
  - 案例
     ```java
     public SeckillActivityRequestVO seckillHandle(SeckillActivityRequestVO request) {
     SeckillActivityRequestVO response;
         String key = "key:" + request.getSeckillId;
         try {
             Boolean lockFlag = redisTemplate.opsForValue().setIfAbsent(key, "val", 10, TimeUnit.SECONDS);
             if (lockFlag) {
                 // HTTP请求用户服务进行用户相关的校验
                 // 用户活动校验
                 
                 // 库存校验
                 Object stock = redisTemplate.opsForHash().get(key+":info", "stock");
                 assert stock != null;
                 if (Integer.parseInt(stock.toString()) <= 0) {
                     // 业务异常
                 } else {
                     redisTemplate.opsForHash().increment(key+":info", "stock", -1);
                     // 生成订单
                     // 发布订单创建成功事件
                     // 构建响应VO
                 }
             }
         } finally {
             // 释放锁
             stringRedisTemplate.delete("key");
             // 构建响应VO
         }
         return response;
     }
     ```
    抢购活动开始的一瞬间，大量的用户校验请求打到了用户服务。导致用户服务网关出现了短暂的响应延迟，有些请求的响应时长超过了10s，但由于HTTP请求的响应超时我们设置的是30s，这就导致接口一直阻塞在用户校验那里，10s后，分布式锁已经失效了，此时有新的请求进来是可以拿到锁的，也就是说锁被覆盖了。这些阻塞的接口执行完之后，又会执行释放锁的逻辑，这就把其他线程的锁释放了，导致新的请求也可以竞争到锁~这真是一个极其恶劣的循环。 这个时候只能依赖库存校验，但是偏偏库存校验不是非原子性的，采用的是get and compare的方式，超卖的悲剧就这样发生了
  - 事故分析
    - 没有其他系统风险容错处理 ，由于用户服务吃紧，网关响应延迟，但没有任何应对方式，这是超卖的导火索。
    - 看似安全的分布式锁其实一点都不安全，虽然采用了set key value [EX seconds] [PX milliseconds] [NX|XX]的方式，但是如果线程A执行的时间较长没有来得及释放，锁就过期了，此时线程B是可以获取到锁的。当线程A执行完成之后，释放锁，实际上就把线程B的锁释放掉了。这个时候，线程C又是可以获取到锁的，而此时如果线程B执行完释放锁实际上就是释放的线程C设置的锁。这是超卖的直接原因。
    - 非原子性的库存校验，非原子性的库存校验导致在并发场景下，库存校验的结果不准确。这是超卖的根本原因。
  - 解决方案
    - 实现相对安全的分布式锁
      - 实现相对安全的分布式锁，必须依赖key的value值。在释放锁的时候，通过value值的唯一性来保证不会勿删。
    - 实现安全的库存校验
      - 可以不用基于LUA脚本实现而是基于redis本身的原子性
    - 改进之后的代码
     ```java
     public SeckillActivityRequestVO seckillHandle(SeckillActivityRequestVO request) {
     SeckillActivityRequestVO response;
         String key = "key:" + request.getSeckillId();
         String val = UUID.randomUUID().toString();
         try {
             Boolean lockFlag = distributedLocker.lock(key, val, 10, TimeUnit.SECONDS);
             if (!lockFlag) {
                 // 业务异常
             }
     
             // 用户活动校验
             // 库存校验，基于Redis本身的原子性来保证
             Long currStock = stringRedisTemplate.opsForHash().increment(key + ":info", "stock", -1);
             if (currStock < 0) { // 说明库存已经扣减完了。
                 // 业务异常。
                 log.error("[抢购下单] 无库存");
             } else {
                 // 生成订单
                 // 发布订单创建成功事件
                 // 构建响应
             }
         } finally {
             distributedLocker.safedUnLock(key, val);
             // 构建响应
         }
         return response;
     }
     ```
    - 深度思考
      - 分布式锁有必要么
      - 分布式锁的选型，有人提出用RedLock来实现分布式锁。RedLock的可靠性更高
      - 网关层基于用户ID做hash算法来决定请求到哪一台服务器。这样就可以基于应用缓存来实现库存的扣减和判断
      ```java
      // 通过消息提前初始化好，借助ConcurrentHashMap实现高效线程安全
      private static ConcurrentHashMap<Long, Boolean> SECKILL_FLAG_MAP = new ConcurrentHashMap<>();
      // 通过消息提前设置好。由于AtomicInteger本身具备原子性，因此这里可以直接使用HashMap
      private static Map<Long, AtomicInteger> SECKILL_STOCK_MAP = new HashMap<>();
      
      ...
      
      public SeckillActivityRequestVO seckillHandle(SeckillActivityRequestVO request) {
      SeckillActivityRequestVO response;
      
          Long seckillId = request.getSeckillId();
          if(!SECKILL_FLAG_MAP.get(requestseckillId)) {
              // 业务异常
          }
           // 用户活动校验
           // 库存校验
          if(SECKILL_STOCK_MAP.get(seckillId).decrementAndGet() < 0) {
              SECKILL_FLAG_MAP.put(seckillId, false);
              // 业务异常
          }
          // 生成订单
          // 发布订单创建成功事件
          // 构建响应
          return response;
      }
      ```
      此方案没有考虑到机器的动态扩容、缩容等复杂场景，如果还要考虑这些话，则不如直接考虑分布式锁的解决方案
- [缓存最终一致性的解决方案](https://mp.weixin.qq.com/s/Y9S89MT0uAobzRKgYVrI9Q)
  - Cache-Aside
    - 在读请求中，首先请求缓存，若缓存命中（ cache hit ），则直接返回缓存中的数据；若缓存未命中（ cache miss ），则查询数据库并将查询结果更新至缓存，然后返回查询出的数据（ demand-filled look-aside ）。在写请求中，先更新数据库，再删除缓存（write-invalidate）。
    - 为什么删除缓存，而不是更新缓存？
      - 首先是性能，当该缓存对应的结果需要消耗大量的计算过程才能得到时，比如需要访问多张数据库表并联合计算，那么在写操作中更新缓存的动作将会是一笔不小的开销。同时，当写操作较多时，可能也会存在刚更新的缓存还没有被读取到，又再次被更新的情况（这常被称为缓存扰动），显然，这样的更新是白白消耗机器性能的，会导致缓存利用率不高。而等到读请求未命中缓存时再去更新，也符合懒加载的思路，需要时再进行计算。删除缓存的操作不仅是幂等的，可以在发生异常时重试，而且写-删除和读-更新在语义上更加对称。
      - 其次是安全，在并发场景下，在写请求中更新缓存可能会引发数据的不一致问题。参考下面的图示，若存在两个来自不同线程的写请求，首先来自线程 1 的写请求更新了数据库（ step 1 ），接着来自线程 2 的写请求再次更新了数据库（ step 3 ），但由于网络延迟等原因，线程 1 可能会晚于线程 2 更新缓存（ step 4 晚于 step 3 ），那么这样便会导致最终写入数据库的结果是来自线程 2 的新值，写入缓存的结果是来自线程 1 的旧值，即缓存落后于数据库，此时再有读请求命中缓存（ step 5 ），读取到的便是旧值。
      - ![img.png](micro_service_cache_consistent.png)
    - 为什么先更新数据库，而不是先删除缓存？
      - 问题仍然出现在并发场景下，首先来自线程 1 的写请求删除了缓存（ step 1 ），接着来自线程 2 的读请求由于缓存的删除导致缓存未命中，根据 Cache-Aside 模式，线程 2 继而查询数据库（ step 2 ），但由于写请求通常慢于读请求，线程 1 更新数据库的操作可能会晚于线程 2 查询数据库后更新缓存的操作（ step 4 晚于 step 3 ），那么这样便会导致最终写入缓存的结果是来自线程 2 中查询到的旧值，而写入数据库的结果是来自线程 1 的新值，即缓存落后于数据库，此时再有读请求命中缓存（ step 5 ），读取到的便是旧值。
      - ![img.png](micro_service_cache_consistent1.png)
      - 另外，先删除缓存，由于缓存中数据缺失，加剧数据库的请求压力，可能会增大缓存穿透出现的概率
    - 如果选择先删除缓存，再更新数据库，那如何解决一致性问题呢？
      - 为了避免“先删除缓存，再更新数据库”这一方案在读写并发时可能带来的缓存脏数据，业界又提出了延时双删的策略，即在更新数据库之后，延迟一段时间再次删除缓存，为了保证第二次删除缓存的时间点在读请求更新缓存之后，这个延迟时间的经验值通常应稍大于业务中读请求的耗时。延迟的实现可以在代码中 sleep 或采用延迟队列。显而易见的是，无论这个值如何预估，都很难和读请求的完成时间点准确衔接，这也是延时双删被诟病的主要原因。
    - 那么 Cache-Aside 存在数据不一致的可能吗？
      - 在 Cache-Aside 中，也存在数据不一致的可能性。在下面的读写并发场景下，首先来自线程 1 的读请求在未命中缓存的情况下查询数据库（ step 1 ），接着来自线程 2 的写请求更新数据库（ step 2 ），但由于一些极端原因，线程 1 中读请求的更新缓存操作晚于线程 2 中写请求的删除缓存的操作（ step 4 晚于 step 3 ），那么这样便会导致最终写入缓存中的是来自线程 1 的旧值，而写入数据库中的是来自线程 2 的新值，即缓存落后于数据库，此时再有读请求命中缓存（ step 5 ），读取到的便是旧值。
      - 这种场景的出现，不仅需要缓存失效且读写并发执行，而且还需要读请求查询数据库的执行早于写请求更新数据库，同时读请求的执行完成晚于写请求。足以见得，这种不一致场景产生的条件非常严格，在实际的生产中出现的可能性较小。
      - ![img.png](micro_service_cache_consistent2.png)
      - 在并发环境下，Cache-Aside 中也存在读请求命中缓存的时间点在写请求更新数据库之后，删除缓存之前，这样也会导致读请求查询到的缓存落后于数据库的情况。
      - ![img.png](micro_service_cache_consistent3.png)
      - 虽然在下一次读请求中，缓存会被更新，但如果业务层面对这种情况的容忍度较低，那么可以采用加锁在写请求中保证“更新数据库&删除缓存”的串行执行为原子性操作（同理也可对读请求中缓存的更新加锁）。加锁势必会导致吞吐量的下降，故采取加锁的方案应该对性能的损耗有所预期。
      - ![img.png](micro_service_cache_consistent4.png)
  - 补偿机制
    - 删除重试机制
      - 由于同步重试删除在性能上会影响吞吐量，所以常通过引入消息队列，将删除失败的缓存对应的 key 放入消息队列中，在对应的消费者中获取删除失败的 key ，异步重试删除。这种方法在实现上相对简单，但由于删除失败后的逻辑需要基于业务代码的 trigger 来触发 ，对业务代码具有一定入侵性。
    - 基于数据库日志（ MySQL binlog ）增量解析、订阅和消费
    - 数据传输服务 DTS
  - Read-Through
    - Read-Through 意为读穿透模式，它的流程和 Cache-Aside 类似，不同点在于 Read-Through 中多了一个访问控制层，读请求只和该访问控制层进行交互，而背后缓存命中与否的逻辑则由访问控制层与数据源进行交互，业务层的实现会更加简洁，并且对于缓存层及持久化层交互的封装程度更高，更易于移植。
  - Write-Through
    - Write-Through 意为直写模式，对于 Write-Through 直写模式来说，它也增加了访问控制层来提供更高程度的封装。不同于 Cache-Aside 的是，Write-Through 直写模式在写请求更新数据库之后，并不会删除缓存，而是更新缓存
    - Write-Through 直写模式适合写操作较多，并且对一致性要求较高的场景，在应用 Write-Through 模式时，也需要通过一定的补偿机制来解决它的问题。首先，在并发环境下，我们前面提到了先更新数据库，再更新缓存会导致缓存和数据库的不一致，那么先更新缓存，再更新数据库呢？这样的操作时序仍然会导致下面这样线程 1 先更新缓存，最后更新数据库的情况，即由于线程 1 和 线程 2 的执行不确定性导致数据库和缓存的不一致。这种由于线程竞争导致的缓存不一致，可以通过分布式锁解决，保证对缓存和数据库的操作仅能由同一个线程完成。对于没有拿到锁的线程，一是通过锁的 timeout 时间进行控制，二是将请求暂存在消息队列中顺序消费。
    - 在下面这种并发执行场景下，来自线程 1 的写请求更新了数据库，接着来自线程 2 的读请求命中缓存，接着线程 1 才更新缓存，这样便会导致线程 2 读取到的缓存落后于数据库。同理，先更新缓存后更新数据库在写请求和读请求并发时，也会出现类似的问题。面对这种场景，我们也可以加锁解决。
    - 在 Write-Through 模式下，不管是先更新缓存还是先更新数据库，都存在更新缓存或者更新数据库失败的情况，上面提到的重试机制和补偿机制在这里也是奏效的。
  - Write-Behind
    - Write behind 意为异步回写模式，它也具有类似 Read-Through/Write-Through 的访问控制层，不同的是，Write behind 在处理写请求时，只更新缓存而不更新数据库，对于数据库的更新，则是通过批量异步更新的方式进行的，批量写入的时间点可以选在数据库负载较低的时间进行。
    - 在 Write-Behind 模式下，写请求延迟较低，减轻了数据库的压力，具有较好的吞吐性。但数据库和缓存的一致性较弱，比如当更新的数据还未被写入数据库时，直接从数据库中查询数据是落后于缓存的。同时，缓存的负载较大，如果缓存宕机会导致数据丢失，所以需要做好缓存的高可用。显然，Write behind 模式下适合大量写操作的场景，常用于电商秒杀场景中库存的扣减。
  - Summary
    - 读多写少的场景下，可以选择采用 Cache-Aside 结合消费数据库日志做补偿”的方案
    - 写多的场景下，可以选择采用 Write-Through 结合分布式锁”的方案
    - 写多的极端场景下，可以选择采用 Write-Behind 的方案
- [大型系统的缓存设计问题](https://mp.weixin.qq.com/s/jFUG6O6vsn_3cWuN-skENg)
  - 缓存的意义
    - 所谓缓存，实际上就是用空间换时间，准确地说是用更高速的空间来换时间，从而整体上提升读的性能。
  - 引入缓存后的一致性挑战
    - 只要使用了缓存，就必然会有不一致的情况出现，只是说这个不一致的时间窗口是否能做到足够的小。有些不合理的设计可能会导致数据持续不一致，这是我们需要改善设计去避免的。
  - 缓存不一致性无法客观地完全消灭
    - 由于数据库和 Redis 之间是没有事务保证的，所以我们无法确保写入数据库成功后，写入 Redis 也是一定成功的；即便 Redis 写入能成功，在数据库写入成功后到 Redis 写入成功前的这段时间里，Redis 数据也肯定是和 MySQL 不一致的。
  - 更新缓存的手段
    - ![img.png](micro_service_cache_design_summar.png)
    - 从一致性的角度来看，采取更新数据库后删除缓存值，是更为适合的策略。因为出现不一致的场景的条件更为苛刻，概率相比其他方案更低
    - 做个简单总结，足以适应绝大部分的互联网开发场景的决策：
      - 针对大部分读多写少场景，建议选择更新数据库后删除缓存的策略。
      - 针对读写相当或者写多读少的场景，建议选择更新数据库后更新缓存的策略。
  - 最终一致性如何保证？
    - 缓存设置过期时间
    - 如何减少缓存删除/更新的失败 - 这时候借助一个可靠的消息中间件就是一个不错的选择。因为消息中间件有 ATLEAST-ONCE 的机制
    - 如何处理复杂的多缓存场景？- 把更新缓存的操作以 MQ 消息的方式发送出去，由不同的系统或者专门的一个系统进行订阅，而做聚合的操作。
    - 通过订阅 MySQL binlog 的方式处理缓存
- [分布式锁](https://mp.weixin.qq.com/s/35aCS_5GqLyzZS3VobL6fg)
  - 分布式锁
    - 属于分布式互斥问题(distributed mutual exclusion)，实际上 Lamport 在那篇经典论文 "Time, clocks, and the ordering of events in a distributed system" 中早就证明了使用状态机能够去中心化解决多进程互斥问题，而共识算法就能实现这样的状态机。但大多时候我们还是会使用一个分布式锁而不是构建一个共识库
  - 分布式锁服务有三个必备的性质
    - 互斥（Mutual Exclusion），这是锁最基本的功能，同一时刻只能有一个客户端持有锁
    - 避免死锁（Dead lock free）. 最常见的是通过设置一个 TTL(Time To Live，存活时间) 来避免死锁
    - 容错（Fault tolerance）. 为避免单点故障，锁服务需要具有一定容错性
      - 一种是锁服务本身是一个集群，能够自动故障切换(ZooKeeper、etcd)
      - 另一种是客户端向多个独立的锁服务发起请求，其中某个锁服务故障时仍然可以从其他锁服务读取到锁信息(Redlock).代价是一个客户端要获取多把锁，并且要求每台机器的时钟都是一样的
  - 分布式锁分为两大类：自旋类和监听类。
    - 自旋类包括基于数据库的实现和基于 Redis 的实现，这类实现需要客户端不停反复请求锁服务查看是否能够获取到锁；
    - 监听类主要包括基于 ZooKeeper 或 etcd 实现的分布式锁，这类实现客户端只需监听(watch) 某个 key，当锁可用时锁服务会通知客户端，无需客户端不停请求锁服务。
  - 基于数据库的实现
    | 字段  | 作用 |
    | ---- | ---- |
    | id	| 自增 id，唯一标识锁 |
    | key	| 锁名称 |
    | value	 | 自定义字段 |
    | ttl	| 存活时间，定时清理，避免死锁 |
  - 基于 Redis 的实现
    - 基于单节点 Redis 的分布式锁 `set key value nx px 10000`
      - [实现可重入锁](https://mp.weixin.qq.com/s/742RZJxeyxhBlqz6eJ5kEg)
        - 当线程拥有锁之后，往后再遇到加锁方法，直接将加锁次数加 1，然后再执行方法逻辑。
        ```lua
        if (redis.call('exists', KEYS[1]) == 0) then
          redis.call('hincrby', KEYS[1], ARGV[2], 1);
          redis.call('pexpire', KEYS[1], ARGV[1]);
          return 1;
        end ;
        if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
          redis.call('hincrby', KEYS[1], ARGV[2], 1);
          redis.call('pexpire', KEYS[1], ARGV[1]);
          return 1;
        end ;
        return 0;
        ```
         ```lua
         -- unlock
         -- 判断 hash set 可重入 key 的值是否等于 0
         -- 如果为 0 代表 该可重入 key 不存在
         if (redis.call('hexists', KEYS[1], ARGV[1]) == 0) then
             return nil;
         end ;
         -- 计算当前可重入次数
         local counter = redis.call('hincrby', KEYS[1], ARGV[1], -1);
         -- 小于等于 0 代表可以解锁
         if (counter > 0) then
             return 0;
         else
             redis.call('del', KEYS[1]);
             return 1;
         end ;
         return nil;
         ```
    - 基于多节点 Redis 的分布式锁 Redlock
  - 基于 ZooKeeper 实现
    - 基于 ZooKeeper 实现的分布式锁依赖以下两个节点属性：
      - sequence：顺序节点，ZooKeeper 会将一个10位带有0填充的序列号附加到客户端设置的 znode 路径之后。例如 locknode/guid-lock- 会返回 locknode/guid-lock-0000000001；
      - ephemeral：临时节点，当客户端和 ZooKeeper 连接断开时，临时节点会被删除，能够避免死锁。但这个断开检测依然有一定心跳延迟，所以仍然需要自增 id 来避免互斥性被破坏。
    - ZooKeeper 官方文档有提供现成的分布式锁实现方法：
      - 首先调用 create()，锁路径例如 locknode/guid-lock-，并设置 sequence 和 ephemeral 标志。guid 是客户端的唯一标识，如果 create() 创建失败可以通过 guid 来进行检查，下面会提到；
      - 调用 getChildren() 获取子节点列表，不要设置 watch 标志（很重要，可以避免 Herd Effect，即惊群效应）；
      - 检查 2 中的子节点列表，如果步骤 1 中创建成功并且返回的序列号后缀是最小的，则客户端持有该分布式锁，到此结束；
      - 如果发现序列不是最小的，则从子节点列表中选择比当前序列号小一位的节点记为 p，客户端调用 exist(p, watch=true)，即监听 p，当 p 被删除时收到通知（该节点只有比自己小一位的节点释放时才能获得锁）；
      - 如果 exist() 返回 null，即前一个分布式锁被释放了，转到步骤 2；否则需要一直等待步骤 4 中 watch 的通知。
  - etcd
    - 我们来看下 etcd 是如何解决分布式锁“三大问题”的：
      - 互斥：etcd 支持事务，通过事务创建 key 和检查 key 是否存在，可以保证互斥性；
      - 容错：etcd 基于 Raft 共识算法，写 key 成功至少需要超过半数节点确认，这就保证了容错性；
      - 死锁：etcd 支持租约(Lease)机制，可以对 key 设置租约存活时间(TTL)，到期后该 key 将失效删除，避免死锁；etc 也支持租约续期，如果客户端还未处理完可以继续续约；同时 etcd 也有自增 id，在下文介绍。
- [Go Micro, Go Kit, Gizmo, Kite 对比](https://mp.weixin.qq.com/s/KQ_qWBrW9MY0a3IL_Bv2kw)
- [微服务架构设计模式](https://mp.weixin.qq.com/s/NpQIgh2VWVXqA6ab_Bo3TA)
- [API Design Google](https://www.bookstack.cn/read/API-design-guide/API-design-guide-01-%E7%AE%80%E4%BB%8B.md)
- [API Improvement Proposal Google](https://google.aip.dev/1)
- [API Desgin Guild](https://cloud.google.com/apis/design)
- [Implement a timing wheel for millions of concurrent tasks](https://faun.pub/implement-a-timing-wheel-for-millions-of-concurrent-tasks-942c8d9e4abe)
  - timing wheel
    - reduces the maintenance of tasks from priority queue O(nlog(n)) to bidirectional linked table O(1), 
    - and the execution of tasks also requires only polling for tasks at one point in time O(N), without putting in and removing elements O(nlog(n)), as in the case of the priority queue.
- [Distributed Lock](https://www.youtube.com/watch?v=VnbC5RG1fEo)
  - Concurrent transaction need be synchronized
    - DB lock is fine, but the Optimistic Lock is great
      - Shared Lock - `SELECT ... LOCK IN SHARE MODE`
      - Exclusive Lock  - `SELECT ... FOR UPDATE`
      - Version Optimistic Lock
    - Sharding the data cannot solve the all of problem
      - Using the shard data solution
      - Consistency can be easy to be guaranteed 
        - Consistent Hash
        - Gossip the indices
      - The issue of Shard
        - Hotspot - Load is not balanced
        - Multiple Entities Transaction
        - Node Failed - Data Replication cause more consistent issue - CAP
  - Redlock
    - the algorithm makes dangerous assumption about timing and system clock
    - Lock expired and deadlock issue
    - Is readlock safe
      - Fencing is great, with this no need distributed lock
      - Clock jump is one big issue, readlock cannnot work correctly under this problem
  - Zookeeper lock
    - Create lock with Sequence and Ephemeral flag
  - Google Chubby
  - Distributed Lock Service need the following features
    - High Availability
      - Data Replicas - strong consistent protocol - Pasox, Raft, zab
      - Master Failover - Leader election
    - Deadlock Detection
      - Keepalive & Lease Timeout
- [Circuit Breakers](https://sirupsen.com/napkin/problem-11-circuit-breakers)
  - There are at least five configuration parameters relevant for circuit breakers:
    - error_threshold. The amount of errors to encounter for the worker before opening the circuit, that is, to start rejecting requests instantly. In our example, it’s been hard-coded to 3.
    - error_timeout. The amount of time in seconds until trying to query the resource again. That’s the time the circuit is open. 15 seconds in our example.
    - success_threshold. The amount of successes on the circuit until closing it again, that is to start accepting all requests to the circuit. In our example above, this is just hard-coded to 1. This requires a bit more logic to have a number > 1, which better implementations like Semian will take care of.
    - resource_timeout. The timeout to the resource/data-store protected by the circuit breaker. 5 seconds in our example.
    - half_open_resource_timeout. Timeout for the resource in seconds when the circuit is checking whether the resource might be back to normal, after the error_timeout. This state is called half_open. Most circuit breaker implementations (including our simple one above) assume that this is the same as the ‘normal’ timeout for the resource. The bet Semian makes is that during steady-state we can tolerate a higher resource timeout, but during failure, we want it to be lower.
  - [Your Circuit Breaker is Misconfigured](https://shopify.engineering/circuit-breaker-misconfigured)
  - Hystrix-Go 熔断算法
    - 当请求失败比率达到一定阈值之后，熔断器开启，并休眠一段时间（由配置决定），这段休眠期过后，熔断器将处于半开状态，在此状态下将试探性的放过一部分流量，如果这部分流量调用成功后，再次将熔断器关闭，否则熔断器继续保持开启并进入下一轮休眠周期。
    - 但这个熔断算法有一个问题，过于一刀切。是否可以做到在熔断器开启状态下（但是后端未 Shutdown）仍然可以放行少部分流量呢?
  - 自适应限流 Google
    - 客户端自行限制请求速度，限制生成请求的数量, 超过这个数量的请求直接在本地回复失败，而不会真是发送到服务端
    - 每个客户端记录过去两分钟内的以下信息（一般代码中以滑动窗口实现）：
      - requests（客户端请求总量）： The number of requests attempted by the application layer(at the client, on top of the adaptive throttling system)
      - accepts（成功的请求总量 - 被 accepted的量）： The number of requests accepted by the backend
    - 该算法的通用描述如下：
      - 在通常情况下（无错误发生时） requests==accepts
      - 当后端出现异常情况时，accepts 的数量会逐渐小于 requests
      - 当后端持续异常时，客户端可以继续发送请求直到 requests=K∗accepts，一旦超过这个值，客户端就启动自适应限流机制，新产生的请求在本地会以 p 概率（下面描述的 Client request rejection probability 定义）被拒绝
      - 当客户端主动丢弃请求时，requests 值会一直增大，在某个时间点会超过 K∗accepts，使概率 p 计算出来的值大于 0，此时客户端会以此概率对请求做主动丢弃
      - 当后端逐渐恢复时，accepts 增加，（同时 requests 值也会增加，但是由于 K 的关系，K×accepts 的放大倍数更快），使得 (requests−K×accepts)/(requests+1) 变为负数，从而概率 p==0，客户端自适应限流结束
    - K 值
      - 降低 K 值会使自适应限流算法更加激进（允许客户端在算法启动时拒绝更多本地请求）
      - 增加 K 值会使自适应限流算法不再那么激进（允许服务端在算法启动时尝试接收更多的请求，与上面相反）
- [设计好接口的36个锦囊](https://mp.weixin.qq.com/s/cijwF_9eqnKuVnvRihhB_w)
  - 接口参数校验
  - 修改老接口时，注意接口的兼容性
  - 设计接口时，充分考虑接口的可扩展性
  - 接口考虑是否需要防重处理
  - 重点接口，考虑线程池隔离, 热点数据隔离性
  - 调用第三方接口要考虑异常和超时处理 - 异常处理, 接口超时, 重试次数
  - 接口实现考虑熔断和降级,限流
  - 日志打印好，接口的关键代码，要有日志保驾护航
  - 接口有些场景，使用异步更合理
  - 优化接口耗时，远程串行考虑改并行调用
  - 接口合并或者说考虑批量处理思想
  - 接口实现过程中，恰当使用缓存
  - 接口考虑[幂等性](https://mp.weixin.qq.com/s?__biz=Mzg3NzU5NTIwNg==&mid=2247497427&idx=1&sn=2ed160c9917ad989eee1ac60d6122855&chksm=cf2229faf855a0ecf5eb34c7335acdf6420426490ee99fc2b602d54ff4ffcecfdab24eeab0a3&token=1260947715&lang=zh_CN&scene=21#wechat_redirect)
    - 设计
      - 全局的唯一性ID - snowflake
      - select+insert+主键/唯一索引冲突 
      - 直接insert + 主键/唯一索引冲突
      - 状态机幂等 - 很多业务表，都是有状态的，比如转账流水表，就会有0-待处理，1-处理中、2-成功、3-失败状态
      - 抽取防重表 - 
      - token
      - 悲观锁(如select for update)
      - 乐观锁
      - 分布式锁
    - [方案](https://mp.weixin.qq.com/s/BZhrSJKbItKZ7zzK3MpSDA)
      - 对于下单等存在唯一主键的，可以使用“唯一主键方案”的方式实现。
      - 对于更新订单状态等相关的更新场景操作，使用“乐观锁方案”实现更为简单。
      - 对于上下游这种，下游请求上游，上游服务可以使用“下游传递唯一序列号方案”更为合理。
      - 类似于前端重复提交、重复下单、没有唯一ID号的场景，可以通过 Token 与 Redis 配合的“防重 Token 方案”实现更为快捷。
  - 读写分离，优先考虑读从库，注意主从延迟问题
  - 接口注意返回的数据量，如果数据量大需要分页
  - 代码锁的粒度控制好
  - 接口状态和错误需要统一明确
  - 接口要考虑异常处理
  - 接口实现过程中，注意大文件、大事务、
  - 保证接口安全性
  - [分布式事务，如何保证](https://mp.weixin.qq.com/s?__biz=Mzg3NzU5NTIwNg==&mid=2247498358&idx=1&sn=aa6c7ceb61b73267d68d1b4fb7ccc2ed&scene=21#wechat_redirect)
  - 考虑线性安全问题
- [七种分布式事务的解决方案](https://mp.weixin.qq.com/s?__biz=Mzg4NTcyNjczNg==&mid=2247504235&idx=1&sn=6ec72cbf907c46d693db10b0835b7528&source=41#wechat_redirect)
  - 分布式事务是指事务的参与者、支持事务的服务器、资源服务器以及事务管理器「分别位于不同的分布式系统的不同节点之上
  - 分布式理论
    - CAP定理
    - BASE理论
      - BASE理论，它是用来对CAP定理进行进一步扩充的。BASE理论指的是：
      -「Basically Available（基本可用）」
      -「Soft state（软状态）」
      -「Eventually consistent（最终一致性）」
  - 分布式事务解决方案
    - 2PC
      - mysql的事务就是通过「日志系统」来完成两阶段提交的。
      - 第一个阶段是「投票阶段」
      - 第二个阶段是「决定阶段」
      - Issue
        - 单点故障」：一旦事务管理器出现故障，整个系统不可用
        - 「数据不一致」：在阶段二，如果事务管理器只发送了部分 commit 消息，此时网络发生异常，那么只有部分参与者接收到 commit 消息，也就是说只有部分参与者提交了事务，使得系统数据不一致。
        - 「响应时间较长」：整个消息链路是串行的，要等待响应结果，不适合高并发的场景
        - 「不确定性」：当事务管理器发送 commit 之后，并且此时只有一个参与者收到了 commit，那么当该参与者与事务管理器同时宕机之后，重新选举的事务管理器无法确定该条消息是否提交成功。
    - 3PC
      - 第一阶段：「CanCommit阶段」这个阶段所做的事很简单，就是协调者询问事务参与者，你是否有能力完成此次事务
      - 第二阶段：「PreCommit阶段」此时协调者会向所有的参与者发送PreCommit请求，参与者收到后开始执行事务操作，并将Undo和Redo信息记录到事务日志中。参与者执行完事务操作后（此时属于未提交事务的状态），就会向协调者反馈“Ack”表示我已经准备好提交了，并等待协调者的下一步指令。
      - 第三阶段：「DoCommit阶段」在阶段二中如果所有的参与者节点都可以进行PreCommit提交，那么协调者就会从“预提交状态”转变为“提交状态”。然后向所有的参与者节点发送"doCommit"请求，参与者节点在收到提交请求后就会各自执行事务提交操作，并向协调者节点反馈“Ack”消息，协调者收到所有参与者的Ack消息后完成事务。相反，如果有一个参与者节点未完成PreCommit的反馈或者反馈超时，那么协调者都会向所有的参与者节点发送abort请求，从而中断事务
      - 相对于2PC来说增加了CanCommit阶段和超时机制。如果段时间内没有收到协调者的commit请求，那么就会自动进行commit，解决了2PC单点故障的问题
      - 但是性能问题和不一致问题仍然没有根本解决
    - 补偿事务TCC
      - 其核心思想是：「针对每个操作，都要注册一个与其对应的确认和补偿（撤销）操作」。它分为三个阶段：「Try,Confirm,Cancel」
      - Try阶段主要是对「业务系统做检测及资源预留」，其主要分为两个阶段
      - Confirm 阶段主要是对「业务系统做确认提交」，Try阶段执行成功并开始执行 Confirm阶段时，默认 Confirm阶段是不会出错的。即：只要Try成功，Confirm一定成功。
      - Cancel 阶段主要是在业务执行错误，需要回滚的状态下执行的业务取消，「预留资源释放」。
      - TCC 事务机制相比于上面介绍的2PC，解决了其几个缺点：
        - 1.「解决了协调者单点」，由主业务方发起并完成这个业务活动。业务活动管理器也变成多点，引入集群。
        - 2.「同步阻塞」：引入超时，超时后进行补偿，并且不会锁定整个资源，将资源转换为业务逻辑形式，粒度变小。
        - 3.「数据一致性」，有了补偿机制之后，由业务活动管理器控制一致性
    - 本地消息表
      - 消息生产方，需要额外建一个消息表，并「记录消息发送状态」。消息表和业务数据要在一个事务里提交，也就是说他们要在一个数据库里面。然后消息会经过MQ发送到消息的消费方。
      - 消息消费方，需要「处理」这个「消息」，并完成自己的业务逻辑。
      - 生产方和消费方定时扫描本地消息表，把还没处理完成的消息或者失败的消息再发送一遍。
    - 消息事务
      - 消息事务的原理是将两个事务「通过消息中间件进行异步解耦」，和上述的本地消息表有点类似，但是是通过消息中间件的机制去做的，其本质就是'将本地消息表封装到了消息中间件中'。
      - RocketMQ
    - 最大努力通知
      - 系统 A 本地事务执行完之后，发送个消息到 MQ；
      - 这里会有个专门消费 MQ 的服务，这个服务会消费 MQ 并调用系统 B 的接口；
    - Sagas 事务模型
- [延时消息的方案](https://mp.weixin.qq.com/s/jurYbCZ5NJMt4RYYtaSnHw)
  - 基于外部存储实现的方案
    - 将 MQ 和 延时模块 区分开来，延时消息模块是一个独立的服务/进程。延时消息先保留到其他存储介质中，然后在消息到期时再投递到 MQ。
  - 基于 数据库（如MySQL）
    - 通过定时线程定时扫描到期的消息，然后进行投递。定时线程的扫描间隔理论上就是你延时消息的最小时间精度。
    - B+Tree索引不适合消息场景的大量写入
  - 基于 RocksDB
    - RocksDB 使用的是LSM Tree，LSM 树更适合大量写入的场景
  - 基于Redis
    - Messages Pool 所有的延时消息存放，结构为KV结构，key为消息ID，value为一个具体的message（这里选择Redis Hash结构主要是因为hash结构能存储较大的数据量，数据较多时候会进行渐进式rehash扩容，并且对于HSET和HGET命令来说时间复杂度都是O(1)）
    - Delayed Queue是16个有序队列（队列支持水平扩展），结构为ZSET，value 为 messages pool中消息ID，score为过期时间（分为多个队列是为了提高扫描的速度）
    - Worker 代表处理线程，通过定时任务扫描 Delayed Queue 中到期的消息
- [Nginx + keepalived 实现高可用](https://mp.weixin.qq.com/s/DBdmwRMgrj0ShvZi-dTkVA)
  - Nginx rewrite 规则
    - Rewrite规则含义就是某个URL重写成特定的URL（类似于Redirect），从某种意义上说为了美观或者对搜索引擎友好，提高收录量及排名等
      ```shell
      rewrite    <regex>    <replacement>                 [flag];
        关键字      正则        替代内容                    flag标记
      ```
  - Nginx 防盗链
    ```shell
    location ~* \.(rmvb|jpg|png|swf|flv)$ { #rmvb|jpg|png|swf|flv表示对rmvb|jpg|png|swf|flv后缀的文件实行防盗链
            valid_referers none blocked  www.dbspread.com; #表示对www.dbspread.com此域名开通白名单，比如在www.test.com的index.html引用download/av123.rmvb,无效
            root   html/b;
            if ($invalid_referer) { #如果请求不是从www.dbspread.com白名单发出来的请求，直接重定向到403.html这个页面或者返回403 
                 #rewrite ^/ http://www.dbspread.com/403.html;
                 return 403;
            }
    }
    ```
  - 动静分离
    ```shell
            location ~ .*\.(php|jsp|cgi|shtml)?$ #动态分离 ~匹配 以.*结尾（以PHP JSP结尾走这段）
             {
                proxy_set_header Host  $host;
                   proxy_set_header X-Real-IP $remote_addr;
                   proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                   proxy_pass http://jvm_web2;
            }
    
            #静态分离 ~匹配 以.*结尾（以html|htm|gif|jpg|jpeg|bmp|png|ico|txt|js|css结尾走这段），当然不是越久越好，如果有10000个用户在线，都保存几个月，系统托跨
            location ~ .*\.(html|htm|gif|jpg|jpeg|bmp|png|ico|txt|js|css)$ 
            {
                root /var/local/static; #静态资源存放在nginx的安装机器上
                #proxy_pass http://www.static.com; #静态资源也可存放在远程服务器上
                expires    30d; #30天之内只要访问过一次就从缓存拿
            }
    ```
  - Nginx+keepalived 高可用
    - keepalived
      - Keepalived软件起初是专为LVS负载均衡软件设计的，用来管理并监控LVS集群系统中各个服务节点的状态，后来又加入了可以实现高可用的VRRP (Virtual Router Redundancy Protocol ,虚拟路由器冗余协议）功能
      - 管理LVS负载均衡软件实现LVS集群节点的健康检查作为系统网络服务的高可用性（failover）
      - Keepalived高可用服务之间的故障切换转移，是通过 VRRP 来实现的。在 Keepalived服务正常工作时，主 Master节点会不断地向备节点发送（多播的方式）心跳消息，用以告诉备Backup节点自己还活着，当主 Master节点发生故障时，就无法发送心跳消息，备节点也就因此无法继续检测到来自主 Master节点的心跳了，于是调用自身的接管程序，接管主Master节点的 IP资源及服务。而当主 Master节点恢复时，备Backup节点又会释放主节点故障时自身接管的IP资源及服务，恢复到原来的备用角色。
    
- [分布式系统设计模式](https://colobu.com/2022/06/26/distributed-system-design-patterns/)
  - 布隆过滤器
    - Bloom 过滤器是一种节省空间的概率数据结构，用于测试元素是否为某集合的成员。它用于我们只需要检查元素是否属于对象的场景。
    - 在BigTable（和Cassandra）中，任何读取操作都必须从组成Tablet的SSTable中读取。如果这些 SSTable 不在内存中，则读取操作可能最终会执行许多磁盘访问以便读取所需的SSTable
  - 一致性哈希
    - 一致的哈希允许您轻松扩展，从而允许以有效的方式复制数据，从而实现更好的可用性和容错能力。
    - 一致散列的主要优点是增量稳定性;节点离开或到达集群仅影响其直接邻居，其他节点不受影响。
  - Quorum
    - 在分布式环境中，quorum是在确认操作成功之前需要成功执行此分布式操作的最小服务器数。
    - Cassandra，为了确保数据一致性，每个写入请求都可以配置为仅当数据已写入至少一个quorum（或大多数）副本节点时才成功。
    - Dynamo 将写入复制到系统中其他节点的草率quorum，而不是像 Paxos 那样的严格多数quorum。所有读/写操作都在首选项列表中的第一个 NN 正常节点上执行，该节点可能并不总是在遍历一致哈希环时遇到的第一个 NN 节点。
  - 领导者(Leader)和追随者(Follower)
  - 心跳
  - Fencing
    - 在领导者-追随者模式中，当领导者失败时，不可能确定领导者已停止工作。例如，慢速网络或网络分区可能会触发新的领导者选举，即使前一个领导者仍在运行并认为它仍然是活动的领导者。
    - 屏蔽是指在以前处于活动状态的领导者周围设置围栏，使其无法访问集群资源，从而停止为任何读/写请求提供服务。
      - 资源屏蔽：系统会阻止以前处于活动状态的领导者访问执行基本任务所需的资源。
      - 节点屏蔽：系统会阻止以前处于活动状态的领导者访问所有资源。执行此操作的常见方法是关闭节点电源或重置节点。
  - WAL (预写日志 Write-ahead Log)
    - 预写日志记录是解决操作系统中文件系统不一致的问题的高级解决方案。受数据库管理系统的启发，此方法首先将要执行的操作的摘要记入“日志”中，然后再将其实际写入磁盘。在发生崩溃的情况下，操作系统只需检查此日志并从中断的位置继续。
  - 分段日志
    - 单个日志文件在启动时读取时可能会增长并成为性能瓶颈。较旧的日志会定期清理，并且很难对单个大文件执行清理操作。
    - 单个日志拆分为多个段。日志文件在指定的大小限制后滚动。使用日志分段，需要有一种将逻辑日志偏移量（或日志序列号）映射到日志段文件的简单方法。
  - 高水位线 (High-Water mark)
    - 跟踪领导者上的最后一个日志条目，该条目已成功复制到追随者的quorum。日志中此条目的索引称为高水位线索引。领导者仅公开到高水位线索引的数据。
    - Kafka：为了处理非可重复读取并确保数据一致性，Kafka broker会跟踪高水位线，这是特定分区的最大偏移量。使用者只能看到高水位线之前的消息。
  - 租约(Lease)
    - 租约就像一个锁，但即使客户端离开，它也能工作。客户端请求有限期限的租约，之后租约到期。如果客户端想要延长租约，它可以在租约到期之前续订租约。
    - Chubby 客户端与领导者保持有时限的会话租约。在此时间间隔内，领导者保证不会单方面终止会话。
  - Gossip 协议
    - Gossip协议是点对点通信机制，其中节点定期交换有关自己和他们所知道的其他节点的状态信息。
    - 每个节点每秒启动一轮Gossip回合，以与另一个随机节点交换有关自己和其他节点的状态信息。
  - Phi 累计故障检测（Phi Accrual Failure Detection）
    - 此算法使用历史检测信号信息使阈值自适应。通用的应计故障检测器不会判断服务器是否处于活动状态，而是输出有关服务器的可疑级别。
    - Cassandra 使用 Phi 应计故障检测器算法来确定群集中节点的状态。
  - 脑裂
    - 分布式系统具有两个或多个活动领导者的场景称为脑裂。
      - 通过使用生成时钟(Generation Clock)可以解决脑裂问题，生成时钟只是一个单调递增的数字，用于指示服务器的生成。
      - 每次选出新领导者时，时钟数字(generation number)都会增加。这意味着，如果旧领导者的时钟数为“1”，则新领导人的时钟数将为“2”。此时钟号包含在从领导发送到其他节点的每个请求中。通过这种方式，节点现在可以通过简单地信任具有最高数字的领导者来轻松区分真正的领导者。
    - Kafka：为了处理脑裂（我们可以有多个active controller broker），Kafka使用“纪元数”（Epoch number），这只是一个单调增加的数字来表示服务器的代次(generation)。
    - HDFS：ZooKeeper用于确保任何时候只有一个NameNode处于活动状态。epoch 编号作为每个事务 ID 的一部分进行维护，以反映 NameNode 的代次。
  - 校验和 (checksum)
    - 要计算校验和，请使用 MD5、SHA-1、SHA-256 或 SHA-512 等加密哈希函数。哈希函数获取输入数据并生成固定长度的字符串（包含字母和数字）;此字符串称为校验和。
  - CAP定理
    - Dynamo：在CAP定理术语中，Dynamo属于AP系统的类别，旨在牺牲强一致性为代价实现高可用性。
    - BigTable：就CAP定理而言，BigTable是一个CP系统，即它具有严格一致的读取和写入。
  - PACELEC 定理
    - PACELC定理指出，在复制数据的系统中：
      - 如果有一个分区（'P'），分布式系统可以在可用性和一致性（即'A'和'C'）之间进行权衡;
      - 否则（'E'），当系统在没有分区的情况下正常运行时，系统可以在延迟（'L'）和一致性（'C'）之间进行权衡。
    - 定理（PAC）的第一部分与CAP定理相同，ELC是扩展。整个论点假设我们通过复制来保持高可用性。因此，当失败时，CAP定理占上风。但如果没有，我们仍然必须考虑复制系统的一致性和延迟之间的权衡。
  - 提示交接 (Hinted Handoff)
    - 如果节点关闭，系统会保留它们错过的所有请求的提示（或注释）。故障节点恢复后，将根据存储的提示将请求转发给它们。
    - 当节点关闭时，领导者会在本地磁盘上的文本文件中写入提示。此提示包含数据及其所属的节点信息。当领导者意识到它为其保留提示的节点已恢复时，它会将每个提示的写入请求转发到该节点。
  - 读取时修复
    - Cassandra和Dynamo使用“读取修复”将最新版本的数据推送到具有旧版本的节点。
  - 默克尔树 (Merkle Trees)
    - “读取修复”可在处理读取请求时消除冲突。但是，如果某个副本明显落后于其他副本，则可能需要很长时间才能解决冲突。
    - 副本可以包含大量数据。单纯地拆分整个范围来计算校验和进行比较并不是很可行;有太多的数据需要传输。相反，我们可以使用Merkle树来比较一个范围的副本。
      - Merkle 树是哈希的二叉树，其中每个内部节点是其两个子节点的哈希，每个叶节点是原始数据一部分的哈希。
      - 比较Merkle树在概念上很简单：
        - 比较两个树的根哈希。
        - 如果它们相等，请停止。
        - 在左边和右边的孩子上递归检查。
    - 为了实现反熵和在后台解决冲突，Dynamo使用Merkle树。
- [GToken替换JWT实现SSO单点登录](https://mp.weixin.qq.com/s/K_k2BkCL1GBeAo5xTVrl8g)
  - jwt的问题
    - jwt无法在服务端主动退出的问题
    - jwt无法作废已颁布的令牌，只能等到令牌过期问题
    - jwt携带大量用户扩展信息导致降低传输效率问题
    - Security Issue
      - The `none` algorithm
      - The algorithms are passed in an array 
        - The solution is to use just one algorithm, or use two different methods with just one algorithm each, and call two methods independently.
      - Claims are optional
        - JWT provides a nice way to organize and ensure different claims that could help improve security, but they are all optional. For example, in the sample code to the right, sub, iss, aud, and so on are all optional.
    - Dangerous for user authentication
      - Logout doesn’t really log you out.
      - Blocking users doesn’t immediately block them
      - JWTs could contain stale data.
      - JWT’s are often not encrypted.
  - gtoken的优势
    - gtoken支持单点应用使用内存存储，支持个人项目文件存储，也支持企业集群使用redis存储，完全适用于个人和企业生产级使用；
    - 有效避免了jwt服务端无法退出问题；
    - 解决jwt无法作废已颁布的令牌，只能等到令牌过期问题；
    - 通过用户扩展信息存储在服务端，有效规避了jwt携带大量用户扩展信息导致降低传输效率问题；
    - 有效避免jwt需要客户端实现续签功能，增加客户端复杂度；支持服务端自动续期，客户端不需要关心续签逻辑；
- [Kitex Proxyless 配合 Istio 与 OpenTelemetry 实现全链路泳道](https://mp.weixin.qq.com/s/SAn-H5p53IfvSy_Y3Mcz_Q)
- [评论系统架构设计](https://mp.weixin.qq.com/s/s99ZWuSi0Lu6ivfpBHvLWQ)
- [How Not To Sort By Average Rating](https://www.evanmiller.org/how-not-to-sort-by-average-rating.html)
  - Score = Lower bound of Wilson score confidence interval for a Bernoulli parameter
- [Clean Architecture]
  - ![img.png](micro_servicee_clean_arch.png)
    ```shell
    ├── adapter // Adapter层，适配各种框架及协议的接入，比如：Gin，tRPC，Echo，Fiber 等
    ├── application // App层，处理Adapter层适配过后与框架、协议等无关的业务逻辑
    │   ├── consumer //（可选）处理外部消息，比如来自消息队列的事件消费
    │   ├── dto // App层的数据传输对象，外层到达App层的数据，从App层出发到外层的数据都通过DTO传播
    │   ├── executor // 处理请求，包括command和query
    │   └── scheduler //（可选）处理定时任务，比如Cron格式的定时Job
    ├── domain // Domain层，最核心最纯粹的业务实体及其规则的抽象定义
    │   ├── gateway // 领域网关，model的核心逻辑以Interface形式在此定义，交由Infra层去实现
    │   └── model // 领域模型实体
    ├── infrastructure // Infra层，各种外部依赖，组件的衔接，以及domain/gateway的具体实现
    │   ├── cache //（可选）内层所需缓存的实现，可以是Redis，Memcached等
    │   ├── client //（可选）各种中间件client的初始化
    │   ├── config // 配置实现
    │   ├── database //（可选）内层所需持久化的实现，可以是MySQL，MongoDB，Neo4j等
    │   ├── distlock //（可选）内层所需分布式锁的实现，可以基于Redis，ZooKeeper，etcd等
    │   ├── log // 日志实现，在此接入第三方日志库，避免对内层的污染
    │   ├── mq //（可选）内层所需消息队列的实现，可以是Kafka，RabbitMQ，Pulsar等
    │   ├── node //（可选）服务节点一致性协调控制实现，可以基于ZooKeeper，etcd等
    │   └── rpc //（可选）广义上第三方服务的访问实现，可以通过HTTP，gRPC，tRPC等
    └── pkg // 各层可共享的公共组件代码
    ```
  - ![img.png](micro_service_proj_layout_process_sample.png)
  - 外部请求首先抵达 Adapter 层。如果是读请求，则携带简单参数调用 App 层；如果是写请求，则携带 DTO 调用 App 层。App 层将收到的 DTO 转化成对应的 Model，调用 Domain 层 gateway 相关业务逻辑接口方法。由于系统初始化阶段已经完成依赖注入，接口对应的来自 Infra 层的具体实现会处理完成并返回 Model 到 Domain 层，再由 Domain 层返回到 App 层，最终经由 Adapter 层将响应内容呈现给外部
- [Shmipc](https://github.com/cloudwego/shmipc-spec)
  - Shmipc is a shared memory IPC library for Go. It provides a simple and efficient way to communicate between processes.
  - 高性能进程间通讯库，它基于共享内存构建，具有零拷贝的特点，同时它引入的同步机制具有批量收割 IO 的能力，相对于其他进程间通讯方式能明显提升性能
  - Shmipc 应用于 Service Mesh 场景下，mesh proxy 进程与业务逻辑进程、与通用 sidecar 进程的通讯， 在大包场景和 IO 密集型场景能够取得显著的性能收益。
  - 设计思路
    - 零拷贝
      - 通过共享内存的方式，避免了进程间通讯的数据拷贝，提升了性能
      - 在 RPC场景下，一次 RPC 流程中在进程间通讯上会有四次的内存拷贝，Request 路径两次， Response 路径两次
      - 基于共享内存通讯零拷贝的特性，我们可以很容易达成这一点。但为了达到零拷贝的效果，围绕共享内存本身，还会产生有许多额外的工作
    - 批量收割IO
      - 在共享内存中构造了一个 IO 队列的来完成批量收割 IO，使其在小包 IO 密集场景也能显现收益。核心思想是：当一个进程把请求写入 IO队列后，会给另外一个进程发通知来处理。
    - 无锁化
      - 通过无锁化的设计，避免了进程间通讯的锁竞争，提升了性能
      - 无锁化的设计，是基于共享内存的特性，通过原子操作来实现的。在共享内存中，我们可以通过原子操作来实现无锁化的设计，这样就可以避免进程间通讯的锁竞争，提升性能。
  - 采坑记录
    - 共享内存泄漏 - IPC 过程共享内存分配和回收涉及到两个进程，稍有不慎就容易发生共享内存的泄漏。问题虽然非常棘手，但只要能够做到泄漏时主动发现，以及泄漏之后有观测手段可以排查即可。
    - 串包 - 排查过程和原因并不具备共性，可以提供更多的参考是增加更多场景的集成测试和单元测试将串包扼杀在摇篮中
    - 共享内存踩踏 - 应该尽可能使用 memfd 来共享内存，而不是 mmap 文件系统中的某个路径。
      - 早期我们通过 mmap 文件系统的路径来共享内存，Shmipc 的开启和共享内存的路径由环境变量指定，启动过程由引导进程注入应用进程。那么存在一种情况是应用进程可能会 fork 出一个进程，该进程继承了应用进程的环境变量并且也集成了 Shmipc，然后 fork 的进程和应用进程 mmap 了同一块共享内存，发现踩踏
    - Sigbus coredump - 使用 memfd 来共享内存
      - 早期我们通过 mmap /dev/shm/路径（tmpfs）下的文件来共享内存，应用服务大都运行在 docker 容器实例中。容器实例对 tmpfs 有容量限制（可以通过 df -h 观测），这会使得 mmap 的共享内存如果超过该限制就会出现 Sigbus
- [Overload Control for Scaling Wechat Microservices](https://www.cs.columbia.edu/~ruigu/papers/socc18-final100.pdf)
  - 如何判断过载
    - 微信使用在请求在队列中的平均等待时间作为判断标准，就是从请求到达，到开始处理的时间. 采用平均等待时间还有一个好处，是这个是独立于服务的，可以应用于任何场景，而不用关联于业务，可以直接在框架上进行改造。
    - 通常判断过载可以使用吞吐量，延迟，CPU 使用率，丢包率，待处理请求数，请求处理事件等等
    - 为啥不使用响应时间？因为响应时间是跟服务相关的，很多微服务是链式调用，响应时间是不可控的，也是无法标准化的，很难作为一个统一的判断依据。
    - 那为什么不使用 CPU 负载作为判断标准呢， 因为 CPU 负载高不代表服务过载，因为一个服务请求处理及时，CPU 处于高位反而是比较良好的表现。实际上 CPU 负载高，监控服务是会告警出来，但是并不会直接进入过载处理流程。
    - 腾讯微服务默认的超时时间是 500ms, 
    - 当平均等待时间大于 20ms 时，以一定的降速因子过滤调部分请求，如果判断平均等待时间小于 20ms，则以一定的速率提升通过率，一般采用快降慢升的策略，防止大的服务波动，整个策略相当于一个负反馈电路
  - 过载保护策略
    - 业务优先级
    - 用户优先级
    - 自适应优先级调整
  - 流程
    - 当用户从微信发起请求，请求被路由到接入层服务，分配统一的业务和用户优先级，所有到下游的字请求都继承相同的优先级。
    - 根据业务逻辑调用 1 个或多个下游服务，当服务收到请求，首先根据自身服务准入优先级判断请求是接受还是丢弃。服务本身根据负载情况周期性的调整准入优先级。
    - 当服务需要再向下游发起请求时，判断本地记录的下游服务准入优先级，如果小于则丢弃，如果没有记录或优先级大于记录则向下游发起请求。
    - 下游服务返回上游服务需要的信息，并且在信息中携带自身准入优先级。
    - 上游接受到返回后解析信息，并更新本地记录的下游服务准入优先级
- [gin]()
  - ![img.png](micro_service_gin.png)
  - gin 将 Engine 作为 http.Handler 的实现类进行注入，从而融入 Golang net/http 标准库的框架之内
  - gin 中基于 handler 链的方式实现中间件和处理函数的协调使用
  - gin 中基于压缩前缀树的方式作为路由树的数据结构，对应于 9 种 http 方法共有 9 棵树
    - 为什么使用压缩前缀树
      - 与压缩前缀树相对的就是使用 hashmap，以 path 为 key，handlers 为 value 进行映射关联，这里选择了前者的原因在于：
      - path 匹配时不是完全精确匹配，比如末尾 ‘/’ 符号的增减、全匹配符号 '*' 的处理等，map 无法胜任（模糊匹配部分的代码于本文中并未体现，大家可以深入源码中加以佐证）
      - 路由的数量相对有限，对应数量级下 map 的性能优势体现不明显，在小数据量的前提下，map 性能甚至要弱于前缀树
      - path 串通常存在基于分组分类的公共前缀，适合使用前缀树进行管理，可以节省存储空间
  - gin 中基于 gin.Context 作为一次 http 请求贯穿整条 handler chain 的核心数据结构
  - gin.Context 是一种会被频繁创建销毁的资源对象，因此使用对象池 sync.Pool 进行缓存复用
- [Kratos transport sample](https://github.com/tx7do/kratos-transport/blob/main/broker/kafka/subscriber.go)
- [Serverless提升资源利用率](https://bytedance.feishu.cn/docx/BY2qdcLTKotr3Sxprujc0vqSnGe)
- [后端架构的演进](https://mp.weixin.qq.com/s/n3VIfMrP4zb_CKJAu7yXMw)
  - MVC
  - DDD
  - 六边形架构（Ports and Adapters
  - 洋葱架构



