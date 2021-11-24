
- [题1](https://mp.weixin.qq.com/s?__biz=MzAxNzY0NDE3NA==&mid=2247484972&idx=1&sn=3ac2c60f30114bef4a4bdd41fd7638a6&chksm=9be329cdac94a0db447cb48a41b609ef8909d78449d30f530609b35f92d0eda386bac675c67c&scene=21#wechat_redirect)
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

- [题2](https://mp.weixin.qq.com/s?__biz=MzAxNzY0NDE3NA==&mid=2247485015&idx=1&sn=4582ca64df8cba44a686ea83299306c9&chksm=9be329b6ac94a0a01fea76c93592ad280805a14cbc4d4227a78aa69c1f347fa583b0aa88d745&cur_album_id=1468728629806153729&scene=189#wechat_redirect)
  ```go
  package main
  
  func main() {
   var a int8 = -1
   var b int8 = -128 / a
  
   println(b)
  }
  ```
  因为 var b int8 = -128 / a 不是常量表达式，因此 untyped 常量 -128 隐式转换为 int8 类型（即和 a 的类型一致），所以 -128 / a 的结果是 int8 类型，值是 128，超出了 int8 的范围。因为结果不是常量，允许溢出，128 的二进制表示是 10000000，正好是 -128 的补码。所以，第一题的结果是 -128

  [Go 语言规范](https://hao.studygolang.com/golang_spec.html#id327)
  > 对于两个整数值 x 和 y ，其整数商 q = x / y 和余数 r = x % y 满足如下关系：
  > x = q*y + r 且 |r| < |y|
    这个规则有一个例外，如果对于 x 的整数类型来说，被除数 x 是该类型中最负的那个值，那么，因为 补码 的 整数溢出 ，商 q = x / -1 等于 x （并且 r = 0 ）。
  ```go
  package main
  
  func main() {
   const a int8 = -1
   var b int8 = -128 / a
  
   println(b)
  }
  ```
  对于 var b int8 = -128 / a，因为 a 是 int8 类型常量，所以 -128 / a 是常量表达式，在编译器计算，结果必然也是常量。因为 a 的类型是 int8，因此 -128 也会隐式转为 int8 类型，128 这个结果超过了 int8 的范围，但常量不允许溢出，因此编译报错。


