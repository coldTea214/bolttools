# bolttools

## 概述

[bolt](https://github.com/boltdb/bolt)是一个比较常用的本地kv存储工具，不过生成的kv文件是不可读的。本工具简单的读取kv文件，读出文件内容，方便测试时查看文件内容，帮助定位问题。如果kv文件略有差错，也可以通过本工具简单修改。

## 编译

go build -o bolttools

## 用法

### 查询命令用法

```
// 查询可用命令
root@community-test:/go/src/github.com/coldTea214/bolttools# ./bolttools -h
BoltView is a tool for reading/writting bolt databases.

Usage:

    boltview command [arguments]

The commands are:

    buckets       list buckets in bolt database
    list          list key-value pairs in bucket
    insert        insert a key-value pair into bucket
    delete        delete a key-value pair from bucket

Use "bolt [command] -h" for more information about a command.

// 查询子命令用法
root@community-test:/go/src/github.com/coldTea214/bolttools# ./bolttools buckets -h
usage: bolt buckets PATH

Buckets prints a table of buckets in bolt database
```

### 读取文件内容

```
// 查询文件中buckets及其条目数
root@community-test:/go/src/github.com/coldTea214/bolttools# ./bolttools buckets local-kv.db     
NAME     ITEMS
======== ========
volume   2

// 查询buckets中具体内容       
root@community-test:/go/src/github.com/coldTea214/bolttools# ./bolttools list local-kv.db volume
KEY          VALUE
============ ============
wx-pv        {"tenantId":"6089d765c34a446e93778e1cd4133f72","volumeId":"91d23b47-99bf-46e4-a952-090d2cdf69b7"}
wx-pv2       {"tenantId":"6089d765c34a446e93778e1cd4133f72","volumeId":"23ff616a-72dd-4e51-ba5e-13e0dca70c4d"}
```

### 调整文件内容

```
// 新增条目
root@community-test:/go/src/github.com/coldTea214/bolttools# ./bolttools insert local-kv.db volume hello world
root@community-test:/go/src/github.com/coldTea214/bolttools# ./bolttools list local-kv.db volume
KEY          VALUE
============ ============
hello        world       
wx-pv        {"tenantId":"6089d765c34a446e93778e1cd4133f72","volumeId":"91d23b47-99bf-46e4-a952-090d2cdf69b7"}
wx-pv2       {"tenantId":"6089d765c34a446e93778e1cd4133f72","volumeId":"23ff616a-72dd-4e51-ba5e-13e0dca70c4d"}

// 删除条目
root@community-test:/go/src/github.com/coldTea214/bolttools# ./bolttools delete local-kv.db volume hello
root@community-test:/go/src/github.com/coldTea214/bolttools# ./bolttools list local-kv.db volume        
KEY          VALUE
============ ============
wx-pv        {"tenantId":"6089d765c34a446e93778e1cd4133f72","volumeId":"91d23b47-99bf-46e4-a952-090d2cdf69b7"}
wx-pv2       {"tenantId":"6089d765c34a446e93778e1cd4133f72","volumeId":"23ff616a-72dd-4e51-ba5e-13e0dca70c4d"}
```
