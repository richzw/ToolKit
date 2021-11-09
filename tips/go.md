
- [Applying Modern Go Concurrency Patterns to Data Pipelines](https://medium.com/amboss/applying-modern-go-concurrency-patterns-to-data-pipelines-b3b5327908d4)
  - A Simple Pipeline
    ```go
    func producer(strings []string) (<-chan string, error) {
        outChannel := make(chan string)
    
        for _, s := range strings {
            outChannel <- s
        }
    
        return outChannel, nil
    }
    
    func sink(values <-chan string) {
        for value := range values {
            log.Println(value)
        }
    }
    
    func main() {
        source := []string{"foo", "bar", "bax"}
    
        outputChannel, err := producer(source)
        if err != nil {
            log.Fatal(err)
        }
    
        sink(outputChannel)
    }
    ```
    you run this with go run main.go you'll see a deadlock
    - The channel returned by producer is not buffered, meaning you can only send values to the channel if someone is receiving values on the other end. But since `sink` is called later in the program, there is no receiver at the point where `outChannel <- s` is called, causing the deadlock.
    - fix it
      - either making the channel buffered, in which case the deadlock will occur once the buffer is full
      - or by running the producer in a Go routine. 
      - whoever creates the channel is also in charge of closing it.
  - Graceful Shutdown With Context
    - with context
    ```go
    func producer(ctx context.Context, strings []string) (<-chan string, error) {
         outChannel := make(chan string)
     
         go func() {
             defer close(outChannel)
    
             for _, s := range strings {
                 time.Sleep(time.Second * 3)
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- s
                }
             }
         }()
     
         return outChannel, nil
     }
     
    
    func sink(ctx context.Context, values <-chan string) {
        for {
            select {
            case <-ctx.Done():
                log.Print(ctx.Err().Error())
                return
            case val, ok := <-values:
    +			log.Print(val)  // for debug
                if ok {
                    log.Println(val)
                }
            }
         }
     }
     
     func main() {
         source := []string{"foo", "bar", "bax"}
     
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()
    
        go func() {
            time.Sleep(time.Second * 5)
            cancel()
        }()
    
        outputChannel, err := producer(ctx, source)
         if err != nil {
             log.Fatal(err)
         }
     
    
        sink(ctx, outputChannel)
     }
    ```
    Issues:
    - This will flood our terminal with empty log messages, like this: 2021/09/08 12:29:30. Apparently the for loop in sink keeps running forever
    
    [Reason](https://golang.org/ref/spec#Receive_operator)
    - A receive operation on a closed channel can always proceed immediately, yielding the element type’s zero value after any previously sent values have been received.

    Fix it
    - The value of ok is true if the value received was delivered by a successful send operation to the channel, or false if it is a zero value generated because the channel is closed and empty.
     ```go
     func sink(ctx context.Context, values <-chan string) {
     
              case val, ok := <-values:
                  if ok {
                      log.Println(val)
                 } else {
                     return
                  }
              }
          }
     ```
  - Adding Parallelism with Fan-Out and Fan-In
    - sending values to a closed channel is a panic
    ```go
    func producer(ctx context.Context, strings []string) (<-chan string, error) {
        outChannel := make(chan string)
    
        go func() {
            defer close(outChannel)
    
            for _, s := range strings {
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- s
                }
            }
        }()
    
        return outChannel, nil
    }
    
    func transformToLower(ctx context.Context, values <-chan string) (<-chan string, error) {
        outChannel := make(chan string)
    
        go func() {
            defer close(outChannel)
    
            for s := range values {
                time.Sleep(time.Second * 3)
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- strings.ToLower(s)
                }
            }
        }()
    
        return outChannel, nil
    }
    
    func transformToTitle(ctx context.Context, values <-chan string) (<-chan string, error) {
        outChannel := make(chan string)
    
        go func() {
            defer close(outChannel)
    
            for s := range values {
                time.Sleep(time.Second * 3)
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- strings.ToTitle(s)
                }
            }
        }()
    
        return outChannel, nil
    }
    
    func sink(ctx context.Context, values <-chan string) {
        for {
            select {
            case <-ctx.Done():
                log.Print(ctx.Err().Error())
                return
            case val, ok := <-values:
                if ok {
                    log.Println(val)
                } else {
                    return
                }
            }
        }
    }
    
    func main() {
        source := []string{"FOO", "BAR", "BAX"}
    
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()
    
        outputChannel, err := producer(ctx, source)
        if err != nil {
            log.Fatal(err)
        }
    
        stage1Channels := []<-chan string{}
    
        for i := 0; i < runtime.NumCPU(); i++ {
            lowerCaseChannel, err := transformToLower(ctx, outputChannel)
            if err != nil {
                log.Fatal(err)
            }
            stage1Channels = append(stage1Channels, lowerCaseChannel)
        }
    
        stage1Merged := mergeStringChans(ctx, stage1Channels...)
        stage2Channels := []<-chan string{}
    
        for i := 0; i < runtime.NumCPU(); i++ {
            titleCaseChannel, err := transformToTitle(ctx, stage1Merged)
            if err != nil {
                log.Fatal(err)
            }
            stage2Channels = append(stage2Channels, titleCaseChannel)
        }
    
        stage2Merged := mergeStringChans(ctx, stage2Channels...)
        sink(ctx, stage2Merged)
    }
    
    func mergeStringChans(ctx context.Context, cs ...<-chan string) <-chan string {
        var wg sync.WaitGroup
        out := make(chan string)
    
        output := func(c <-chan string) {
            defer wg.Done()
            for n := range c {
                select {
                case out <- n:
                case <-ctx.Done():
                    return
                }
            }
        }
    
        wg.Add(len(cs))
        for _, c := range cs {
            go output(c)
        }
    
        go func() {
            wg.Wait()
            close(out)
        }()
    
        return out
    }
    ```
  - Error Handling
    - The most common way of propagating errors that I’ve seen is through a separate error channel. Unlike the value channels that connect pipeline stages, the error channels are not passed to downstream stages.
  - Removing Boilerplate With Generics
  - Maximum Efficiency With Semaphores
    - What if our input list only had a single element in it? Then we only need a single Go routine, not NumCPU() Go routines. 
    - Instead of creating a fixed number of Go routines, we will range over the input channel. For every value we receive from it, we will spawn a Go routine (see the example in the semaphore package)
    
    ```go
    package tips
    
    import (
        "errors"
        "context"
        "log"
        "runtime"
        "strings"
        "time"
    
        "golang.org/x/sync/semaphore"
    )
    
    func producer(ctx context.Context, strings []string) (<-chan string, error) {
        outChannel := make(chan string)
    
        go func() {
            defer close(outChannel)
    
            for _, s := range strings {
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- s
                }
            }
        }()
    
        return outChannel, nil
    }
    
    func sink(ctx context.Context, cancelFunc context.CancelFunc, values <-chan string, errors <-chan error) {
        for {
            select {
            case <-ctx.Done():
                log.Print(ctx.Err().Error())
                return
            case err := <-errors:
                if err != nil {
                    log.Println("error: ", err.Error())
                    cancelFunc()
                }
            case val, ok := <-values:
                if ok {
                    log.Printf("sink: %s", val)
                } else {
                    log.Print("done")
                    return
                }
            }
        }
    }
    
    func step[In any, Out any](
        ctx context.Context,
        inputChannel <-chan In,
        outputChannel chan Out,
        errorChannel chan error,
        fn func(In) (Out, error),
    ) {
        defer close(outputChannel)
    
        limit := runtime.NumCPU()
        sem1 := semaphore.NewWeighted(limit)
    
        for s := range inputChannel {
            select {
            case <-ctx.Done():
                log.Print("1 abort")
                break
            default:
            }
    
            if err := sem1.Acquire(ctx, 1); err != nil {
                log.Printf("Failed to acquire semaphore: %v", err)
                break
            }
    
            go func(s In) {
                defer sem1.Release(1)
                time.Sleep(time.Second * 3)
    
                result, err := fn(s)
                if err != nil {
                    errorChannel <- err
                } else {
                    outputChannel <- result
                }
            }(s)
        }
    
        if err := sem1.Acquire(ctx, limit); err != nil {
            log.Printf("Failed to acquire semaphore: %v", err)
        }
    }
    
    func main() {
        source := []string{"FOO", "BAR", "BAX"}
    
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()
    
        readStream, err := producer(ctx, source)
        if err != nil {
            log.Fatal(err)
        }
    
        stage1 := make(chan string)
        errorChannel := make(chan error)
    
        transformA := func(s string) (string, error) {
            return strings.ToLower(s), nil
        }
    
        go func() {
            step(ctx, readStream, stage1, errorChannel, transformA)
        }()
    
        stage2 := make(chan string)
    
        transformB := func(s string) (string, error) {
            if s == "foo" {
                return "", errors.New("oh no")
            }
    
            return strings.Title(s), nil
        }
    
        go func() {
            step(ctx, stage1, stage2, errorChannel, transformB)
        }()
    
        sink(ctx, cancel, stage2, errorChannel)
    }
    ```

- [Handling 1 Million Requests per Minute with Go](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/)
  - The web handler would receive a JSON document that may contain a collection of many payloads that needed to be written to Amazon S3
  - Naive approach to GO routine
    ```go
    func payloadHandler(w http.ResponseWriter, r *http.Request) {
    
        if r.Method != "POST" {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }
    
        // Read the body into a string for json decoding
        var content = &PayloadCollection{}
        err := json.NewDecoder(io.LimitReader(r.Body, MaxLength)).Decode(&content)
        if err != nil {
            w.Header().Set("Content-Type", "application/json; charset=UTF-8")
            w.WriteHeader(http.StatusBadRequest)
            return
        }
    
        // Go through each payload and queue items individually to be posted to S3
        for _, payload := range content.Payloads {
            go payload.UploadToS3()   // <----- DON'T DO THIS
        }
    
        w.WriteHeader(http.StatusOK)
    }
    ```
  - Try again
    - create a buffered channel where we could queue up some jobs and upload them to S3
     ```go
     var Queue chan Payload
     
     func init() {
         Queue = make(chan Payload, MAX_QUEUE)
     }
     
     func payloadHandler(w http.ResponseWriter, r *http.Request) {
         ...
         // Go through each payload and queue items individually to be posted to S3
         for _, payload := range content.Payloads {
             Queue <- payload
         }
         ...
     }
     func StartProcessor() {
         for {
             select {
             case job := <-Queue:
                 job.payload.UploadToS3()  // <-- STILL NOT GOOD
             }
         }
     }
     ```
    Issue: since the rate of incoming requests were much larger than the ability of the single processor to upload to S3, our buffered channel was quickly reaching its limit and blocking the request handler ability to queue more items.
  - Better solution
    - create a 2-tier channel system, one for queuing jobs and another to control how many workers operate on the JobQueue concurrently.

    ```go
    var (
        MaxWorker = os.Getenv("MAX_WORKERS")
        MaxQueue  = os.Getenv("MAX_QUEUE")
    )
    
    // Job represents the job to be run
    type Job struct {
        Payload Payload
    }
    
    // A buffered channel that we can send work requests on.
    var JobQueue chan Job
    
    // Worker represents the worker that executes the job
    type Worker struct {
        WorkerPool  chan chan Job
        JobChannel  chan Job
        quit    	chan bool
    }
    
    func NewWorker(workerPool chan chan Job) Worker {
        return Worker{
            WorkerPool: workerPool,
            JobChannel: make(chan Job),
            quit:       make(chan bool)}
    }
    
    // Start method starts the run loop for the worker, listening for a quit channel in
    // case we need to stop it
    func (w Worker) Start() {
        go func() {
            for {
                // register the current worker into the worker queue.
                w.WorkerPool <- w.JobChannel
    
                select {
                case job := <-w.JobChannel:
                    // we have received a work request.
                    if err := job.Payload.UploadToS3(); err != nil {
                        log.Errorf("Error uploading to S3: %s", err.Error())
                    }
    
                case <-w.quit:
                    // we have received a signal to stop
                    return
                }
            }
        }()
    }
    
    // Stop signals the worker to stop listening for work requests.
    func (w Worker) Stop() {
        go func() {
            w.quit <- true
        }()
    }
    
    // handler
    func payloadHandler(w http.ResponseWriter, r *http.Request) {
    
        if r.Method != "POST" {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }
    
        // Read the body into a string for json decoding
        var content = &PayloadCollection{}
        err := json.NewDecoder(io.LimitReader(r.Body, MaxLength)).Decode(&content)
        if err != nil {
            w.Header().Set("Content-Type", "application/json; charset=UTF-8")
            w.WriteHeader(http.StatusBadRequest)
            return
        }
    
        // Go through each payload and queue items individually to be posted to S3
        for _, payload := range content.Payloads {
    
            // let's create a job with the payload
            work := Job{Payload: payload}
    
            // Push the work onto the queue.
            JobQueue <- work
        }
    
        w.WriteHeader(http.StatusOK)
    }
    
    dispatcher := NewDispatcher(MaxWorker)
    dispatcher.Run()
    
    type Dispatcher struct {
        // A pool of workers channels that are registered with the dispatcher
        WorkerPool chan chan Job
    }
    
    func NewDispatcher(maxWorkers int) *Dispatcher {
        pool := make(chan chan Job, maxWorkers)
        return &Dispatcher{WorkerPool: pool}
    }
    
    func (d *Dispatcher) Run() {
        // starting n number of workers
        for i := 0; i < d.maxWorkers; i++ {
            worker := NewWorker(d.pool)
            worker.Start()
        }
    
        go d.dispatch()
    }
    
    func (d *Dispatcher) dispatch() {
        for {
            select {
            case job := <-JobQueue:
                // a job request has been received
                go func(job Job) {
                    // try to obtain a worker job channel that is available.
                    // this will block until a worker is idle
                    jobChannel := <-d.WorkerPool
    
                    // dispatch the job to the worker job channel
                    jobChannel <- job
                }(job)
            }
        }
    }
    ```
- [Go timer 是如何被调度的](https://mp.weixin.qq.com/s/zy354p9MQq10fpTL20uuCA)
  - 概述
    - 不管用 **NewTimer**, **timer.After**，还是 **timer.AfterFun** 来初始化一个 timer, 这个 timer 最终都会加入到一个全局 timer 堆中，由 Go runtime 统一管理。
    - Go 1.9 版本之前，所有的计时器由全局唯一的四叉堆维护，协程间竞争激烈。
    - Go 1.10 - 1.13，全局使用 64 个四叉堆维护全部的计时器，没有本质解决 1.9 版本之前的问题
    - Go 1.14 版本之后，每个 P 单独维护一个四叉堆。
  - 原理
    - 四叉堆原理
      - 四叉树顾名思义最多有四个子节点，为了兼顾四叉树插、删除、重排速度，所以四个兄弟节点间并不要求其按触发早晚排序。
    - timer 是如何被调度的
      - 调用 NewTimer，timer.After, timer.AfterFunc 生产 timer, 加入对应的 P 的堆上。
      - 调用 timer.Stop, timer.Reset 改变对应的 timer 的状态。
      - GMP 在调度周期内中会调用 checkTimers ，遍历该 P 的 timer 堆上的元素，根据对应 timer 的状态执行真的操作。
    - timer 是如何加入到 timer 堆上的
      - 通过 NewTimer, time.After, timer.AfterFunc 初始化 timer 后，相关 timer 就会被放入到对应 p 的 timer 堆上。
      - timer 已经被标记为 timerRemoved，调用了 timer.Reset(d)，这个 timer 也会重新被加入到 p 的 timer 堆上
      - timer 还没到需要被执行的时间，被调用了 timer.Reset(d)，这个 timer 会被 GMP 调度探测到，先将该 timer 从 timer 堆上删除，然后重新加入到 timer 堆上
      - STW 时，runtime 会释放不再使用的 p 的资源，p.destroy()->timer.moveTimers，将不再被使用的 p 的 timers 上有效的 timer(状态是：timerWaiting，timerModifiedEarlier，timerModifiedLater) 都重新加入到一个新的 p 的 timer 上
    - Reset 时 timer 是如何被操作的
      - 被标记为 timerRemoved 的 timer，这种 timer 是已经从 timer 堆上删除了，但会重新设置被触发时间，加入到 timer 堆中
      - 等待被触发的 timer，在 Reset 函数中只会修改其触发时间和状态（timerModifiedEarlier或timerModifiedLater）。这个被修改状态的 timer 也同样会被重新加入到 timer堆上，不过是由 GMP 触发的，由 checkTimers 调用 adjusttimers 或者 runtimer 来执行的。
    - Stop 时 timer 是如何被操作的
      - time.Stop 为了让 timer 停止，不再被触发，也就是从 timer 堆上删除。不过 timer.Stop 并不会真正的从 p 的 timer 堆上删除 timer，只会将 timer 的状态修改为 timerDeleted。然后等待 GMP 触发的 adjusttimers 或者 runtimer 来执行。
    - Timer 是如何被真正执行的
      - timer 的真正执行者是 GMP。GMP 会在每个调度周期内，通过 runtime.checkTimers 调用 timer.runtimer(). timer.runtimer 会检查该 p 的 timer 堆上的所有 timer，判断这些 timer 是否能被触发。
      - 如果该 timer 能够被触发，会通过回调函数 sendTime 给 Timer 的 channel C 发一个当前时间，告诉我们这个 timer 已经被触发了。
      - 如果是 ticker 的话，被触发后，会计算下一次要触发的时间，重新将 timer 加入 timer 堆中。
  - Timer 使用中的坑
    - 错误创建很多 timer，导致资源浪费
      ```go
      func main() {
          for {
              // xxx 一些操作
              timeout := time.After(30 * time.Second)
              select {
              case <- someDone:
                  // do something
              case <-timeout:
                  return
              }
          }
      }
      ```
      因为 timer.After 底层是调用的 timer.NewTimer，NewTimer 生成 timer 后，会将 timer 放入到全局的 timer 堆中。
      for 会创建出来数以万计的 timer 放入到 timer 堆中，导致机器内存暴涨，同时不管 GMP 周期 checkTimers，还是插入新的 timer 都会疯狂遍历 timer 堆，导致 CPU 异常。
       ```go
       func main() {
           timer := time.NewTimer(time.Second * 5)    
           for {
               timer.Reset(time.Second * 5)
       
               select {
               case <- someDone:
                   // do something
               case <-timer.C:
                   return
               }
           }
       }
       ```
    - 程序阻塞，造成内存或者 goroutine 泄露
       ```go
       func main() {
           timer1 := time.NewTimer(2 * time.Second)
           <-timer1.C
           println("done")
       }
       ```
      只有等待 timer 超时 "done" 才会输出，原理很简单：程序阻塞在 <-timer1.C 上，一直等待 timer 被触发时，回调函数 time.sendTime 才会发送一个当前时间到 timer1.C 上，程序才能继续往下执行。
      ```go
      func main() {
          timer1 := time.NewTimer(2 * time.Second)
          go func() {
              timer1.Stop() // refer to doc
          }()
          <-timer1.C
      
          println("done")
      }
      ```
      程序就会一直死锁了，因为 timer1.Stop 并不会关闭 channel C，使程序一直阻塞在 timer1.C 上。

      Stop 的正确的使用方式：
       ```go
       func main() {
           timer1 := time.NewTimer(2 * time.Second)
           go func() {
               if !timer1.Stop() {
                   <-timer1.C
               }
           }()
       
           select {
           case <-timer1.C:
               fmt.Println("expired")
           default:
           }
           println("done")
       }
       ```
- [panic](https://mp.weixin.qq.com/s/sGdTVSRxqxIezdlEASB39A)
  - 什么时候会产生 panic
    - 主动方式：
      - 程序猿主动调用 panic 函数；
    - 被动的方式：
      - 编译器的隐藏代码触发
        ```go
        func divzero(a, b int) int {
            c := a/b
            return c
        }
        ```
        用 dlv 调试断点到 divzero 函数，然后执行 disassemble ，你就能看到秘密了
        编译器偷偷加上了一段 if/else 的判断逻辑，并且还给加了 runtime.panicdivide  的代码。
      - 内核发送给进程信号触发
      
        最典型的是非法地址访问，比如， nil 指针 访问会触发 panic
        
        在 Go 进程启动的时候会注册默认的信号处理程序（ sigtramp ）

        在 cpu 访问到 0 地址会触发 page fault 异常，这是一个非法地址，内核会发送 SIGSEGV 信号给进程，所以当收到 SIGSEGV 信号的时候，就会让 sigtramp 函数来处理，最终调用到 panic 函数 ：
         ```
         // 信号处理函数回
         sigtramp （纯汇编代码）
           -> sigtrampgo （ signal_unix.go ）
             -> sighandler  （ signal_sighandler.go ）
                -> preparePanic （ signal_amd64x.go ）
         
                   -> sigpanic （ signal_unix.go ）
                     -> panicmem 
                       -> panic (内存段错误)
         ```
        在进程初始化的时候，创建 M0（线程）的时候用系统调用 sigaction 给信号注册处理函数为 sigtramp
    - Summary
      - panic( ) 函数内部会产生一个关键的数据结构体 _panic ，并且挂接到 goroutine 之上；
      - panic( ) 函数内部会执行 _defer 函数链条，并针对 _panic 的状态进行对应的处理；
      - 循环执行 goroutine 上面的 _defer 函数链，如果执行完了都还没有恢复 _panic 的状态，那就没得办法了，退出进程，打印堆栈。
      - 如果在 goroutine 的 _defer 链上，有个朋友 recover 了一下，把这个 _panic 标记成恢复，那事情就到此为止，就从这个 _defer 函数执行后续正常代码即可，走 deferreturn 的逻辑。

- [如何限定Goroutine数量](https://juejin.cn/post/7017286487502766093)
  - 用有 buffer 的 channel 来限制
  - channel 与 sync 同步组合方式实现控制 goroutine
  - 利用无缓冲 channel 与任务发送/执行分离方式
    ```go
    var wg = sync.WaitGroup{}
    
    func doBusiness(ch chan int) {
    
        for t := range ch {
            fmt.Println("go task = ", t, ", goroutine count = ", runtime.NumGoroutine())
            wg.Done()
        }
    }
    
    func sendTask(task int, ch chan int) {
        wg.Add(1)
        ch <- task
    }
    
    func main() {
    
        ch := make(chan int)   //无buffer channel
    
        goCnt := 3              //启动goroutine的数量
        for i := 0; i < goCnt; i++ {
            //启动go
            go doBusiness(ch)
        }
    
        taskCnt := math.MaxInt64 //模拟用户需求业务的数量
        for t := 0; t < taskCnt; t++ {
            //发送任务
            sendTask(t, ch)
        }
    
        wg.Wait()
    }
    ```

- [Sync Once Source Code](https://mp.weixin.qq.com/s/nkhZyKG4nrUulpliMKdgRw)
  - 问题
    - 为啥源码引入Mutex而不是CAS操作
    - 为啥要有fast path, slow path 
    - 加锁之后为啥要有done==0，为啥有double check，为啥这里不是原子读
    - store为啥要加defer
    - 为啥是atomic.store，不是直接赋值1
  - 演进
    - 开始
      ```go
      type Once struct {
       m    Mutex
       done bool
      }
      
      func (o *Once) Do(f func()) {
       o.m.Lock()
       defer o.m.Unlock()
       if !o.done {
        o.done = true
        f()
       }
      }
      ```
      缺点：每次都要执行Mutex加锁操作，对于Once这种语义有必要吗，是否可以先判断一下done的value是否为true，然后再进行加锁操作呢？
    - 进化
      ```go
      type Once struct {
       m    Mutex
       done int32
      }
      
      func (o *Once) Do(f func()) {
       if atomic.AddInt32(&o.done, 0) == 1 {
        return
       }
       // Slow-path.
       o.m.Lock()
       defer o.m.Unlock()
       if o.done == 0 {
        f()
        atomic.CompareAndSwapInt32(&o.done, 0, 1)
       }
      }
      ```
      进化点
      - 在slow-path加锁后，要继续判断done值是否为0，确认done为0后才要执行f()函数，这是因为在多协程环境下仅仅通过一次atomic.AddInt32判断并不能保证原子性，比如俩协程g1、g2，g2在g1刚刚执行完atomic.CompareAndSwapInt32(&o.done, 0, 1)进入了slow path，如果不进行double check，那g2又会执行一次f()。
      - 用一个int32变量done表示once的对象是否已执行完，有两个地方使用到了atomic包里的方法对o.done进行判断，分别是，用AddInt32函数根据o.done的值是否为1判断once是否已执行过，若执行过直接返回；f()函数执行完后，对o.done通过cas操作进行赋值1。
      - 问到atomic.CompareAndSwapInt32(&o.done, 0, 1)可否被o.done == 1替换， 答案是不可以
        - 现在的CPU一般拥有多个核心，而CPU的处理速度快于从内存读取变量的速度，为了弥补这俩速度的差异，现在CPU每个核心都有自己的L1、L2、L3级高速缓存，CPU可以直接从高速缓存中读取数据，但是这样一来内存中的一份数据就在缓存中有多份副本，在同一时间下这些副本中的可能会不一样，为了保持缓存一致性，Intel CPU使用了MESI协议
        - AddInt32方法和CompareAndSwapInt32方法(均为amd64平台 runtime/internal/atomic/atomic_amd64.s)底层都是在汇编层面调用了LOCK指令，LOCK指令通过总线锁或MESI协议保证原子性（具体措施与CPU的版本有关），提供了强一致性的缓存读写保证，保证LOCK之后的指令在带LOCK前缀的指令执行之后才执行，从而保证读到最新的o.done值。
    - 小优化1
      - 把done的类型由int32替换为uint32,用CompareAndSwapUint32替换了CompareAndSwapInt32, 用LoadUint32替换了AddInt32方法
      - LoadUint32底层并没有LOCK指令用于加锁，我觉得能这么写的主要原因是进入slow path之后会继续用Mutex加锁并判断o.done的值，且后面的CAS操作是加锁的，所以可以这么改
    - 小优化2
      - 用StoreUint32替换了CompareAndSwapUint32操作，CAS操作在这里确实有点多余，因为这行代码最主要的功能是原子性的done = 1
      - Store命令的底层是，其中关键的指令是XCHG，有的同学可能要问了，这源码里没有LOCK指令啊，怎么保证happen before呢，Intel手册有这样的描述: The LOCK prefix is automatically assumed for XCHG instruction.，这个指令默认带LOCK前缀，能保证Happen Before语义。
    - 小优化3
      - 在StoreUint32前增加defer前缀，增加defer是保证 即使f()在执行过程中出现panic，Once仍然保证f()只执行一次，这样符合严格的Once语义。
      - 除了预防panic，defer还能解决指令重排的问题：现在CPU为了执行效率，源码在真正执行时的顺序和代码的顺序可能并不一样，比如这段代码中a不一定打印"hello, world"，也可能打印空字符串。
        ```go
        var a string
        var done bool
        
        func setup() {
         a = "hello, world"
         done = true
        }
        
        func main() {
         go setup()
         for !done {
         }
         print(a)
        }
        ```
    - 小优化4
      - 用函数区分开了fast path和slow path，对fast path做了内联优化




