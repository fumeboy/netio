package netio

import (
	"net"
	"syscall"
)

type Handler struct {
	WhenOpen func() interface{}
	WhenRead func(c *Conn, bytes []byte)
}

type server struct {
	connections map[int]*Conn
	handler Handler
}

var globalServer server

func (h Handler) Run(addr string) (error) {
	defer closeAll()
	if err := globalEpoll.init(); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return  err
	}
	defer listener.Close()

	globalServer.connections = map[int]*Conn{}
	globalServer.handler = h

	go globalEpoll.run()

	for {
		if raw_conn,err := listener.Accept();err == nil{
			rc := raw_conn.(*net.TCPConn)
			file, err := rc.File()
			if err != nil{

			}else{
				fd :=int(file.Fd())
				globalServer.connections[fd] = (&Conn{}).new(fd)
				globalEpoll.add(fd)
			}
		}else{
		}
	}
}

func closeAll() {
	globalEpoll.close()
	<- globalEpoll.finishChan
	for _,c := range globalServer.connections {
		syscall.Close(c.Fd)
	}
	globalServer = server{}
	globalEpoll = epoll{}
}

