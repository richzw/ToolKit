
- [Go 编码建议——性能篇](https://dablelv.blog.csdn.net/article/details/122281882)
  - 1.反射虽好，切莫贪杯
    - 优先使用 strconv 而不是 fmt
    - 少量的重复不比反射差
    - 慎用 binary.Read 和 binary.Write - binary.Read 和 binary.Write 使用反射并且很慢
  - 避免重复的字符串到字节切片的转换
  - 指定容器容量
    - 指定 map 容量提示
    - 指定切片容量
  - 字符串拼接方式的选择
    - 行内拼接字符串推荐使用运算符+ 
      - 如果待拼接的变量不涉及类型转换且数量较少（<=5），行内拼接字符串推荐使用运算符 +，反之使用 fmt.Sprintf()
    - 非行内拼接字符串推荐使用 strings.Builder
  - 遍历 []struct{} 使用下标而不是 range
    - range 在迭代过程中返回的是元素的拷贝，index 则不存在拷贝。
    - 如果 range 迭代的元素较小，那么 index 和 range 的性能几乎一样，如基本类型的切片 []int。
    - 但如果迭代的元素较大，如一个包含很多属性的 struct 结构体，那么 index 的性能将显著地高于 range，有时候甚至会有上千倍的性能差异。
      - 对于这种场景，建议使用 index。
      - 如果使用 range，建议只迭代下标，通过下标访问元素，这种使用方式和 index 就没有区别了。
      - 如果想使用 range 同时迭代下标和值，则需要将切片/数组的元素改为指针，才能不影响性能。
  - 使用空结构体节省内存 - 不占内存空间
    - 实现集合（Set）
    - 不发送数据的信道
    - 仅包含方法的结构体 
  - struct 布局考虑内存对齐
    - 为什么需要内存对齐
      - 合理的内存对齐可以提高内存读写的性能，并且便于实现变量操作的原子性
    - Go 内存对齐规则
    - 合理的 struct 布局可以减少内存占用，提高程序性能
    - 空结构与空数组对内存对齐的影响 
    - 在对内存特别敏感的结构体的设计上，我们可以通过调整字段的顺序，将字段宽度从小到大由上到下排列，来减少内存的占用
  - 减少逃逸，将变量限制在栈上
    - 变量逃逸一般发生在如下几种情况：
       - 变量较大（栈空间不足）
       - 变量大小不确定（如 slice 长度或容量不定）
       - 返回地址
       - 返回引用（引用变量的底层是指针）
       - 返回值类型不确定（不能确定大小）
       - 闭包
    - 局部切片尽可能确定长度或容量
    - 返回值 VS 返回指针
      - 值传递会拷贝整个对象，而指针传递只会拷贝地址，指向的对象是同一个。
      - 传指针可以减少值的拷贝，但是会导致内存分配逃逸到堆中，增加垃圾回收（GC）的负担。在对象频繁创建和删除的场景下，返回指针导致的 GC 开销可能会严重影响性能。
    - 小的拷贝好过引用
    - 返回值使用确定的类型
  - sync.Pool 复用对象
    - 简介 - sync.Pool 是可伸缩的，同时也是并发安全的，其容量仅受限于内存的大小。存放在池中的对象如果不活跃了会被自动清理。
    - 作用 - 用来保存和复用临时对象，减少内存分配，降低 GC 压力。
  - 1.关于锁
    - 无锁化
      - 无锁数据结构
      - 串行无锁
        - 串行无锁是一种思想，就是避免对共享资源的并发访问，改为每个并发操作访问自己独占的资源，达到串行访问资源的效果，来避免使用锁。不同的场景有不同的实现方式。
        - 比如网络 I/O 场景下将单 Reactor 多线程模型改为主从 Reactor 多线程模型，避免对同一个消息队列锁读取。
    - 减少锁竞争 - 可以采用分片的形式，减少对资源加锁的次数，这样也可以提高整体的性能
    - 优先使用共享锁而非互斥锁
    - 限制协程数量
      - 协程数过多的问题
        - 内存开销
        - 调度开销
        - GC 开销
      - 限制协程数量
      - 协程池化
      - 小结

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
    - [使用更高效的序列化库](https://segmentfault.com/a/1190000041591284)
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
    - [数据攒批 减少调用](https://mp.weixin.qq.com/s/ntNGz6mjlWE7gb_ZBc5YeA)
      - 背景
        - 在观察 pprof 图后发现写 hbase 占用了近 50% 的相对 CPU，经过进一步分析后，发现每次在序列化一个字段时 Thrift 都会调用一次 socket->syscall，带来频繁的上下文切换开销。
      - 优化
        - 原代码中使用了 Thrift 的 TTransport 实现，其功能是包装 TSocket，裸调 Syscall，每次 Write 时都会调用 socket 写入进而调用 Syscall。
        - 发现其中有多种 Transport 实现，而 TTBufferedTransport 是符合我们编码习惯的。
        - 数据攒批：将数据先写入用户态内存中，而后统一调用 syscall 进行写入，常用在数据落盘、网络传输中，可降低系统调用次数、利用磁盘顺序写特性等，是一种空间换时间的做法。有时也会牺牲一定的数据实时性
    - [语法调整](https://mp.weixin.qq.com/s/Lv2XTD-SPnxT2vnPNeREbg)
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
- [80x caching improvement in Go](https://www.ksred.com/80x-caching-improvement-in-go/)
  - By moving away from JSON and using [gob](https://pkg.go.dev/encoding/gob) instead, I made an 84x improvement on reads, and a 20x improvement on writes.
    ```go
    // Using JSON
    err = json.Unmarshal(obj, result)
    if err != nil {
        return
    }
    
    // Using gob
    buf := bytes.NewBuffer(obj)
    dec := gob.NewDecoder(buf)
    err = dec.Decode(&result)
    if err != nil {
        return
    }
    ```
- [IO 密集型服务 性能优化实战记录](https://mp.weixin.qq.com/s/83M0j8lIALdF-eOdD5GukA)
  - 背景
    - 项目背景
      - Feature 服务作为特征服务，产出特征数据供上游业务使用。服务压力：高峰期 API 模块 10wQPS，计算模块 20wQPS。服务本地缓存机制
    - 面对问题
      - 服务 API 侧存在较严重的 P99 耗时毛刺问题（固定出现在每分钟第 0-10s），导致上游服务的访问错误率达到 1‰ 以上，影响到业务指标；
      - 目标：解决耗时毛刺问题，将 P99 耗时整体优化至 15ms 以下；
    - 导致服务的个别部分出现高尾部延迟的响应时间的变异性（耗时长尾的原因）可能由于许多原因而产生，包括：
      - 共享的资源。机器可能被不同的应用程序共享，争夺共享资源（如CPU核心、处理器缓存、内存带宽和网络带宽）（在云上环境中这个问题更甚，如不同容器资源争抢、Sidecar 进程影响）；在同一个应用程序中，不同的请求可能争夺资源。
      - 守护程序。后台守护程序可能平均只使用有限的资源，但在安排时可能产生几毫秒的中断。
      - 全局资源共享。在不同机器上运行的应用程序可能会争夺全球资源（如网络交换机和共享文件系统（数据库））。
      - 维护活动。后台活动（如分布式文件系统中的数据重建，BigTable等存储系统中的定期日志压缩（此处指 LSM Compaction 机制，基于 RocksDB 的数据库皆有此问题），以及垃圾收集语言中的定期垃圾收集（自身和上下游都会有 GC 问题 1. Codis proxy 为 GO 语言所写，也会有 GC 问题；2. 此次 Feature 服务耗时毛刺即时因为服务本身 GC 问题，详情见下文）会导致周期性的延迟高峰；以及排队。中间服务器和网络交换机的多层排队放大了这种变化性。
  - 解决方案
    - 服务 CPU 优化
      - 从提高服务 CPU Idle 角度入手，对服务耗时毛刺问题展开优化
      - 通过减少反序列化操作、更换 JSON 序列化库（json-iterator）两种方式进行了优化
        - 反序列化时的开销减少，使单个请求中的计算时间得到了减少；
        - 单个请求的处理时间减少，使同时并发处理的请求数得到了减少，减轻了调度切换、协程/线程排队、资源竞争的开销
      - [json-iterator 库为什么快](https://cloud.tencent.com/developer/article/1064753)
        - 标准库 json 库使用 reflect.Value 进行取值与赋值，但 reflect.Value 不是一个可复用的反射对象，每次都需要按照变量生成 reflect.Value 结构体，因此性能很差。
        - json-iterator 实现原理是用 reflect.Type 得出的类型信息通过「对象指针地址+字段偏移」的方式直接进行取值与赋值，而不依赖于 reflect.Value，reflect.Type 是一个可复用的对象，同一类型的 reflect.Type 是相等的，因此可按照类型对 reflect.Type 进行 cache 复用。
        - 总的来说其作用是减少内存分配和反射调用次数，进而减少了内存分配带来的系统调用、锁和 GC 等代价，以及使用反射带来的开销。
    - 调用方式优化 - 对冲请求
      - Feature 服务 API 模块访问计算模块 P99 显著高于 P95
      - 经观察计算模块不同机器之间毛刺出现时间点不同，单机毛刺呈偶发现象，所有机器聚合看呈规律性毛刺
      - 优化 - [The Tail at Scale](https://mp.weixin.qq.com/s/BKMPNix-zn64-0MP3XVLeA)
        - 针对 P99 高于 P95 现象，提出对冲请求方案，对毛刺问题进行优化
        - 对冲请求(Hedged requests.)：把对下游的一次请求拆成两个，先发第一个，n毫秒超时后，发出第二个，两个请求哪个先返回用哪个
        - 对冲请求是从概率的角度消除偶发因素的影响，从而解决长尾问题，因此需要考量耗时是否为业务侧自身固定因素导致
        - 局限性
          - 请求需要幂等，否则会造成数据不一致
          - 对冲请求超时时间并非动态调整，而是人为设定，因此极端情况下会有雪崩风险；
            - BRPC 实践：对冲请求会消耗一次对下游的重试次数；
            - bilibili 实践：
              - 对 retry 请求下游会阻断级联；
              - 本身要做熔断；
              - 在 middleware 层实现窗口统计，限制重试总请求占比，比如 1.1 倍；
    - 语言 GC 优化
      - 观察现象，初步定位原因对 Feature 服务早高峰毛刺时的 Trace 图进行耗时分析后发现，在毛刺期间程序 GC pause 时间（GC 周期与任务生命周期重叠的总和）长达近 50+ms（见左图），绝大多数 goroutine 在 GC 时进行了长时间的辅助标记（mark assist，见右图中浅绿色部分），GC 问题严重，因此怀疑耗时毛刺问题是由 GC 导致
      - 从原因出发，进行针对性分析
        - 根据观察计算模块服务平均每 10 秒发生 2 次 GC，GC 频率较低，但在每分钟前 10s 第一次与第二次的 GC 压力大小（做 mark assist 的 goroutine 数）呈明显差距，因此怀疑是在每分钟前 10s 进行第一次 GC 时的压力过高导致了耗时毛刺。
        - 根据 Golang GC 原理分析可知，G 被招募去做辅助标记是因为该 G 分配堆内存太快导致，而 计算模块每分钟缓存失效机制会导致大量的下游访问，从而引入更多的对象分配，两者结合互相印证了为何在每分钟前 10s 的第一次 GC 压力超乎寻常；
      - 按照分析结论，设计优化操作从减少对象分配数角度出发，对 Pprof heap 图进行观察
        - 在 inuse_objects 指标下 cache 库占用最大；
        - 在 alloc_objects 指标下 json 序列化占用最大；
      - 通过对业界开源的 [json 和 cache 库调研后](https://segmentfault.com/a/1190000041591284)，采用性能较好、低分配的 GJSON 和 0GC 的 BigCache 对原有库进行替换；
      - Golang GC
        - 在通俗意义上常认为，GO GC 触发时机为堆大小增长为上次 GC 两倍时。但在 GO GC 实际实践中会按照 [Pacer 调频算法](https://golang.design/under-the-hood/zh-cn/part2runtime/ch08gc/pacing/)根据堆增长速度、对象标记速度等因素进行预计算，使堆大小在达到两倍大小前提前发起 GC，最佳情况下会只占用 25% CPU 且在堆大小增长为两倍时，刚好完成 GC。
        - 但 Pacer 只能在稳态情况下控制 CPU 占用为 25%，一旦服务内部有瞬态情况，例如定时任务、缓存失效等等，Pacer 基于稳态的预判失效，导致 GC 标记速度小于分配速度，为达到 GC 回收目标（在堆大小到达两倍之前完成 GC），会导致大量 Goroutine 被招募去执行 Mark Assist 操作以协助回收工作，从而阻碍到 Goroutine 正常的工作执行。因此目前 GO GC 的 Marking 阶段对耗时影响时最为严重的。
- [The Tail At Scale论文解读](https://mp.weixin.qq.com/s/BKMPNix-zn64-0MP3XVLeA)
  - 讲述的是google内部的一些长尾耗时优化相关的经验
  - 服务耗时为什么会产生抖动
    - 服务A实例通过服务发现模块找到下游服务B上的实例，通过调度算法决定调用服务B上的具体实例的接口
    - 服务A实例调用服务B实例的耗时= 网络往返的时间+服务B实例执行请求的耗时
  - 网络因素的影响
    - 传输链路上的耗时差异
    - 数据排队
  - 服务实例本身对耗时的影响
    - 全局共享资源。服务内部可能会对一些全局的资源进行竞争。当竞争激烈的时候可能会存在线程饥饿的状态，长时间无法获得锁会导致请求耗时明显增大。
    - CPU过载。现代CPU会有保护自己的措施，当CPU过热的时候就会有降低执行指令的速度，从而达到保护CPU的作用。
    - GC。STW会停止所有正在工作的线程。
  - 组件的耗时抖动对集群的影响
    - 99分位耗时远比95分位高. 99分位过高，但是95分位表现是正常的，这样会给我们造成误判，认为服务状态并不健康，从而扩容我们的服务集群，虽然这对降低99分位是有帮助的（因为扩容的机器分摊了一部分流量），但是这样做的性价比并不高，因为大多数的请求处理情况是正常的。所以优化耗时抖动是必要的
  - 减少服务组件的耗时抖动
    - 服务等级分类和请求优先队列。一个服务会提供一个或者多个接口，可以定义接口的优先级，让优先级高的接口优先请求，优先级低或者对耗时不敏感的接口请求靠后执行。
    - 减少线头阻塞。在网络交换机里面，有输入端口，交换单元，输出端口，如果一个输入端口之中的数据要输出到多个输出端口，就需要排队，通过减少线头阻塞，可以降低网络传输的耗时。
    - 管理后台任务和请求并行化。对一些后台任务进行有效的管控，比如日志压缩，GC，可以在服务状态良好的时候进行。并且一些对下游的请求如果两者之间没有互相依赖，是可以并行执行的。
  - 请求维度耗时抖动优化
    - 对冲请求
      - 既然99分位耗时比95分位耗时大一倍，那么如果在请求等待响应时间已经大于95分位耗时的时候可以重发一个相同的请求，采用两个请求中首先返回的结果。在google的相关实践中，对冲请求带来的效果是很明显的。
    - 并行请求
  - 集群维度耗时抖动优化
    - 微分区。服务中的多个实例可以组成一个小型的分区，当分区中一个实例出现耗时抖动，可以往该实例中其他实例转移流量。比如实例A所在分区中有20个实例，当A出现耗时抖动，将流量转移到其他19个实例上。对于其他19个实例来说增加了大概5%的流量负载，却有效保证了分区内的耗时维持在一个较低的水位。
    - 分区状态探测与预测。在上面优化的前提下，可以探测每个分区的服务耗时情况以及预测出耗时抖动，及时的做流量的转移。
    - 实例监测与流量摘除。如果一个服务实力状态异常，可以把实例的流量摘掉，分摊到分区中别的实例上面去，这样做可以提高集群的整体健康状态。
- [优化redis写入而降低cpu使用率的一次经历](https://mp.weixin.qq.com/s/ntNGz6mjlWE7gb_ZBc5YeA)
  - 异步批量写入
    ```go
    type CounterCache struct {
       rwMu        sync.RWMutex
       redisClient redis.Cmdable
    
       countCache   map[string]int64
       hasUpdateExpire map[string]struct{}
    }
    
    func NewCounterCache(redisClient redis.Cmdable) *CounterCache {
       c := &CounterCache{
          redisClient: redisClient,
          countCache:    make(map[string]int64),
       }
       go c.startFlushTicker()
       return c
    }
    
    func (c *CounterCache) IncrBy(key string, value int64) int64 {
       val := c.incrCacheBy(key, value)
       redisCount, _ := c.redisClient.Get(key).Int64()
       return val + redisCount
    }
    
    func (c *CounterCache) incrCacheBy(key string, value int64) int64 {
       c.rwMu.Lock()
       defer c.rwMu.Unlock()
        
       count := c.countCache[key]
       count += value
       c.countCache[key] = count
       return count
    }
    
    func (c *CounterCache) Get(key string) (int64, error) {
       cacheVal := c.get(key)
       redisValue, err := c.redisClient.Get(key).Int64()
       if err != nil && err != redis.Nil {
          return cacheVal, err
       }
    
       return redisValue + cacheVal, nil
    }
    
    func (c *CounterCache) get(key string) int64 {
       c.rwMu.RLock()
       defer c.rwMu.RUnlock()
       return c.countCache[key]
    }
    
    func (c *CounterCache) startFlushTicker() {
       ticker := time.NewTicker(time.Second * 5)
       for {
          select {
          case <-ticker.C:
             c.flush()
          }
       }
    }
    
    func (c *CounterCache) flush() {
       var oldCountCache map[string]int64
       c.rwMu.Lock()
       oldCountCache = c.countCache
       c.countCache = make(map[string]int64)
       c.rwMu.Unlock()
    
       for key, value := range oldCountCache {
          c.redisClient.IncrBy(key, value)
           if _, ok := c.hasUpdateExpire[key]; !ok {
             err := c.redisClient.Expire(key, DefaultExpiration)
             if err == nil {
                 c.hasUpdateExpire[key] = struct{}{}
             }
          }
       }
    }
    ```
- [Go 应用的性能优化](https://xargin.com/go-perf-optimization/)
  - 优化的前置知识
    - 对于计算密集型的程序来说，优化的主要精力会放在 CPU 上，要知道 CPU 基本的流水线概念，知道怎么样在使用少的 CPU 资源的情况下，达到相同的计算目标。
    - 对于 IO 密集型的程序(后端服务一般都是 IO 密集型)来说，优化可以是降低程序的服务延迟，也可以是提升系统整体的吞吐量. IO 密集型应用主要与磁盘、内存、网络打交道
      - 要了解内存的多级存储结构：L1，L2，L3，主存
      - 要知道基本的文件系统读写 syscall，批量 syscall，数据同步 syscall。
      - 要熟悉项目中使用的网络协议，至少要对 TCP, HTTP 有所了解。
  - 优化越靠近应用层效果越好
    - [How I cut GTA Online loading times by 70%](https://nee.lv/2021/02/28/How-I-cut-GTA-Online-loading-times-by-70/)
  - 优化的工作流程
    - 建立评估指标，例如固定 QPS 压力下的延迟或内存占用，或模块在满足 SLA 前提下的极限 QPS
    - 通过自研、开源压测工具进行压测，直到模块无法满足预设性能要求:如大量超时，QPS 不达预期，OOM
    - 通过内置 profile 工具寻找性能瓶颈
    - 本地 benchmark 证明优化效果
    - 集成 patch 到业务模块，回到 2
  - 工具
    - pprof
      - memory profiler
        - 四个相应的指标
          - inuse_objects：当我们认为内存中的驻留对象过多时，就会关注该指标
          - inuse_space：当我们认为应用程序占据的 RSS 过大时，会关注该指标
          - alloc_objects：当应用曾经发生过历史上的大量内存分配行为导致 CPU 或内存使用大幅上升时，可能关注该指标
          - alloc_space：当应用历史上发生过内存使用大量上升时，会关注该指标
        - 网关类应用因为海量连接的关系，会导致进程消耗大量内存，所以我们经常看到相关的优化文章，主要就是降低应用的 inuse_space。
        - 当我们进行 GC 调优时，会同时关注应用分配的对象数、正在使用的对象数，以及 GC 的 CPU 占用的指标。
      - cpu profiler
        - CPU profiler 使用 setitimer 系统调用，操作系统会每秒 100 次向程序发送 SIGPROF 信号。在 Go 进程中会选择随机的信号执行 sigtrampgo 函数。该函数使用 sigprof 或 sigprofNonGo 来记录线程当前的栈。
        - Go 语言内置的 cpu profiler 是在性能领域比较常见的 On-CPU profiler，对于瓶颈主要在 CPU 消耗的应用，我们使用内置的 profiler 也就足够了
    - fgprof
      - 我们碰到的问题是应用的 CPU 使用不高，但接口的延迟却很大，那么就需要用上 Off-CPU profiler，遗憾的是官方的 profiler 并未提供该功能，我们需要借助社区的 fgprof。
      - fgprof 是启动了一个后台的 goroutine，每秒启动 99 次，调用 runtime.GoroutineProfile 来采集所有 gorooutine 的栈。
      - 但调用 GoroutineProfile 函数的开销并不低，如果线上系统的 goroutine 上万，每次采集 profile 都遍历上万个 goroutine 的成本实在是太高了。所以 fgprof 只适合在测试环境中使用
    - trace
      - 一般情况下我们是不需要使用 trace 来定位性能问题的，通过压测 + profile 就可以解决大部分问题，除非我们的问题与 runtime 本身的问题相关
      - 采集 trace 对系统的性能影响还是比较大的，即使我们只是开启 gctrace，把 gctrace 日志重定向到文件，对系统延迟也会有一定影响，因为 gctrace 的 print 是在 stw 期间来做的 
      - [gctrace引起runtime调度阻塞](https://xiaorui.cc/archives/6232)



