
结构体多字段的原子操作
-----------

```go
type Person struct {
    name string
    age  int
}

var p Person

func update(name string, age int) {
    p.name = name
    // 加点随机性
    time.Sleep(time.Millisecond*200)
    p.age = age
}

wg := sync.WaitGroup{}
wg.Add(10)
// 10 个协程并发更新
for i := 0; i < 10; i++ {
    name, age := fmt.Sprintf("nobody:%v", i), i
    go func() {
        defer wg.Done()
        update(name, age)
    }()
}
wg.Wait()
```

```shell
# time ./atomic_test
p.name=nobody:8
p.age=7

real 0m0.203s
user 0m0.000s
sys 0m0.000s
```

这个 200 毫秒是因为奇伢在 update 函数中故意加入了一点点时延，这样可以让程序估计跑慢一点。

每个协程跑 update 的时候至少需要 200 毫秒，10 个协程并发跑，没有任何互斥，时间重叠，所以整个程序的时间也是差不都 200 毫秒左右。

确保正确性

- 锁互斥

  ```go
    var p Person
    // 互斥锁，保护变量更新
    var mu sync.Mutex
  ```
  
  ```shell
    time ./atomic_test
    p.name=nobody:8
    p.age=8
    
    real 0m2.017s
    user 0m0.000s
  ```

  程序串行执行了 10 次 update 函数，时间是累加的。程序 2 秒的运行时延就这样来的。

  加锁不怕，抢锁等待才可怕。在大量并发的时候，由于锁的互斥特性，这里的性能可能堪忧。

- 原子操作

  ```go
    // 全局变量（简单处理）
    var p atomic.Value
    
    func update(name string, age int) {
        lp := &Person{}
        // 更新第一个字段
        lp.name = name
        // 加点随机性
        time.Sleep(time.Millisecond * 200)
        // 更新第二个字段
        lp.age = age
        // 原子设置到全局变量
        p.Store(lp)
    }

    _p := p.Load().(*Person)
    fmt.Printf("p.name=%s\np.age=%v\n", _p.name, _p.age)
  ```
  这 10 个协程还是并发的，没有类似于锁阻塞等待的操作，只有最后 p.Store(lp) 调用内才有做状态的同步

  - atomic.Value结构体
    ```go
    type Value struct {
        v interface{}
    }
    ```
    `interface {}` 是给程序猿用的，`eface`  是 Go 内部自己用的，位于不同层面的同一个东西，而 `atomic.Value` 就利用了这个特性，在 value.go 定义了一个 `ifaceWords` 的结构体

    `interface {}`，`eface`，`ifaceWords` 这三个结构体内存布局完全一致，只是用的地方不同而已，本质无差别。这给类型的强制转化创造了前提

  - Value.Store
    - atomic.Value 使用 ^uintptr(0) 作为第一次存取的标志位，这个标识位是设置在 type 字段里，这是一个中间状态；
    - 通过 CompareAndSwapPointer 来确保 ^uintptr(0)  只能被一个执行体抢到，其他没抢到的走 continue ，再循环一次；
    - atomic.Value 第一次写入数据时，将当前协程设置为不可抢占，当存储完毕后，即可解除不可抢占；
    - 真正的赋值，无论是第一次，还是后续的 data 赋值，在 Store 内，只涉及到指针的原子操作，不涉及到数据拷贝

    Value.Store()  的**参数必须是个局部变量**

  `atomic.Value` 的 `Store` 和 `Load` 方法都不涉及到数据拷贝，只涉及到指针操作

  `atomic.Value` 使用 `cas` 操作只在初始赋值的时候，一旦赋值过，后续赋值的原子操作更简单，依赖于 `StorePointer` ，指针值得原子赋值


