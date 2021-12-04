
- [sync.Once](https://mp.weixin.qq.com/s/Eu6T5-4v82kdh-Url4O5_w)
  - `sync.Once`底层通过`defer`标记初始化完成，所以无论初始化是否成功都会标记初始化完成，即**不可重入**。
  - G1和G2同时执行时，G1执行失败后，G2不会执行初始化逻辑，因此需要double check。
  ```go
      if atomic.LoadUint32(&o.done) == 0 {
          // Outlined slow-path to allow inlining of the fast-path.
          o.doSlow(f)
      }
  }
  
  func (o *Once) doSlow(f func()) {
      o.m.Lock()
      defer o.m.Unlock()
      if o.done == 0 {
          defer atomic.StoreUint32(&o.done, 1)
          f()
      }
  }
  ```
  
  可重入且并发安全的sync.Once
  ```go
  type IOnce struct {
   done uint32
   m    sync.Mutex
  }
  // Do方法传递的函数增加一个error返回值
  func (o *IOnce) Do(f func() error) {
   if atomic.LoadUint32(&o.done) == 0 {
    o.doSlow(f)
   }
  }
  
  // 不使用defer控制don标识，而通过也无妨的返回值来控制
  func (o *IOnce) doSlow(f func() error) {
   o.m.Lock()
   defer o.m.Unlock()
   if o.done == 0 {
    if f() != nil {
     return
    }
    // 执行成功后才将done置为1
    atomic.StoreUint32(&o.done, 1)
   }
  }
  
  ```

- [零值](https://mp.weixin.qq.com/s/o7SwVDpscSDbi6ePxpjjLQ)
  - 零值channels
    
    给定一个nil channel c:
      - <-c 从c 接收将永远阻塞
      - c <- v 发送值到c 会永远阻塞
      - close(c) 关闭c 引发panic

- [InterfaceSlice](https://github.com/golang/go/wiki/InterfaceSlice)
  Given that you can assign a variable of any type to an interface{}, often people will try code like the following.
  ```go
  var dataSlice []int = foo()
  var interfaceSlice []interface{} = dataSlice
  ```
  
  [This gets the error](https://mp.weixin.qq.com/s/rAIhapDHrA7jQVr_uvQpEg)
  
  `cannot use dataSlice (type []int) as type []interface { } in assignment`

  A variable with type []interface{} has a specific memory layout, known at compile time.

  Each interface{} takes up two words (one word for the type of what is contained, the other word for either the contained data or a pointer to it). As a consequence, a slice with length N and with type []interface{} is backed by a chunk of data that is N*2 words long.

  This is different than the chunk of data backing a slice with type []MyType and the same length. Its chunk of data will be N*sizeof(MyType) words long.

  - fix it
    ```go
    var dataSlice []int = foo()
    var interfaceSlice []interface{} = make([]interface{}, len(dataSlice))
    for i, d := range dataSlice {
        interfaceSlice[i] = d
    }
    ```
- [golang redis pipeline管道引发乱序串读](http://xiaorui.cc/archives/7179)
  - 在使用 golang redigo pipeline 模式下，错误使用会引发乱序串读的问题。简单说，发了一组 pipeline命 令，但由于只发送而没有去解析接收结果，那么后面通过连接池重用该连接时，会拿到了上次的请求结果，乱序串读了。
  - 看源码可得知，Flush() 只是把buffer缓冲区的数据写到连接里，而没有从连接读取的过程。所以说，在redigo的pipeline里，有几次的写，就应该有几次的 Receive() 。Receive是从连接读缓冲区里读取解析数据。 receive() 是不可或缺的！ 不能多，也不能少，每个 send() 都对应一个 receive()。
  - 使用了 pipeline 批量确实有效的减少了时延，也减少了 redis 压力。不要再去使用 golang redigo 这个库了，请直接选择 go-redis 库。

- [defer 坑过](https://mp.weixin.qq.com/s/1T6Z74Wri27Ap8skeJiyWQ)
  - case 1
     ```go
     package main
     
     import (
         "fmt"
     )
     
     func a() (r int) {
         defer func() {
             r++
         }()
         return 0
     }
     
     func b() (r int) {
         t := 5
         defer func() {
             t = t + 5
         }()
         return t
     }
     
     func c() (r int) {
         defer func(r int) {
             r = r + 5
         }(r)
         return 1
     }
     
     func main() {
         fmt.Println("a = ", a())
         fmt.Println("b = ", b())
         fmt.Println("c = ", c())
     }
     ```
  - case 2

  defer 表达式的函数如果在 panic 后面，则这个函数无法被执行。
   ```go
   func main() {
       panic("a")
       defer func() {
           fmt.Println("b")
       }()
   }
   ```
- [defer Close() 的风险](https://mp.weixin.qq.com/s?__biz=MzkyMzIyNjIxMQ==&mid=2247484680&idx=1&sn=5df7d6a7b410fcdec01982470ca2158d&chksm=c1e91c04f69e9512628968de16dd081582fea5aae45b17f0facdc33322c9d447d9c09529f36c&scene=21#wechat_redirect)
  - 问题：
    - `defer x.Close()` 会忽略它的返回值，但在执行 x.Close() 时，我们并不能保证 x 一定能正常关闭，万一它返回错误应该怎么办？这种写法，会让程序有可能出现非常难以排查的错误。
  - Close() 方法会返回什么错误呢
    - 在 POSIX 操作系统中，例如 Linux 或者 maxOS，关闭文件的 Close() 函数最终是调用了系统方法 close()，我们可以通过 man close 手册，查看 close() 可能会返回什么错误
      - [EBADF]            fildes is not a valid, active file descriptor.
      - [EINTR]            Its execution was interrupted by a signal.
      - [EIO]              A previously-uncommitted write(2) encountered an input/output error.
    - EIO 的错误是指未提交写。 EIO 错误的确是我们需要提防的错误。这意味着如果我们尝试将数据保存到磁盘，在 defer x.Close() 执行时，操作系统还并未将数据刷到磁盘，这时我们应该获取到该错误提示
  - 改造方案
    - 不使用 defer
      ```go
       1func solution01() error {
       2    f, err := os.Create("/home/golangshare/gopher.txt")
       3    if err != nil {
       4        return err
       5    }
       6
       7    if _, err = io.WriteString(f, "hello gopher"); err != nil {
       8        f.Close()
       9        return err
      10    }
      11
      12    return f.Close()
      13}
      ```
      需要在每个发生错误的地方都要加上关闭语句 f.Close()，如果对 f 的写操作 case 较多，容易存在遗漏关闭文件的风险。
    - 通过命名返回值 err 和闭包来处理
       ```go
       1func solution02() (err error) {
       2    f, err := os.Create("/home/golangshare/gopher.txt")
       3    if err != nil {
       4        return
       5    }
       6
       7    defer func() {
       8        closeErr := f.Close()
       9        if err == nil {
       10            err = closeErr
       11        }
       12    }()
       13
       14    _, err = io.WriteString(f, "hello gopher")
       15    return
       16}
       ```
      如果有更多 if err !=nil 的条件分支，这种模式可以有效降低代码行数。
    - 在函数最后 return 语句之前，显示调用一次 f.Close()
      ```go
      1func solution03() error {
       2    f, err := os.Create("/home/golangshare/gopher.txt")
       3    if err != nil {
       4        return err
       5    }
       6    defer f.Close()
       7
       8    if _, err := io.WriteString(f, "hello gopher"); err != nil {
       9        return err
      10    }
      11
      12    if err := f.Close(); err != nil {
      13        return err
      14    }
      15    return nil
      16}
      ```
      这种解决方案能在 io.WriteString 发生错误时，由于 defer f.Close() 的存在能得到 close 调用。也能在 io.WriteString 未发生错误，但缓存未刷新到磁盘时，得到 err := f.Close() 的错误，而且由于 defer f.Close() 并不会返回错误，所以并不担心两次 Close() 调用会将错误覆盖。
    - 函数 return 时执行 f.Sync()
      ```go
      func solution04() error {
       2    f, err := os.Create("/home/golangshare/gopher.txt")
       3    if err != nil {
       4        return err
       5    }
       6    defer f.Close()
       7
       8    if _, err = io.WriteString(f, "hello world"); err != nil {
       9        return err
      10    }
      11
      12    return f.Sync()
      13}
      ```
      由于 fsync 的调用，这种模式能很好地避免 close 出现的 EIO。可以预见的是，由于强制性刷盘，这种方案虽然能很好地保证数据安全性，但是在执行效率上却会大打折扣。
- [字符串底层原理](https://mp.weixin.qq.com/s/1ozBAEYpf07aei0xh3_kNQ)
  - 字符串
    ```go
    type stringStruct struct {
        str unsafe.Pointer // 一个指向底层数据的指针. 使用 utf-8 编码方式
        len int            // 字符串的字节长度，非字符个数
    }
    ```
  - [rune](https://kvs-vishnu23.medium.com/what-exactly-is-a-rune-datatype-in-go-f652093a88eb)
    - 在 unicode 字符集中，每一个字符都有一个对应的编号，我们称这个编号为 code point，而 Go 中的rune 类型就代表一个字符的 code point。
    ![img.png](go_rune.png)
    - rune is an alias for int32 value which represents a single Unicode point. From the above program to print the rune equivalent of 128513, we used Printf() function with %c control string
     ```go
         oo := '中'
         fmt.Println("As an int value ", oo) // 20013, this is rune
         fmt.Printf("As a string: %s, %s and as a char: %c \n", oo, string(oo), oo)
     ```
- [一个打点引发的事故](https://mp.weixin.qq.com/s?__biz=MjM5MDUwNTQwMQ==&mid=2257486531&idx=1&sn=a996f5932ca9018f78ce0049f3a7baa8&chksm=a539e215924e6b03ab2c6d2ab370ca184afa85d0f6fac0af644dc3bef4fbdd319d1f0a7a932e&cur_album_id=1690026440752168967&scene=189#wechat_redirect)
  - 查到了 OOM 的实例 goroutine 数暴涨，接口 QPS 有尖峰，比正常翻了几倍。所以，得到的结论就是接口流量太多，超过服务极限，导致 OOM
  - 由于 metrics 底层是用 udp 发送的，有文件锁，大量打点的情况下，会引起激烈的锁冲突，造成 goroutine 堆积、请求堆积，和请求关联的 model 无法释放，于是就 OOM 了
  - 由于这个地方的打点非常多，几十万 QPS，一冲突，goroutine 都 gopark 去等锁了，持有的内存无法释放，服务一会儿就 gg 了
- [mutex 出问题怎么办？强大的 goroutine 诊断工具](https://mp.weixin.qq.com/s/JadUu7odckhNfcXDULoXVA)
  - 我们如何才能看到所有 goroutine 在任何给定时刻都在做什么
    - 每个 Go 程序都带有[一个默认 SIGQUIT 信号处理程序](https://pkg.go.dev/os/signal#hdr-Default_behavior_of_signals_in_Go_programs)的开箱即用的解决方案 。收到此信号后，程序将堆栈转储打印到 stderr 并退出
      ```go
      func main() {
          ch := make(chan (bool), 1)
      
          go func() {
              readForever(ch)
          }()
      
          writeForever(ch)
      }
      
      func readForever(ch chan (bool)) {
          for {
              <-ch
          }
      }
      
      func writeForever(ch chan (bool)) {
          for {
              ch <- true
          }
      }
      ```
      - 运行这个程序并通过 `CTRL+\` 发送一个 SIGQUIT 来终止它.
      - 我们可以使用以下 kill命令发送信号：
       `kill -SIGQUIT <process id>`
      - 对于 Docker，我们需要向正在运行的容器发送 SIGQUIT。没问题：
        ```shell
        docker kill --signal=SIGQUIT <container_id>
        # Then, grab the stack dump from the container logs.
        docker logs <container_id>
        ```
  - [goroutine-inspect](https://github.com/linuxerwang/goroutine-inspect)
    - goroutine-inspect 的工具。它是一个 pprof 风格的交互式 CLI，它允许你操作堆栈转储、过滤掉不相关的跟踪或搜索特定功能
    - 













