
- [Netty 如何应对 TCP 连接的正常关闭，异常关闭，半关闭场景](https://mp.weixin.qq.com/s/2MW5xscY7j9_0byBFN47Vw)
  - 正常 TCP 连接关闭
    - Netty 处理 TCP 正常关闭流程（ Socket 接收缓冲区中只有 EOF ，没有其他正常接收数据）可以看出，这种情况下只会触发 ChannelReadComplete 事件而不会触发 ChannelRead 事件。
  - Netty 对 TCP 连接正常关闭的处理
    - close 方法发起 TCP 连接关闭流程
      - Netty 这里使用一个 boolean closeInitiated 变量来防止 Reactor 线程来重复执行关闭流程，因为 Channel 的关闭操作可以在多个业务线程中发起，这样就会导致多个业务线程向 Reactor 线程提交多个关闭 Channel 的任务。
      - 通过 isActive() 获取 Channel 的状态 boolean wasActive ，由于此时我们还没有关闭 Channel，所以 Channel 现在的状态肯定是 active 的。之所以在关闭流程的一开始就获取 Channel 是否 active 的状态，是因为当我们关闭 Channel 之后，需要通过这个状态来判断 channel 是否是第一次从 active 变为 inactive ，如果是第一次，则会触发 ChannelInactive 事件在 Channel 对应的 pipeline 中传播。
      - 针对 SO_LINGER 选项的处理
        - 默认情况下，当我们调用 Socket 的 close 方法后 ，close 方法会立即返回，剩下的事情会交给内核协议栈帮助我们处理，如果此时 Socket 对应的发送缓冲区还有数据待发送，接下来内核协议栈会将 Socket 发送缓冲区的数据发送出去，随后会向对端发送 FIN 包关闭连接
        - 而 SO_LINGER 选项会控制调用 close 方法关闭 Socket 的行为
          - l_onoff = 0 时 l_linger 的值会被忽略，属于我们上边讲述的默认关闭行为。
          - l_onoff = 1，l_linger > 0：这种情况下，应用程序调用 close 方法后就不会立马返回，无论 Socket 是阻塞模式还是非阻塞模式，应用程序都会阻塞在这里。直到以下两个条件其中之一发生，才会解除阻塞返回。随后进行正常的四次挥手关闭流程。
            - 当 Socket 发送缓冲区的数据全部发送出去，并等到对端 ACK 后，close 方法返回。
            - 应用程序在 close 方法上的阻塞时间到达 l_linger 设置的值后，close 方法返回。
            - ![img.png](netty_so_linger1.png)
          - l_onoff = 1，l_linger = 0：这种情况下，当应用程序调用 close 方法后会立即返回，随后内核直接清空 Socket 的发送缓冲区，并向对端发送 RST 包，主动关闭方直接跳过四次挥手进入 CLOSE 状态，注意这种情况下是不会有 TIME_WAIT 状态的。
            - ![img.png](netty_so_linger2.png)
        - 当我们设置了 SO_LINGER 选项之后，Channel 的关闭动作会被阻塞并延时关闭，在延时关闭期间，Reactor 线程依然可以响应 OP_READ 事件和 OP_WRITE 事件，这可能会导致 Reactor 线程不断的自旋循环浪费 CPU 资源，所以基于这个原因，netty 这里需要将 Channel 从 Reactor 上注销掉。这样 Reactor 线程就不会在响应 Channel 上的 IO 事件了。
  - TCP 连接的异常关闭
    - 







