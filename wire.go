package utmd

import (
	"bytes"
	"encoding/binary"
	"io"
)

func readUint8(r io.Reader) (n uint8, err error) {
	b1 := make([]byte, 1)
	err = readMsgData(r, b1)
	if err == nil {
		n = b1[0]
	}
	return
}

func readUint32(r io.Reader) (n uint32, err error) {
	b4 := make([]byte, 4)
	err = readMsgData(r, b4)
	if err != nil {
		return
	}
	n = binary.BigEndian.Uint32(b4)
	return
}

func readMsgData(r io.Reader, b []byte) (err error) {
	var nn int
	for nn < len(b) {
		n, err := r.Read(b[nn:])
		if err != nil {
			break
		}
		nn += n
	}
	return
}

func discardMsgData(r io.Reader, n int) (err error) {
	b := make([]byte, 1024)
	for n > 0 {
		k := n
		if n > 1024 {
			k = 1024
		}
		i, err := r.Read(b[:k])
		if err != nil {
			break
		}
		n -= i
	}
	return
}

type handshake struct {
	p   string
	r   []byte
	tor []byte
	id  []byte
}

func readHandshake(r io.Reader) (h *handshake, err error) {
	l, err := readUint8(r)
	if err != nil {
		return
	}
	b := make([]byte, l+48)
	err = readMsgData(r, b)
	if err != nil {
		return
	}
	h = new(handshake)
	h.p = string(b[:l])
	h.r = b[l : l+8]
	h.tor = b[l+8 : l+28]
	h.id = b[l+28 : l+48]
	return
}

func writeHandshake(w io.Writer, h *handshake) (err error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(byte(len(h.p)))
	buf.WriteString(h.p)
	buf.Write(h.r)
	buf.Write(h.tor)
	buf.Write(h.id)
	_, err = w.Write(buf.Bytes())
	return
}

type message struct {
	id    uint8
	index uint8
	data  []byte
}

func readMessage(r io.Reader) (msg *message, err error) {
	n, err := readUint32(r)
	if err != nil {
		return
	}
	if n > 0 {
		b := make([]byte, n)
		err = readMsgData(r, b)
		if err != nil {
			return
		}
		msg = new(message)
		msg.id = uint8(b[0])
		if n > 1 {
			msg.index = uint8(b[1])
		}
		if n > 2 {
			msg.data = b[2:]
		}
	}
	return
}

func writeMessage(w io.Writer, msg *message) (err error) {
	b4 := make([]byte, 4)
	n := uint32(2 + len(msg.data))
	binary.BigEndian.PutUint32(b4, n)
	buf := bytes.NewBuffer(nil)
	buf.Write(b4)
	buf.WriteByte(msg.id)
	buf.WriteByte(msg.index)
	buf.Write(msg.data)
	_, err = w.Write(buf.Bytes())
	return
}
