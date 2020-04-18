package netio

import (
	"syscall"
)

func (this *connection) write(req *request) {
	var n int
	var err error
	var res *result
	for {
		buf := req.buf.ReadRestBytes()
		n, err = syscall.Write(this.file_desc, buf)
		if err == syscall.EAGAIN {
			res = &result{
				conn: this,
				Err:  nil,
			}
			this.ln.write_failed_callback_fn(res)
			break
		}
		// if err == syscall.EINTR {
		// 	continue
		// }
		// 手动更新 buf
		req.buf.MoveReadOffset(n)
		if err == nil {
			// all bytes written
			if req.buf.ReadOffset() == req.buf.UsedSize() {
				this.write_request = nil
				res = &result{
					conn: this,
					Err:  nil,
				}
				this.ln.write_success_callback_fn(res)
				break
			}
		} else {
			this.write_request = nil
			res = &result{
				conn: this,
				Err:  err,
			}
			this.ln.write_failed_callback_fn(res)
			break
		}
	}
}
