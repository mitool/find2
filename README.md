# find2
用golang写的查找/替换文件内容的工具。

# 特性
1. 支持按内容搜索文件
2. 支持替换文件内容
3. 支持将匹配到的文件保存到指定文件夹
4. 支持将匹配到的文件进行zip压缩(保存为compress.zip)
5. 匹配规则统一使用正则表达式

## 用法
执行以下命令
```
go build

# 查看用法
find2 -h
```
example:
```
./find2 -findPath ./www -fileRule \.php$ -contentRule "a\.user|a\.profile" -savePath ./found -compress true
```
本例中，压缩后的文件保存在"./found/compress.zip"。

## 用参数组合实现特定功能
1. 如果不设置`-contentRule`,并且设置`-compress`为`true`,则会将符合规则的文件添加到zip压缩包。
2. 如果不设置`-contentRule`,并且设置`-compress`为`false`或者不设置,则会将符合规则的文件复制到`-savePath`指定的目录中，
如果不指定`-savePath`的值则复制到当前目录中。