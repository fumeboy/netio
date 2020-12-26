package http

import (
	"fmt"
	"github.com/fumeboy/netio"
)

var handlerFunc func(req *Request, resp *Response)

var http = netio.Handler{
	WhenOpen: func() interface{} {
		return &httpAgreement{}
	},
	WhenRead: func(conn *netio.Conn, data []byte) {
		agreement := conn.Agreement.(*httpAgreement)

		if agreement.request != nil {
			agreement.requestBuffer = append(agreement.requestBuffer, data...)

			if agreement.request.BodyNum > len(agreement.requestBuffer) {
				return
			}

			agreement.request.Body = agreement.requestBuffer[:agreement.request.BodyNum]
			agreement.requestBuffer = agreement.requestBuffer[agreement.request.BodyNum:]

			resp := productionHttpResponse(agreement.request)
			handlerFunc(agreement.request, resp)
			conn.Send(resp.change2bytes())
			agreement.request = nil
			return
		}

		var (
			temData   []byte
			appendLen int
		)

		if len(agreement.requestBuffer) < 4 {
			appendLen = len(agreement.requestBuffer)
			temData = append(agreement.requestBuffer, data...)
		} else {
			appendLen = 3
			temData = append(agreement.requestBuffer[len(agreement.requestBuffer)-3:], data...)
		}

		headerEndNum, flag := findAgreementSpilt(temData)
		if flag == false {
			conn.Agreement = append(agreement.requestBuffer, data...)
			return
		}

		headerEndNum -= appendLen
		contentBytes := append(agreement.requestBuffer, data...)
		headerBytes := contentBytes[:headerEndNum]
		bodyBytes := contentBytes[headerEndNum:]
		agreement.request = parseHttpHeader(headerBytes)

		if agreement.request == nil {
			return
		}

		if agreement.request.BodyNum <= len(bodyBytes) {
			agreement.request.Body = bodyBytes[:agreement.request.BodyNum]

			resp := productionHttpResponse(agreement.request)
			handlerFunc(agreement.request, resp)
			conn.Send(resp.change2bytes())
			agreement.request = nil
		}

		agreement.requestBuffer = bodyBytes
	},
}

func Run(addr string, handler func(req *Request, resp *Response))  {
	handlerFunc = handler
	if err := http.Run(addr); err != nil{
		fmt.Println(err)
	}
}
