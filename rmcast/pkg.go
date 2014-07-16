package rmcast

import (
	//"fmt"
)

/*   +------------------------+
 *   | type | len | seq | val |
 *   +------------------------+
 *     2B     4B    4B
 *
 *   type:
 *     1 -- data pkg
 *     2 -- ack
 *     3 -- nak
 *
 *   len:
 *     val len
 *
 *   seq:
 *     send sequence, increase monotonicly by 1
 *
 *   if type == ack:
 *     no val, receiver must ack first pkg, then ack per millsecond normally, sender
 *     can assure reciever is alive by ack
 *   
 *   if type == nak:
 *     val format:
 *   +-----------------------------+
 *   | seq | len | ... | seq | len |
 *   +-----------------------------+
 *     4B    4B          4B     4B
 *
 *     seq: 
 * 		first seq of lost range, len is length of range. every receiver hole occupy a
 *      <seq, len> pair
 */
import (
	"encoding/binary"
	"net"
	"log"
)

const (
	TYPE_DATA = iota
	TYPE_ACK
	TYPE_NAK
)

const (
	PKG_HEADER_LEN = 10
)

type PKG struct {
	buf []byte
	buflen int
	Addr net.UDPAddr
}

func NewPKG () *PKG {
	return &PKG{make ([]byte, 4096), PKG_HEADER_LEN, net.UDPAddr{}}
}

func (pkg *PKG) GetType () uint16 {
	return binary.BigEndian.Uint16 (pkg.buf[:2])
}

func (pkg *PKG) SetType (t uint16) {
	binary.BigEndian.PutUint16 (pkg.buf[:2], t)
}

func (pkg *PKG) GetSeq () uint32 {
	return binary.BigEndian.Uint32 (pkg.buf[6:10])
}

func (pkg *PKG) SetSeq (seq uint32) {
	binary.BigEndian.PutUint32 (pkg.buf[6:10], seq)
}

func (pkg *PKG) GetLen () uint32 {
	return binary.BigEndian.Uint32 (pkg.buf[2:6])
}

func (pkg *PKG) SetLen (l int) {
	binary.BigEndian.PutUint32 (pkg.buf[2:6], uint32 (l))
}

func (pkg *PKG) SetBufLen (l int) {
	pkg.buflen = l
	pkg.buf = pkg.buf[:pkg.buflen]
}

func (pkg *PKG) GetVal () []byte {
	return pkg.buf[10:]
}

func (pkg *PKG) SetVal (val []byte) {
	copy (pkg.buf[10:], val)
	pkg.buflen += len (val)
	if pkg.buflen >= 4096 {
		log.Fatal ("SetVal pkg.buf overflow")
	}
	pkg.buf = pkg.buf[:pkg.buflen]
}

// overwrite all buf
func (pkg *PKG) SetBuf (val []byte) {
	copy (pkg.buf[:], val)
	pkg.buflen = len (val)
	if pkg.buflen >= 4096 {
		log.Fatal ("SetBuf pkg.buf overflow")
	}
	pkg.buf = pkg.buf[:pkg.buflen]
}

func (pkg *PKG) GetBuf () []byte {
	/// fmt.Println ("buf size: ", 10 + binary.BigEndian.Uint32 (pkg.buf[2:6]))
	/// return pkg.buf[:10 + binary.BigEndian.Uint32 (pkg.buf[2:6])]
	return pkg.buf[:]
}

func (pkg *PKG) GetBufLen () int {
	return pkg.buflen
}

func (pkg *PKG) Reset () {
	pkg.buf = pkg.buf[0:10]
	pkg.buflen = 10
}
/*
func (pkg *PKG) String () string {
	buf := make ([]byte, 1024)
	
	return string (buf)
}
*/
