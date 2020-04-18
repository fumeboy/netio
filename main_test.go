package netio

import (
	"fmt"
	"net"
	_ "net/http/pprof"
	"testing"
)

func TestEcho(t *testing.T) {
	tcpaddr, _ := net.ResolveTCPAddr("tcp", "localhost:8990")
	ln, _ := net.ListenTCP("tcp", tcpaddr)
	w, _ := NewServer()
	w.AddListener(
		func(res *result) {
			fmt.Println("ln 1")
			res.conn.read_buffer.Write([]byte("pong!"))
			go res.conn.write(&request{
				buf: res.conn.read_buffer,
			})
			return
		},
		func(res *result) {
			fmt.Println(2, res.Err)
			w.Close(res.conn)
		},
		func(res *result) {
			fmt.Println(3)
			w.Close(res.conn)
		},
		func(res *result) {
			fmt.Println(4, res.Err)
			w.Close(res.conn)
		},
		ln,
	)
	w.Run()
}

func TestEchoClient(t *testing.T) {
	conn, _ := net.Dial("tcp", "localhost:8990")
	conn.Write([]byte("ping"))
	buf := make([]byte,128)
	n, _ := conn.Read(buf)

	fmt.Println(n,string(buf[:n]))
	conn.Close()
}