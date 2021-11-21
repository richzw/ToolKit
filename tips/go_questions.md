
- [1](https://mp.weixin.qq.com/s?__biz=MzAxNzY0NDE3NA==&mid=2247484972&idx=1&sn=3ac2c60f30114bef4a4bdd41fd7638a6&chksm=9be329cdac94a0db447cb48a41b609ef8909d78449d30f530609b35f92d0eda386bac675c67c&scene=21#wechat_redirect)
  ```go
  package main
  
  const s = "Go101.org"
  // len(s) == 9
  // 1 << 9 == 512
  // 512 / 128 == 4
  
  var a byte = 1 << len(s) / 128
  var b byte = 1 << len(s[:]) / 128
  
  func main() {
    println(a, b)
  }
  ```
  如果 const s = "Go101.org” 改为 var s = "Go101.org" 结果又会是什么呢？ // 0 0 
  然则
   ```go
   package main
   
   var s = [9]byte{'G', 'o', '1', '0', '1', '.', 'o', 'r', 'g'}
   
   var a byte = 1 << len(s) / 128
   var b byte = 1 << len(s[:]) / 128
   
   func main() {
    println(a, b)
   }
   ```
  Go 语言规范中关于长度和容量的说明
   > 内置函数 len 和 cap 获取各种类型的实参并返回一个 int 类型结果。实现会保证结果总是一个 int 值。
   如果 s 是一个字符串常量，那么 len(s) 是一个常量 。如果 s 类型是一个数组或到数组的指针且表达式 s 不包含 信道接收 或（非常量的） 函数调用的话， 那么表达式 len(s) 和 cap(s) 是常量；这种情况下， s 是不求值的。否则的话， len 和 cap 的调用结果不是常量且 s 会被求值。

  ```go
  var a byte = 1 << len(s) / 128
  var b byte = 1 << len(s[:]) / 128
  ```
  第一句的 len(s) 是常量（因为 s 是字符串常量）；而第二句的 len(s[:]) 不是常量。这是这两条语句的唯一区别：两个 len 的返回结果数值并无差异，都是 9，但一个是常量一个不是

  位移运算这里。Go 语言规范中有这么一句
  > The right operand in a shift expression must have integer type or be an untyped constant representable by a value of type uint. If the left operand of a non-constant shift expression is an untyped constant, it is first implicitly converted to the type it would assume if the shift expression were replaced by its left operand alone.
  
  > If the left operand of a constant shift expression is an untyped constant, the result is an integer constant; otherwise it is a constant of the same type as the left operand, which must be of integer type.
  
  - 因此对于 var a byte = 1 << len(s) / 128，因为 1 << len(s) 是一个常量位移表达式，因此它的结果也是一个整数常量，所以是 512，最后除以 128，最终结果就是 4。
  - 而对于 var b byte = 1 << len(s[:]) / 128，因为 1 << len(s[:]) 不是一个常量位移表达式，而做操作数是 1，一个无类型常量，根据规范定义它是 byte 类型（根据：如果一个非常量位移表达式的左侧的操作数是一个无类号常量，那么它会先被隐式地转换为假如位移表达式被其左侧操作数单独替换后的类型）。

  常量是在编译的时候进行计算的。在 Go 语言中，常量分两种：无类型和有类型。
  - Go 规范上说，字面值常量， true , false , iota 以及一些仅包含无类型的恒定操作数的 
  - 常量表达式 是无类型的。
    
  所以 `var b byte = 1 << len(s[:]) / 128` 中，根据规范定义，1 会隐式转换为 byte 类型，因此 `1 << len(s[:])` 的结果也是 byte 类型，而 byte 类型最大只能表示 255，很显然 512 溢出了，结果为 0，因此最后 b 的结果也是 0。



