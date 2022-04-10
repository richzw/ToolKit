
- do not overuse `fmt.Sprintf` in your hot path. It is costly due to maintaining the buffer pool and dynamic dispatches for interfaces.
    - if you are doing `fmt.Sprintf("%s%s", var1, var2)`, consider simple string concatenation.
    - if you are doing `fmt.Sprintf("%x", var)`, consider using `hex.EncodeToString` or `strconv.FormatInt(var, 16)`
- 如果需要把数字转换成字符串，使用 strconv.Itoa() 比 fmt.Sprintf() 要快一倍左右。
- 使用StringBuffer 或是StringBuild 来拼接字符串，性能会比使用 + 或 += 高出三到四个数量级。
- String to []byte
  ```go
  func cstring(s string) []byte {
    b := make([]byte, len(s)+1)
    copy(b, s)
    return b
  }
  ```
- sync.Pool
  - 临时对象池应该是对可读性影响最小且优化效果显著的手段
  - 还有一种利用sync.Pool特性，来减少锁竞争的优化手段，也非常巧妙。另外，在优化前要善用go逃逸检查分析对象是否逃逸到堆上，防止负优化
- goroutine pool
  - 可以限制goroutine数量，避免无限制的增长。
  - 减少栈扩容的次数。
  - 频繁创建goroutine的场景下，资源复用，节省内存。（需要一定规模。一般场景下，效果不太明显）
- reflect
  - 缓存反射结果，减少不必要的反射次数。例如[json-iterator]（https://github.com/json-iterator/go）
  - 直接使用unsafe.Pointer根据各个字段偏移赋值。
  - 消除一般的struct反射内存消耗go-reflect（https://github.com/goccy/go-reflect）
  - 避免一些类型转换，如interface->[]byte。
- lock
  - 减小锁粒度:
     go标准库当中，math.rand就有这么一处隐患。当我们直接使用rand库生成随机数时，实际上由全局的globalRand对象负责生成。globalRand加锁后生成随机数，会导致我们在高频使用随机数的场景下效率低下。
  - atomic: 适当场景下，用原子操作代替互斥锁也是一种经典的lock-free技巧
- golink
  - golink（https://golang.org/cmd/compile/）使用格式：
  ```go
  //go:linkname FastRand runtime.fastrand
  func FastRand() uint32
  ```
  - 主要功能就是让编译器编译的时候，把当前符号指向到目标符号。上面的函数FastRand被指向到runtime.fastrand,runtime包生成的也是伪随机数，和math包不同的是，它的随机数生成使用的上下文是来自当前goroutine的，所以它不用加锁。正因如此，一些开源库选择直接使用runtime的随机数生成函数。
  - 另外，标准库中的`time.Now()`，这个库在会有两次系统调用runtime.walltime1和runtime.nanotime，分别获取时间戳和程序运行时间。大部分场景下，我们只需要时间戳，这时候就可以直接使用`runtime.walltime1`。
  - 系统调用在go里面相对来讲是比较重的。runtime会切换到g0栈中去执行这部分代码，time.Now方法在go<=1.16中有两次连续的系统调用
  ```go
  //go:linkname nanotime1 runtime.nanotime1
  func nanotime1() int64
  func main() {
      defer func( begin int64) {
          cost := (nanotime1() - begin)/1000/1000
          fmt.Printf("cost = %dms \n" ,cost)
      }(nanotime1())
      
      time.Sleep(time.Second)
  }
  ```
- log-函数名称行号的获取
  - 在runtime中，函数行号和函数名称的获取分为两步：
    - runtime回溯goroutine栈，获取上层调用方函数的的程序计数器（pc）。
    - 根据pc，找到对应的funcInfo,然后返回行号名称。
  - 经过pprof分析。第二步性能占比最大，约60%。针对第一步，我们经过多次尝试，并没有找到有效的办法。但是第二步很明显，我们不需要每次都调用runtime函数去查找pc和函数信息的，我们可以把第一次的结果缓存起来，后面直接使用。这样，第二步约60%的消耗就可以去掉。
  ```go
  var(
      m sync.Map
  )
  func Caller(skip int)(pc uintptr, file string, line int, ok bool){
      rpc := [1]uintptr{}
      n := runtime.Callers(skip+1, rpc[:])
      if n < 1 {
          return
      }
      var (
          frame  runtime.Frame
          )
      pc  = rpc[0]
      if item,ok:=m.Load(pc);ok{
          frame = item.(runtime.Frame)
      }else{
          tmprpc := []uintptr{
              pc,
          }
          frame, _ = runtime.CallersFrames(tmprpc).Next()
          m.Store(pc,frame)
      }
      return frame.PC,frame.File,frame.Line,frame.PC!=0
  }
  ```
