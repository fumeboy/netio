package http

type (
	httpAgreement struct {
		requestBuffer []byte
		request       *Request
	}
)

func findAgreementSpilt(data []byte) (int, bool) {
	for i := 0; i < len(data)-3; i++ {
		if data[i] == '\r' && data[i+1] == '\n' && data[i+2] == '\r' && data[i+3] == '\n' {
			return i + 3, true
		}
	}

	return 0, false
}
