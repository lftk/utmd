package utmd

import (
	"bytes"
	"fmt"

	"github.com/4396/dht"
)

var protocol = "BitTorrent protocol"
var idformat = "-%s%04d-123456789000"
var reserved = []byte{0, 0, 0, 0, 0x10, 0, 0, 0}

func encodeHandshake(name string, ver int, tor *dht.ID) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(byte(19))
	buf.WriteString(protocol)
	buf.Write(reserved)
	buf.Write(tor.Bytes())
	buf.WriteString(fmt.Sprintf(idformat, name, ver))
	return buf.Bytes()
}