- epoll
  - runtime对网络io，以及定时器的管理，会放到自己维护的一个epoll里，具体可以参考runtime/netpool。在一些高并发的网络io中，有以下几个问题：
    - 需要维护大量的协程去处理读写事件。
    - 对连接的状态无感知，必须要等待read或者write返回错误才能知道对端状态，其余时间只能等待。
    - 原生的netpool只维护一个epoll，没有充分发挥多核优势。
  - 基于此，有很多项目用x/unix扩展包实现了自己的基于epoll的网络库，比如gnet, 还有字节跳动的netpoll。
  - 在我们的项目中，也有尝试过使用。最终我们还是觉得基于标准库的实现已经足够。理由如下：
    - 用户态的goroutine优先级没有gonetpool的调度优先级高。带来的问题就是毛刺多了。近期字节跳动也开源了自己的netpool，并且通过优化扩展包内epoll的使用方式来优化这个问题，具体效果未知。
    - 效果不明显，我们绝大部分业务的QPS主要受限于其他的RPC调用，或者CPU计算。收发包的优化效果很难体现。
    - 增加了系统复杂性，虽然标准库慢一点点，但是足够稳定和简单。
- [如何高效地拼接字符串](https://mp.weixin.qq.com/s/9328Ju9pF80djNtRXqfSXQ)
  - **+** 操作符，也叫级联符
    拼接过程：
    - 1.编译器将字符串转换成字符数组后调用 runtime/string.go 的 concatstrings() 函数
    - 2.在函数内遍历字符数组，得到总长度
    - 3.如果字符数组总长度未超过预留 buf(32字节)，使用预留，反之，生成新的字符数组，根据总长度一次性分配内存空间
    - 4.将字符串逐个拷贝到新数组，并销毁旧数组
  - **+=** 追加操作符
    - 与 + 操作符相同，也是通过 runtime/string.go的concatstrings() 函数实现拼接，区别是它通常用于循环中往字符串末尾追加，每追加一次，生成一个新的字符串替代旧的，效率极低
  - strings.Builder 在 Golang 1.10 更新后，替代了byte.Buffer，成为号称效率最高的拼接方法。
    拼接过程：
    - 1.创建 []byte，用于缓存需要拼接的字符串
    - 2.通过 append 将数据填充到前面创建的 []byte 中
    - 3.append 时，如果字符串超过初始容量 8 且小于 1024 字节时，按乘以 2 的容量创建新的字节数组，超过 1024 字节时，按 1/4 增加
    - 4.将老数据复制到新创建的字节数组中 5.追加新数据并返回
  - strings.Join() 主要适用于以指定分隔符方式连接成一个新字符串，分隔符可以为空，在字符串一次拼接操作中，性能仅次于 + 操作符。
    拼接过程：
    - 1.接收的是一个字符切片
    - 2.遍历字符切片得到总长度，据此通过 builder.Grow 分配内存
    - 3.底层使用了 strings.Builder，每使用一次 strings.Join() ，都会创建新的 builder 对象
  - fmt.Sprintf()，返回使用 format 格式化的参数。除了字符串拼接，函数内还有很多格式方面的判断，性能不高，但它可以拼接多种类型，字符串或数字等

  结论：
  - 在待拼接字符串确定，可一次完成字符串拼接的情况下，推荐使用 + 操作符，即便 strings.Builder 用 Grow() 方法预先扩容，其性能也是不如 + 操作符的，另外，Grow()也不可设置过大。
  - 在拼接字符串不确定、需要循环追加字符串时，推荐使用 strings.Builder。但在使用时，必须使用 Grow() 预先扩容，否则性能不如 strings.Join()。
- [Set 的最佳实现方案](https://mp.weixin.qq.com/s/pcwCW7jtr2_CJ_k58He-6Q)
  - 使用 map 来实现 Set，意味着我们只关心 key 的存在，其 value 值并不重要
  - 我们选择了以下常用的类型作为 value 进行测试：bool、int、interface{}、struct{}。
  - 从内存开销而言，struct{} 是最小的，反映在执行时间上也是最少的。由于 bool 类型仅占一个字节，它相较于空结构而言，相差的并不多。但是，如果使用 interface{} 类型，那差距就很明显了
- [优化 Golang 分布式行情推送的性能瓶颈](https://mp.weixin.qq.com/s?__biz=MjM5MDUwNTQwMQ==&mid=2257486432&idx=1&sn=dea96a7309a8228bed48c80c3d675957&chksm=a539e2b6924e6ba06cdd8be0c410a5c3e0eb400b681b7f1d90a9cb3b6aad9fcc90ea74d8d99c&cur_album_id=1680805216599736323&scene=190#rd)
  - 并发操作map带来的锁竞争及时延
    - 推送的服务需要维护订阅关系，一般是用嵌套的map结构来表示，这样造成map并发竞争下带来的锁竞争和时延高的问题。
    - 解决方法：在每个业务里划分256个map和读写锁，这样锁的粒度降低到1/256。
      ```go
      sync.RWMutex
      map[string]map[string]client
      
      改成这样
      m *shardMap.shardMap
      ```
  - 串行消息通知改成并发模式
    - 在推送服务维护了某个topic和1w个客户端chan的映射，当从mq收到该topic消息后，再通知给这1w个客户端chan
      ```go
      notifiers := []*mapping.StreamNotifier{}
      // conv slice
      for _, notifier := range notifierMap {
          notifiers = append(notifiers, notifier)
      }
      
      
      // optimize: direct map struct
      taskChunks := b.splitChunks(notifiers, batchChunkSize)
      
      
      // concurrent send chan
      wg := sync.WaitGroup{}
      for _, chunk := range taskChunks {
          chunkCopy := chunk // slice replica
          wg.Add(1)
          b.SubmitBlock(
              func() {
                  for _, notifier := range chunkCopy {
                      b.directSendMesg(notifier, mesg)
                  }
                  wg.Done()
              },
          )
      }
      wg.Wait()
      ```
  - 过多的定时器造成cpu开销加大
    - go在1.9之后把单个timerproc改成多个timerproc，减少了锁竞争，但四叉堆数据结构的时间复杂度依旧复杂，高精度引起的树和锁的操作也依然频繁。
    - 改用[时间轮](https://github.com/rfyiamcool/go-timewheel)解决上述的问题。数据结构改用简单的循环数组和map，时间的精度弱化到秒的级别，业务上对于时间差是可以接受的。
  - 多协程读写chan会出现send closed panic的问题
    - 解决的方法很简单，就是不要直接使用channel，而是封装一个触发器，当客户端关闭时，不主动去close chan，而是关闭触发器里的ctx，然后直接删除topic跟触发器的映射。
      ```go
      // 触发器的结构
      type StreamNotifier struct {
          Guid  string
          Queue chan interface{}
      
      
          closed int32
          ctx    context.Context
          cancel context.CancelFunc
      }
      
      
      func (sc *StreamNotifier) IsClosed() bool {
          if sc.ctx.Err() == nil {
              return false
          }
          return true
      }
      ```
  - 提高grpc的吞吐性能
    - 内网的两个节点使用单连接就可以跑满网络带宽，无性能问题。但在golang里实现的grpc会有各种锁竞争的问题。
    - 如何优化？多开grpc客户端，规避锁竞争的冲突概率。[测试](https://github.com/rfyiamcool/grpc_batch_test)下来qps提升很明显，从8w可以提到20w左右。
  - 减少协程数量
    - 有朋友认为等待事件的协程多了无所谓，只是占内存，协程拿不到调度，不会对runtime性能产生消耗。这个说法是错误的。虽然拿不到调度，看起来只是占内存，但是会对 GC 有很大的开销。所以，不要开太多的空闲的协程，比如协程池开的很大。
    - 在推送的架构里，push-gateway到push-server不仅几个连接就可以，且几十个stream就可以。我们自己实现大量消息在十几个stream里跑，然后调度通知。在golang grpc streaming的实现里，每个streaming请求都需要一个协程去等待事件。所以，共享stream通道也能减少协程的数量。
  - GC
    - 对于频繁创建的结构体采用sync.Pool进行缓存。
    - 有些业务的缓存先前使用list链表来存储，在不断更新新数据时，会不断的创建新对象，对 GC 造成影响，所以改用可复用的循环数组来实现热缓存。
- [更快的时间解析](https://colobu.com/2021/10/10/faster-time-parsing/)
    ```go
    func BenchmarkParseRFC3339(b *testing.B) {
        now := time.Now().UTC().Format(time.RFC3339Nano)
        for i := 0; i < b.N; i++ {
            if _, err := time.Parse(time.RFC3339, now); err != nil {
                b.Fatal(err)
            }
        }
    }
    ```
  如果我们采样 CPU profile,我们观察到很多时间都花费在调用strconv.Atoi上
    ```go
    > go test -run ^$ -bench BenchmarkParseRFC3339 -cpuprofile cpu.prof 
    > go tool pprof cpu.prof
    Type: cpu
    Time: Oct 1, 2021 at 7:19pm (BST)
    Duration: 1.22s, Total samples = 960ms (78.50%)
    Entering interactive mode (type "help" for commands, "o" for options)
    (pprof) top
    Showing nodes accounting for 950ms, 98.96% of 960ms total
    Showing top 10 nodes out of 24
          flat  flat%   sum%        cum   cum%
         380ms 39.58% 39.58%      380ms 39.58%  strconv.Atoi
         370ms 38.54% 78.12%      920ms 95.83%  github.com/philpearl/blog/content/post.parseTime
          60ms  6.25% 84.38%      170ms 17.71%  time.Date
    ```
  我们的大部分数字正好有2个字节长，或者正好有4个字节长。我们可以编写数字解析函数，针对我们的特殊情况做优化，不需要任何令人讨厌的慢循环:
    ```go
    func atoi2(in string) (int, error) {
        a, b := int(in[0]-'0'), int(in[1]-'0')
        if a < 0 || a > 9 || b < 0 || b > 9 {
            return 0, fmt.Errorf("can't parse number %q", in)
        }
        return a*10 + b, nil
    }
    func atoi4(in string) (int, error) {
        a, b, c, d := int(in[0]-'0'), int(in[1]-'0'), int(in[2]-'0'), int(in[3]-'0')
        if a < 0 || a > 9 || b < 0 || b > 9 || c < 0 || c > 9 || d < 0 || d > 9 {
            return 0, fmt.Errorf("can't parse number %q", in)
        }
        return a*1000 + b*100 + c*10 + d, nil
    }
    ```
  让我们在看一眼现在的CPU profile, 并且看一些汇编代码。在atoi2中有两个slice长度检查(下面绿色的汇编代码,调用panicIndex之前)，不是有一个[边界检查](https://go101.org/article/bounds-check-elimination.html)的技巧吗？

  以下是根据此技巧进行修正后的代码。函数开始处的_ = in[1]给了编译器充足的提示，这样我们在调用它的时候不用每次都检查是否溢出了:
    ```go
    func atoi2(in string) (int, error) {
        _ = in[1] // This helps the compiler reduce the number of times it checks `in` is long enough
        a, b := int(in[0]-'0'), int(in[1]-'0')
        if a < 0 || a > 9 || b < 0 || b > 9 {
            return 0, fmt.Errorf("can't parse number %q", in)
        }
        return a*10 + b, nil
    }
    ```
  atoi2非常短。为什么它不被内联的？如果我们简化错误处理，是不是有效果？如果我们删除对fmt.Errorf的调用，并将其替换为一个简单的错误类型，这将降低atoi2函数的复杂性。这可能足以让Go编译器决定不作为单独的代码块而是直接在调用函数中内联这个函数。
    ```go
    var errNotNumber = errors.New("not a valid number")
    func atoi2(in string) (int, error) {
        _ = in[1]
        a, b := int(in[0]-'0'), int(in[1]-'0')
        if a < 0 || a > 9 || b < 0 || b > 9 {
            return 0, errNotNumber
        }
        return a*10 + b, nil
    }
    ```
- [逃逸分析来提升程序性能](https://mp.weixin.qq.com/s/exQy5I7RQBVADFNe1wcbqw)
  - 逃逸分析
    - 在对变量放到堆上还是栈上进行分析，该分析在编译阶段完成。如果一个变量超过了函数调用的生命周期，也就是这个变量在函数外部存在引用，编译器会把这个变量分配到堆上，这时我们就说这个变量发生逃逸了。
  - 如何确定是否逃逸
    - `go run -gcflags '-m -l' main.go`
  - 可能出现逃逸的场景
    - **interface{}** 赋值，会发生逃逸，优化方案是将类型设置为固定类型
      ```go
      type Student struct {
       Name interface{}  // ---> String
      }
      
      func main()  {
       stu := new(Student)
       stu.Name = "tom"
      
      }
      ```
    - 返回指针类型，会发生逃逸
      - 函数传递指针和传值哪个效率高吗？我们知道传递指针可以减少底层值的拷贝，可以提高效率，但是如果拷贝的数据量小，由于指针传递会产生逃逸，可能会使用堆，也可能会增加 GC 的负担，所以传递指针不一定是高效的
    - 栈空间不足，会发生逃逸，优化方案尽量设置容量
- [Deep Dive into The Escape Analysis in Go](https://slides.com/jalex-chang/go-esc)
  - The escape analysis is a mechanism to automatically decide whether a variable should be allocated in the heap or not in compile time.
    - It tries to keep variables on the stack as much as possible.
    - If a variable would be allocated in the heap, the variable is escaped (from the stack).
    - ESC would consider assignment relationships between declared variables.
    - Generally, a variable scapes if:
      - its address has been captured by ​the address-of operand (&).
      - and at least one of the related variables has already escaped.
  - Basically, ESC determines whether variables escape or not by
    - the data-flow analysis (shortest path analysis)
      - Data-flow is a directed weighted graph
      - Constructed from the abstract syntax tree (AST).
      - It is used to represent relationships between variables.
    - and other additional rules
      - Huge Objects
        - For explicit declarations (var or :=)  The variables escape if their sizes are over 10MB
        - For implicit declarations (new or make). The variables escape if their sizes are over 64KB 
      - A slice variable escapes if its size of the capacity is non-constant. 
      - Map
        - A variable escapes if it is referenced by a map's key or value.
        - The escape happens no matter the map escape or not.
      - Returning values is a backward behavior that
        - the referenced variables escape if the return values are pointers
        - the values escape if they are map or slice
      - Passing arguments is a forward behavior that
        - the arguments escape if input parameters have leaked (to heap)
      - A variable escapes if
        - the source variable is captured by a closure function
        - and their relationship is address-of (derefs = -1 )
  - Observations
    - Through understanding the concept of ESC, we can find that
      - variables usually escape
        - when their addresses are captured by other variables.
        - when ESC does not know their object sizes in compile time.
      - And passing arguments to a function is safer than returning values from the function. 
    - Initialize slice with constants
    - Passing variables to closure as arguments instead of accessing the variables directly.
    - Injecting changes to the passed parameters instead of return values back
- [TiDB TPS 提升 1000 倍的性能优化之旅](https://gocn.vip/topics/20825)
  - TPS 从 1 到 30
    - 第一个 SQL 优化例子是解决索引缺失的问题
    - 第二个 SQL 优化的例子是解决有索引却用不上的问题
  - TPS 从 30 到 320
    - 测试环境是六台 ARM 服务器，每台 16 个 Numa，每个 Numa 是 8C 16GB。
    - 我们对这个组网的方式做了调整，部署了 36 个 TiDB + 6 个 TiKV，每个 TiDB 会绑两个 Numa ，每个 TiKV 有四个 Numa
  - TPS 从 320 到 600
    - 我们对整体的火焰图和网络做了一些分析。由下方火焰图可见，整个系统的 CPU 20% 是消耗在一个叫 finish_task_switch 的，做进程切换，调度相关的系统调用上，说明系统在内核态存在资源争抢和串行点。
    - 我们使用 `mpstat -P ALL 5` 命令对所有 CPU 的利用率进行确认，发现了一个比较有趣的现象 —— 所有的网卡的软中断（%soft），都打到了第一个 Numa（CPU 0-7）上
    - 又因为我们在第一个 Numa 上面还跑着 TiDB、PD 和 Haproxy 等，用户 CPU（%usr）是 2% 到将近 40%，第一个 Numa 的 CPU 都被打满了（%idle 接近 0）。其他的 Numa 使用率仅 55% 左右。
    - 对于没有绑核的程序 —— PD 和 Haproxy，我们在火焰图里面观察到关于内存的访问或者内存的加锁等系统调用占比非常高。对于开启 Numa 的系统，其实 CPU 访问内存的速度是不平等的。通常访问远端 Numa 的内存延迟是访问本地 Numa 内存的十倍。硬件厂商也推荐应用最好不要进行跨 Numa 部署
    - 我们进行了组网方式的调整。对于六台机器
      - 1）第一个 Numa 都空出来专门处理网络软中断，不跑任何的程序；
      - 2）所有的程序都需要绑核，每个 TiDB 只绑一个 Numa，TiDB 的数据翻倍， PD 和 Haproxy 也进行绑核
  - PS 从 600 到 880
    - 数据库最大连接数稳定在 2000，应用加大并发连接数也没有提升。
    - 使用 mysql 连接 Haproxy 地址会报错。因为 Haproxy 单个 proxy 后台 session 限制默认两千，通过把 Haproxy 从多线程模式改成了多进程的模式可以解除这个限制。
  - TPS 抖动解决
    - TPS 880 时应用出现明显的波动，事务处理延迟出现巨大的波动. 查看 P9999 延迟，发现波动巨大
    - SQL 执行计划稳定性 - 永不准确的统计信息. 统计信息是否具有代表性，取决于统计信息更新时，数据的状态
  - TPS 880 到 1200+
    - 使用一台 ARM 服务器，同样是 16 个 Numa，部署 15 个应用，每个应用 jvm 绑定一个 Numa，连接到 TiDB 集群
- [A 5x reduction in RAM usage with Zoekt memory optimizations](https://about.sourcegraph.com/blog/zoekt-memory-optimizations-for-sourcegraph-cloud/)
  - Measure how a server’s RAM is being used
    - you can set the GOGC environment variable to more aggressively reduce the maximum overhead. We run Zoekt with `GOGC=50` to reduce the likelihood that it will exceed its available memory.
    - built-in profiling tools. Digging into the code, this turned out to be a function that builds a map from trigrams to the location of a posting list on disk. It’s building a big mapping from 64-bit trigrams (three 21-bit Unicode characters) to 32-bit offsets and lengths.
  - Implement a more compact data structure for locating postings lists
    - Go maps provide O(1) access times, but they consume a fair amount of memory per entry— roughly 40 bytes each.
    - Storing these mappings as two slices instead of a map reduces its memory usage from 15GB to 5GB
      ```go
      type arrayNgramOffset struct {
             ngrams []ngram
             // offsets is values from simpleSection.off. simpleSection.sz is computed by subtracting adjacent offsets.
             offsets []uint32
      }
      ```
  - metadata optimizations
    - you copy a slice that grew dynamically into a precisely sized one, you don’t waste the unused trailing capacity
      ```go
      // shrinkUint32Slice copies slices with excess capacity to precisely sized ones
      // to avoid wasting memory. It should be used on slices with long static durations.
      func shrinkUint32Slice(a []uint32) []uint32 {
             if cap(a)-len(a) < 32 {
                     return a
             }
             out := make([]uint32, len(a))
             copy(out, a)
             return out
      }
      ```
- [简单的服务响应时长优化方法](https://mp.weixin.qq.com/s/YP06ErRfydZ1R6-_J5-fcQ)
  - 如果是串行调用的话响应时间会随着 rpc 调用次数呈线性增长，所以我们要优化性能一般会将串行改并行. 简单的场景下使用 waitGroup 也能够满足需求，但是如果我们需要对 rpc 调用返回的数据进行校验、数据加工转换、数据汇总呢？继续使用 waitGroup 就有点力不从心了
  - 通过 MapReduce 把正交（不相关）的请求并行化，你就可以大幅降低服务响应时长
- [优化redis写入而降低cpu使用率](https://mp.weixin.qq.com/s/16Fn7LahXSadTHS0NXcapQ)
  - 背景
    - 项目中基于redis记录实时请求量的一个功能，因流量上涨造成redis服务器的CPU高于80%而触发了自动报警机制，经分析将实时写入redis的方式变更成批量写入的方式，从而将CPU使用率降低了30%左右的经历
  - v1
    - 第一个版本很简单，就是将最大值存放在redis中，然后按天的维度记录每个国家流量的实时请求数量。每次流量来了之后，先查询出该国家流量的最大值，以及当天的实时请求数，然后做比较，如果实时数已经超过了最大值，就直接返回，否则就对实时数进行+1操作即可。
     ```go
         maxReq := redis.Get(key)
     
         day := time.Now().Format("20060102")
         dailyKey := "CN:"+day+":req"
         dailyReq := redis.Get(dailyKey)
     
         if dailyReq > maxReq {
             return true
         }
     
         redis.Incr(dailyKey, dailyReq)
         redis.Expire(dailyKey, 7*24*time.Hour)
     ```
  - v2
    - 我们通过使用一个hasUpdateExpire的map类型，来记录某个key是否已经被设置了有效期的标识. 减少Expire的执行次数
  - v3 异步批量写入
    - 我们的技术不直接写入redis，而是写在内存缓存中，即一个全局变量中，同时启动一个定时器，每隔一段时间就将内存中的数据批量写入到redis中
      ```go
      type CounterCache struct {
         rwMu        sync.RWMutex
         redisClient redis.Cmdable
      
         countCache   map[string]int64
         hasUpdateExpire map[string]struct{}
      }
      ```
  - v4 maybe
    - redis 多个命令采用pipeline方式执行
- [Golang simple optimization notes](https://medium.com/scum-gazeta/golang-simple-optimization-notes-70bc64673980)
  - Arrays and slices
    - Don’t forget to use “copy” 
      - We try not to use append when copying or, for example, when merging two or more slices.
    - We iterate correctly
      - If we have a slice with many elements, or with large elements, we try to use “for” or range with a single element. With this approach, we will avoid unnecessary copying.
    - Reusing slices
      - If we need to carry out some kind of manipulation with the incoming slice and return the result, we can return it, but already modified. This way we avoid new memory allocations.
    - We do not leave unused slices
      - If we need to cut off a small piece from a slice and use only it, remember that the main part will also remain with you forever. We use copy for a new piece to send the old one to the GC.
  - strings
    - Doing concatenation correctly
      - If gluing strings can be done in one statement, then we use “+”, if we need to do this in a loop, then we use string.Builder. Specify the size for the builder in advance through “Grow”
    - Using transformation optimization
      - Since strings under the hood consist of a slice of bytes, sometimes conversions between these two types allow you to avoid memory allocation.
    - Using Internment
      - We can pool strings, thereby helping the compiler store identical strings only once.
    - Avoiding Allocations
      - We can use a map (concatenation) instead of a composite key, we can use a slice of bytes. We try not to use the fmt package, because all of its functions use reflection.
  - structures
    - Avoid copying large structures
      - Standard copy cases
        - cast to interface
        - receiving and sending to channels
        - replacing an entry in a map
        - adding an element to a slice
        - iteration (range)
    - Avoid accessing struct fields through pointers
      - Dereferencing is expensive, we can do it as little as possible especially in a loop. We also lose the ability to use fast registers.
    - Work with small structures
      - This work is optimized by the compiler, which means it is cheap.
    - Reduce structure size with alignment
      - We can align our structures (arrange the fields in the right order, depending on their size) and thus we can reduce the size of the structure itself.
  - func
    - Use inline functions or inline them yourself
      - We try to write small functions available for inlining by the compiler — it’s fast, but it’s even faster to embed code from functions yourself. This is especially true for hot path functions.
      - What won’t inlined?
        - recovery func
        - select blocks
        - type declarations
        - defer
        - goroutine
        - for-range
    - Choose your function arguments wisely
      - We try to use “small” arguments, as their copying will be specially optimized. We also try to keep a balance between copying and growing the stack with a load on the GC.
      - Avoid a large number of arguments — let your program use super fast registers (there are a limited number of them)
    - Use “defer” carefully
      - Try not to use defer, or at least not use it in a loop.
    - Facilitating the “hot path”
      - Avoid allocating memory in these places, especially for short-lived objects. Make the most common branches first (if, switch).
  - map
    - Using an empty structure as values
      - struct{} is nothing, so using this approach for example for signal values is very beneficial.
    - Clearing the map
      - The map can only grow and cannot shrink. We need to control this — reset the maps completely and explicitly, because. deleting all of its elements won’t help.
    - We try not to use pointers in keys and values
      - If the map does not contain pointers, then the GC will not waste its precious time on it. And know that strings are also pointers — use an array of bytes instead of strings for keys.
    - Reducing the number of changes
      - Again, we do not want to use a pointer, but we can use a composite of a map and a slice and store the keys in the map, and in the slice the values ​​that we can already change without restrictions.
  - interface
    - Counting memory allocations
      - Remember, to assign a value to an interface, you first need to copy it somewhere and then paste a pointer to it. The keyword is copy. And it turns out that the cost of boxing and unboxing will be approximate to the size of the structure and one allocation
    - Choosing the optimal types
      - There are some cases when there will be no allocations during boxing / unboxing. For example, small and boolean values of variables and constants, structures with one simple field, pointers (map, chan, func including)
    - Avoiding memory allocation
      - As elsewhere, we try to avoid unnecessary allocations. For example, to assign an interface to an interface, instead of boxing twice.
    - Use only when needed
      - Avoid using interfaces in the parameters and results of small, frequently called functions. We do not need extra packing and unpacking.
      - Use interface method calls less frequently, if only because it prevents inlining.
  - Pointers, channels, BCE
    - Avoid unnecessary dereferences
      - Especially in a loop, because it turns out to be too expensive. Dereferencing is a whole complex of necessary actions that we do not want to perform at our expense.
    - Channel usage is inefficient
      - Channels are slower than other synchronization methods. In addition, the more cases in select, the slower our program. But select, case + default are optimized.
    - Try to avoid unnecessary boundary checks
      - This is also expensive and we should avoid it in every possible way. For example, it is more correct to check (get) the maximum slice index once, instead of several checks. It is better to immediately try to get extreme options.
- [计算密集型服务 性能优化实战始末](https://mp.weixin.qq.com/s/aIKNqQAaI37iJPEEi9z8YQ)
  - 面对问题
    - worker 服务消费上游数据（工作日高峰期产出速度达近 200 MB/s，节假日高峰期可达 300MB/s 以上）。基于快慢隔离的思想，以三个不同的 consumer group 消费同一 Topic，隔离三种数据处理链路
    - worker 服务在高峰期时 CPU Idle 会降至 60%，因其属于数据处理类计算密集型服务，CPU Idle 过低会使服务吞吐降低，在数据处理上产生较大延时，且受限于 Kafka 分区数，无法进行横向扩容；
  - 性能优化
    - 服务与存储之间置换压力
      - 背景
        - 对于 Apollo 有较严重的大 Key 问题，再结合 RocksDB 特有的写放大问题，会进一步加剧存储压力。在这个背景下，我们采用 zlib 压缩算法，对消息体先进行压缩后再写入 Apollo，减缓读写大 Key 对 Apollo 的压力。
      - 优化
        - 在 CPU 的优化过程中，我们发现服务在压缩操作上占用了较多的 CPU，于是对压缩等级进行调整，以减小压缩率、增大下游存储压力为代价，减少压缩操作对服务 CPU 的占用，提升服务 CPU 。
      - 关于压缩等级
        - 在压缩等级的设置上可能存在较为严重的边际效用递减问题。
        - 在进行基准测试时发现，将zlib压缩等级由 BestCompression 调整为 DefaultCompression 后，压缩率只有近 1‱ 的下降，但压缩方面的 CPU 占用却相对提高近 **50%**。
    - 使用更高效的序列化库
      - 背景
        - worker 服务在设计之初基于快慢隔离的思想，使用三个不同的 consumer group 进行分开消费，导致对同一份数据会重复消费三次，而上游产出的数据是在 PB 序列化之后写入 Kafka，消费侧亦需要进行 PB 反序列化方能使用，因此导致了 PB 反序列化操作在 CPU 上的较大开销。
      - 优化
        - 采用 gogo/protobuf 库替换掉原生的 golang/protobuf 库
        - gogo/protobuf 为什么快？
          - 通过对每一个字段都生成代码的方式，取消了对反射的使用；
          - 采用预计算方式，在序列化时能够减少内存分配次数，进而减少了内存分配带来的系统调用、锁和 GC 等代价。
        - [用过去或未来换现在的时间](https://mp.weixin.qq.com/s/S8KVnG0NZDrylenIwSCq8g)
          - 页面静态化、池化技术、预编译、代码生成等都是提前做一些事情，用过去的时间，来降低用户在线服务的响应时间；
          - 另外对于一些在线服务非必须的计算、存储的耗时操作，也可以异步化延后进行处理，这就是用未来的时间换现在的时间
    - 数据攒批 减少调用
      - 背景
        - 在观察 pprof 图后发现写 hbase 占用了近 50% 的相对 CPU，经过进一步分析后，发现每次在序列化一个字段时 Thrift 都会调用一次 socket->syscall，带来频繁的上下文切换开销。
      - 优化
        - 原代码中使用了 Thrift 的 TTransport 实现，其功能是包装 TSocket，裸调 Syscall，每次 Write 时都会调用 socket 写入进而调用 Syscall。
        - 发现其中有多种 Transport 实现，而 TTBufferedTransport 是符合我们编码习惯的。
        - 数据攒批：将数据先写入用户态内存中，而后统一调用 syscall 进行写入，常用在数据落盘、网络传输中，可降低系统调用次数、利用磁盘顺序写特性等，是一种空间换时间的做法。有时也会牺牲一定的数据实时性
    - 语法调整
      - slice、map 预初始化，减少频繁扩容导致的内存拷贝与分配开销
      - 字符串连接使用 strings.builder(预初始化) 代替 fmt.Sprintf()
      - buffer 修改返回 string([]byte) 操作为 []byte，减少内存 []byte -> string 的内存拷贝开销
      - string <-> []byte 的另一种优化，需确保 []byte 内容后续不会被修改
    - GC 调优
      - 背景
        - 在上次优化完成之后，系统已经基本稳定，CPU Idle 高峰期也可以维持在 80% 左右，但后续因业务诉求对上游数据采样率调整至 100%，CPU.Idle 高峰期指标再次下降至近 70%，且由于定时任务的问题，存在 CPU.Idle 掉 0 风险；
      - 优化
        - 经过对 pprof 的再次分析，发现 runtime.gcMarkWorker 占用不合常理，达到近 30%，于是开始着手对 GC 进行优化
        - 方法一：使用 sync.pool()
          - 我们在项目中使用其对 bytes.buffer 对象进行缓存复用，意图减少 GC 开销，但实际上线后 CPU Idle 却略微下降，且 GC 问题并无缓解。原因有二：
          - sync.pool 是全局对象，读写存在竞争问题，因此在这方面会消耗一定的 CPU，但之所以通常用它优化后 CPU 会有提升，是因为它的对象复用功能对 GC 带来的优化，因此 sync.pool 的优化效果取决于锁竞争增加的 CPU 消耗与优化 GC 减少的 CPU 消耗这两者的差值；
          - GC 压力的大小通常取决于 inuse_objects，与 inuse_heap 无关，也就是说与正在使用的对象数有关，与正在使用的堆大小无关；
          - 本次优化时选择对 bytes.buffer 进行复用，是想做到减少堆大小的分配，出发点错了，对 GC 问题的理解有误，对 GC 的优化因从 pprof heap 图 inuse_objects 与 alloc_objects 两个指标出发。
        - 方法二：设置 GOGC
          - GOGC 默认值是 100，也就是下次 GC 触发的 heap 的大小是这次 GC 之后的 heap 的一倍，通过调大 GOGC 值（gcpercent）的方式，达到减少 GC 次数的目的
          - 问题：GOGC 参数不易控制，设置较小提升有限，设置较大容易有 OOM 风险，因为堆大小本身是在实时变化的，在任何流量下都设置一个固定值，是一件有风险的事情。这个问题目前已经有解决方案，Uber 发表的文章中提到了一种自动调整 GOGC 参数的方案，用于在这种方式下优化 GO 的 GC CPU 占用
            `debug.SetGCPercent(1000)`
        - 方法三：[GO ballast 内存控制](https://blog.twitch.tv/en/2019/04/10/go-memory-ballast-how-i-learnt-to-stop-worrying-and-love-the-heap/)
          - 仍然是从利用了下次 GC 触发的 heap 的大小是这次 GC 之后的 heap 的一倍这一原理，初始化一个生命周期贯穿整个 Go 应用生命周期的超大 slice，用于内存占位，增大 heap_marked 值降低 GC 频率；实际操作有以下两种方式
          - `stub = make([]byte, 100MB)  stub[0]=1` - 会实际占用物理内存，在可观测性上会更舒服一点
          - `ballast = make([]byte, 100MB)  runtime.KeepAlive(ballast)` - 并不会实际占用物理内存
      - 关于 GC 调优
        - GC 优化手段的优先级：设置 GOGC、GO ballast 内存控制等操作是一种治标不治本略显 trick 的方式，在做 GC 优化时还应先从对象复用、减少对象分配角度着手，在确无优化空间或优化成本较大时，再选择此种方式；
        - 设置 GOGC、GO ballast 内存控制等操作本质上也是一种空间换时间的做法，在内存与 CPU 之间进行压力置换；
        - 在 GC 调优方面，还有很多其他优化方式，如 bigcache 在堆内定义大数组切片自行管理、fastcache 直接调用 syscall.mmap 申请堆外内存使用、offheap 使用 cgo 管理堆外内存等等。








