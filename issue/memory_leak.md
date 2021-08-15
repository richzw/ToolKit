
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
    
  -  互斥锁忘记解锁
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



