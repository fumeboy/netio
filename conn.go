package netio

import (
	"bytes"
	"sync"
	"syscall"
)

type Conn struct {
	closeOnce     sync.Once
	readData    []byte // TODO 缓存池
	readBuffer  bytes.Buffer
	writeBuffer bytes.Buffer
	rLock       sync.Mutex
	wLock       sync.Mutex
	Fd          int

	Agreement interface{} //存储协议数据
}

func (c *Conn) new(fd int) *Conn {
	c.Fd = fd
	c.readData = make([]byte, 1024)
	c.Agreement = globalServer.handler.WhenOpen()
	return c
}

func (c *Conn) Send(data []byte) {
	c.wLock.Lock()
	defer c.wLock.Unlock()
	//如果 当前 write 缓冲区 还有 数据未发送,那直接 append 到缓冲区中
	if c.writeBuffer.Len() != 0 {
		c.writeBuffer.Write(data)
		return
	}
	//尝试对 socket 进行写数据
	nums, err := syscall.Write(c.Fd, data)
	if err != nil && err != syscall.EAGAIN {
		c.close()
		return
	}
	if nums >= len(data) {
		return
	}
	c.writeBuffer.Write(data[nums:])
}

func (c *Conn) write() { // 可写
	/*
		ET模式下，EPOLLOUT触发条件有：

			1.缓冲区满-->缓冲区非满；
			2.同时监听EPOLLOUT和EPOLLIN事件 时，当有IN 事件发生，都会顺带一个OUT事件；
			3.一个客户端connect过来，accept成功后会触发一次OUT事件。

		其中2最令人费解，内核代码这块有注释，说是一般有IN 时候都能OUT，就顺带一个，多给了个事件。。
	*/
	c.wLock.Lock()
	defer c.wLock.Unlock()
	nums, err := syscall.Write(c.Fd, c.writeBuffer.Bytes())
	if err != nil {
		if err == syscall.EAGAIN {
			return
		}
		c.close()
		return
	}
	c.writeBuffer.Next(nums)
}

func (c *Conn) read() { // 可读
	c.rLock.Lock()
	defer c.rLock.Unlock()

	for {
		nums, err := syscall.Read(c.Fd, c.readData)
		if err == syscall.EAGAIN {
			// 对于非阻塞的网络连接的文件描述符，如果错误是EAGAIN
			// 说明Socket的缓冲区为空，未读取到任何数据
			return
		}
		if err != nil || nums == 0 { // EOF
			c.close()
			return
		}
		if nums == len(c.readData) {
			c.readBuffer.Write(c.readData)
			continue
		}
		c.readBuffer.Write(c.readData[:nums])
		break
	}

	globalServer.handler.WhenRead(c, c.readBuffer.Bytes())
	c.readBuffer.Reset()
}

func (c *Conn) close() {
	c.closeOnce.Do(func() {
		globalEpoll.del(c.Fd)
		delete(globalServer.connections, c.Fd)
		if err := syscall.Close(c.Fd); err != nil {
			//log.Errorf("unix close err %v", err)
		}
	})
}
