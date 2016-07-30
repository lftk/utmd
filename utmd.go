package utmd

import (
	"bytes"
	"errors"
	"net"

	"github.com/4396/dht"
)

var (
	peerid   = []byte("-ZJ4396-123456789000")
	reserved = []byte{0, 0, 0, 0, 0, 0x10, 0, 0}
	protocol = []byte("BitTorrent protocol")
	protolen = byte(19)
)

var (
	errPeerID = errors.New("unmatched peer id")
)

func handshake(conn net.Conn, id, tor []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(protolen)
	buf.Write(protocol)
	buf.Write(reserved)
	buf.Write(tor)
	buf.Write(id)
	b1 := buf.Bytes()
	_, err := conn.Write(b1)
	if err != nil {
		return nil, err
	}

	b2 := make([]byte, buf.Len())
	n, err := conn.Read(b2)
	if err != nil {
		return nil, err
	}
	if n != buf.Len() {
		return nil, errors.New("invalid handshake message")
	}
	if bytes.Compare(b1[:20], b2[:20]) != 0 {
		return nil, errors.New("invalid protocol")
	}
	if b2[25] != 0x10 {
		return nil, errors.New("unsupported extension")
	}
	if bytes.Compare(b1[28:48], b2[28:48]) != 0 {
		return nil, errors.New("unmatched info hash")
	}
	if bytes.Compare(b1[48:56], b2[48:56]) != 0 {
		return b2[48:56], errPeerID
	}
	return nil, nil
}

// Handshake bt peer
func Handshake(addr string, tor []byte) (*Peer, error) {
	id := []byte("-ZJ4396-123456789000")
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		name, err := handshake(conn, id, tor)
		if err == nil {
			return &Peer{tor: tor, conn: conn}, nil
		}
		if conn.Close(); err != errPeerID {
			return nil, err
		}
		copy(id, name)
	}
}

// Download torrent metadata
func Download(addr *net.TCPAddr, tor *dht.ID) error {

	return nil
}
