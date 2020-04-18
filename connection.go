package netio

import (
	"net"
	"syscall"

	SmartBuffer "github.com/fumeboy/netio/bufferpool"
)

type request struct {
	buf *SmartBuffer.Buffer
}

type connection struct {
	write_request *request // 主动写
	                       // conn 不可能同时间有多个 request
	                       // 不存在同一时间的并发读和并发写

	file_desc     int
	inner         net.Conn

	user        *SmartBuffer.User
	read_buffer *SmartBuffer.Buffer // 被动读

	ln *listener
}

func (this *Server) Close(conn *connection){
	delete(this.conn_set, conn.file_desc)
	syscall.Close(conn.file_desc)
}