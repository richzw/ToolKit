
- Socket Buffer
  - 一个socket ，会带有两个缓冲区，一个用于发送，一个用于接收
    - 用户发送消息的时候写给 send buffer（发送缓冲区）
    - 用户接收消息的时候写给 recv buffer（接收缓冲区）
  - 查看 socket 缓冲区，可以在linux环境下执行 `netstat -nt` 命令
    - Send-Q 是发送缓冲区
    - Recv-Q 代表接收缓冲区
  - TCP
    - 执行 send 发送的字节，会立马发送吗
    ![img.png](socket_sent.png)
    - 如果缓冲区满了会怎么办
    ![img_1.png](socket_recv.png)
    - 如果socket缓冲区还有数据，执行close了，会怎么样
      - 如果接收缓冲区有数据时，执行close了，会怎么样
        - 如果接收缓冲区还有数据未读，会先把接收缓冲区的数据清空，然后给对端发一个RST。
      - 如果接收缓冲区是空的，那么就调用 tcp_send_fin() 开始进行四次挥手过程的第一次挥手
    - 如果发送缓冲区有数据时，执行close了，会怎么样
      - 内核会把发送缓冲区最后一个数据块拿出来。然后置为 FIN。
      - socket 缓冲区是个先进先出的队列，这种情况是指内核会等待TCP层安静把发送缓冲区数据都发完，最后再执行四次挥手的第一次挥手（FIN包）
  - UDP
    - 我们大部分情况下，都不会用  MSG_MORE，也就是来一个数据包就直接发一个数据包。从这个行为上来说，虽然UDP用上了发送缓冲区，但实际上并没有起到"缓冲"的作用

