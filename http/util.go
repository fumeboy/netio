package http

import (
	"reflect"
	"unsafe"
)

func BytesToStringFast(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StringToBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{sh.Data, sh.Len, 0}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
