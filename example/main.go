package main

import (
	"fmt"

	"github.com/4396/dht"
	"github.com/4396/utmd"
)

func download(s string) (b []byte, err error) {
	tor, err := dht.ResolveID(s[:40])
	if err != nil {
		return
	}
	b, err = utmd.Download(s[41:], tor.Bytes(), func(down, total int) {
		fmt.Println(s, down, total)
	})
	return
}

func tryDownload(s string, n int) (data []byte, err error) {
	for i := 0; i < n; i++ {
		data, err = download(s)
		if err == nil {
			return
		}
	}
	return
}

func main() {
	data, err := tryDownload("497673c9d0ca952492dc3a092115ac587e3a01d9 111.182.197.21:11101", 3)
	fmt.Println(err, string(data))
}
