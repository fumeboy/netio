netio

依赖 epoll 实现的异步网络IO库

特性：
    被动读，主动写

使用的对象：
    server
    listner
    connection
    poller
    event
    Request （导出）
    Result  （导出）

流程：
    创建 server
    在 server 上注册 listener
    listener 接受 connection，并将 conn 注册到 epoll
    epoll 返回 events， server 接收并根据 events 进行读写
    依据读写的结果，分别调用 listener 初始化时绑定的四个函数：
                                              read_success
                                              read_failed
                                              write_success
                                              write_failed

疑问：
    网络这方面知道的还是很少，但是单个conn的读写应该是顺序执行的吧？
    应该没有并发写单个conn的情景才对
    所以这个程序也是认为『只有顺序读顺序写』来设计的

参考：
    https://github.com/xtaci/gaio

