# netio

一个依赖 epoll 实现的网络IO库

## 实例

请见 ./http/test/main.go

## 流程

初始化
    创建全局单例 globalServer
    创建全局单例 globalEpoll

go 协程 执行 globalEpoll.run()

创建 TCP listener
    循环执行 Accept， 除非程序被强制退出
        accept 得到 conn 后，添加到 connections 中， connections 是一个 map
        同时添加 conn.fd 到 epoll 中进行监听

globalEpoll 通过 wait() 得到 conn fd 上发生的 事件
    归纳这些事件为 可读 / 可写 / 错误 事件
        接收到 可读事件 后， 读出数据， 并传给 业务 函数
        业务函数通常会有发送数据的操作
        如果 可写， 则写出要发送的数据

退出时，
    globalEpoll.close()
    遍历 connections 执行 conn.close()
    对 globalServer 和 globalEpoll 置空
