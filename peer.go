package utmd

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"

	"github.com/4396/dht"
	"github.com/zeebo/bencode"
)

// Peer bt peer
type Peer struct {
	conn *net.TCPConn
}

func randomReerID() []byte {
	var id [20]byte
	copy(id[:], []byte("-XL2016-"))
	rand.Read(id[8:])
	return id[:]
}

// Dial bt peer
func (p *Peer) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	p.conn = conn.(*net.TCPConn)
	return nil
}

type exthandshake struct {
	//M map[string]int `bencode:"m"`
	P int64  `bencode:"p"`
	V string `bencode:"v"`
}

// Handshake with bt peer
func (p *Peer) Handshake(tor *dht.ID) (err error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(byte(19))
	buf.Write([]byte("BitTorrent protocol"))
	var ext [8]byte
	ext[5] = 0x10
	buf.Write(ext[:])
	buf.Write(tor.Bytes())
	buf.Write(randomReerID())
	_, err = p.conn.Write(buf.Bytes())
	if err != nil {
		return
	}
	b := make([]byte, 1024)
	n, err := p.conn.Read(b)
	if err != nil {
		return
	}
	fmt.Println(n)
	fmt.Println(string(b[74:n]))
	p.decodeHandshake(b)
	var msg exthandshake
	err = bencode.DecodeBytes(b[22:n], &msg)
	if err != nil {
		return
	}
	fmt.Println(msg.V)
	return
}

func (p *Peer) decodeHandshake(b []byte) {
	fmt.Println(b[0], b[1], b[2], b[3], b[4], b[5])
}

func (p *Peer) RecvHandleshake() {
}

// ExtHandshake with bt peer
func (p *Peer) ExtHandshake() {

}

// RecvData from bt peer
func (p *Peer) RecvData(b []byte) (int, error) {
	return p.conn.Read(b)
}
