
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








