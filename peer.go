package utmd

import (
	"bytes"
	"errors"
	"fmt"
	"net"

	"github.com/zeebo/bencode"
)

// Peer has the metadata of torrent
type Peer struct {
	conn  net.Conn
	tor   []byte
	index uint8
	size  int
}

func newPeer(tor []byte, conn net.Conn) *Peer {
	return &Peer{tor: tor, conn: conn}
}

// Close connection
func (p *Peer) Close() {
	p.conn.Close()
}

func (p *Peer) readHeader() (n uint32, id, eid uint8, err error) {
	n, err = readUint32(p.conn)
	if err != nil {
		return
	}
	if n > 0 {
		id, err = readUint8(p.conn)
		if err != nil {
			return
		}
	}
	if n > 1 {
		eid, err = readUint8(p.conn)
		if err != nil {
			return
		}
	}
	return
}

func (p *Peer) readMessage(id uint8) (eid uint8, data []byte, err error) {
	var (
		cnt       int
		n0        uint32
		id0, eid0 uint8
	)

LOOP:
	n, id2, eid, err := p.readHeader()
	if err != nil {
		return
	}

	if id == id2 {
		if n > 2 {
			data = make([]byte, n-2)
			err = readMsgData(p.conn, data)
		}
		return
	}

	if n > 2 {
		discardMsgData(p.conn, int(n-2))
	}

	if n0 != n || id0 != id2 || eid0 != eid {
		n0, id0, eid0 = n, id2, eid
	} else if cnt++; cnt > 16 {
		err = fmt.Errorf("too many duplicate messages(%d,%d,%d)", n0, id0, eid0)
		return
	}

	goto LOOP
}

func (p *Peer) readMessageEx(id, eid uint8) (data []byte, err error) {
LOOP:
	eid2, data, err := p.readMessage(id)
	if err != nil || eid == eid2 {
		return
	}
	goto LOOP
}

func (p *Peer) writeMessage(id uint8, index uint8, data []byte) error {
	msg := &message{id, index, data}
	return writeMessage(p.conn, msg)
}

func (p *Peer) handshake() (err error) {
	// send handshake message
	err = p.writeMessage(20, 0, []byte("d1:md11:ut_metadatai1eee"))
	if err != nil {
		return
	}
	// recv handshake message
	i, b, err := p.readMessage(20)
	if err != nil {
		return
	}
	if i != 0 {
		err = errors.New("handshake failure")
		return
	}
	var h struct {
		M    map[string]int64 `bencode:"m"`
		Size int64            `bencode:"metadata_size"`
	}
	err = bencode.DecodeBytes(b, &h)
	if err != nil {
		return
	}
	p.size = int(h.Size)
	p.index = uint8(h.M["ut_metadata"])
	return
}

type metadata struct {
	Type  int64 `bencode:"msg_type"`
	Piece int64 `bencode:"piece"`
}

// Size return metadata size
func (p *Peer) Size() int {
	return p.size
}

// NumBlocks returns the number of block
func (p *Peer) NumBlocks() (n int) {
	for s := p.size; s > 0; s -= 16384 {
		n++
	}
	return
}

// ReadBlock download block
func (p *Peer) ReadBlock(i int) (b []byte, err error) {
	// request metadata
	var md metadata
	md.Type = 0
	md.Piece = int64(i)
	data, err := bencode.EncodeBytes(&md)
	if err != nil {
		return
	}
	err = p.writeMessage(20, p.index, data)
	if err != nil {
		return
	}
	// download metadata
	data, err = p.readMessageEx(20, 1)
	if err != nil {
		return
	}
	n := bytes.Index(data, []byte("ee"))
	if n != -1 {
		b = data[n+2:]
	}
	return
}
