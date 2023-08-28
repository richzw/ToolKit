
Memory Leak 
------------

- Goroutine 泄漏

  概述
  - Goroutine 内正在进行 channel/mutex 等读写操作，但由于逻辑问题，某些情况下会被一直阻塞。 
  - Goroutine 内的业务逻辑进入死循环，资源一直无法释放。
  - Goroutine 内的业务逻辑进入长时间等待，有不断新增的 Goroutine 进入等待

  详情

  - channel 使用不当
    - 发送不接收
      ```go
        func main() {
            for i := 0; i < 4; i++ {
                queryAll()
                fmt.Printf("goroutines: %d\n", runtime.NumGoroutine())
            }
        }
        
        func queryAll() int {
            ch := make(chan int)
            for i := 0; i < 3; i++ {
                go func() { ch <- query() }()
            }
            return <-ch
        }
        
        func query() int {
            n := rand.Intn(100)
            time.Sleep(time.Duration(n) * time.Millisecond)
            return n
        }
      ```
    - 接收不发送
      ```go
        func main() {
            defer func() {
                fmt.Println("goroutines: ", runtime.NumGoroutine())
            }()
        
            var ch chan struct{}
            go func() {
                ch <- struct{}{}
            }()
            
            time.Sleep(time.Second)
        }
        ```
    - nil channel
    ```go
        func main() {
            defer func() {
                fmt.Println("goroutines: ", runtime.NumGoroutine())
            }()
        
            var ch chan int
            go func() {
                <-ch
            }()
            
            time.Sleep(time.Second)
        }
        ```
    
  - 慢等待
    ```go
    func main() {
        for {
        go func() {
            _, err := http.Get("https://www.xxx.com/")
            if err != nil {
                fmt.Printf("http.Get err: %v\n", err)
            }
        // do something...
        }()
    
        time.Sleep(time.Second * 1)
        fmt.Println("goroutines: ", runtime.NumGoroutine())
        }
    }
    ```
    第三方接口，有时候会很慢，久久不返回响应结果。恰好，Go 语言中默认的 `http.Client `是没有设置超时时间的。
    
    在 Go 工程中，我们一般建议至少对 `http.Client` 设置超时时间：

    ```go
    httpClient := http.Client{
        Timeout: time.Second * 15,
    }
    ```
    
  - 互斥锁忘记解锁
    ```go
    var mutex sync.Mutex
    for i := 0; i < 10; i++ {
        go func() {
            mutex.Lock()
            total += 1
        }()
    }
    ```
    第一个互斥锁 `sync.Mutex` 加锁了，但是他可能在处理业务逻辑，又或是忘记 `Unlock` 了。
    因此导致后面的所有 `sync.Mutex` 想加锁，却因未释放又都阻塞住了

  - 同步锁使用不当
    ```go
    func handle(v int) {
        var wg sync.WaitGroup
        wg.Add(5)
        for i := 0; i < v; i++ {
            fmt.Println("脑子进煎鱼了")
            wg.Done()
        }
        wg.Wait()
    }
    ```
    由于 `wg.Add` 的数量与 `wg.Done` 数量并不匹配，因此在调用 `wg.Wait` 方法后一直阻塞等待。

  - time.After引起OOM
    ```go
    ch := make(chan int, 10)
    go func() {
        in := 1
        for {
            in++
            ch <- in
        }
    }()
    
    for {
        select {
        case _ = <-ch:
            // do something...
            continue
        case <-time.After(3 * time.Minute):
            fmt.Printf("now %d", time.Now().Unix())
        }
    }
    ```
    被遗弃的 `time.After` 定时任务还是在时间堆里面，定时任务未到期之前，是不会被 GC 清理的。
    在 `for` 循环里**不要**使用 `select + time.After` 的组合
    
    另外，从select 语义来看，每次执行select的时候，`time.After`总会执行一次。[参见](https://jishuin.proginn.com/p/763bfbd651ad)
    ```go
    func main() {
     ch := make(chan int)
     go func() {
      select {
      case ch <- getVal(1):
       fmt.Println("in first case")
      case ch <- getVal(2):
       fmt.Println("in second case")
      default:
       fmt.Println("default")
      }
     }()
    
     fmt.Println("The val:", <-ch)
    }
    
    func getVal(i int) int {
     fmt.Println("getVal, i=", i)
     return i
    }
    ```
    
    __Solution__
    ```go
    timer := time.NewTimer(3 * time.Minute)
    defer timer.Stop()
    
    ...
    for {
        select {
        ...
        case <-timer.C:
            fmt.Printf("now %d", time.Now().Unix())
        }
    }
    ```



- 排查方法

  在业务服务的运行场景中，Goroutine 内导致的泄露，大多数处于生产、测试环境，因此更多的是使用 `PProf`：

  ```go
  import (
  "net/http"
  _ "net/http/pprof"
  )
  
  http.ListenAndServe("localhost:6060", nil))
  ```

  只要我们调用 `http://localhost:6060/debug/pprof/goroutine?debug=1`，PProf 会返回所有带有堆栈跟踪的 Goroutine 列表。

  case 2:
  我们直接找一台能ping通容器并且装了golang的机器，直接用下面的命令看看当前服务的内存分配情况：
  ```shell
  $ go tool pprof -inuse_space http://ip:amdin_port/debug/pprof/heap
  ```
  `-inuse_space`参数就是当前服务使用的内存情况，还有一个`-alloc_space`参数是指服务启动以来总共分配的内存情况. 进入交互界面后我们用top命令看下当前占用内存最高的部分

  再抓一下当前内存分配的详情：
  ```shell
  $ wget http://ip:admin_port/debug/pprof/heap?debug=1
  ```

  这个命令其实就是把当前内存分配的详情文件抓了下来，本地会生成一个叫heap?debug=1的文件，看一看服务内存分配的具体情况

  说不定是goroutine泄漏呢！于是赶在内存暴涨结束之际，又火速敲下以下命令：
  ```shell
  $ wget http://ip:admin_port/debug/pprof/goroutine?debug=1
  $ wget http://ip:admin_port/debug/pprof/goroutine?debug=2
  ```

  debug=1就是获取服务当前goroutine的数目和大致信息，debug=2获取服务当前goroutine的详细信息

  再看看服务线程挂的子线程有多少：
  ```shell
  $ ps mp 3030923 -o THREAD,tid | wc -l
  ```

  - 原因1，channel堵塞导致goroutine泄漏
  ```go
   ch := make(chan struct{})
  
    go func () {
      wg.Wait()
      ch <- struct{}{}
    }()
  
    // 接收完成或者超时
    select {
    case <- ch:
      return
    case <- time.After(time.Second * 10):
      return
    }
  ```

  - 原因2 http超时导致goroutine泄漏

  ```go
  // 默认的httpClient
  var DefaultCli *http.Client
  
  func init() {
    DefaultCli = &http.Client{
      // --> should set timeout
      Transport: &http.Transport{
        DialContext: (&net.Dialer{
          Timeout:   2 * time.Second,
          KeepAlive: 30 * time.Second,
        }).DialContext,
    }}
  }
  ```
  
  - 原因3 go新版本内存管理问题

    Go1.12中使用的新的MADV_FREE模式，这个模式会更有效的释放无用的内存，但可能会让RSS增高
  ```shell
    GODEBUG=madvdontneed=1
  ```
  
- [Deep Dive goroutine leak](https://mourya-g9.medium.com/deep-dive-on-goroutine-leaks-and-best-practices-to-avoid-them-a35021383f64)
  - an unbuffered channel blocks write to channel until consumer consumes the message from that channel. 
  - there are packages like https://github.com/uber-go/goleak which helps you to find goroutine leaks
  
- [切片使用不当会造成内存泄漏](https://gocn.vip/topics/17387)
  
  `100 Go Mistackes：How to Avoid Them`
  
  - 因切片容量而导致内存泄漏
    ```go
    func consumeMessages() {
        for {
            msg := receiveMessage() ①
            storeMessageType(getMessageType(msg)) ②
            // Do something with msg
        }
    }
    
    func getMessageType(msg []byte) []byte { ③
        return msg[:5]
    }
    ```
    我们使用 **msg[:5]** 对 msg 进行切分操作时，实际上是创建了一个长度为 5 的新切片。因为新切片和原切片共享同一个底层数据。所以它的容量依然是跟源切片 msg 的容量一样。即使实际的 msg 不再被引用，但剩余的元素依然在内存中

    我们该如何解决呢？最简单的方法就是在 getMessageType 函数内部将消息类型拷贝到一个新的切片上，来替代对 msg 进行切分
    ```go
    func getMessageType(msg []byte) []byte {
        msgType := make([]byte, 5)
        copy(msgType, msg)
        return msgType
    }
    ```
  - 因指针类型导致内存泄露
    ```go
    func keepFirstElementOnly(ids []string) []string {
        return string[:1]
    }
    ```
    如果我们传递给 keepFirstElementOnly 函数一个有 100 个字符串的切片，那么，剩下的 99 个字符串会被 GC 回收吗？在该例子中是**会被回收的**。容量将保持为 100 个元素，但会收集剩余的 99 个字符串将减少所消耗的内存

    我们通过指针的方式传递元素，看看会发生什么：
    ```go
    func keepFirstElementOnly(ids []*string) []*string {
       return customers[:1]
    }
    ```
    
    现在剩余的 99 个元素还会被 GC 回收吗？在该示例中是不可以的。
    
    规则如下：**若切片的元素类型是指针或带指针字段的结构体，那么元素将不会被 GC 回收**。如果我们想返回一个容量为 1 的切片，我们可以使用 copy 函数或使用满切片表达式（s[:1:1]）。另外，如果我们想保持容量，则需要将剩余的元素填充为 nil：
    ```go
    func keepFirstElementOnly(ids []*string) []*string {
        for i := 1; i < len(ids); i++ {
            ids[i] = nil
        }
        return ids[:1]
    }
    ```
    
    对于剩余所有的元素，我们手动的填充为 nil。在本示例中，我们会返回一个具有和输入参数切片的容量大小一致的切片，但剩下的 *string 类型的元素会被 GC 自动回收。
- [定位并修复 Go 中的内存泄露](https://mp.weixin.qq.com/s/zcQxmqN0LT9L0qQsp2nypg)
  [Source](https://dev.to/googlecloud/finding-and-fixing-memory-leaks-in-go-1k1h)
  - Google Cloud Go 客户端库[1] 通常在后台使用 gRPC 来连接 Google Cloud API。创建 API 客户端时，库会初始化与 API 的连接，然后保持该连接处于打开状态，直到你调用 Client.Close。
    ````go
    client, err := api.NewClient()
    // Check err.
    defer client.Close()
    ````
  - 如果在应该 Close 的时候不 Close client 会发生什么呢？
    - 通过向服务器添加 pprof.Index 处理程序开始调试： `mux.HandleFunc("/debug/pprof/", pprof.Index)`
    - 然后向服务器发送一些请求：
      ```go
      for i in {1..5}; do
        curl --header "Content-Type: application/json" --request POST --data '{"name": "HelloHTTP", "type": "testing", "location": "us-central1"}' localhost:8080/v0/cron
        echo " -- $i"
      done
      ```
    - 收集了一些初始pprof数据： `curl http://localhost:8080/debug/pprof/heap > heap.0.pprof`
      ```shell
      $ go tool pprof heap.0.pprof
      (pprof) top10
      ```
    - google.golang.org/grpc/internal/transport.newBufWriter使用大量内存真的很突出！这是泄漏与什么相关的第一个迹象：gRPC
    - 使用grep，我们可以获得包含NewClient样式调用的所有文件的列表，然后将该列表传递给另一个调用grep以仅列出不包含 Close 的文件，同时忽略测试文件：
      `$ grep -L Close $(grep -El 'New[^(]*Client' **/*.go) | grep -v test`
    - `$ grep -L Close $(grep -El 'New[^(]*Client' **/*.go) | grep -v test | xargs sed -i '/New[^(]*Client/,/}/s/}/}\ndefer client.Close()/'`
- [glibc导致的堆外内存泄露](https://mp.weixin.qq.com/s/55slokngVRgqEav6c3TxOA)
- [一次神奇的崩溃](https://mp.weixin.qq.com/s/vMlK7oIQH62VV6qHSPHnQQ)
  - 一个崩溃案例的分析过程。回顾了C++多态和类内存布局、pc指针与芯片异常处理、内存屏障的相关知识。
  - 编译器进行了reorder优化，我们就可以使用内存屏障禁止编译器相关优化，可以在addObserver代码中插入一行表示内存屏障的汇编`__asm__ __volatile__("":::"memory")`




  

