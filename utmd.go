package utmd

import (
	"bytes"
	"errors"
	"net"
)

var (
	reserved = []byte{0, 0, 0, 0, 0, 0x10, 0, 0}
	protocol = "BitTorrent protocol"
)

func checkHandshake(h *handshake, tor []byte) error {
	if h.p != protocol {
		return errors.New("invalid protocol")
	}
	if bytes.Compare(h.tor, tor) != 0 {
		return errors.New("unmatched info hash")
	}
	for i := 0; i < 8; i++ {
		if r := reserved[i]; r != 0 && h.r[i] != r {
			return errors.New("unmatched reserved")
		}
	}
	return nil
}

func hello(addr string, id, tor []byte) (net.Conn, []byte, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	h := &handshake{protocol, reserved, tor, id}
	err = writeHandshake(conn, h)
	if err != nil {
		goto EXIT
	}
	h, err = readHandshake(conn)
	if err != nil {
		goto EXIT
	}
	err = checkHandshake(h, tor)
	if err != nil {
		goto EXIT
	}
	return conn, h.id, nil

EXIT:
	conn.Close()
	return nil, nil, err
}

func newPeerID(addr string, tor []byte) ([]byte, error) {
	conn, id, err := hello(addr, []byte("-ZJ4396-123456789000"), tor)
	if err != nil {
		return nil, err
	}
	conn.Close()
	copy(id[8:], []byte("123456789000"))
	return id, nil
}

// Handshake connection with addr and returns a peer
func Handshake(addr string, tor []byte) (p *Peer, err error) {
	id, err := newPeerID(addr, tor)
	if err != nil {
		return
	}
	conn, _, err := hello(addr, id, tor)
	if err != nil {
		return
	}
	p = newPeer(tor, conn)
	if err = p.handshake(); err != nil {
		conn.Close()
	}
	return
}

// Download metadata of torrent(tor infohash)
func Download(addr string, tor []byte, cb func(down, total int)) (b []byte, err error) {
	p, err := Handshake(addr, tor)
	if err != nil {
		return
	}
	defer p.Close()
	if size := p.Size(); size > 0 {
		b = make([]byte, size)
		if cb != nil {
			cb(0, size)
		}
		for i, n := 0, 0; i < p.NumBlocks(); i++ {
			data, err := p.ReadBlock(i)
			if err != nil {
				return nil, err
			}
			copy(b[n:], data)
			n += len(data)
			if cb != nil {
				cb(n, size)
			}
		}
	}
	return
}
