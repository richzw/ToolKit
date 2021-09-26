
- orDone 是一种并发控制模式, 在多任务场景下实现，有一个任务成功返回即立即结束等待
  - 递归 利用二分法递归， 将所有待监听信号的chan都select起来
  ```go
  func Or(channels ...<-chan interface{}) <-chan interface{} {
   // 只有零个或者1个chan
   switch len(channels) {
   case 0:
          // 返回nil， 让读取阻塞等待
    return nil
   case 1:
    return channels[0]
   }
  
   orDone := make(chan interface{})
   go func() {
          // 返回时利用close做结束信号的广播
    defer close(orDone)
  
          // 利用select监听第一个chan的返回
    switch len(channels) {
    case 2: // 直接select
     select {
     case <-channels[0]:
     case <-channels[1]:
     }
    default: // 二分法递归处理
     m := len(channels) / 2
     select {
     case <-Or(channels[:m]...):
     case <-Or(channels[m:]...):
     }
    }
   }()
  
   return orDone
  }
  ```
  - 利用反射 这里要用到reflect.SelectCase, 他可以描述一种select的case, 来指明其接受的是chan的读取或发送
  ```go
  func OrInReflect(channels ...<-chan interface{}) <-chan interface{} {
   // 只有0个或者1个
   switch len(channels) {
   case 0:
    return nil
   case 1:
    return channels[0]
   }
  
   orDone := make(chan interface{})
   go func() {
    defer close(orDone)
    // 利用反射构建SelectCase，这里是读取
    var cases []reflect.SelectCase
    for _, c := range channels {
     cases = append(cases, reflect.SelectCase{
      Dir:  reflect.SelectRecv,
      Chan: reflect.ValueOf(c),
     })
    }
  
    // 随机选择一个可用的case
    reflect.Select(cases)
   }()
  
   return orDone
  }
  ```
  - Test

  大量并发chan场景下， 反射使用内存更多些，但速度更快

  ```go
  func repeat(
   done <-chan interface{},
      // 外部传入done控制是否结束
   values ...interface{},
  ) <-chan interface{} {
   valueStream := make(chan interface{})
   go func() {
          // 返回时释放
    defer close(valueStream)
    for {
     for _, v := range values {
      select {
      case <-done:
       return
      case valueStream <- v:
      }
     }
    }
   }()
   return valueStream
  }
  
  func BenchmarkOr(b *testing.B) {
  done := make(chan interface{})
  defer close(done)
  num := 100
  streams := make([]<-chan interface{}, num)
  for i := range streams {
  streams[i] = repeat(done, []int{1, 2, 3})
  }
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
  <-Or(streams...)
  }
  }
  ```
