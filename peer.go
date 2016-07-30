package utmd

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/zeebo/bencode"
)

// Peer bt peer
type Peer struct {
	tor   []byte
	conn  net.Conn
	index int64
	size  int64
	data  []byte
}

var hiMsg []byte

func init() {
	data := []byte("d1:md11:ut_metadatai1eee")
	prefix := make([]byte, 4)
	binary.BigEndian.PutUint32(prefix, uint32(len(data)+2))
	buf := bytes.NewBuffer(nil)
	buf.Write(prefix)
	buf.WriteByte(20)
	buf.WriteByte(0)
	buf.Write(data)
	hiMsg = buf.Bytes()
}

func (p *Peer) Close() {
	p.conn.Close()
}

func (p *Peer) handshake() (err error) {
	// send handshake message
	_, err = p.conn.Write(hiMsg)
	if err != nil {
		return
	}
	// recv handshake message
	header := make([]byte, 6)
	_, err = p.conn.Read(header)
	if err != nil {
		return
	}
	n := binary.BigEndian.Uint16(header) - 2
	if header[4] != 20 || header[5] != 0 {
		err = errors.New("get metadata failed")
		return
	}
	b := make([]byte, int(n))
	_, err = p.conn.Read(b)
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
	p.size = h.Size
	p.index = h.M["ut_metadata"]
	return
}

type metadata struct {
	Type  int64 `bencode:"msg_type"`
	Piece int64 `bencode:"piece"`
}

// Download metadata
func (p *Peer) Download() (err error) {
	err = p.handshake()
	if err != nil {
		return
	}
	for i := 0; i < int((p.size+1)/16*1024); i++ {
		err = p.download(i)
		if err != nil {
			return
		}
	}
	fmt.Println(string(p.data))
	return
}

func (p *Peer) download(piece int) (err error) {
	var md metadata
	md.Type = 0
	md.Piece = int64(piece)
	data, err := bencode.EncodeBytes(&md)
	if err != nil {
		return
	}

	index := byte(p.index)
	prefix := make([]byte, 4)
	binary.BigEndian.PutUint32(prefix, uint32(len(data)+2))
	buf := bytes.NewBuffer(nil)
	buf.Write(prefix)
	buf.WriteByte(20)
	buf.WriteByte(index)
	buf.Write(data)

READ:
	size, id, index := readHeader(p.conn)
	if size == 0 {
		err = errors.New("-----")
		return
	}
	if size > 0 {
		b := make([]byte, size-2)
		n, err := p.conn.Read(b)
		if err != nil {
			return err
		}
		if id != 20 || index != 1 {
			goto READ
		}
		p.data = append(p.data, b[:n]...)
	}
	return
}

func readHeader(conn net.Conn) (size uint32, id, index uint8) {
	b := make([]byte, 6)
	_, err := conn.Read(b)
	if err != nil {
		return
	}
	size = binary.BigEndian.Uint32(b)
	id = uint8(b[4])
	index = uint8(b[5])
	return
}

func sendMsg(conn *net.TCPConn, msg []byte) error {
	var nn int
	for nn < len(msg) {
		n, err := conn.Write(msg[nn:])
		if err != nil {
			return err
		}
		nn += n
	}
	return nil
}

func recvMsg(conn *net.TCPConn, msg []byte) error {
	var nn int
	for nn < len(msg) {
		n, err := conn.Read(msg[nn:])
		if err != nil {
			return err
		}
		nn += n
	}
	return nil
}
