CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go

docker build -t netio_http_test .

docker run -p 8001:8000  netio_http_test