
- do not overuse `fmt.Sprintf` in your hot path. It is costly due to maintaining the buffer pool and dynamic dispatches for interfaces.
    - if you are doing `fmt.Sprintf("%s%s", var1, var2)`, consider simple string concatenation.
    - if you are doing `fmt.Sprintf("%x", var)`, consider using `hex.EncodeToString` or `strconv.FormatInt(var, 16)`
- 如果需要把数字转换成字符串，使用 strconv.Itoa() 比 fmt.Sprintf() 要快一倍左右。
- 使用StringBuffer 或是StringBuild 来拼接字符串，性能会比使用 + 或 += 高出三到四个数量级。
- String to []byte
  ```go
  func cstring(s string) []byte {
    b := make([]byte, len(s)+1)
    copy(b, s)
    return b
  }
  ```
