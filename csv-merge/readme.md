### 一个对比文件的小工具

主要目的是工作中处理一个对比csv文件的工作，根据某几个特殊字段做校验，将目标数据数据里有标记的条目标记到源数据上，并生成新的csv文件

```golang
./csv-merge -origin file1.csv -target file2.csv,file3.csv -result result.csv
```