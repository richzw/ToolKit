
- [find 命令的 7 种用法](https://mp.weixin.qq.com/s/XS2KOhBeGeviusIqIBnxKg)
  - 如果你的 Linux 服务器上有一个名为 logs 的目录，如何删除该目录下最后一次访问时间超过一年的日志文件呢 `find . -type f -atime +365 -exec rm -rf {} \`
  - 按名称或正则表达式查找文件 `find ./yang/books -type f -name "*.pdf"`
  - 查找不同类型的文件 - 指定 -type 选项来搜索其他类型的文件 ``
  - 按指定的时间戳查找文件
    - 访问时间戳（atime）：最后一次读取文件的时间。
    - 修改时间戳（mtime）：文件内容最后一次被修改的时间。
    - 更改时间戳（ctime）：上次更改文件元数据的时间（如，所有权、位置、文件类型和权限设置）
    - + 表示“大于”，- 表示“小于”。所以我们可以搜索 ctime 在 5~10 天前的文 `find . -type f -ctime +5 -ctime -10`
  - 按大小查找文件 - 要查找大小为 10 MB ~ 1 GB 的文件 `find . -type f -size +10M -size -1G`
  - 按所有权查找文件 - 以下命令将查找所有属于 yang 的文件 `find -type f -user yang`
  - 在找到文件后执行命令 - `-exec` 选项后面的命令必须以分号（;）结束 `find . -type f -atime +5 -exec ls {} \;`





