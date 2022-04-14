# gdc (google drive cli)

## 服务帐号全网域授权
`https://www.googleapis.com/auth/drive`
`https://www.googleapis.com/auth/admin.directory.group`

## 命令

单文件命令:
* `cat` 读取文件
* `cp` 复制文件
* `mv` 移动文件
* `rm` 删除文件

目录命令:
* `ls` 列出共享盘或共享盘的文件 
* `mb` 创建共享盘
* `rb` 删除共享盘 
* `sync` 同步文件夹到共享盘

群组命令:
* `group`
  * `adduser` 添加用户
  

示例:

```bash
gdc cat --range 0-1024 --count 1 filename.txt
```




