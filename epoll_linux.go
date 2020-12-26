package netio

import (
	"sync"
	"syscall"
	"unsafe"
)

const (
	maxEvents = 1024

	EPOLLET_ = 0x80000000
	// syscall.EPOLLET 的值是 -0x80000000，应该是错误的， 很奇怪
	// ET 边缘触发
)

var globalEpoll epoll

type epoll struct {
	fd                     int
	wakeUpEventFd          int
	wakeUpEventFdSignalOut []byte

	toAdd      []int // 可以考虑改成 chan
	toAddMutex sync.Mutex

	closeChan  chan struct{}
	closeOnce  sync.Once
	finishChan chan struct{}
}

func (p *epoll) init() error {
	fd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return err
	}
	r0, _, e0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0) // 调用 eventfd 函数
	if e0 != 0 {
		syscall.Close(fd)
		return err
	}
	if err := syscall.EpollCtl(fd, syscall.EPOLL_CTL_ADD, int(r0),
		&syscall.EpollEvent{Fd: int32(r0),
			Events: syscall.EPOLLIN,
		},
	); err != nil {
		syscall.Close(fd)
		syscall.Close(int(r0))
		return err
	}
	p.fd = fd
	p.wakeUpEventFd = int(r0)
	p.wakeUpEventFdSignalOut = make([]byte, 8)
	p.closeChan = make(chan struct{})
	p.finishChan = make(chan struct{})
	return err
}

func (p *epoll) add(fd int) error {
	p.toAddMutex.Lock()
	p.toAdd = append(p.toAdd, fd)
	p.toAddMutex.Unlock()

	return p.wakeup() // 使用 wakeup 使 epoll wait 结束阻塞
}

func (p *epoll) del(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (p *epoll) close() error {
	p.closeOnce.Do(func() {
		close(p.closeChan)
	})
	return p.wakeup()
}

func (p *epoll) wakeup() error {
	var x uint64 = 1 // 非 0 值
	_, err := syscall.Write(p.wakeUpEventFd, (*(*[8]byte)(unsafe.Pointer(&x)))[:])
	return err
}

func (p *epoll) run() {
	defer func() {
		syscall.Close(p.fd)
		syscall.Close(p.wakeUpEventFd)
		p.fd = -1
		p.wakeUpEventFd = -1
		p.finishChan <- struct{}{}
	}()
	events := make([]syscall.EpollEvent, maxEvents)
	for {
		select {
		case <-p.closeChan:
			return
		default:
			p.toAddMutex.Lock()
			for _, fd := range p.toAdd {
				syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, int(fd), &syscall.EpollEvent{Fd: int32(fd), Events: syscall.EPOLLRDHUP | syscall.EPOLLIN | syscall.EPOLLOUT | EPOLLET_})
			}
			p.toAdd = p.toAdd[:0]
			p.toAddMutex.Unlock()

			n, err := syscall.EpollWait(p.fd, events, -1)
			if err == syscall.EINTR {
				continue
			}
			if err != nil {
				return
			}
			for i := 0; i < n; i++ {
				ev := &events[i]
				if int(ev.Fd) == p.wakeUpEventFd {
					syscall.Read(p.wakeUpEventFd, p.wakeUpEventFdSignalOut)
				} else if c, ok := globalServer.connections[int(ev.Fd)]; ok {
					/*
						EPOLLIN ：表示对应的文件描述符可以读（包括对端SOCKET正常关闭）；
						EPOLLOUT：表示对应的文件描述符可以写；
						EPOLLPRI：表示对应的文件描述符有紧急的数据可读（这里应该表示有带外数据到来）；
						EPOLLERR：表示对应的文件描述符发生错误；
						EPOLLHUP：表示对应的文件描述符被挂断；
						EPOLLET： 将EPOLL设为边缘触发(Edge Triggered)模式，这是相对于水平触发(Level Triggered)来说的。
						EPOLLONESHOT：只监听一次事件，当监听完这次事件之后，如果还需要继续监听这个socket的话，需要再次把这个socket加入到EPOLL队列里
					*/
					if ((events[i].Events & syscall.EPOLLHUP) != 0) && ((events[i].Events & syscall.EPOLLIN) == 0) {
						c.close()
						continue
					}
					if events[i].Events&(syscall.EPOLLIN|syscall.EPOLLPRI|syscall.EPOLLRDHUP) != 0 {
						c.read()
					}
					if (events[i].Events&syscall.EPOLLERR != 0) || (events[i].Events&syscall.EPOLLOUT != 0) {
						c.write()
					}
				}
			}
		}
	}
}
