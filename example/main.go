package main

import (
	"fmt"

	"github.com/4396/dht"
	"github.com/4396/utmd"
)

func main() {
	peer := new(utmd.Peer)
	err := peer.Dial("95.213.229.154:54949")
	if err != nil {
		fmt.Println(err)
		return
	}
	tor, _ := dht.ResolveID("e94e84f2919507ea99f61cb3d733e266e047b6b2")
	err = peer.Handshake(tor)
	if err != nil {
		fmt.Println(err)
		return
	}
	/*
		for {
			b := make([]byte, 16*1024)
			n, err := peer.RecvData(b)
			fmt.Println(string(b[:n]), n, err)
			if err != nil {
				break
			}
		}
	*/
}
