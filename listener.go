package netio

import (
	"net"

	SmartBuffer "github.com/fumeboy/netio/bufferpool"
)

type listener struct {
	read_success_callback_fn  func(res *result)
	read_failed_callback_fn   func(res *result)
	write_success_callback_fn func(res *result)
	write_failed_callback_fn  func(res *result)

	inner net.Listener
	user  *SmartBuffer.User

	server *Server
}

func (this *Server) AddListener(
	read_success_callback_fn func(res *result),
	read_failed_callback_fn func(res *result),
	write_success_callback_fn func(res *result),
	write_failed_callback_fn func(res *result),
	rawln net.Listener,
) {
	ln := &listener{}
	ln.inner = rawln
	ln.read_success_callback_fn = read_success_callback_fn
	ln.read_failed_callback_fn = read_failed_callback_fn
	ln.write_success_callback_fn = write_success_callback_fn
	ln.write_failed_callback_fn = write_failed_callback_fn
	ln.user = SmartBuffer.NewUser(SmartBuffer.LV_1024x16)
	ln.server = this
	this.listeners = append(this.listeners, ln)
}

func (this *listener) keep_accepting() {
	for {
		raw_conn, err := this.inner.Accept()
		if err != nil {
			// handle error
		}
		conn := &connection{
			inner: raw_conn,
			user:  this.user,
			ln:    this,
		}
		if conn.file_desc, err = dupconn(raw_conn); err != nil {

		} else {
			err = this.server.poller.watch(conn.file_desc)
			if err != nil {
			}
			this.server.conn_set[conn.file_desc] = conn
		}
	}
}
