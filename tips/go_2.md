
- [unsafe 包](https://mp.weixin.qq.com/s/wdFdPv3Bdnhy5pc8KL6w6w)
  - unsafe 实现原理
    - Sizeof(x ArbitrayType)方法主要作用是用返回类型x所占据的字节数，但并不包含x所指向的内容的大小，与C语言标准库中的Sizeof()方法功能一样，比如在32位机器上，一个指针返回大小就是4字节。
    - Offsetof(x ArbitraryType)方法主要作用是返回结构体成员在内存中的位置离结构体起始处(结构体的第一个字段的偏移量都是0)的字节数，即偏移量，我们在注释中看一看到其入参必须是一个结构体，其返回值是一个常量。
    - Alignof(x ArbitratyType)的主要作用是返回一个类型的对齐值，也可以叫做对齐系数或者对齐倍数。对齐值是一个和内存对齐有关的值，合理的内存对齐可以提高内存读写的性能。一般对齐值是2^n，最大不会超过8(受内存对齐影响).获取对齐值还可以使用反射包的函数，也就是说：unsafe.Alignof(x)等价于reflect.TypeOf(x).Align()。
  - 三种指针类型
    - ***T**：普通类型指针类型，用于传递对象地址，不能进行指针运算。
    - **unsafe.poniter**：通用指针类型，用于转换不同类型的指针，不能进行指针运算，不能读取内存存储的值(需转换到某一类型的普通指针)
    - **uintptr**：用于指针运算，GC不把uintptr当指针，uintptr无法持有对象。uintptr类型的目标会被回收。
    
    三者关系就是：unsafe.Pointer是桥梁，可以让任意类型的指针实现相互转换，也可以将任意类型的指针转换为uintptr进行指针运算，也就说uintptr是用来与unsafe.Pointer打配合
  - Sizeof、Alignof、Offsetof三个函数的基本使用
    ```go
     // sizeof
     fmt.Println(unsafe.Sizeof(true))
     fmt.Println(unsafe.Sizeof(int8(0)))
     fmt.Println(unsafe.Sizeof(int16(10)))
     fmt.Println(unsafe.Sizeof(int(10)))
     fmt.Println(unsafe.Sizeof(int32(190)))
     fmt.Println(unsafe.Sizeof("asong"))
     fmt.Println(unsafe.Sizeof([]int{1,3,4}))
     // Offsetof
     user := User{Name: "Asong", Age: 23,Gender: true}
     userNamePointer := unsafe.Pointer(&user)
    
     nNamePointer := (*string)(unsafe.Pointer(userNamePointer))
     *nNamePointer = "Golang梦工厂"
    
     nAgePointer := (*uint32)(unsafe.Pointer(uintptr(userNamePointer) + unsafe.Offsetof(user.Age)))
     *nAgePointer = 25
    
     nGender := (*bool)(unsafe.Pointer(uintptr(userNamePointer)+unsafe.Offsetof(user.Gender)))
     *nGender = false
    
     fmt.Printf("u.Name: %s, u.Age: %d,  u.Gender: %v\n", user.Name, user.Age,user.Gender)
     // Alignof
     var f32 float32
     var s string
     var m map[string]string
     var p *int32
    
     fmt.Println(unsafe.Alignof(f32))
     fmt.Println(unsafe.Alignof(s))
     fmt.Println(unsafe.Alignof(m))
     fmt.Println(unsafe.Alignof(p))
    ```
  - 内存对齐
    - 对齐的作用和原因：CPU访问内存时，并不是逐个字节访问，而是以字长（word size)单位访问。比如32位的CPU，字长为4字节，那么CPU访问内存的单位也是4字节。这样设计可以减少CPU访问内存的次数，加大CPU访问内存的吞吐量。假设我们需要读取8个字节的数据，一次读取4个字节那么就只需读取2次就可以。内存对齐对实现变量的原子性操作也是有好处的，每次内存访问都是原子的，如果变量的大小不超过字长，那么内存对齐后，对该变量的访问就是原子的，这个特性在并发场景下至关重要。
    
- Splitting a go array or slice in a defined number of chunks.
  ```go
  func SplitSlice(array []int, numberOfChunks int) [][]int {
      if len(array) == 0 {
          return nil
      }
      if numberOfChunks <= 0 {
          return nil
      }
  
      if numberOfChunks == 1 {
          return [][]int{array}
      }
  
      result := make([][]int, numberOfChunks)
  
      // we have more splits than elements in the input array.
      if numberOfChunks > len(array) {
          for i := 0; i < len(array); i++ {
              result[i] = []int{array[i]}
          }
          return result
      }
  
      for i := 0; i < numberOfChunks; i++ {
  
          min := (i * len(array) / numberOfChunks)
          max := ((i + 1) * len(array)) / numberOfChunks
  
          result[i] = array[min:max]
  
      }
  
      return result
  }
  ```
