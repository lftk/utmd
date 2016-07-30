package main

import (
	"fmt"

	"github.com/4396/dht"
	"github.com/4396/utmd"
)

func main() {
	tor, _ := dht.ResolveID("497673c9d0ca952492dc3a092115ac587e3a01d9")
	p, err := utmd.Handshake("111.182.197.21:11101", tor.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer p.Close()
	fmt.Println("----------")
	if err = p.Download(); err != nil {
		fmt.Println(err)
	}
}