- [SO_REUSEPORT vs SO_REUSEADDR](https://idea.popcount.org/2014-04-03-bind-before-connect/)
  - SO_REUSEADDR - There are at least three situations when this flag is useful
    - Normally after binding to a port and stopping a server it's necessary to wait for a socket to time out before another server can bind to the same port. With SO_REUSEADDR set it's possible to rebind immediately, even if the socket is in a TIME_WAIT state.
    - When one server binds to INADDR_ANY, say 0.0.0.0:1234, it's impossible to have another server binding to a specific address like 192.168.1.21:1234. With SO_REUSEADDR flag this behaviour is allowed.
    - When using the bind before connect trick only a single connection can use a single outgoing source port. With this flag, it's possible for many connections to reuse the same source port, given that they connect to different destination addresses.
  - SO_REUSEPORT
    - It was introduced for UDP multicast sockets. Initially, only a single server was able to use a particular port to listen to a multicast group. This flag allowed different sockets to bind to exactly the same IP and port, and receive datagrams for selected multicast groups.
    - More generally speaking, setting SO_REUSEPORT informs a kernel of an intention to share a particular bound port between many processes, but only for a single user. For multicast datagrams are distributed based on multicast groups, for usual UDP datagrams are distributed in round-robin way. For a long time this flag wasn't available for TCP sockets, but recently Google submitted patches that fix it and distribute incoming connections in round-robin way between listening sockets.
  - EADDRNOTAVAIL vs EADDRINUSE
    - Check bind() for EADDRINUSE errors, in case we run out of available ports.
    - Check connect() for EADDRNOTAVAIL errors in case there is a connection conflict and retry if necessary.
    - If you establish 64k connections using connect, bind will fail with EADDRINUSE. 
    - when thousands of connections are using bind before connect straight connect might fail with EADDRNOTAVAIL.
    - In such case connect() will fail with EADDRNOTAVAIL error. Here's a code handling this situation:
      ```c
      for i in range(RETRIES):
          s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
          s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
          s.bind(("192.168.1.21", 0))
          try:
              s.connect(("www.google.com", 80))
              break
          except socket.error, e:
              if e.errno != errno.EADDRNOTAVAIL:
                  raise
      else:
          raise Exception("Failed to find an unused source port")
      ```
- [Outgoing connections on Linux](https://blog.cloudflare.com/how-to-stop-running-out-of-ephemeral-ports-and-start-to-love-long-lived-connections/)
  - TCP
    - naive method

     | technique description	| errno on port exhaustion	| possible src 2-tuple reuse |
     | ----- | ------ | ------ |
     | connect(dst_IP, dst_port)	| EADDRNOTAVAIL	| yes (good!)|
    - Manually selecting source IP address

     |technique description	|errno on port exhaustion	|possible src 2-tuple reuse |
     | ----- | ------ | ------ |
     |bind(src_IP, 0) <br> connect(dst_IP, dst_port)	| EADDRINUSE	| no (bad!) |
    - IP_BIND_ADDRESS_NO_PORT

      | technique description	| errno on port exhaustion	| possible src 2-tuple reuse |
      | ----- | ------ | ------ |
      |IP_BIND_ADDRESS_NO_PORT <br> bind(src_IP, 0) <br> connect(dst_IP, dst_port)	|EADDRNOTAVAIL	 | yes (good!) |
    - Explicitly selecting a source port

      | technique description	| errno on port exhaustion	| possible src 2-tuple reuse |
      | ----- | ------ | ------ |
      | SO_REUSEADDR <br> bind(src_IP, src_port) <br> connect(dst_IP, dst_port)	| EADDRNOTAVAIL	| yes (good!) |
  - UDP
    - Vanilla UDP is limited

      | technique description	| errno on port exhaustion	| possible src 2-tuple reuse	| risk of overshadowing |
      | ----- | ------ | ------ | -------  |
      | connect(dst_IP, dst_port)	| EAGAIN	| no (bad!)	 | no |
    - SO_REUSEADDR is hard

      | technique description	| errno on port exhaustion	| possible src 2-tuple reuse	| risk of overshadowing |
      | ----- | ------ | ------ | -------  |
      | SO_REUSEADDR <br> connect(dst_IP, dst_port)	| EAGAIN	| yes	 | yes (bad!) |
- [epoll Edge Triggered vs Level Triggered](https://stackoverflow.com/a/41730177/3011380)
  - Edge Triggered: 
    - Once you set EPOLLET, events are triggered only if they change the state of the fd - meaning that only the first event is triggered and no new events will get triggered until that event is fully handled.
    - This design is explicitly meant to prevent epoll_wait from returning due to an event that is in the process of being handled (i.e., when new data arrives while the EPOLLIN was already raised but read hadn't been called or not all of the data was read).
    - The edge-triggered event rule is simple all same-type (i.e. EPOLLIN) events are merged until all available data was processed.
    - In the case of a listening socket, the EPOLLIN event won't be triggered again until all existing listen "backlog" sockets have been accepted using accept.
    - In the case of a byte stream, new events won't be triggered until all the the available bytes have been read from the stream (the buffer was emptied).
  - Level Triggered: 
    - On the other hand, level triggered events will behave closer to how legacy select (or poll) operates, allowing epoll to be used with older code.
    - The event-merger rule is more complex: events of the same type are only merged if no one is waiting for an event (no one is waiting for epoll_wait to return), or if multiple events happen before epoll_wait can return... otherwise any event causes epoll_wait to return.
    - In the case of a listening socket, the EPOLLIN event will be triggered every time a client connects... unless no one is waiting for epoll_wait to return, in which case the next call for epoll_wait will return immediately and all the EPOLLIN events that occurred during that time will have been merged into a single event.
    - In the case of a byte stream, new events will be triggered every time new data comes in... unless, of course, no one is waiting for epoll_wait to return, in which case the next call will return immediately for all the data that arrive util epoll_wait returned (even if it arrived in different chunks / events).
  - Exclusive return: 
    - The EPOLLEXCLUSIVE flag is used to prevent the "thundering heard" behavior, so only a single epoll_wait caller is woken up for each fd wake-up event.
    - As I pointed out before, for edge-triggered states, an fd wake-up event is a change in the fd state. So all EPOLLIN events will be raised until all data was read (the listening socket's backlog was emptied).
    - On the other hand, for level triggered events, each EPOLLIN will invoke a wake up event. If no one is waiting, these events will be merged.



