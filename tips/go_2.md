
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
  - Go 显然是不支持可重入互斥锁的
    - Russ Cox 于 2010 年在《Experimenting with GO》就给出了答复，认为递归（又称：重入）互斥是个坏主意，这个设计并不好。
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









