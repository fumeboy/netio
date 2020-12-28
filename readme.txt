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

## 下一步
    0
        在 net/http 中， 对每一个 conn 都开了一个协程，这被认为是浪费的
        需要写一个协程池用于处理 conn 读写
    1
        每个 conn 有一个上下文，用于存储必要的数据， 比如 http 请求
        需要写一个对象池，用于复用这些上下文结构体
    2
        实现应用层协议，比如 HTTP
        这是最重要的一步， 绝大多数同这个项目一类的项目都放弃了做这件事
        因为真正要处理协议太麻烦了！
        如果有信心，可以借鉴 net/http 中的程序自己实现对 HTTP1.1 TLS HTTP2 这些协议的支持

