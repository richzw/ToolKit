
- [gRPC flow control](https://www.youtube.com/watch?v=EEEGBwEA8yA)
  - TCP congestion control
    - Common Algorithm
      - Reno
      - BIC
      - CUBIC
    - General Strategy
      - Incr sending rate if ACK
      - Decr if ACK missed
  - TCP flow control
    - Stop reading kernel buffer
    - Receiver drop further packets
    - Receiver being protected
    - Trigger Impact
      - Reduce throughput
      - Degraded multiplexing
  - gRPC flow control
    - Algorithm similar to Token Bucket (WINDOW_UPDATE Frame)
    - Features
      - High performance
      - Fine grained throttling - Stream (RPC)/ Connection
      - Frame Priority
  - Window Size
    - Solution: BDP estimator - Bandwidth Delay Product: the amount of data that can be in transit in the network
    - Goal: intelligently avoid triggering flow control
    - Measure BDP through PING frame and PID controller
    - Set the _init window size_ to BDP
  - Challenge
    - Fairness between RPCs  - HTTP2 flow control supports multiplexing
    - Throttle based on performance - gRPC has BDP estimator
    - Flow control From End to End - gRPC has built-in Flow control




