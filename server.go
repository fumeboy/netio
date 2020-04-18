package netio

import (
	"fmt"
)

type Server struct {
	poller   *poller
	conn_set map[int]*connection

	listeners []*listener

	stop chan int8
}

func NewServer() (*Server, error) {
	var err error
	ser := &Server{}
	ser.poller, err = openPoll()
	ser.stop = make(chan int8, 1)
	ser.listeners = []*listener{}
	ser.conn_set = map[int]*connection{}
	return ser, err
}

func (this *Server) Run() {
	fmt.Println("start!")
	go this.wait()
	for _, ln := range this.listeners {
		go ln.keep_accepting()
	}
	for {
		select {
		case <-this.stop:
			break
		}
	}
}

func (this *Server) handle_events(events []event) {
	var conn *connection
	var ok bool
	for i, l := 0, len(events); i < l; i++ {
		e := events[i]
		conn, ok = this.conn_set[e.ident]
		if !ok {
			continue
		}
		if e.r {
			go conn.read()
		}
		if e.w {

		}
	}
}
