
- [unsafe 包](https://mp.weixin.qq.com/s/wdFdPv3Bdnhy5pc8KL6w6w)
  - unsafe 实现原理
    - Sizeof(x ArbitrayType)方法主要作用是用返回类型x所占据的字节数，但并不包含x所指向的内容的大小，与C语言标准库中的Sizeof()方法功能一样，比如在32位机器上，一个指针返回大小就是4字节。
    - Offsetof(x ArbitraryType)方法主要作用是返回结构体成员在内存中的位置离结构体起始处(结构体的第一个字段的偏移量都是0)的字节数，即偏移量，我们在注释中看一看到其入参必须是一个结构体，其返回值是一个常量。
    - Alignof(x ArbitratyType)的主要作用是返回一个类型的对齐值，也可以叫做对齐系数或者对齐倍数。对齐值是一个和内存对齐有关的值，合理的内存对齐可以提高内存读写的性能。一般对齐值是2^n，最大不会超过8(受内存对齐影响).获取对齐值还可以使用反射包的函数，也就是说：unsafe.Alignof(x)等价于reflect.TypeOf(x).Align()。
  - 三种指针类型
    - ***T**：普通类型指针类型，用于传递对象地址，不能进行指针运算。
    - **unsafe.poniter**：通用指针类型，用于转换不同类型的指针，不能进行指针运算，不能读取内存存储的值(需转换到某一类型的普通指针)
    - **uintptr**：用于指针运算，GC不把uintptr当指针，uintptr无法持有对象。uintptr类型的目标会被回收。
    
    三者关系就是：unsafe.Pointer是桥梁，可以让任意类型的指针实现相互转换，也可以将任意类型的指针转换为uintptr进行指针运算，也就说uintptr是用来与unsafe.Pointer打配合
  - Sizeof、Alignof、Offsetof三个函数的基本使用
    ```go
     // sizeof
     fmt.Println(unsafe.Sizeof(true))
     fmt.Println(unsafe.Sizeof(int8(0)))
     fmt.Println(unsafe.Sizeof(int16(10)))
     fmt.Println(unsafe.Sizeof(int(10)))
     fmt.Println(unsafe.Sizeof(int32(190)))
     fmt.Println(unsafe.Sizeof("asong"))
     fmt.Println(unsafe.Sizeof([]int{1,3,4}))
     // Offsetof
     user := User{Name: "Asong", Age: 23,Gender: true}
     userNamePointer := unsafe.Pointer(&user)
    
     nNamePointer := (*string)(unsafe.Pointer(userNamePointer))
     *nNamePointer = "Golang梦工厂"
    
     nAgePointer := (*uint32)(unsafe.Pointer(uintptr(userNamePointer) + unsafe.Offsetof(user.Age)))
     *nAgePointer = 25
    
     nGender := (*bool)(unsafe.Pointer(uintptr(userNamePointer)+unsafe.Offsetof(user.Gender)))
     *nGender = false
    
     fmt.Printf("u.Name: %s, u.Age: %d,  u.Gender: %v\n", user.Name, user.Age,user.Gender)
     // Alignof
     var f32 float32
     var s string
     var m map[string]string
     var p *int32
    
     fmt.Println(unsafe.Alignof(f32))
     fmt.Println(unsafe.Alignof(s))
     fmt.Println(unsafe.Alignof(m))
     fmt.Println(unsafe.Alignof(p))
    ```
  - 内存对齐
    - 对齐的作用和原因：CPU访问内存时，并不是逐个字节访问，而是以字长（word size)单位访问。比如32位的CPU，字长为4字节，那么CPU访问内存的单位也是4字节。这样设计可以减少CPU访问内存的次数，加大CPU访问内存的吞吐量。假设我们需要读取8个字节的数据，一次读取4个字节那么就只需读取2次就可以。内存对齐对实现变量的原子性操作也是有好处的，每次内存访问都是原子的，如果变量的大小不超过字长，那么内存对齐后，对该变量的访问就是原子的，这个特性在并发场景下至关重要。
    
- 





