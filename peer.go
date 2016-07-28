package utmd

import (
	"bytes"
	"math/rand"
	"net"

	"github.com/4396/dht"
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

// Handshake with bt peer
func (p *Peer) Handshake(tor *dht.ID) error {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(byte(19))
	buf.Write([]byte("BitTorrent protocol"))
	var ext [8]byte
	ext[5] = 0x10
	buf.Write(ext[:])
	buf.Write(tor.Bytes())
	buf.Write(randomReerID())
	_, err := p.conn.Write(buf.Bytes())
	return err
}

// RecvData from bt peer
func (p *Peer) RecvData(b []byte) (int, error) {
	return p.conn.Read(b)
}
