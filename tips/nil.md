
nil
--------

`nil` a predeclared identifier representing the zero value for a pointer, channel, func, interface, map or slice type

- `nil` hs no type
- `nil` is not a keyword

- kind of nil

    |  type |   means |
    | ----- | ------ |
    | pointers | point to nothing |
    | slices   | have no backing array |
    | maps    | are not initialized |
    | channels |  are not initialized |
    | functions |  are not initialized  |
    | interfaces | have no value assigned, not even a nil pointer |

- nil is useful

  |  type |   details |
  | ----- | ------ |
  | pointers | methods can be called on nil receivers |
  | slices   | perfectly valid zero values |
  | maps     | perfect as read-only values |
  | channels | essential for some some cocurrency patterns |
  | functions | needed for completeness |
  | interfaces | the more used signal in Go (err != nil) |

- nil == nil ?
  
  `invalid operation: nil == nil (operator == not defined on nil)
  ==符号对于nil来说是一种未定义的操作, 因为nil是没有类型的，是在编译期根据上下文确定的，所以要比较nil的值也就是比较不同类型的nil
  
  











