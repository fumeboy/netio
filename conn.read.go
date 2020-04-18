package netio

import (
	"io"
	"syscall"

	SmartBuffer "github.com/fumeboy/netio/bufferpool"
)

func (this *connection) read() {
	var nr int
	var err error
	var buffer *SmartBuffer.Buffer
	for {
		if this.read_buffer == nil {
			buffer = this.user.Get()
			this.read_buffer = buffer
		} else {
			buffer = this.read_buffer
		}
		buf_rest, _ := buffer.RestBytes()
		nr, err = syscall.Read(this.file_desc, buf_rest)
		buffer.AfterWriteToRestBytes(nr)
		// 对于非阻塞的网络连接的文件描述符，如果错误是EAGAIN
		// 说明Socket的缓冲区为空，未读取到任何数据
		if err != nil {
			if err == syscall.EAGAIN {
				continue
			}

			// On MacOS we can see EINTR here if the user
			// pressed ^Z.
			if err == syscall.EINTR {
				continue
			}

			break
		}
		// proper setting of EOF
		if nr == 0 {
			err = io.EOF
		}
		break
	}
	res := &result{
		conn: this,
		Err:  err,
	}
	if err == nil || err == io.EOF {
		this.ln.read_success_callback_fn(res)
	} else {
		this.ln.read_failed_callback_fn(res)
	}
}