- [如何保留 Go 程序崩溃现场](https://mp.weixin.qq.com/s/RktnMydDtOZFwEFLLYzlCA)
  - core dump
    - 可以使用`ulimit -c [size]`命令指定记录 core dump 文件的大小
    - GOTRACEBACK `GOTRACEBACK=system go run main.go`
      - none，不显示任何 goroutine 堆栈信息
      - single，默认级别，显示当前 goroutine 堆栈信息
      - all，显示所有 user （不包括 runtime）创建的 goroutine 堆栈信息
      - system，显示所有 user + runtime 创建的 goroutine 堆栈信息
      - crash，和 system 打印一致，但会生成 core dump 文件（Unix 系统上，崩溃会引发 SIGABRT 以触发core dump）
        如果想获取 core dump 文件，那么就应该把 GOTRACEBACK 的值设置为 crash 。当然，我们还可以通过 runtime/debug 包中的 SetTraceback 方法来设置堆栈打印级别
    - dlv core 命令来调试 core dump
      ```shell
      go get -u github.com/go-delve/delve/cmd/dlv
      ```
      - 通过 dlv 调试器来调试 core 文件，执行命令格式 dlv core 可执行文件名 core文件
      ```shell
      (dlv) goroutines
      (dlv) goroutine 1
      (dlv) bt
      (dlv) frame 5
      ```
- [Make synchronous code asynchronous with context.Context and channels](https://pauldigian.hashnode.dev/advanced-go-make-synchronous-code-asynchronous-with-contextcontext-and-channels)
  - The main point of this pattern is to wait, at the same time, for either the result of the blocking function, or for a context cancellation. If the context is cancelled, the result from the blocking function is not needed anymore, and we can gracefully and quickly shut down the code path and free valuable resources. If the result comes in time, we can move on with the default execution.
    ```go
    ctx := request.Context()
    
    r := io.SomeReader() // not important which reader
    b := make([]byte, 1024)
    type readResult struct {
        readBytes int
        err             error
    }
    resultCh := make(chan readResult)
    
    go func() {
        if err := ctx.Err(); err != nil {
            return
        }
        n, err := r.Read(b)
        select {
            case <-ctx.Done():
                // do nothing
            case resultCh <- readResult{readBytes: n, err: err}:
                // do nothing
        }
    }()
    
    var result readResult
    select {
        case <-ctx.Done():
            // do something in here
            // likely return from the function with an error
            return ctx.Err()
        case result <- resultCh:
            // great we go our result in time and we can move on
            // from now on, `b` is populated, with the read bytes
    }
    
    // here we can use `result` as `readResult` and the buffer `b`
    
    ```
- [Golang GC Marker](https://mp.weixin.qq.com/s/n-4YxL_irIBqd2fxszmDeg)
  - 每个P都有  Mark worker
  - 三色标记
  - 混合写屏障： 在栈外设置 加入写屏障 + 删除写屏障

- [map](https://mp.weixin.qq.com/s/SGv5vuh9aU2mqViC4Kj0YQ)
  - hmap 由很多 bmap（bucket） 构成，每个 bmap 都保存了 8 个 key/value 对
    ![img.png](go_map.png)
  - 我们仔细看 mapextra 结构体里对 overflow 字段的注释. map 的 key 和 value 都不包含指针的话，在 GC 期间就可以避免对它的扫描。在 map 非常大（几百万个 key）的场景下，能提升不少性能
  - bmap 这个结构体里有一个 overflow 指针，它指向溢出的 bucket。因为它是一个指针，所以 GC 的时候肯定要扫描它，也就要扫描所有的 bmap。
  - 而当 map 的 key/value 都是非指针类型的话，扫描是可以避免的，直接标记整个 map 的颜色（三色标记法）就行了，不用去扫描每个 bmap 的 overflow 指针
  - 于是就利用 hmap 里的 extra 结构体的 overflow 指针来 “hold” 这些 overflow 的 bucket，并把 bmap 结构体的 overflow 指针类型变成一个 unitptr 类型（这些是在编译期干的）。于是整个 bmap 就完全没有指针了，也就不会在 GC 期间被扫描
  - 当我们知道上面这些原理后，就可以利用它来对一些场景进行性能优化：
    `map[string]int -> map[[12]byte]int`
    因为 string 底层有指针，所以当 string 作为 map 的 key 时，GC 阶段会扫描整个 map；而数组 [12]byte 是一个值类型，不会被 GC 扫描。
  - Go语言使用 map 时尽量不要在 big map 中保存指针
  - map 的 key 和 value 要不要在 GC 里扫描，和类型是有关的。数组类型是个值类型，string 底层也是指针。
  - 不过要注意，key/value 大于 128B 的时候，会退化成指针类型。 那么问题来了，什么是指针类型呢？**所有显式 *T 以及内部有 pointer 的对像都是指针类型。
- [channel](https://mp.weixin.qq.com/s?__biz=MzAxMTA4Njc0OQ==&mid=2651445085&idx=3&sn=2aecb5560dec2c0128ddc7cc3403a5a5&chksm=80bb09afb7cc80b97c989d35c925350121d6164c5dd65eb5bef59aebc811f95614d41c4314fc&scene=21#wechat_redirect)
  - 基本特性
    - 双向和单向；三种表现方式，分别是：声明双向通道：`chan T`、声明只允许发送的通道：`chan <- T`、声明只允许接收的通道：`<- chan T`
    - channel 中还分为 “无缓冲 channel” 和 “缓冲 channel”
      - 无缓冲的 channel（unbuffered channel），其缓冲区大小则默认为 0。在功能上其接受者会阻塞等待并阻塞应用程序，直至收到通信和接收到数据
      - 有缓存的 channel（buffered channel），其缓存区大小是根据所设置的值来调整。在功能上，若缓冲区未满则不会阻塞，会源源不断的进行传输。当缓冲区满了后，发送者就会阻塞并等待。而当缓冲区为空时，接受者就会阻塞并等待，直至有新的数据
  - 基本原理
    - channel 是一个有锁的环形队列
      ![img.png](go_channel.png)
      - dataqsiz：循环队列的长度。
      - buf：指向长度为 dataqsiz 的底层数组，仅有当 channel 为缓冲型的才有意义
      - sendx：已发送元素在循环队列中的索引位置。
      - recvx：已接收元素在循环队列中的索引位置。
      - recvq：接受者的 sudog 等待队列（缓冲区不足时阻塞等待的 goroutine）。
      - sendq：发送者的 sudog 等待队列。
      - sudog 是 Go 语言中用于存放协程状态为阻塞的 goroutine 的双向链表抽象，你可以直接理解为一个正在等待的 goroutine 就可以了
    - 发送
      - 使用 ch <- i 表达式向 Channel 发送数据时遇到的几种情况：
        - 如果当前 Channel 的 recvq 上存在已经被阻塞的 Goroutine，那么会直接将数据发送给当前的 Goroutine 并将其设置成下一个运行的协程； 
        - 如果 Channel 存在缓冲区并且其中还有空闲的容量，我们就会直接将数据直接存储到当前缓冲区 sendx 所在的位置上； 
        - 如果都不满足上面的两种情况，就会创建一个 sudog 结构并加入 Channel 的 sendq 队列，同时当前的 Goroutine 就会陷入阻塞等待其他的协程向 Channel 中发送数据以被唤醒；
      - 发送数据的过程中包含几个会触发 Goroutine 调度的时机，首先是发送数据时发现 Channel 上存在等待接收数据的 Goroutine，这是会立刻设置处理器的 runnext 属性，但是并不会立刻触发调度，第二个时机是发送数据时并没有找到接收方并且缓冲区已经满了，这时就会将自己加入 Channel 的 sendq 队列并立刻调用 goparkunlock 触发 Goroutine 的调度让出处理器的使用权。
    - 接收
      - 从 Channel 中接收数据时的几种情况：
        - 如果 Channel 是空的，那么就会直接调用 gopark 挂起当前的 Goroutine；
        - 如果 Channel 已经关闭并且缓冲区没有任何数据，chanrecv 函数就会直接返回；
        - 如果 Channel 上的 sendq 队列中存在挂起的 Goroutine，就会将recvx 索引所在的数据拷贝到接收变量所在的内存空间上并将 sendq 队列中 Goroutine 的数据拷贝到缓冲区中；
        - 如果 Channel 的缓冲区中包含数据就会直接从 recvx 所在的索引上进行读取；
        - 在默认情况下会直接挂起当前的 Goroutine，将 sudog 结构加入 recvq 队列并等待调度器的唤醒；
- [如何高效的进行字符串拼接](https://mp.weixin.qq.com/s/TDU_lnxbQmlnGosN4Ss7Tw)
  - string类型本质上就是一个byte类型的数组
    ```go
    //go:nosplit
    func gostringnocopy(str *byte) string {
     ss := stringStruct{str: unsafe.Pointer(str), len: findnull(str)}
     s := *(*string)(unsafe.Pointer(&ss))
     return s
    }
    ```
  - 字符串拼接的6种方式
    - 原生拼接方式"+"
    - 字符串格式化函数fmt.Sprintf
    - Strings.builder - 提供的String方法就是将[]]byte转换为string类型，这里为了避免内存拷贝的问题
    - bytes.Buffer 
    - strings.join - 基于strings.builder来实现的, join方法内调用了b.Grow(n)方法，这个是进行初步的容量分配
    - 切片append 
  - benchmark 比较
    - 当进行少量字符串拼接时，直接使用+操作符进行拼接字符串，效率还是挺高的
    - 当要拼接的字符串数量上来时，+操作符的性能就比较低了；函数fmt.Sprintf还是不适合进行字符串拼接，无论拼接字符串数量多少，性能损耗都很大
    - strings.Builder无论是少量字符串的拼接还是大量的字符串拼接，性能一直都能稳定
    - strings.join方法的benchmark就可以发现，因为使用了grow方法，提前分配好内存，在字符串拼接的过程中，不需要进行字符串的拷贝，也不需要分配新的内存
    - bytes.Buffer方法性能是低于strings.builder的，bytes.Buffer 转化为字符串时重新申请了一块空间，存放生成的字符串变量
    - strings.join ≈ strings.builder > bytes.buffer > []byte转换string > "+" > fmt.sprintf

- [Go Ballast 让内存控制更加丝滑](https://mp.weixin.qq.com/s/SlQkv74hXZzZEdUAhTobuw)
  - GO 的 GC 是标记-清除方式，当 GC 会触发时全量遍历变量进行标记，当标记结束后执行清除，把标记为白色的对象执行垃圾回收。值得注意的是，这里的回收仅仅是标记内存可以返回给操作系统，并不是立即回收，这就是你看到 Go 应用 RSS 一直居高不下的原因。在整个垃圾回收过程中会暂停整个 Go 程序（STW），Go 垃圾回收的耗时还是主要取决于标记花费的时间的长短，清除过程是非常快的。
  - Go GC 优化的手段你知道的有哪些？
    - 设置 GOGC
      - 设置 GOGC 的弊端
      - GOGC 设置比率的方式不精确 - 我们很难精确的控制我们想要的触发的垃圾回收的阈值
      - GOGC 设置的非常小，会频繁触发 GC 导致太多无效的 CPU 浪费
      - 对某些程序本身占用内存就低，容易触发 GC - 对 API 接口耗时比较敏感的业务，如果  GOGC 置默认值的时候，也可能也会遇到接口的周期性的耗时波动
      - GOGC 设置很大，有的时候又容易触发 OOM
    - 设置 debug.SetGCPercent()
      
    这两种方式的原理和效果都是一样的，GOGC 默认值是 100，也就是下次 GC 触发的 heap 的大小是这次 GC 之后的 heap 的一倍
  - [GO 内存 ballast](https://blog.twitch.tv/en/2019/04/10/go-memory-ballast-how-i-learnt-to-stop-worrying-and-love-the-heap/) [issue 23044](https://github.com/golang/go/issues/23044)
    - 什么是 Go ballast，其实很简单就是初始化一个生命周期贯穿整个 Go 应用生命周期的超大 slice。
      ```go
      func main() {
        ballast := make([]byte, 10*1024*1024*1024) // 10G
        // do something
        runtime.KeepAlive(ballast)
      }
      ```
      上面的代码就初始化了一个 ballast，利用 runtime.KeepAlive 来保证 ballast 不会被 GC 给回收掉。
      利用这个特性，就能保证 GC 在 10G 的一倍时才能被触发，这样就能够比较精准控制 GO GC 的触发时机
    - 这里初始化一个 10G 的数组，不就占用了 10 G 的物理内存呢？ 答案其实是不会的。
       ```go
       func main() {
           ballast := make([]byte, 10*1024*1024*1024)
       
           <-time.After(time.Duration(math.MaxInt64))
           runtime.KeepAlive(ballast)
       }
       
       $ ps -eo pmem,comm,pid,maj_flt,min_flt,rss,vsz --sort -rss | numfmt --header --to=iec --field 5 | numfmt --header --from-unit=1024 --to=iec --field 6 | column -t | egrep "[t]est|[P]I"
       
       ```
    - 当怀疑我们的接口的耗时是由于 GC 的频繁触发引起的，我们需要怎么确定呢？
      - 首先你会想到周期性的抓取 pprof 的来分析，这种方案其实也可以，但是太麻烦了。
      - 其实可以根据 GC 的触发时间绘制这个曲线图，GC 的触发时间可以利用 runtime.Memstats 的 LastGC 来获取。
- [runtime.KeepAlive 有什么用](https://mp.weixin.qq.com/s/1KlMbvnflFwQS2e-SFG-IA)
  - 有些同学喜欢利用 runtime.SetFinalizer 模拟析构函数，当变量被回收时，执行一些回收操作，加速一些资源的释放。在做性能优化的时候这样做确实有一定的效果，不过这样做是有一定的风险的。
    ```go
    type File struct { d int }
    
    func main() {
        p := openFile("t.txt")
        content := readFile(p.d)
    
        println("Here is the content: "+content)
    }
    
    func openFile(path string) *File {
        d, err := syscall.Open(path, syscall.O_RDONLY, 0)
        if err != nil {
          panic(err)
        }
    
        p := &File{d}
        runtime.SetFinalizer(p, func(p *File) {
          syscall.Close(p.d)
        })
    
        return p
    }
    
    func readFile(descriptor int) string {
        doSomeAllocation()
    
        var buf [1000]byte
        _, err := syscall.Read(descriptor, buf[:])
        if err != nil {
          panic(err)
        }
    
        return string(buf[:])
    }
    
    func doSomeAllocation() {
        var a *int
    
        // memory increase to force the GC
        for i:= 0; i < 10000000; i++ {
          i := 1
          a = &i
        }
    
        _ = a
    }
    ```
    - doSomeAllocation 会强制执行 GC，当我们执行这段代码时会出现下面的错误。
    - 因为 syscall.Open 产生的文件描述符比较特殊，是个 int 类型，当以值拷贝的方式在函数间传递时，并不会让 File.d 产生引用关系，于是 GC 发生时就会调用 runtime.SetFinalizer(p, func(p *File) 导致文件描述符被 close 掉
  - 什么是 runtime.KeepAlive
    - 我们如果才能让文件描述符不被 gc 给释放掉呢？其实很简单，只需要调用 runtime.KeepAlive 即可
      ```go
      func main() {
          p := openFile("t.txt")
          content := readFile(p.d)
          
          runtime.KeepAlive(p)
      
          println("Here is the content: "+content)
      }
      ```
      runtime.KeepAlive 能阻止 runtime.SetFinalizer 延迟发生，保证我们的变量不被 GC 所回收
- [Goroutine 泄漏检查器](https://mp.weixin.qq.com/s/eSa6B1Z1cnpUJ1Vn3bxhUA)
  - 具有监控存活的 goroutine 数量功能的 APM (Application Performance Monitoring) 应用程序性能监控可以轻松查出 goroutine 泄漏。例如 NewRelic APM 中 goroutine 的监控
  - [goroutine 泄漏检测器](https://github.com/uber-go/goleak)
    ```go
    func leak() error {
     go func() {
      time.Sleep(time.Minute)
     }()
    
     return nil
    }
    
    func TestLeakFunction(t *testing.T) {
      defer goleak.VerifyNone(t)
    
      if err := leak(); err != nil {
        t.Fatal("error not expected")
      }
    }
    ```
    从报错信息中我们可以提取出两个有用的信息：
    - 报错信息顶部为泄漏的 goroutine 的堆栈信息，以及 goroutine 的状态，可以帮我们快速调试并了解泄漏的 goroutine
    - 之后为 goroutineID，在使用 trace 可视化的时候很有用，以下是通过 go test -trace trace.out 生成的用例截图：
  - 运行原理
    - goleak 检测了所有的 goroutine 而不是只检测泄漏的 goroutine
    - goroutine 的堆栈信息由 golang 标准库中的 runtime.Stack，它可以被任何人取到。不过，[Goroutine 的 ID 是拿不到的](https://groups.google.com/forum/#!topic/golang-nuts/0HGyCOrhuuI)
    - 之后，goleak 解析所有的 goroutine 出并通过以下规则过滤 go 标准库中产生的 goroutine
      - 由 go test 创建来运行测试逻辑的 goroutine。
      - 由 runtime 创建的 goroutine，例如监听信号接收的 goroutine。想要了解更多相关信息，请参阅Go: [gsignal, Master of goroutine](https://medium.com/a-journey-with-go/go-gsignal-master-of-signals-329f7ff39391)
      - 当前运行的 goroutine
    - 经过此次过滤后，如果没有剩余的 goroutine，则表示没有发生泄漏。但是 goleak 还是存在一下缺陷：
      - 三方库或者运行在后台中，遗漏的 goroutine 将会造成虚假的结果(无 goroutine 泄漏)
      - 如果在其他未使用 goleak 的测试代码中使用了 goroutine，那么泄漏结果也是错误的。如果这个 goroutine 一直运行到下次使用 goleak 的代码， 则结果也会被这个 goroutine 影响，发生错误。
- [Recover](https://mp.weixin.qq.com/s/y6bLqjevvqlP3AEjTaztYw)
  - 多帧情况
    ```go
    func level1() {
        defer fmt.Println("defer func 3")
        defer func() {
            if err := recover(); err != nil {
                fmt.Println("recovering...")
            }
        }()
        defer fmt.Println("defer 2")
    
        level2()
    }
    
    func level2() {
        defer fmt.Println("defer func 4")
        panic("level2")
    }
    
    func main() {
        level1()
    ```
    - 由于一个函数 recover 了 panic，Go 需要一种跟踪，并恢复这个程序的方法。为了达到这个目的，每一个 Goroutine 嵌入了一个特殊的属性，指向一个代表该 panic 的对象
    - 当 panic 发生的时候，该对象会在运行 defer 函数前被创建。然后，recover 这个 panic 的函数仅仅返回这个对象的信息，同时将这个 panic 标记为已恢复（recovered
    - 一旦 panic 被认为已经恢复，Go 需要恢复当前的工作。但是，由于运行时处于 defer 函数的帧中，它不知道恢复到哪里。出于这个原因，当 panic 标记已恢复的时候，Go 保存当前的程序计数器和当前帧的堆栈指针，以便 panic 发生后恢复该函数
    - 我们也可以使用 objdump 查看 程序计数器的指向e.g. `objdump -D my-binary | grep 105acef`
    - 该指令指向函数调用 runtime.deferreturn，这个指令被编译器插入到每个函数的末尾，而它运行 defer 函数。在前面的例子中，这些 defer 函数中的大多数已经运行了——直到恢复，因此，只有剩下的那些会在调用者返回前运行
  - goexit
    - 函数 runtime.Goexit 使用完全相同的工作流程。runtime.Goexit 实际上创造了一个 panic 对象，且有着一个特殊标记来让它与真正的 panic 区别开来。这个标记让运行时可以跳过恢复以及适当的退出，而不是直接停止程序的运行
- [怎么让goroutine跑一半就退出](https://mp.weixin.qq.com/s/KBDXzcPLXFovnuY6WSIdoA)
  - Ans: 插入一个 `runtime.Goexit()`， 协程就会直接结束。并且结束前还能执行到defer 函数
  - runtime.Goexit()是什么
    - 从代码上看，runtime.Goexit()会先执行一下defer里的方法，这里就解释了开头的代码里为什么在defer里的打印2能正常输出
    - 然后代码再执行goexit1。本质就是对goexit0的简单封装
    - goexit0 做的事情就是将当前的协程G置为_Gdead状态，然后把它从M上摘下来，尝试放回到P的本地队列中。然后重新调度一波，获取另一个能跑的G，拿出来跑。
    - 总结一下，只要执行 goexit 这个函数，当前协程就会退出，同时还能调度下一个可执行的协程出来跑
  - goexit的用途
    - 每个堆栈底部都是这个方法
    - main函数也是个协程，栈底也是goexit
      - main函数也是由newproc创建的，只要通过newproc创建的goroutine，栈底就会有一个goexit
  - os.Exit() 指的是整个进程退出；而runtime.Goexit()指的是协程退出。
- [gowatch监听文件变动](https://mp.weixin.qq.com/s/f3q5ryWvLonOKMOZEJnXNQ)
  - 在linux内核中，有一种用于通知用户空间程序文件系统变化的机制—Inotify
  - Golang的标准库syscall实现了该机制 [ref](github.com/silenceper/gowatch)
  ![img.png](go_watcher.png)
- [用 kqueue 实现一个简单的 TCP Server](https://dev.to/frosnerd/writing-a-simple-tcp-server-using-kqueue-cah)
- [Golang 程序启动过程](https://juejin.cn/post/7035633561805783070)
- [Go: gsignal, Master of Signals](https://medium.com/a-journey-with-go/go-gsignal-master-of-signals-329f7ff39391)
  - Each `os.Signal` channel listens to their own set of events. Here is a diagram with the subscription workflow of the previous example:
  ![img.png](go_signal.png)
  - Go also gives the ability for a channel to stop being notified — function `Stop(os.Signal) `— or to ignore signals — function `Ignore(...os.Signal)`
  - gsignal
    - During the initialization phase, the signal spawns a goroutine that runs in a loop and act as a consumer to process the signals.
    - Then, when a signal reaches the program, the signal handler delegates it to a special goroutine called gsignal
    - Each thread (represented by M) has an internal gsignal goroutine to handle the signals. 
    - gsignal analyzes the signal to check if it processable, and wakes up the sleeping goroutine along with sending the signal to the queue
- [Go 服务中 HTTP 请求的生命周期](https://mp.weixin.qq.com/s/8j-hzmxs9NlaPDptldHJDg)
  - ListenAndServe 监听给定地址的 TCP 端口，之后循环接受新的连接。对于每一个新连接，它都会调度一个 goroutine 来处理这个连接（稍后详细说明）。处理连接涉及一个这样的循环：
    - 从连接中解析 HTTP 请求；产生 http.Request
    - 将这个 http.Request 传递给用户定义的 handler
  - http.ServeMux 是一个实现了 http.Handler 接口的类
    - ServeMux 维护了一个（根据长度）排序的 {pattern, handler} 的切片。
    - Handle 或 HandleFunc 向该切片增加新的 handler。
    - ServeHTTP：
      - （通过查找这个排序好的 handler 对的切片）为请求的 path 找到对应的 handler
      - 调用 handler 的 ServeHTTP 方法
    - mux 可以被看做是一个转发 handler；这种模式在 HTTP 服务开发中极为常见，这就是中间件。
  - 中间件只是另一个 HTTP handler
    - 它包裹了一个其他的 handler。中间件 handler 通过调用 ListenAndServe 被注册进来
    - 当调用的时候，它可以执行任意的预处理，调用自身包裹的 handler 然后可以执行任意的后置处理
    - 中间件最大的优点是可以组合。被中间件所包裹“用户 handler” 也可以是另一个中间件，依次类推。这是一个互相包裹的 http.Handler 链
      ```go
      func politeGreeting(w http.ResponseWriter, req *http.Request) {
       fmt.Fprintf(w, "Welcome! Thanks for visiting!\n")
      }
      
      func loggingMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, req)
        log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
       })
      }
      
      func main() {
       lm := loggingMiddleware(http.HandlerFunc(politeGreeting))
       log.Fatal(http.ListenAndServe(":8090", lm))
      }
      ```
      loggingMiddleware 利用 http.HandlerFunc 和闭包使代码更加简洁，同时保留了相同的功能。更重要的是这个例子展示了中间件事实上的标准签名：一个函数传入一个 http.Handler，有时还有其他状态，之后返回一个不同的 http.Handler。返回的 handler 现在应该替换掉传入中间件的那个 handler，之后会“神奇地”执行它原有的功能，并且与中间件的功能包装在一起
  - 并发和 panic 处理
    - 每个连接由 http.Server.Serve 在一个新的 goroutine 中处理。
    - net/http 包（在 conn.serve 方法中）内置对每个服务 goroutine 有 recovery
- [Go error 处理最佳实践](https://mp.weixin.qq.com/s/XojOIIZfKm_wXul9eSU1tQ)
  - https://github.com/pkg/errors
    - Wrap 封装底层 error, 增加更多消息，提供调用栈信息，这是原生 error 缺少的
    - WithMessage 封装底层 error, 增加更多消息，但不提供调用栈信息
    - Cause 返回最底层的 error, 剥去层层的 wrap
  - errors.Is 会递归的 Unwrap err
    - Is 是做的指针地址判断，如果错误 Error() 内容一样，但是根 error 是不同实例，那么 Is 判断也是 false, 这点就很扯
  - 官方库如何生成一个 wrapper error
    - fmt.Errorf 格式化时使用 %w
  - golang.org/x/sync/errgroup
    - 适用如下场景：并发场景下，如果一个 goroutine 有错误，那么就要提前返回，并取消其它并行的请求
  - 线上实践注意的几个问题
    - 所有异步的 goroutine 都要用 recover 去兜底处理
    - 数据传输和退出控制，需要用单独的 channel 不能混, 我们一般用 context 取消异步 goroutine, 而不是直接 close channels
    - error 级联使用问题。 如果复用 err 变量的情况下， Call2 返回的 error 是自定义类型，此时 err 类型是不一样的，导致经典的 error is not nil, but value is nil
      ```go
      type myError struct {
       string
      }
      
      func (i *myError) Error() string {
       return i.string
      }
      
      func Call1() error {
       return nil
      }
      
      func Call2() *myError {
       return nil
      }
      
      func main() {
       err := Call1()
       if err != nil {
        fmt.Printf("call1 is not nil: %v\n", err)
       }
      
       err = Call2()
       if err != nil {
        fmt.Printf("call2 err is not nil: %v\n", err)
       }
      }
      ```
    - 并发问题
      ```go
      var FIRST error = errors.New("Test error")
      var SECOND error = nil
      
      func main() {
          var err error
          go func() {
              i := 1
              for {
                  i = 1 - i
                  if i == 0 {
                      err = FIRST
                  } else {
                      err = SECOND
                  }
                  time.Sleep(10)
              }
          }()
          for {
              if err != nil {
                  fmt.Println(err.Error())
              }
              time.Sleep(10)
          }
      ```
      go 内置类型除了 channel 大部分都是非线程安全的，error 也不例外
    - 官方库无法 wrap 调用栈，所以 fmt.Errorf %w 不如 pkg/errors 库实用，但是errors.Wrap 最好保证只调用一次，否则全是重复的调用栈
    - 如果 err 为 nil 的时候，也会返回 nil. 所以 Wrap 前最好做下判断
- 切片拷贝
  - `=、[:]`是浅拷贝，`copy()`拷贝是深拷贝
- [Go语言也有隐式转型](https://mp.weixin.qq.com/s/NCM-RrzxYiAUlAAYshdAaQ)
  - Question
    ```go
    type MyInt int
    type MyMap map[string]int
    
    func main() {
        var x MyInt
        var y int 
        x = y     // 会报错: cannot use y (type int) as type MyInt in assignment
        _ = x 
    
        var m1 MyMap
        var m2 map[string]int
        m1 = m2 // 不会报错
        m2 = m1 // 不会报错
    }
    ```
  - Deep dive
    - `type T1 int` 使用上述类型声明语句定义的类型T1、T2被称为defined type (named type)
      - 所有数值类型都是defined type；(这里面就包含int)
      - 字符串类型string是defined type；
      - 布尔类型bool是defined type。
      - map、数组、切片、结构体、channel等原生复合类型(composite type)都不是defined type
    - Assignability
      - x's type V and T have identical underlying types and at least one of V or T is not a defined type.
      - 它和Go的[无类型常量隐式转型]()类似
        ```go
        type MyInt int
        const a = 1234
        var n MyInt = a
        ```
- [Zero down restart and deploy](https://bunrouter.uptrace.dev/guide/go-zero-downtime-restarts.html#systemd-socket)
- [新 IP 包的设计思路](https://mp.weixin.qq.com/s/VxJvLRTt3zoGRTzAhZTUsQ) [Source](https://tailscale.com/blog/netaddr-new-ip-type-for-go/)
  - 标准库 net.IP 的问题
    - 可变的。net.IP 的底层类型是 []byte，它的定义是：type IP []byte，这意味着你可以随意修改它。不可变数据结构更安全、更简单。
    - 不可比较的。因为 Go 中 slice 类型是不可比较的，也就是说 net.IP 不支持 ==，也不能作为 map 的 key。
    - 有两个 IP 地址类型，net.IP 表示基本的 IPv4 或 IPv6 地址，而 net.IPAddr 表示支持 zone scopes 的 IPv6。因为有两个类型，使用时就存在选择问题，到底使用哪个。标准库存在两个这样的方法：Resolver.LookupIP vs Resolver.LookupIPAddr。（关于什么是 IPv6 zone scopes 见维基百科：https://en.wikipedia.org/wiki/IPv6_address#Scoped_literal_IPv6_addresses_(with_zone_index 。）
    - 太大。在 Go 中，64 位机器上，slice 类型占 24 个字节，这只是 slice header。因此，net.IP 的大小实际包含两部分：24 字节的 slice header 和 4 或 6 字节的 IP 地址。而 net.IPAddr 更有额外的字符串类型 Zone 字段，占用空间更多。
    - 不是 allocates free 的，会增加 GC 的工作。当你调用 net.ParseIP 或接收一个 UDP 包时，它为了记录 IP 地址会分配底层数组的内存，然后指针放入 net.IP 的 slice header 中。
    - 当解析一个字符串形式的 IP 地址时，net.IP 无法区分 IPv4 映射的 IPv6 地址[2]和 IPv4 地址。因为 net.IP 不会记录原始的地址族（address family）。见 issue 37921
- [这些 //go: 指令](https://mp.weixin.qq.com/s/KK_rWHqTTy4zzqG96RKbsQ)
  - go:linkname
    ```go
    //go:linkname localname importpath.name
    ```
    该指令指示编译器使用 importpath.name 作为源代码中声明为 localname 的变量或函数的目标文件符号名称。但是由于这个伪指令，可以破坏类型系统和包模块化，只有引用了 unsafe 包才可以使用。
  - go:noescape - 该指令指定下一个有声明但没有主体（意味着实现有可能不是 Go）的函数，不允许编译器对其做逃逸分析。
  - go:nosplit - 该指令指定文件中声明的下一个函数不得包含堆栈溢出检查。
  - go:nowritebarrierrec - 该指令表示编译器遇到写屏障时就会产生一个错误，并且允许递归。也就是这个函数调用的其他函数如果有写屏障也会报错。
  - go:yeswritebarrierrec
  - go:noinline
  - go:norace - 该指令表示禁止进行竞态检测。
  - go:notinheap - 该指令常用于类型声明，它表示这个类型不允许从 GC 堆上进行申请内存
- [Go 语言类型可比性](https://mp.weixin.qq.com/s/_AYOAtNhPGZy4ttfsDDw8w)
  - 那哪些类型是可比较的呢
    - Boolean（布尔值）、Integer（整型）、Floating-point（浮点数）、Complex（复数）、String（字符）这些类型是毫无疑问可以比较的。
    - Poniter (指针) 可以比较：如果两个指针指向同一个变量，或者两个指针类型相同且值都为 nil，则它们相等。注意，指向不同的零大小变量的指针可能相等，也可能不相等
    - Channel （通道）具有可比性
    - Interface （接口值）具有可比性
  - 哪些类型是不可比较的
    - slice、map、function 这些是不可以比较的，但是也有特殊情况，那就是当他们值是 nil 时，可以与 nil 进行比较。
  - 如果我们的变量中包含不可比较类型，或者 interface 类型（它的动态类型可能存在不可比较的情况），那么我们直接运用比较运算符 == ，会引发程序错误。此时应该选用 reflect.DeepEqual 函数（当然也有特殊情况，例如 []byte，可以通过 bytes. Equal 函数进行比较）。
    ```go
    type Data struct {
     UUID    string
     Content interface{}
    }
    var x, y Data
    x = Data{
    UUID:    "856f5555806443e98b7ed04c5a9d6a9a",
    Content: 1,
    }
    bytes, _ := json.Marshal(x)
    _ = json.Unmarshal(bytes, &y)
    fmt.Println(x)  // {856f5555806443e98b7ed04c5a9d6a9a 1}
    fmt.Println(y)  // {856f5555806443e98b7ed04c5a9d6a9a 1}
    fmt.Println(reflect.DeepEqual(x, y)) // false ???
    
    ```
    原来此 1 非彼 1，Content 字段的数据类型由 int 转换为了 float64 。而在接口中，其动态类型不一致时，它的比较是不相等的。
- [Go 为什么不支持可重入锁](https://mp.weixin.qq.com/s/pQBsAxnaBXkk7G1cdUgsww)
  - 可重入锁
    - 在加锁上：如果是可重入互斥锁，当前尝试加锁的线程如果就是持有该锁的线程时，加锁操作就会成功。
    - 在解锁上：可重入互斥锁一般都会记录被加锁的次数，只有执行相同次数的解锁操作才会真正解锁。
    - Java
      - Java 的 ReentrantReadWriteLock 支持锁降级，但不能升级，即获取了写锁的线程，可以继续获取读锁，但获取读锁的线程无法再获取写锁；
      - ReentrantReadWriteLock 实现了公平和非公平两种锁，公平锁的情况下，获取读锁、写锁前需要看同步队列中是否先线程在我之前排队；非公平锁的情况下：写锁可以直接抢占锁，但是读锁获取有一个让步条件，如果当前同步队列 head.next 是一个写锁在等待，并且自己不是重入的，就要让步等待。
  - Go 显然是不支持可重入互斥锁的
    - Russ Cox 于 2010 年在《Experimenting with GO》就给出了答复，认为递归（又称：重入）互斥是个坏主意，这个设计并不好。
    - Go 的锁是不知道协程或者线程信息的，只知道代码调用先后顺序，即读写锁无法升级或降级。[issue](https://github.com/golang/go/issues/30657)
  - [读锁必然是可重入的?](https://mp.weixin.qq.com/s/T5FQ7z02L60g3ss3YR7axA)
  - RWLock
    - 如果一个协程持有读锁，另一个协程可能会调用 Lock 加写锁，那么再也没有一个协程可以获得读锁，直到前一个读锁释放，这是为了禁止读锁递归。也确保了锁最终可用，一个阻塞的写锁调用会将新的读锁排除在外。
- [用Go实现可重入锁](https://mp.weixin.qq.com/s/LFkPlsLVj24OWZKvanUNVA)
  - 实现一个可重入锁需要这两点：
    - 记住持有锁的线程
    - 统计重入的次数
   ```go
   type ReentrantLock struct {
    lock *sync.Mutex
    cond *sync.Cond
    recursion int32  //记录当前goroutine的重入次数
    host     int64   // 记录当前持有锁的goroutine id
   }
   ```
- [系统中钱的精度](https://mp.weixin.qq.com/s/7Jd5m1pPfivi727R6TTIpA)
  - 精度的问题
   ![img.png](go_float6.png)
  - 用浮点数计算 - 判断两个浮点数是否相等往往采用a - b <= 0.00001的形式
  - 用整型计算
    - 事先定好小数保留8位精度，则0.1和0.2分别表示成整数为10000000和20000000
    - 但，表示2.3在计算机中实际的存储值，因此使用float2Int函数进行转换时的结果是229999999而不是230000000
      - 要解决这个问题也很简单，只需引入github.com/shopspring/decimal即可
        ````go
        const prec = 100000000
        var decimalPrec = decimal.NewFromFloat(prec)
        func float2Int(f float64) int64 {
           return decimal.NewFromFloat(f).Mul(decimalPrec).IntPart()
        }
        ```
    - 整型表示浮点数的范围 
      - 以int64为例，数值范围为-9223372036854775808～9223372036854775807，如果我们对小数部分精度保留8位，则剩余表示整数部分依旧有11位，即只表示钱的话仍旧可以存储上百亿的金额
    - 整型表示浮点数的除法 
      - 在Go中没有隐式的整型转浮点的说法，即整型和整型相除得到的结果依旧是整型。我们以整型表示浮点数时，就尤其需要注意整型的除法运算会丢失所有的小数部分，所以一定要先转换为浮点数再进行相除。
  - 浮点和整型的最大精度
    - int64的范围为-9223372036854775808～9223372036854775807，则用整型表示浮点型时，整数部分和小数部分的有效十进制位最多为19位。
    - uint64的范围为0~18446744073709551615，则用整型表示浮点型时，整数部分和小数部分的有效十进制位最多为20位，因为系统中表示金额时一般不会存储负数，所以和int64相比，更加推荐使用uint64
    - float64根据IEEE754标准，并参考维基百科知其整数部分和小数部分的有效十进制位为15-17位。推荐使用整型表示浮点数
  - 除法和减法的结合
    ```go
    // 1元钱分给3个人，每个人分多少？
    var m float64 = float64(1) / 3
    fmt.Println(m, m+m+m)
    ```
    计算结果知，每人分得0.3333333333333333元，而将每人分得的钱再次汇总时又变成了1元，那么 这0.0000000000000001元是从石头里面蹦出来的嘛
    ```go
    // 1元钱分给3个人，每个人分多少？
    var m float64 = float64(1) / 3
    fmt.Println(m, m+m+m)
    // 最后一人分得的钱使用减法
    m3 := 1 - m - m
    fmt.Println(m3, m+m+m3)
    ```
- [Shallow copy and Deep copy in Go](https://echorand.me/posts/go-values-references-etc/)
  - Basic data types
    - When it comes to the basic types, numbers and strings, it is always deep copy.
    - There is no shallow copy when it comes to these types. Another way of saying that is that, when we want a shallow copy, use memory addresses for basic data types
  - Slice of integers and strings
    - When it comes to a slice, we are always working with shallow copies.
    - If you want to create a deep copy, you will find the copy() function useful. 
  - Arrays of strings and integers
    - When it comes to arrays, we are always working with deep copies. 
    - if we want to pass an array which we want to modify in another function and want the updated result to be reflected in the original array, we should pass the array by reference
  - Elements in Maps and Struct types
    - A map is by default call by reference and struct is by default call by value. 
    - if you want call by value behavior:
      - map: Create a deep copy by creating a new map and copying the key value pairs. Be careful of also ensuring that you deep copy the elements themselves
      - struct: Create a deep copy by creating a new struct and copying the elements. Be careful of also ensuring that you deep copy the elements themselves
- [并发编程指南](https://mp.weixin.qq.com/s/V0krCjWrndzz71cVOPBxdg)
  - channel 特性
  ![img.png](go_channel1.png)
    ```go
    v, ok := <-a  // 检查是否成功关闭(ok = false：已关闭)
    ```
    - 等待一个事件，也可以通过 close 一个 channel 就足够了
    - 利用 channel 阻塞的特性和带缓冲的 channel 来实现控制并发数量
    - singlelFlight - 一般系统重要的查询增加了缓存后，如果遇到缓存击穿，那么可以通过任务计划，加索等方式去解决这个问题，singleflight 这个库也可以很不错的应对这种问题
    - 
  - 有锁的地方就去用 channel 优化 [demo](https://github.com/LinkinStars/simple-chatroom)
- [Go的内存布局和分配原理](https://mp.weixin.qq.com/s/gCDxWzslfPXayJ_RFQVb7g)
  ![img.png](go_memory.png)
  - mheap
    - Go 在程序启动时，首先会向操作系统申请一大块内存，并交由mheap结构全局管理
    - mheap 会将这一大块内存，切分成不同规格的小内存块，我们称之为 mspan，根据规格大小不同，mspan 大概有 70类左右，划分得可谓是非常的精细，足以满足各种对象内存的分配。
  - mcentral
    - 启动一个 Go 程序，会初始化很多的 mcentral ，每个 mcentral 只负责管理一种特定规格的 mspan
    - 相当于 mcentral 实现了在 mheap 的基础上对 mspan 的精细化管理
    - 但是 mcentral 在 Go 程序中是全局可见的，因此如果每次协程来 mcentral 申请内存的时候，都需要加锁
  - mcache
    - 每个P都会绑定一个叫 mcache 的本地缓存
    - 当前运行的goroutine会从mcache中查找可用的mspan。从本地mcache里分配内存时不需要加锁，这种分配策略效率更高
  - 对于那些超过 64KB 的内存申请，会直接从堆上(mheap)上分配对应的数量的内存页（每页大小是 8KB）给程序。
- [逃逸分析](https://slides.com/jalex-chang/go-esc#/9)
  - 根据变量的使用范围
    - 通过 `go build -gcflags '-m -l' demo.go` 来查看逃逸分析的结果，其中 -m 是打印逃逸分析的信息，-l 则是禁止内联优化。
    - 返回任意引用型的变量：Slice 和 Map
    - 在闭包函数中使用外部变量
  - 根据变量类型是否确定程序自己监控自己
    - fmt.Println() 函数, 其接收的参数类型是 interface{} ，对于这种编译期不能确定其参数的具体类型，编译器会将其分配于堆上
  - 根据变量的占用大小
    - 以 64KB 为分界线，我们将内存块分为 小内存块 和 大内存块。 小内存块走常规的 mspan 供应链申请，而大内存块则需要直接向 mheap，在堆区申请。
  - 根据变量长度是否确定
    - 由于逃逸分析是在编译期就运行的，而不是在运行时运行的。因此避免有一些不定长的变量可能会很大，而在栈上分配内存失败，Go 会选择把这些变量统一在堆上申请内存，这是一种可以理解的保险的做法。
- [Go内存管理](https://mp.weixin.qq.com/s?__biz=MzAxMTA4Njc0OQ==&mid=2651449810&idx=2&sn=0c31fe1bf035e4505a17e5686b3457b2&chksm=80bb3720b7ccbe361853d23e873ac94623f78c31887e87aae7e2e2eadae26e3f9d9a3ccbfffd&scene=21#wechat_redirect) [Origin](https://povilasv.me/go-memory-management/#)
  - 内存的基本知识
    - 我们通过 ps 命令观察这个正在运行的程序 `ps -u -p <pid>`, 这个程序居然耗掉了 379.39M 虚拟内存，实际使用内存为 5.11M
      - 虚拟内存大小(VSZ)是进程可以访问的所有内存，包括换出的内存、分配但未使用的内存和共享库中的内存。(stackoverflow 上有很好的解释。)
      - 驻留集大小(RSS)是进程在实际内存中的内存页数乘以内存页大小，这里不包括换出的内存页（译者注：包含共享库占用的内存）
    - 虚拟内存可以使用基于 CPU 体系结构和操作系统的段或页表来实现
    - 在分页虚拟内存中，我们将虚拟内存划分为块，称为页。页的大小可以根据硬件的不同而有所不同，但是页的大小通常是 4-64 KB，此外，通常还能够使用从 2MB 到 1GB 的巨大的页。分块很有用，因为单独管理每个内存槽需要更多的内存，而且会降低计算机的性能。
    - 为了实现分页虚拟内存，计算机通常有一个称为内存管理单元(MMU) 的芯片，它位于 CPU 和内存之间。MMU 在一个名为页表的表(它存储在内存中)中保存了从虚拟地址到物理地址的映射，其中每页包含一个页表项(PTE)。MMU 还有一个物理缓存旁路转换缓冲(TLB)，用来存储最近从虚拟内存到物理内存的转换
    - 每个进程都有一个线性虚拟地址空间，地址从 0 到最大值。虚拟地址空间不需要是连续的，因此并非所有这些虚拟地址实际上都用于存储数据，并且它们不占用 RAM 或磁盘中的空间
  - 操作系统相关
    - 在 Linux 系统中，你可以通过 execve() 系统调用来调用你的程序加载器。
    - 可执行或不可执行的目标文件通常采用容器格式，例如可执行文件和可链接格式（ELF）
      - 查看 ELF 文件信息，如：size --format=sysv main 或 readelf -l main
    - 要动态分配内存，你有几个选择。其中一个选项是调用操作系统（syscall 或通过 libc）。操作系统提供各种功能，如：
      - mmap/munmap - 分配/解除分配固定块内存页面。
      - brk/sbrk - 更改/获取数据分段大小。
      - madvise - 提供操作系统如何管理内存的建议。
      - set_thread_area/get_thread_area - 适用于线程本地存储。
  - 内存分配器
    - Go 语言不使用 malloc 来获取内存, 最初基于 TCMalloc
      - TCMalloc 比 glibc 2.3 malloc 更快
      - TCMalloc 还减少了多线程程序的锁争用
        - 对于小型对象，几乎没有争用。
        - 对于大型对象，TCMalloc 尝试使用细粒度和高效的自旋锁。
    - Go 语言的内存分配器
      - Go 语言的内存分配器与 TCMalloc 类似，它在页运行（spans/mspan 对象）中工作，使用线程局部缓存并根据大小划分分配。跨度是 8K 或更大的连续内存区域
      - Spans 有 3 种类型 - free span, using span, stack span
      - mcache
      - mcentral
      - mheap
  - Debug
    - cat /proc/30376/status
    - cat /proc/31086/maps
- [Json.Unmarshal精度丢失](https://mp.weixin.qq.com/s/36CqC1U54LUd4-izt4iZ1g)
  - demo
    ```go
    var request = `{"id":7044144249855934983,"name":"demo"}`
    
     var test interface{}
     err := json.Unmarshal([]byte(request), &test)
     if err != nil {
      fmt.Println("error:", err)
     }
    
     obj := test.(map[string]interface{})
     dealStr, err := json.Marshal(test)
     if err != nil {
      fmt.Println("error:", err)
     }
    
     id := obj["id"]
    
     // 反序列化之后重新序列化打印
     fmt.Println(string(dealStr))
     fmt.Printf("%+v\n", reflect.TypeOf(id).Name())
     fmt.Printf("%+v\n", id.(float64))
    ```
    ```shell
    {"id":7044144249855935000,"name":"demo"}
    float64
    7.044144249855935e+18
    ```
    - 原来是这样的：
      - 在json的规范中，对于数字类型是不区分整形和浮点型的。
      - 在使用json.Unmarshal进行json的反序列化的时候，如果没有指定数据类型，使用interface{}作为接收变量，其默认采用的float64作为其数字的接受类型
      - 当数字的精度超过float能够表示的精度范围时就会造成精度丢失的问题
    - 解决方案有两种：
      - 上游将id改为string传给下游
      - 下游使用json.number类型来避免对float64的使用
       ```go
        var request = `{"id":7044144249855934983}`
       
        var test interface{}
        decoder := json.NewDecoder(strings.NewReader(request))
        decoder.UseNumber()
        err := decoder.Decode(&test)
        if err != nil {
         fmt.Println("error:", err)
        }
       
        objStr, err := json.Marshal(test)
        if err != nil {
         fmt.Println("error:", err)
        }
       
        fmt.Println(string(objStr))
       ```
  - 探究
    - 为什么json.unmarshal使用float64来处理就可能出现精度缺失呢？ 缺失的程度是怎样的？
      - int64是将64bit的数据全部用来存储数据，但是float64需要表达的信息更多，因此float64单纯用于数据存储的位数将小于64bit，这就导致了float64可存储的最大整数是小于int64的。
    - 什么时候出现精度缺失？ 里面有什么规律吗？
      - ![img.png](go_float7.png)
      - 尾数部分全部为1时就已经拉满了，再多1位尾数就要向指数发生进位，此时就会出现精度缺失，因此对于float64来说：
        - 最大的安全整数是52位尾数全为1且指数部分为最小 0x001F FFFF FFFF FFFF
        - float64可以存储的最大整数是52位尾数全位1且指数部分为最大 0x07FEF FFFF FFFF FFFF
      - 也就是理论上数值超过9007199254740991就可能会出现精度缺失
      - 10进制数值的有效数字是16位，一旦超过16位基本上缺失精度是没跑了，回过头看我处理的id是20位长度，所以必然出现精度缺失。
    - [反序列化时decoder和unmarshal如何选择呢？](https://stackoverflow.com/questions/21197239/decoding-json-using-json-unmarshal-vs-json-newdecoder-decode)
      - float64存在精度缺失的问题，因此go单独对此给出了一个解决方案：
        - 使用 json.Decoder 来代替 json.Unmarshal 方法
        - 该方案首先创建了一个 jsonDecoder，然后调用了 UseNumber 方法
        - 使用 UseNumber 方法后，json 包会将数字转换成一个内置的 Number 类型（本质是string），Number类型提供了转换为 int64、float64 等多个方法
      - json.NewDecoder是从一个流里面直接进行解码，代码更少，可以用于http连接与socket连接的读取与写入，或者文件读取
      - json.Unmarshal是从已存在与内存中的json进行解码
- [HttpClient读取Body超时](https://juejin.cn/post/7051451783909998623)
  - HttpClient 请求有 30%概率超时， 报context deadline exceeded (Client.Timeout or context cancellation while reading body) 异常
    ![img.png](go_http_read_timeout.png)
  - 吐槽iotil.ReadAll的性能了。
    客户端使用 iotil.ReadAll 读取大的响应体，会不断申请内存(源码显示会从 512B->50M)，耗时较长，性能较差、并且有内存泄漏的风险， [针对大的响应体替换iotil.ReadAll的方案](https://stackoverflow.com/questions/52539695/alternative-to-ioutil-readall-in-go)
  - 替换ioutil ReadAll
    - a much more efficient way of parsing JSON, which is to simply use the Decoder type.
      ```go
      err := json.NewDecoder(r).Decode(&v)
      if err != nil {
         return err
      }
      ```
    - Writing data to a file
      ```go
      f, err := os.Create("file")
      if err != nil {
          return err 
      }
      defer f.Close()
      
      // Copy will put all the data from Body into f, without creating a huge buffer in memory
      // (moves chunks at a time)
      io.Copy(f, resp.Body)
      ```
- [动态调整 GOGC 优化 Go 的 GC 标记 CPU 占用](https://mp.weixin.qq.com/s/XR1KAeCW930i-Qxv6N2kaA)
  - [在核心服务上动态调整 GOGC 来降低 GC 的 mark 阶段 CPU 占用](https://eng.uber.com/how-we-saved-70k-cores-across-30-mission-critical-services/)
  - 起因
    - 经过一段时间的线上 profile 采集发现 GC 是很多核心服务的一个很大的 CPU 消耗点，比如 runtime.scanobject 方法消耗了很大比例的计算资源
  - GOGC Tuner
    - Go 的 runtime 会间隙性地调用垃圾收集器来并发进行垃圾回收。这个启动是由内存的压力反馈来决定何时启动 GC 的。所以 Go 的服务可以通过增加内存的用量来降低 GC 的频率以降低 GC 的总 CPU 占用
    - Go 的 GC 触发算法可以简化成下面这样的公式： `hard_target = live_dataset + live_dataset * (GOGC / 100).` 由 pacer 算法来计算每次最合适触发的 heap 内存占用
    - 固定的 GOGC 值没法满足 Uber 内部所有的服务。具体的挑战包括：
      - 对于容器内的可用最大内存并没有进行考虑，理论上存在 OOM 的可能性。
      - 不同的微服务对内存的使用情况完全不同。
  - 自动化
    - Uber 内部搞了一个叫 GOGCTuner 的库。这个库简化了 Go 的 GOGC 参数调整流程，并且能够可靠地自动对其进行调整。
    - 默认的 GOGC 参数是 100%，这个值对于 GO 的开发者来说并不明确，其本身还是依赖于活跃的堆内存。GOGCTuner 会限制应用使用 70% 的内存。并且能够将内存用量严格限制住。
    - 可以保护应用不发生 OOM：该库会读取 cgroup 下的应用内存限制，并且强制限制只能使用 70% 的内存，从我们的经验来看这样还是比较安全的。
    - 使用 MADV_FREE 内存策略会导致错误的内存指标。所以使用 Go 1.12-Go 1.15 的同学注意设置 madvdontneed 的环境变量
  - 可观测性 - 对垃圾回收的一些关键指标进行了监控
    - 垃圾回收触发的时间间隔：_可以知道是否还需要进一步的优化。比如 Go 每两分钟强制触发一次垃圾回收。
    - GC 的 CPU 使用量: 使我们能知道哪些服务受 GC 影响最大
    - 活跃的对象大小: 帮我们来诊断内存泄露
    - GOGC 的动态值: 能知道 tuner 是不是在干活。
  - 实现
    - Go 有一个 finalizer 机制，在对象被 GC 时可以触发用户的回调方法。Uber 实现了一个自引用的 finalizer 能够在每次 GC 的时候进行 reset，这样也可以降低这个内存检测的 CPU 消耗
    - 在 finalizerHandler 里调用 runtime.SetFinalizer(f, finalizerHandler) 能让这个 handler 在每次 GC 期间被执行；这样就不会让引用真的被干掉了，这样使该对象存活也并不需要太高的成本，只是一个指针而已
    ![img.png](go_finalizer.png)

- [空结构体struct{}](https://mp.weixin.qq.com/s/YU45CYk7Q-Na2WISMrUYCQ)
  - 空结构体类型的变量占用的空间为0
     ```go
     var s struct{}
     fmt.Println(unsafe.Sizeof(s)) // prints 0
     typ := reflect.TypeOf(s)
     fmt.Println(typ.Size()) // 0
     ```
  - 所有空结构体类型的变量地址都是一样的 - zerobase 的地址
    ```go
        a := struct{}{}
        b := struct{}{}
     
        c := emptyStruct{}
     
        fmt.Println(a)
        fmt.Printf("%pn", &a) //0x116be80
        fmt.Printf("%pn", &b) //0x116be80
        fmt.Printf("%pn", &c) //0x116be80
     
        fmt.Println(a == b) //true
    ```
  - 空结构体的应用场景
    - 基于map实现集合功能
      ```go
      var CanSkipFuncs = map[string]struct{}{
          "Email":   {},
          "IP":      {},
          "Mobile":  {},
          "Tel":     {},
          "Phone":   {},
          "ZipCode": {},
      }
      ```
    - 与channel组合使用，实现一个信号 - 基于缓冲channel实现并发限速
      ```go
      var limit = make(chan struct{}, 3)
      
      func main() {
          // …………
          for _, w := range work {
              go func() {
                  limit <- struct{}{}
                  w()
                  <-limit
              }()
          }
          // …………
      }
      ```
- [slice tricks](https://mp.weixin.qq.com/s/IQRHWNUnxiaCDleayNVRVg)
  - [slice tricks official](https://github.com/golang/go/wiki/SliceTricks)
  - [slice tricks legend](https://ueokande.github.io/go-slice-tricks/)
- [线程安全的map](https://mp.weixin.qq.com/s/H5HDrwhxZ_4v6Vf5xXUsIg)
  - sync.map
    - 使用场景
      - sync.Map更适合读多更新多而插入新值少的场景, 因为在key存在的情况下读写删操作可以不用加锁直接访问readOnly
      - 不适合反复插入与读取新值的场景，因为这种场景会频繁操作dirty，需要频繁加锁和更新read
    - 设计点:expunged
      - entry.p取值有3种，nil、expunged和指向真实值
      - 当用Store方法插入新key时，会加锁访问dirty，并把readOnly中的未被标记为删除的所有entry指针复制到dirty，此时之前被Delete方法标记为软删除的entry（entry.p被置为nil）都变为expunged，那这些被标记为expunged的entry将不会出现在dirty中。
      - 如果没有expunged，只有nil会出现什么结果呢？
        - 直接删掉entry==nil的元素，而不是置为expunged：在用Store方法插入新key时，readOnly数据拷贝到dirty时直接把为ni的entry删掉。但这要对readOnly加锁，sync.map设计理念是读写分离，所以访问readOnly不能加锁。
        - 不删除entry==nil的元素，全部拷贝：在用Store方法插入新key时，readOnly中entry.p为nil的数据全部拷贝到dirty中。那么在dirty提升为readOnly后这些已被删除的脏数据仍会保留，也就是说它们会永远得不到清除，占用的内存会越来越大。
        - 不拷贝entry.p==nil的元素：在用Store方法插入新key时，不把readOnly中entry.p为nil的数据拷贝到dirty中，那在用Store更新值时，就会出现readOnly和dirty不同步的状态，即readOnly中存在dirty中不存在的key，那dirty提升为readOnly时会出现数据丢失的问题。
  - orcanman/concurrent-map
    - orcaman/concurrent-map的适用场景是：反复插入与读取新值，
    - 其实现思路是:对go原生map进行分片加锁，降低锁粒度，从而达到最少的锁等待时间(锁冲突)。
- [Use buffered channel as mutex](https://mp.weixin.qq.com/s/DRE38mOYYqURMFVkqYPu3w)
  - 通过 channel 和通信更好地完成更高级别的同步
  - 无缓冲 channel 及其不足之处
    - 如果没有接收方，发送者将会阻塞；相同地，如果没有发送方，接收者将会阻塞。基于这种特性，所以我们不能将无缓冲的 channel 作为锁来使用。
  - 缓冲为 1 的 channel 的特性及其可取之处
    - 缓冲大小为 1 的 channel 具有如下的特性：如果缓冲满了，发送时将会阻塞；如果缓存腾空，发送时就会解除阻塞。
    - 缓冲满时 <--> 上锁
    - 缓冲腾空 <--> 解锁
  - sample
     ```go
      chanLock := make(chan int, 1) //1
      var wg sync.WaitGroup
      for _, str := range ss { //2
       wg.Add(1) 
       go func(aString string) {
     
        chanLock <- 1 //3
        for i := 0; i < 1000; i++ {
         file.WriteString(aString + "\n")
        }
        <-chanLock //4
        wg.Done() //5
       }(str) //pass by value
      }
      wg.Wait()
     ```
- [API Design](https://go-talks.appspot.com/github.com/matryer/golanguk/building-apis.slide#10)
  - OK pattern
  - Public pattern / Adapter
- [流处理场景下的最小化内存使用](https://mp.weixin.qq.com/s/RWDyDmeI1YhstAh-rHd2-A)
  - [Source](https://engineering.be.com.vn/large-stream-processing-in-golang-with-minimal-memory-usage-c1f90c9bf4ce)
  - Multipart 文件转发
    - io.Copy 从 src 复制副本到 dst，直到在 src 上到达 EOF 或发生错误。它返回复制的字节数和复制时遇到的第一个错误(如果有的话)
    - 在文件离线处理时，你可以打开一个带缓冲的 writer 然后完全复制 reader 中内容，并且不用担心任何其他影响。然而，Copy 操作将持续地将数据复制到 Writer，直到 Reader 读完数据。但这是一个无法控制的过程，如果你处理 writer 中数据的速度不能与复制操作一样快，那么它将很快耗尽你的缓冲区资源
    - Pipe 提供一对 writer 和 reader，并且读写操作都是同步的。利用内部缓冲机制，直到之前写入的数据被完全消耗掉才能写到一个新的 writer 数据快。这样你就可以完全控制如何读取和写入数据。现在，数据吞吐量取决于处理器读取文本的方式，以及 writer 更新数据的速度。
        ```go
        r, w := io.Pipe()
        m := multipart.NewWriter(w)
        go func() {
           defer w.Close()
           defer m.Close()
           part, err := m.CreateFormFile("file", "textFile.txt")
           if err != nil {
              return
           }
           file, err := os.Open(name)
           if err != nil {
              return
           }
           defer file.Close()
           if _, err = io.Copy(part, file); err != nil {
              return
           }
        }()
        http.Post(url, m.FormDataContentType(), r)
        ```
  - 预取和补偿文件流
    - 一个可行的解决方案是使用 io.TeeReader，它会将从 reader 读取的数据写入另一个 writer 中。TeeReader 最常见的用例是将一个流克隆成一个新的流，在保持流不被破坏的情况下为 reader 提供服务
    - 但问题是，如果在将其传递给 GCP 文件处理程序之前同步运行它，它最终还是会将所有数据复制到准备好的缓冲区。一个可行的方法是再次使用 Pipe 来操作它，达到无本地缓存效果。但另一个问题是，TeeReader 要求在完成读取过程之前必须完成写入过程，而 Pipe则相反。
     ```go
     type prefetchReader struct {
        reader   io.Reader
        prefetch []byte
        size     int
     }
     
     func newPrefetchReader(reader io.Reader, prefetch []byte) *prefetchReader {
        return &prefetchReader{
           reader:   reader,
           prefetch: prefetch,
        }
     }
     
     func (r *prefetchReader) Read(p []byte) (n int, err error) {
        if len(p) == 0 {
           return 0, fmt.Errorf("empty buffer")
        }
        defer func() {
           r.size += n
        }()
        if len(r.prefetch) > 0 {
           if len(p) >= len(r.prefetch) {
              copy(p, r.prefetch)
              n := len(r.prefetch)
              r.prefetch = nil
              return n, nil
           } else {
              copy(p, r.prefetch[:len(p)])
              r.prefetch = r.prefetch[len(p):]
              return len(p), nil
           }
        }
        return r.reader.Read(p)
     }
     ```
- [Further Dangers of Large Heaps in Go](https://syslog.ravelin.com/further-dangers-of-large-heaps-in-go-7a267b57d487)
  - To keep the amount of GC work down you essentially have two choices as follows.
    - Make sure the memory you allocate contains no pointers. That means no slices, no strings, no time.Time, and definitely no pointers to other allocations. If an allocation has no pointers it gets marked as such and the GC does not scan it.
    - Allocate the memory off-heap by directly calling the mmap syscall yourself. Then the GC knows nothing about the memory. This has upsides and downsides. The downside is that this memory can’t really be used to reference objects allocated normally, as the GC may think they are no longer in-use and free them.
- [Implementing Graceful Shutdown in Go](https://dev.to/rudderstack/implementing-graceful-shutdown-in-go-1a1b)
  - Anti-patterns
    - Block artificially
    - os.Exit(): Calling os.Exit(1) while other go routines are still running is essentially equal to SIGKILL, no chance for closing open connections and finishing inflight requests and processing.
  - How to make it graceful in Go
    - How to wait for all the running go routines to exit
      - Channel: This is mostly useful when waiting on a single go routine.
      - WaitGroup: Waiting multiple go routines
      - errgroup
        - The two errgroup's methods .Go and .Wait are more readable and easier to maintain in comparison to WaitGroup.
        - In addition, as its name suggests it does error propagation and cancels the context in order to terminate the other go-routines in case of an error.
    - How to propagate the termination signal to multiple go routines
      - channel failed to do that, whereas context could done. Refer to snippet
- [基于channel实现的并发安全的字节池](https://mp.weixin.qq.com/s/91_FxpV5qbR-XNqh0Dh8EA)
  - MinIO
    ```go
    type BytePoolCap struct {
        c    chan []byte
        w    int
        wcap int
    }
    ```
     ```go
     func (bp *BytePoolCap) Get() (b []byte) {
         select {
         case b = <-bp.c:
         // reuse existing buffer
         default:
             // create new buffer
             if bp.wcap > 0 {
                 b = make([]byte, bp.w, bp.wcap)
             } else {
                 b = make([]byte, bp.w)
             }
         }
         return
     }
     ```
- [Understanding Allocations in Go](https://medium.com/eureka-engineering/understanding-allocations-in-go-stack-heap-memory-9a2631b5035d)
- [A visual guide to Go Memory Allocator from scratch ](https://medium.com/@ankur_anand/a-visual-guide-to-golang-memory-allocator-from-ground-up-e132258453ed)
  - [Chinese](https://www.linuxzen.com/go-memory-allocator-visual-guide.html)
  - 内存分配器
    - 如果堆上有足够的空间的满足我们代码的内存申请，内存分配器可以完成内存申请无需内核参与，否则将通过操作系统调用（brk）进行扩展堆，通常是申请一大块内存。
    - 内存分配器除了更新 brk address 还有其他职责, 如何减少 内部（internal）和外部（external）碎片和如何快速分配当前块
    - Go 内存分配器建模相近的内存分配器： TCMalloc。核心思想是将内存分为多个级别缩小锁的粒度。在 TCMalloc 内存管理内部分为两个部分：线程内存（thread memory)和页堆（page heap）
  - Go 内存分配器
    - Go 实现的 TCMalloc 将内存页（Memory Pages）分为 67 种不同大小规格的块, 这些页通过 mspan 结构体进行管理
    - mspan: 是一个包含页起始地址、页的 span 规格和页的数量的双端链表
    - mcache: 一个本地线程缓存（Local Thread Cache）称作 mcache
      - mcache 包含所有大小规格的 mspan 作为缓存
      - 由于每个 P 都拥有各自的 mcache，所以从 mcache 分配内存无需持有锁
      - <=32K 字节的对象直接使用相应大小规格的 mspan 通过 mcache 分配
      - 当 mcache 没有可用空间时会从 mcentral 的 mspans 列表获取一个新的所需大小规格的 mspan
    - mcentral: mcentral 对象收集所有给定规格大小的 span。
      - 每一个 mcentral 都包含两个 mspan 的列表
        - empty mspanList -- 没有空闲对象或 span 已经被 mcache 缓存的 span 列表
        - nonempty mspanList -- 有空闲对象的 span 列表
      - 对齐填充（Padding）用于确保 mcentrals 以 CacheLineSize 个字节数分隔，所以每一个 MCentral.lock 都可以获取自己的缓存行（cache line），以避免伪共享（false sharing）问题。
      - 每一个 mcentral 结构体都维护在 mheap 结构体内
    - mheap: Go 使用 mheap 对象管理堆，只有一个全局变量。持有虚拟地址空间。
      - 由于我们有各个规格的 span 的 mcentral，当一个 mcache 从 mcentral 申请 mspan 时，只需要在独立的 mcentral 级别中使用锁，所以其它任何 mcache 在同一时间申请不同大小规格的 mspan 将互不受影响可以正常申请。
      - 大于 32K 的对象被定义为大对象，直接通过 mheap 分配。这些大对象的申请是以一个全局锁为代价的，因此任何给定的时间点只能同时供一个 P 申请。
- [Mutex vs Atomic](https://ms2008.github.io/2019/05/12/golang-data-race/)
  - Mutexes do no scale. Atomic loads do.
  - mutex 由操作系统实现，而 atomic 包中的原子操作则由底层硬件直接提供支持。在 CPU 实现的指令集里，有一些指令被封装进了 atomic 包，这些指令在执行的过程中是不允许中断（interrupt）的，因此原子操作可以在 lock-free 的情况下保证并发安全，并且它的性能也能做到随 CPU 个数的增多而线性扩展。
- [The Escape Analysis in Go](https://slides.com/jalex-chang/go-esc)
  - Introduction
    - Allocating objects on the stack is faster than in the heap.
    - The escape analysis is a mechanism to automatically decide whether a variable should be allocated in the heap or not in compile time
  - When does ESC happen
  - ESC - concept - Generally, a variable scapes if:
    - its address has been captured by the address-of operand (&).
    - and at least one of the related variables has already escaped.
  - How does ESC work - Basically, ESC determines whether variables escape or not by
    - the data-flow analysis (shortest path analysis)
    - and other additional rules
      - Huge objects
        - For explicit declarations (var or :=)
          - The variables escape if their sizes are over 10MB
        - For implicit declarations (new or make)
          - The variables escape if their sizes are over 64KB 
      - Slice
        - A slice variable escapes if its size of the capacity is non-constant
      - Map
        - A variable escapes if it is referenced by a map's key or value.
        - The escape happens no matter the map escape or not
      - Return values - Returning values is a backward behavior that
        - the referenced variables escape if the return values are pointers
        - the values escape if they are map or slice 
      - Input parameters -  Passing arguments is a forward behavior that
        - the arguments escape if input parameters have leaked (to heap)
      - Closure function - A variable escapes if
        - the source variable is captured by a closure function
        - and their relationship is address-of (derefs = -1 )
  - How to utilize ESC to benefit our programs
    - Observations - Through understanding the concept of ESC, we can find that
      - variables usually escape
        - when their addresses are captured by other variables.
        - when ESC does not know their object sizes in compile time.
      - And passing arguments to a function is safer than returning values from the function. 
    - the first and most important suggestion is: try not to use pointers as much as possible
    - Initialize slice with constants
    - Passing variables to closure functions
    - Argument injection 
      - Injecting changes to the passed parameters instead of return values back - For exmaple: Reader.Read in pkg bufio
- [golang本地缓存(bigcache/freecache/fastcache等)选型对比](https://zhuanlan.zhihu.com/p/487455942)
  - 本地缓存需求
    - 需要较高的读写性能 + 命中率
    - 支持按写入时间过期
    - 支持淘汰策略
    - 解决GC问题，否则大量对象写入会引起STW扫描标记时间过长，CPU毛刺严重
  - 可选的开源本地缓存组件汇总
    ![img.png](go_local_cache.png)
    - 上述本地缓存组件中，实现零GC的方案主要就两种：
      - 无GC：分配堆外内存(Mmap)
      - 避免GC：map非指针优化(map[uint64]uint32)或者采用slice实现一套无指针的map
      - 避免GC：数据存入[]byte slice(可考虑底层采用环形队列封装循环使用空间)
    - 实现高性能的关键在于：
      - 数据分片(降低锁的粒度)
  - 主流缓存组件实现原理剖析
    - freecache实现原理
      - 在freecache中它通过segment来进行对数据分片，freecache内部包含256个segment，每个segment维护一把互斥锁，每一条kv数据进来后首先会根据k进行计算其hash值，然后根据hash值决定当前的这条数据落入到哪个segment中。
      - 每个segment而言，它由索引、数据两部分构成。
        - 索引：其中索引最简单的方式采用map来维护，例如map[uint64]uint32这种。而freecache并没有采用这种做法，而是通过采用slice来底层实现一套无指针的map，以此避免GC扫描。
        - 数据：数据采用环形缓冲区来循环使用，底层采用[]byte进行封装实现。数据写入环形缓冲区后，记录写入的位置index作为索引，读取时首先读取数据header信息，然后再读取kv数据。
      - ![img.png](go_local_cache_freecache.png)
    - [bigcache实现原理](https://blog.allegro.tech/2016/03/writing-fast-cache-service-in-go.html)
      - bigcache同样是采用分片的方式构成，一个bigcache对象包含2^n 个cacheShard对象，默认是1024个。每个cacheShard对象维护着一把sync.RWLock锁(读写锁)。所有的数据会分散到不同的cacheShard中。
      - 每个cacheShard同样由索引和数据构成。索引采用map[uint64]uint32来存储，数据采用entry([]byte)环形队列存储。索引中存储的是该条数据在entryBuffer写入的位置pos。每条kv数据按照TLV的格式写入队列。
      - 和bigcache和freecache不同的一点在于它的环形队列可以自动扩容。同时bigcache中数据的过期是通过全局的时间窗口维护的，每个单独的kv无法设置不同的过期时间。
      - ![img.png](go_local_cache_bigcace.png)
      - 堆上有 4 千万个对象，GC 的扫描过程就超过了 4 秒钟，这就不能忍了。 主要的优化思路有：
        - offheap（堆外内存），GC 只会扫描堆上的对象，那就把对象都搞到栈上去，但是这样这个缓存库就高度依赖 offheap 的 malloc 和 free 操作了
        - 参考 freecache 的思路，用 ringbuffer 存 entry，绕过了 map 里存指针，简单瞄了一下代码，后面有空再研究一下（继续挖坑
        - 利用 Go 1.5+ 的特性： 当 map 中的 key 和 value 都是基础类型时，GC 就不会扫到 map 里的 key 和 value
      - 最终他们采用了 map[uint64]uint32 作为 cacheShard 中的关键存储。key 是 sharding 时得到的 uint64 hashed key，value 则只存 offset ，整体使用 FIFO 的 bytes queue，也符合按照时序淘汰的需求，非常精巧。
    - fastcache实现原理
      - 它的灵感来自于bigcache。所以整体的思路和bigcache很类似，数据通过bucket进行分片。fastcache由512个bucket构成。每个bucket维护一把读写锁。
      - 在bucket内部数据同理是索引、数据两部分构成。索引用map[uint64]uint64存储。数据采用chunks二维的切片(二维数组)存储。
      - 它的内存分配是在堆外分配的，而不是在堆上分配的。堆外分配的内存。这样做也就避免了golang GC的影响。
      - ![img.png](go_local_cache_fastcache.png)
- [Goroutine 数量控制在多少合适，会影响 GC 和调度](https://mp.weixin.qq.com/s?__biz=MzUxMDI4MDc1NA==&mid=2247487250&idx=1&sn=3004324a9d2ba99233c4af48843dba64&scene=21#wechat_redirect)
  - M 的限制
    - 第一，要知道在协程的执行中，真正干活的是 GPM 中的哪一个？
      - 那势必是 M（系统线程） 了，因为 G 是用户态上的东西，最终执行都是得映射，对应到 M 这一个系统线程上去运行。
    - 那么 M 有没有限制呢？
      - 答案是：有的。在 Go 语言中，M 的默认数量限制是 10000, 可以通过 debug.SetMaxThreads 方法进行设置
    - [GPM 模型的 M 实际数量受什么影响](https://mp.weixin.qq.com/s/q9fafsQlhm-CLUsDYQAhbg)
      - 本质上与 M 是否空闲和是否忙碌有关。
      - 如果在调度时，发现没有足够的 M 来绑定 P，P 中又有需要就绪的任务，就会创建新的 M 来绑定。
      - 如果有空闲的 M，自然也就不会创建全新的 M 了，会优先使用。
  - G 的限制
    - 第二，那 G 呢，Goroutine 的创建数量是否有限制？
      - 答案是：没有。但理论上会受内存的影响，假设一个 Goroutine 创建需要 4k（via @GoWKH）：
  - P 的限制
    - 第三，那 P 呢，P 的数量是否有限制，受什么影响？
      - 答案是：有限制。P 的数量受环境变量 GOMAXPROCS 的直接影响。
    - 环境变量 GOMAXPROCS 又是什么？
      - 在 Go 语言中，通过设置 GOMAXPROCS，用户可以调整调度中 P（Processor）的数量。
      - 另一个重点在于，与 P 相关联的的 M（系统线程），是需要绑定 P 才能进行具体的任务执行的，因此 P 的多少会影响到 Go 程序的运行表现。
- [GO 1.17调用规约到底优化了多少](https://mp.weixin.qq.com/s/7T3Gh4H-qyUrv-WBxT9jcg)
  - 什么是调用规约
    - 当函数A带着3个int类型的实参调用函数B的时候，3个参数放在哪
    - 基于寄存器/平台的调用规约又分为Caller-saved registers 和 Callee-saved registers两种：
      - Caller-saved registers：也叫易失性寄存器，用来保存不需要跨函数传递的参数,比如A(q, w) -> B(q, w int) -> C(w int)，也就是说寄存器保存的值在保证程序正确的情况下会发生变化
      - Callee-saved registers：也叫持久性寄存器，调用方调用函数之后存储在寄存器的值不会变（不会让被调用方改掉），因为需要持久保存对于带GC的语言又是一个挑战。
    - Go目前实现的是Caller-saved registers，为什么要有 Callee-saved registers呢，最典型的就是保存调用栈
  - 为什么Go要从基于栈的调用规约切换为基于寄存器的调用规约 
    - 官网说优化了5%
    - 可能还不如intel优化栈操作来的实际，所有好多人吐槽，“要只提升5%，不如考虑考虑将所有代码内内联吧”
  - Go如何实现的
    - 函数参数和返回值对于int类型只用了9个寄存器分别是： RAX, RBX, RCX, RDI, RSI, R8, R9, R10, R11
    - 那到底什么是stack spill/register spill呢？ 一个寄存器（比如AX），如果长时间存放某个值不适用就会被踢栈上。
- [JSON 与 Cache 库 调研与选型](https://mp.weixin.qq.com/s/2WVBYJjeDkTBr9dDkbouqg)
  - JSON
    - GO 1.14 标准库 JSON大量使用反射获取值，首先 go 的反射本身性能较差，其次频繁分配对象，也会带来内存分配和 GC 的开销
    - valyala/fastjson star: 1.4k
      - 它将 JSON 解析划分为两部分：Parse、Get。 Parse 负责将 JSON 串解析成为一个结构体并返回，然后通过返回的结构体来获取数据。在 Parse 解析的过程是无锁的，所以如果想要在并发地调用 Parse 进行解析需要使用 ParserPool
      - 通过遍历 json 字符串找到 key 所对应的 value，返回其值 []byte，由业务方自行处理。同时可以返回一个 parse 对象用于多次解析；
      - 只提供了简单的 get 接口，不提供 Unmarshal 到结构体或 map 的接口；
      - 没有常用的如 JSON 转 Struct 或 JSON 转 map 的操作。如果只是想简单的获取 JSON 中的值，那么使用这个库是非常方便的，但是如果想要把 JSON 值转化成一个结构体就需要自己动手一个个设值了。
    - tidwall/gjson star: 9.5k
      - 原理与 fastjson 类似，但不会像 fastjson 一样将解析的内容保存在一个 parse 对象中，后续可以反复的利用，所以当调用 GetMany 想要返回多个值的时候，需要遍历 JSON 串多次，因此效率会比较低；
      - 提供了 get 接口和 Unmarshal 到 map 的接口，但没有提供 Unmarshal 到 struct 的接口；
    - buger/jsonparser star: 4.4k
      - 原理与 gjson 类似，有一些更灵活的 api； 只提供了简单的 get 接口，不提供 Unmarshal 到结构体或 map 的接口；
      - 性能如此高的原因可以总结为：
        - 使用 for 循环来减少递归的使用；
        - 相比标准库而言没有使用反射；
        - 在查找相应的 key 值找到了便直接退出，可以不用继续往下递归；
        - 所操作的 JSON 串都是已被传入的，不会去重新再去申请新的空间，减少了内存分配；
    - json-iterator star: 10.3k
      - 兼容标准库；
      - 其之所以快，一个是尽量减少不必要的内存复制，另一个是减少 reflect 的使用——同一类型的对象，jsoniter 只调用 reflect 解析一次之后即缓存下来。
      - 不过随着 go 版本的迭代，原生 json 库的性能也越来越高，jsonter 的性能优势也越来越窄，但仍有明显优势。
    - [sonic](https://mp.weixin.qq.com/s?__biz=MzI1MzYzMjE0MQ==&mid=2247491325&idx=1&sn=e8799316d55c0951b0b54b404a3d87b8&scene=21#wechat_redirect) star: 2k
      - 兼容标准库；
      - 通过JIT（即时编译）和SIMD（单指令-多数据）加速；需要 go 1.15 及以上的版本，提供完成的 json 操作的 API，是一个比 json-iterator 更优的选择。
      - 已经在抖音内部大范围使用，且 github 库维护给力，issues 解决积极，安全性有保证。
      - sonic ：基于 JIT 技术的开源全场景高性能 JSON 库
    - easyjson star: 3.5k
      - 支持序列化和反序列化;
      - 通过代码生成的方式，达到不使用反射的目的；
    - 业务场景
      - 需要 Unmarshal map；
      - json 导致的 GC 与 CPU 压力较大；
      - 业务较为重要，需要一个稳定的序列化库；
    - 选型思路
      - easyjson 需要生成代码，丧失了 json 的灵活性，增加维护成本，因此不予考虑；
      - sonic 需要 go 1.15 及以上的版本，且业务场景无 Unmarshal 到结构体的操作，因此暂时不做选择；
      - json-iterator 的优势在于兼容标准库接口，但因为使用到了反射，性能相对较差，且业务场景没有反序列化结构体的场景，因此不予考虑；
      - fastjson、gjson、jsonparser 由于没有用到反射，因此性能要高于 json-iterator。所以着重在这三个中选择；
      - fastjson 实现了 0 分配的开销，但是 star 数较少，不予考虑；
      - gjson 与 jsonparser 类似，速度及内存分配上各擅胜场，灵活性上也各有长处，比较难抉择，但业务场景下不需要使用到其提供的灵活 API，而有 json 序列化到 map 的场景，所以 gjson 会有一些优势，再结合 star 数后选择 gjson；
  - Cache
    - go-cache star: 5.7k
      - 最简单的 cache，可以直接存储指针，下面的部分 Cache 都需要先把对象序列化为 []byte，会引入一定的序列化开销，但可以用高效的序列化库减少开销；
      - 可以对每个 key 设置 TTL；
      - 无淘汰机制；
    - freecache star: 3.6k
      - 0 GC;
      - 可以对每个 key 设置 TTL；
      - 近 LRU 淘汰；
    - bigcache star: 5.4k
      - 0 GC；
      - 只有全局 TTL，不能对每个 key 设置 TTL；
      - 如果超过内存最大值（也可以不设置，内存使用无上限），采用的是 FIFO 策略；
      - 产生 hash 冲突会导致旧值被覆盖；
      - 会在内存中分配大数组用以达到 0 GC 的目的，一定程度上会影响到 GC 频率；
    - fastcache star: 1.3k
      - 0 GC；
      - 不支持 TTL；
      - 如果超过设置最大值，底层是 ring buffer，缓存会被覆盖掉， 采用的是 FIFO 策略；
      - 调用 mmap 分配堆外内存，因此不会影响到 gc 频率；
    - groupcache star: 11k
      - 一个较为复杂的 cache 实现，本质上是个 LRU cache；
      - 是一个lib库形式的进程内的分布式缓存，也可以认为是本地缓存，但不是简单的单机缓存，不过也可以作为单机缓存；
      - 特性如下：单机缓存和基于HTTP的分布式缓存；最近最少访问（LRU，Least Recently Used）缓存策略；使用Golang锁机制防止缓存击穿；使用一致性哈希选择节点以实现负载均衡；使用Protobuf优化节点间二进制通信；
    - goburrow star: 468
      - Go 中 Guava Cache 的部分实现；
      - 没有对 GC 做优化，内部使用 sync.map；
      - 支持淘汰策略：LRU、Segmented LRU (default)、TinyLFU (experimental)；
    - ristretto star: 3.6k
      - 在 GC 方面做了少量优化；
      - 可以对每个 key 设置 TTL；
      - 在吞吐方面做了较多优化，使得在复杂的淘汰策略下仍具有较好的吞吐水平；
      - 在命中率方面，具备出色的准入政策和 SampledLFU 驱逐政策，因此高于其他 cache 库；
    - 业务场景 - Feature 服务
      - key 分钟固定窗口失效，且 key 中自带分钟级时间戳；
      - 内存容量足够，有全局 TTL 即可，不需要额外的淘汰机制；
      - 缓存 Key 数量较多，对 GC 压力较大；
      - Value 是 string，另外可以通过不安全方式无开销转换为 []byte；
      - 业务较为重要，需要一个稳定的 cache 库；
    - 选型思路
      - goburrow、ristretto 两个 cache 的主打的是固定内存情况下的命中率，对 GC 无优化，且 Feature 服务的 Cache 是分钟固定窗口失效，机器内存容量远大于窗口内的缓存 value 之和，因此不需要用到更好的淘汰机制，而且 Feature 服务本次更换 cahce 要解决的是缓存中对象数量太多，导致的 GC 问题，因此不考虑这两种；
      - groupcache 是一个 LRU Cache，且功能较重，Feature 服务只需要一个本地 Cache 库，不需要用到这些特性，因此不考虑这个 Cahce；
      - fastcache 最大的问题是不支持 TTL，这个是 Feature 服务所不能接受的，因此不考虑这个Cahce；
      - go-cache 类似于 Feature 服务中的 beego/cache 库，最简单的 Cache 库，对 GC 无优化，且 Feature 服务的 value 本身就为 string 类型，不会引入序列化开销，且可以通过不安全的方式实现 string 与 []byte 之间 0 开销转换；
      - freecache、bigcache 比较适合 Feature 服务，freecache 的优势在于近 LRU 的淘汰，并且可以对每个 Key 设置 TTL，但 Feature 服务内存空间足够无需进行缓存淘汰，且 key 名中自带分钟级时间戳，key 有效期都为 1min，因此无需使用 freecache；
      - bigcache 相对于 freecache 的优势之一是您不需要提前知道缓存的大小，因为当 bigcache 已满时，它可以为新条目分配额外的内存，而不是像 freecache 当前那样覆盖现有的。摘自：bigcache[7]
- [Go AST](https://mp.weixin.qq.com/s/pCcNtUykXAwb-BN_prPGpA)
  - uber-go 的 gopatch 也非常强大，假如你的代码有很多 go func 开启的 goroutine, 你想批量加入 recover 逻辑，如果数据特别多人工加很麻烦，这时可以用 gopatcher
    ```go
    var patchTemplateString = `@@
    @@
    + import "runtime/debug"
    + import "{{ Logger }}"
    + import "{{ Statsd }}"
    
    go func(...) {
    +    defer func(){
    +        if err := recover(); err != nil {
    +            statsd.Count1("{{ StatsdTag }}", "{{ FileName }}")
    +            logging.Error("{{ LoggerTag }}", "{{ FileName }} recover from panic, err=%+v, stack=%v", err, string(debug.Stack()))
    +        }
    +    }()
    ...
    }()
    `
    ```
  - 大部分 linter 工具都是用 go ast 实现的，比如对于大写的 Public 函数，如果没有注释报错
    ```go
    // BuildArgs write a
    func BuildArgs() {
        var a int
        a = a + bbb.c
        return a
    }
    
    29  .  .  1: *ast.FuncDecl {
    30  .  .  .  Doc: *ast.CommentGroup {
    31  .  .  .  .  List: []*ast.Comment (len = 1) {
    32  .  .  .  .  .  0: *ast.Comment {
    33  .  .  .  .  .  .  Slash: foo:7:1
    34  .  .  .  .  .  .  Text: "// BuildArgs write a"
    35  .  .  .  .  .  }
    36  .  .  .  .  }
    37  .  .  .  }
    38  .  .  .  Recv: nil
    39  .  .  .  Name: *ast.Ident {
    40  .  .  .  .  NamePos: foo:8:6
    41  .  .  .  .  Name: "BuildArgs"
    42  .  .  .  .  Obj: *ast.Object {
    43  .  .  .  .  .  Kind: func
    44  .  .  .  .  .  Name: "BuildArgs"
    45  .  .  .  .  .  Decl: *(obj @ 29)
    46  .  .  .  .  .  Data: nil
    47  .  .  .  .  .  Type: nil
    48  .  .  .  .  }
    49  .  .  .  }
    ```
    - linter 只需要检查 FuncDecl 的 Name 如果是可导出的，同时 Doc.CommentGroup 不存在，或是注释不以函数名开头，报错即可
- [Uber工程师对真实世界并发问题的研究](https://mp.weixin.qq.com/s/AOTSLMXpdD9R7YuHdDQCXA)
  - [Doc](https://arxiv.org/abs/2204.00764)
  - Go标准库常见并发原语不允许在使用后Copy, go vet也能检查出来. 比如下面的代码，两个goroutine想共享mutex,需要传递&mutex,而不是mutex。
     ```go
      var a int
      // CriticalSection receives a copy of mutex .
      func CriticalSection ( m sync . Mutex ) {
      m.Lock ()
       a ++
      m.Unlock ()
      }
      func main () {
      mutex := sync . Mutex {}
      // passes a copy of m to A .
      go CriticalSection ( mutex )
      go CriticalSection ( mutex )
      }
     ```
  - 混用消息传递和共享内存两种并发方式. 如果context因为超时或者主动cancel被取消的话，Start中的goroutine中的f.ch <- 1可能会被永远阻塞，导致goroutine泄露。
    ```go
    func ( f * Future ) Start () {
     go func () {
     resp , err := f.f () // invoke a registered function
      f.response = resp
      f.err = err
      f.ch <- 1 // may block forever !
     }()
     }
     func ( f * Future ) Wait ( ctx context . Context ) error {
     select {
     case <-f.ch :
     return nil
     case <- ctx.Done () :
      f.err = ErrCancelled
     return ErrCancelled
     }
    ```
- [Go 处理大数组：使用 for range 还是 for 循环](https://mp.weixin.qq.com/s/dHjGn3gwxnYqlrsUkgc0dg)
  - 既然 for range 使用的是副本数据，那 for range 会比经典的 for 循环消耗更多的资源并且性能更差吗
  - 我们使用 for 循环和 for range 分别遍历一个包含 10 万个 int 类型元素的数组。
    - for range 的确会稍劣于 for 循环，当然这其中包含了编译器级别优化的结果（通常是静态单赋值，或者 SSA 链接）
    - 让我们关闭优化开关 `go test -c -gcflags '-N -l'`，再次运行压力测试. 两种循环的性能都明显下降， for range 下降得更为明显，性能也更加比经典 for 循环差。
  - 遍历结构体数组
    - 不管是什么类型的结构体元素数组，经典的 for 循环遍历的性能比较一致，但是 for range 的遍历性能会随着结构字段数量的增加而降低。
  - 由于在 Go 中切片的底层都是通过数组来存储数据，尽管有 for range 的副本复制问题，但是切片副本指向的底层数组与原切片是一致的。这意味着，当我们将数组通过切片代替后，不管是通过 for range 或者 for 循环均能得到一致的稳定的遍历性能
- [为什么 Go 用起来会难受](https://mp.weixin.qq.com/s/qDJAOYy6kYrELHuL-MHlrg)
  - 浅拷贝和泄露
    - 我们经常要用到 slice、map 等基础类型。但有一个比较麻烦的点，就是会涉及到浅拷贝。
  - nil 接口不是 nil
    - 强行将一段 Go 程序的变量值赋为 nil，并进行 nil 与 nil 的判断
  - 垃圾回收
    - 垃圾回收唯一的可调节的是 GC 频率，可以通过 GOGC 变量设置初始垃圾收集器的目标百分比值
    - GOGC 的值设置的越大，GC 的频率越低，但每次最终所触发到 GC 的堆内存也会更大。
  - [依赖管理](https://xargin.com/go-mod-hurt-gophers/)
- [Go 语言中的一些非常规优化](https://xargin.com/unusual-opt-in-go/)
  - 网络方面
    - 当前 Go 的网络抽象在效率上有些低，每一个连接至少要有一个 goroutine 来维护，有些协议实现可能有两个。因此 goroutine 总数 = 连接数 * 1 or 连接数 * 2。当连接数超过 10w 时，goroutine 栈本身带来的内存消耗就有几个 GB。
    - 大量的 goroutine 也会给调度和 GC 均带来很大压力
    - 由于用户的 syscall.EpollWait 是运行在一个没有任何优先级的 goroutine 中，当 CPU idle 较低时，系统整体的延迟不可控，比标准库的延迟还要高很多。
- [memory ballast 和 auto gc tuner 成为历史](https://mp.weixin.qq.com/s/ry9HpZqFt4nZD_BZYLUBeA)
  - [memory ballast](https://blog.twitch.tv/en/2019/04/10/go-memory-ballast-how-i-learnt-to-stop-worrying-and-love-the-heap/)
    - 通过在堆上分配一个巨大的对象(一般是几个 GB)来欺骗 GOGC，让 Go 能够尽量充分地利用堆空间来减少 GC 触发的频率。
  - [auto gc tuner](https://eng.uber.com/how-we-saved-70k-cores-across-30-mission-critical-services/)
    - 设定程序的内存占用阈值，通过 GC 期间对用户 finalizer 函数的回调来达成每次 GC 触发时都动态设置 GOGC，以使应用使用内存与目标逐渐趋近的目的。
         ```go
         type finalizerRef struct {
             parent *finalizer
         }
         
         type finalizer struct {
             ch  chan time.Time
             ref *finalizerRef
         }
         
         func finalizerHandler(f *finalizerRef) {
             select {
             case f.parent.ch <- time.Time{}:
             default:
             }
             runtime.SetFinalizer(f, finalizerHandler)
         }
         
         func NewTicker() (*finalizer) {
             f := &finalizer{ch: make(chan time.Time, 1)}
             f.ref = &finalizerRef{parent: f}
             runtime.SetFinalizer(f.ref, finalizerHandler)
             f.ref = nil
             return f
         }
         ```
      - [gogctuner sample](https://github.com/cch123/gogctuner)
  - memory ballast 和 auto gc tuner 是为了解决 Go 在没有充分利用内存的情况下，频繁触发 GC 导致 GC 占用 CPU 过高的优化手段。
  - 在 Go 1.19 中新增加的 debug.SetMemoryLimit 从根本上解决了这个问题，可以直接把 memory ballast 和 gc tuner 丢进垃圾桶了。
     - OOM 的场景：SetMemoryLimit 设置内存上限即可
     - GC 触发频率高的场景：SetMemoryLimit 设置内存上限，GOGC = off
  - 1.19 另一个更新也比较有意思，按栈统计确定初始大小 会在 GC 期间将扫描过的 goroutine 和其栈大小求一个平均值，下次 newproc 创建 goroutine 的时候，就用这个平均值来创建新的 goroutine，而不是以前的 2KB
- [怎么使用 SetMemoryLimit](https://mp.weixin.qq.com/s/EIuM073G7VV1rIsnTXWyEw)
  - `GOMEMLIMIT=10737418240 GOGC=off GODEBUG=gctrace=1 ./soft_memory_limit -depth=21`
  - 通过SetMemoryLimit设置一个较大的值，再加上 GOGC=off，可以实现ballast的效果
  - 但是在没有关闭GOGC的情况下，还是有可能会触发很多次的GC,影响性能，这个时候还得GOGC Tuner调优，减少触达MemoryLimit之前的GC次数。
- [聊聊两个Go即将过时的GC优化策略](https://www.luozhiyun.com/archives/680)
  - GC
    - GC几个阶段
      - sweep termination（清理终止）：会触发 STW ，所有的 P（处理器） 都会进入 safe-point（安全点）；
      - the mark phase（标记阶段）：恢复程序执行，GC 执行根节点的标记，这包括扫描所有的栈、全局对象以及不在堆中的运行时数据结构；
      - mark termination（标记终止）：触发 STW，扭转 GC 状态，关闭 GC 工作线程等；
      - the sweep phase（清理阶段）：恢复程序执行，后台并发清理所有的内存管理单元
    - 由于标记阶段是要从根节点对堆进行遍历，对存活的对象进行着色标记，因此标记的时间和目前存活的对象有关，而不是与堆的大小有关，也就是堆上的垃圾对象并不会增加 GC 的标记时间
    - 在什么时候会触发 GC
      - 监控线程 runtime.sysmon 定时调用；
        - 后台运行一个线程定时执行 runtime.sysmon 函数，这个函数主要用来检查死锁、运行计时器、调度抢占、以及 GC 等。
      - 手动调用 runtime.GC 函数进行垃圾收集；
        - 会获取当前的 GC 循环次数，然后设值为 gcTriggerCycle 模式调用 gcStart 进行循环
      - 申请内存时 runtime.mallocgc 会根据堆大小判断是否调用；
      - 上面这三个触发 GC 的地方最终都会调用 gcStart 执行 GC，但是在执行 GC 之前一定会先判断这次调用是否应该被执行，并不是每次调用都一定会执行 GC， 这个时候就要说一下 runtime.gcTrigger中的 test 函数，这个函数负责校验本次 GC 是否应该被执行。
      - runtime.gcTrigger中的 test 函数最终会根据自己的三个策略
        - gcTriggerHeap：按堆大小触发，堆大小和上次 GC 时相比达到一定阈值则触发；
          - 触发 GC 的时机是由上次 GC 时的堆内存大小，和当前堆内存大小值对比的增长率来决定的，这个增长率就是环境变量 GOGC，默认是 100
          - 可以通过 GODEBUG=gctrace=1,gcpacertrace=1 打印出来
        - gcTriggerTime：按时间触发，如果超过 forcegcperiod（默认2分钟） 时间没有被 GC，那么会执行GC；
        - gcTriggerCycle：没有开启垃圾收集，则触发新的循环；
    - 如果收集器确定它需要减慢分配速度，它将招募应用程序 Goroutines 来协助标记工作。这称为 Mark assist 标记辅助。这也就是为什么在分配内存的时候还需要判断要不要执行 mallocgc 进行 GC
    - 在进行 Mark assist 的时候 Goroutines 会暂停当前的工作，进行辅助标记工作，这会导致当前 Goroutines 工作的任务有一些延迟。
  - Go Memory Ballast
  - Go GC Tuner
  - Soft Memory Limit
    - 通过内置的 debug.SetMemoryLimit 函数我们可以调整触发 GC 的堆内存目标值，从而减少 GC 次数，降低GC 时 CPU 占用的目的。
- [简单的 redis get 为什么也会有秒级的延迟](https://mp.weixin.qq.com/s/GtVtgTBZxW2ecs7QyDTcEg)
  - redis server 6.0 给 client 返回的 command 命令的响应，在 go-redis/redis v6 版本 parse 会出错：degradation for not caching command response[1]。
  - 每次 parse 都出错，那自然每次 once.Do 都会进 slow path 了，redis cluster 的 client 是全局公用，所以这里的锁是个全局锁，并且锁内有较慢的网络调用
- [Go 实现 REST API 部分更新](http://russellluo.com/)
  - 使用指针 [sample](https://go.dev/play/p/XaTbJkJOAk4)
  - 客户端维护的 FieldMask 
     ```go
     type Person struct {
         Name    string  `json:"name"`
         Age     int     `json:"age"`
         Address Address `json:"address"`
     }
     
     type UpdatePersonRequest struct {
         Person    Person `json:"person"`
         FieldMask string `json:"field_mask"`
     }
     ```
  - 改用 JSON Patch
     ```go
     PATCH /people/1 HTTP/1.1
     
     [
         { 
             "op": "replace", 
             "path": "/age", 
             "value": 25
         },
         {
             "op": "replace",
             "path": "/address/city",
             "value": "Guangzhou"
         }
     ]
     ```
  - 服务端维护的 FieldMask
    - Go 的 JSON 反序列化其实有两种：
      - 将 JSON 反序列化为结构体（优势：操作直观方便；不足：有零值问题）
      - 将 JSON 反序列化为 map[string]interface{}（优势：能够准确表达 JSON 中有无特定字段；不足：操作不够直观方便）
    - 如果我们只是把 map[string]interface{} 作为一个反序列化的中间结果呢？比如：
      - 首先将 JSON 反序列化为 map[string]interface{}
      - 然后用 map[string]interface{} 来充当（服务端维护的）FieldMask
      - 最后将 map[string]interface{} 解析为结构体（幸运的是，已经有现成的库 [mapstructure](https://github.com/mitchellh/mapstructure) 可以做到！）
    - [sample](https://github.com/RussellLuo/fieldmask/blob/master/example_partial_update_test.go)
- [noCopy 是什么机制](https://mp.weixin.qq.com/s/ptVPpsWsQkgmRZmlMlFfdA)
  - 当我们用 go vet 检查静态问题的时候，你是否遇到 noCopy 相关的错误  assignment copies lock value to m: sync.Mutex
  - 为什么锁不能拷贝
    - 变量资源本身带状态且操作要配套的不能拷贝。
    - 针对需要配套操作的变量类型，基本上都会要求 noCopy 的，否则拷贝出来就乱套了
  - Go 里面通过实现一个 noCopy 的结构体，然后嵌入这个结构体就能让 go vet 检查出来。
  - Mutex Lock，Cond，Pool，WaitGroup 。这些资源都严格要求操作要配套
- [A Guide to the Go Garbage Collector](https://colobu.com/2022/07/16/A-Guide-to-the-Go-Garbage-Collector/)
- [提升Go语言开发效率的小技巧](https://mp.weixin.qq.com/s/cdc1NCSkvAU4Urk2wVFDMw)
  - 声明不定长数组 - 使用...操作符声明数组时，你只管填充元素值，其他的交给编译器自己去搞就好了
    ```go
    a := [...]int{1, 3, 5} // 数组长度是3，等同于 a := [3]{1, 3, 5}
    
    a := [...]int{1: 20, 999: 10} // 数组长度是100, 下标1的元素值是20，下标999的元素值是10，其他元素值都是0
    ```
  - init函数
    - 从当前包开始，如果当前包包含多个依赖包，则先初始化依赖包，层层递归初始化各个包，在每一个包中，按照源文件的字典序从前往后执行，每一个源文件中，优先初始化常量、变量，最后初始化init函数，当出现多个init函数时，则按照顺序从前往后依次执行，每一个包完成加载后，递归返回，最后在初始化当前包
    - init函数实现了sync.Once，无论包被导入多少次，init函数只会被执行一次，所以使用init可以应用在服务注册、中间件初始化、实现单例模式








