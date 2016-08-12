package utmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

func (p *Peer) readMessageEx(id, eid uint8, f func(io.Reader, uint32) (uint32, error)) (err error) {
	var found bool
	for !found {
		err = p.readMessage(id, func(r io.Reader, eid2 uint8, size uint32) (uint32, error) {
			if eid == eid2 {
				found = true
				return f(r, size)
			}
			return 0, nil
		})
	}
	return
}

func (p *Peer) readMessage(id uint8, f func(io.Reader, uint8, uint32) (uint32, error)) (err error) {
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
	k := n - 2

	if id == id2 {
		r, err := f(p.conn, eid, k)
		if err != nil {
			return err
		}
		k = k - r
	}

	if k > 0 {
		err = discardMsgData(p.conn, int(k))
		if err != nil {
			return err
		}
	}

	if id == id2 {
		return nil
	}

	if n0 != n || id0 != id2 || eid0 != eid {
		n0, id0, eid0 = n, id2, eid
	} else if cnt++; cnt > 16 {
		err = fmt.Errorf("too many duplicate messages(%d,%d,%d)", n0, id0, eid0)
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
	var b []byte
	err = p.readMessageEx(20, 0, func(r io.Reader, size uint32) (uint32, error) {
		var err error
		if size > 0 {
			b = make([]byte, size)
			err = readMsgData(r, b)
		}
		return size, err
	})
	if err != nil {
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

// NumChunk returns the number of chunk
func (p *Peer) NumChunk() (n int) {
	for s := p.size; s > 0; s -= 16384 {
		n++
	}
	return
}

// ReadChunk download chunk
func (p *Peer) ReadChunk(i int, b []byte) (n int, err error) {
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
	err = p.readMessageEx(20, 1, func(r io.Reader, size uint32) (n uint32, err error) {
		defer func() {
			if x := recover(); x != nil {
				err = errors.New("happen panic when read message data")
			}
		}()

		mn := min(size, 64)
		mb := make([]byte, mn)
		err = readMsgData(r, mb)
		if err != nil {
			return
		}

		i := bytes.Index(mb, []byte("ee"))
		if i == -1 {
			err = errors.New("read metadata failure")
			return
		}
		m := new(metadata)
		err = bencode.DecodeBytes(mb[:i+2], m)
		if err != nil {
			return
		}
		if m.Type != 1 {
			err = errors.New("rejected the request")
			return
		}

		n0 := uint32(i) + 2
		if n0 < mn {
			copy(b, mb[n0:])
		}
		n1 := mn - n0
		n2 := min(size-n0, uint32(len(b)))
		err = readMsgData(r, b[n1:n2])
		n = n0 + n2
		return
	})
	return
}

func min(n1, n2 uint32) uint32 {
	if n1 < n2 {
		return n1
	}
	return n2
}
