
- [map 缩容？](https://mp.weixin.qq.com/s/Slvgl3KZax2jsy2xGDdFKw)

  在 Go 底层源码 src/runtime/map.go 中，扩缩容的处理方法是 grow 为前缀的方法来处理的. 无论是扩容还是缩容，其都是由 hashGrow 方法进行处理
  若是扩容，则 bigger 为 1，也就是 B+1。代表 hash 表容量扩大 1 倍。不满足就是缩容，也就是 hash 表容量不变。

  可以得出结论：map 的扩缩容的主要区别在于 hmap.B 的容量大小改变。_而缩容由于 hmap.B 压根没变，内存空间的占用也是没有变化的_。

  若要实现 ”真缩容“，唯一可用的解决方法是：**创建一个新的 map 并从旧的 map 中复制元素**。

  - [为什么不支持？](https://github.com/golang/go/issues/20135)
  - 简单来讲，就是没有找到一个很好的方法实现，存在明确的实现成本问题，没法很方便的 告诉 Go 运行时，我要：
    - 记得保留存储空间，我要立即重用 map。
    - 赶紧释放存储空间，map 从现在开始会小很多。


