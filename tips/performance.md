
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
  - 频繁创建goroutine的场景下，资源复用，节省内存。（需要一定规模。一般场景下，效果不太明显。）

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

  golink（https://golang.org/cmd/compile/）使用格式：
  ```go
  //go:linkname FastRand runtime.fastrand
  func FastRand() uint32
  ```
  主要功能就是让编译器编译的时候，把当前符号指向到目标符号。上面的函数FastRand被指向到runtime.fastrand,runtime包生成的也是伪随机数，和math包不同的是，它的随机数生成使用的上下文是来自当前goroutine的，所以它不用加锁。正因如此，一些开源库选择直接使用runtime的随机数生成函数。

  另外，标准库中的`time.Now()`，这个库在会有两次系统调用runtime.walltime1和runtime.nanotime，分别获取时间戳和程序运行时间。大部分场景下，我们只需要时间戳，这时候就可以直接使用`runtime.walltime1`。
  
  系统调用在go里面相对来讲是比较重的。runtime会切换到g0栈中去执行这部分代码，time.Now方法在go<=1.16中有两次连续的系统调用
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

  在runtime中，函数行号和函数名称的获取分为两步：

  - runtime回溯goroutine栈，获取上层调用方函数的的程序计数器（pc）。
  - 根据pc，找到对应的funcInfo,然后返回行号名称。
  
  经过pprof分析。第二步性能占比最大，约60%。针对第一步，我们经过多次尝试，并没有找到有效的办法。但是第二步很明显，我们不需要每次都调用runtime函数去查找pc和函数信息的，我们可以把第一次的结果缓存起来，后面直接使用。这样，第二步约60%的消耗就可以去掉。

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

  runtime对网络io，以及定时器的管理，会放到自己维护的一个epoll里，具体可以参考runtime/netpool。在一些高并发的网络io中，有以下几个问题：

  - 需要维护大量的协程去处理读写事件。
  - 对连接的状态无感知，必须要等待read或者write返回错误才能知道对端状态，其余时间只能等待。
  - 原生的netpool只维护一个epoll，没有充分发挥多核优势。
  
  基于此，有很多项目用x/unix扩展包实现了自己的基于epoll的网络库，比如gnet, 还有字节跳动的netpoll。

  在我们的项目中，也有尝试过使用。最终我们还是觉得基于标准库的实现已经足够。理由如下：

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

    与 + 操作符相同，也是通过 runtime/string.go的concatstrings() 函数实现拼接，区别是它通常用于循环中往字符串末尾追加，每追加一次，生成一个新的字符串替代旧的，效率极低
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

  使用 map 来实现 Set，意味着我们只关心 key 的存在，其 value 值并不重要

  我们选择了以下常用的类型作为 value 进行测试：bool、int、interface{}、struct{}。
  - 从内存开销而言，struct{} 是最小的，反映在执行时间上也是最少的。由于 bool 类型仅占一个字节，它相较于空结构而言，相差的并不多。但是，如果使用 interface{} 类型，那差距就很明显了

  


