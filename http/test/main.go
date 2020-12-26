package main

import (
	"fmt"
	"github.com/fumeboy/netio/http"
)

func main(){
	http.Run(":8000", func(req *http.Request, resp *http.Response) {
		fmt.Println(req.Url)
		resp.Write([]byte("hello world "))
	})
}
