package rmcast

import (
	//"fmt"
)

/*   +------------------------------+
 *   | type | id  | len | seq | val |
 *   +------------------------------+
 *     2B     4B    4B    4B 
 *
 *   type:
 *     1 -- data pkg
 *     2 -- ack
 *     3 -- nak
 *     4 -- protocl control 
 *
 *	 id:
 *    every login user will allocate a id, then communication with this id
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
 *
 *	 if type == protocl
 *	  val format:
 *   +--------------------------------------
 *   | protocol type |
 *   +--------------------------------------
 *         2B
 *   protocol type:
 *     1 - login
 *     2 - logout
 *     3 - forcelogout
 *
 */
import (
	"encoding/binary"
	"net"
	"log"
)

const (
	TYPE_DATA = iota + 1
	TYPE_ACK
	TYPE_NAK
	TYPE_PROTOCOL
)

const (
	TYPE_PROTO_LOGIN = iota + 1
	TYPE_PROTO_LOGOUT
	TYPE_PROTO_FORCELOGOUT
)

const (
	PKG_HEADER_LEN = 14
	TYPE_OFFSET = 0
	ID_OFFSET = 2
	LEN_OFFSET = 6
	SEQ_OFFSET = 10
	VAL_OFFSET = 14
	PROTO_VAL_OFFSET = 16
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
	return binary.BigEndian.Uint16 (pkg.buf[:ID_OFFSET])
}

func (pkg *PKG) SetType (t uint16) {
	binary.BigEndian.PutUint16 (pkg.buf[:ID_OFFSET], t)
}

func (pkg *PKG) GetProtoType () uint16 {
	return binary.BigEndian.Uint16 (pkg.buf[VAL_OFFSET:PROTO_VAL_OFFSET])
}

func (pkg *PKG) SetProtoType (t uint16) {
	binary.BigEndian.PutUint16 (pkg.buf[VAL_OFFSET:PROTO_VAL_OFFSET], t)
}

func (pkg *PKG) GetSeq () uint32 {
	return binary.BigEndian.Uint32 (pkg.buf[SEQ_OFFSET:VAL_OFFSET])
}

func (pkg *PKG) SetSeq (seq uint32) {
	binary.BigEndian.PutUint32 (pkg.buf[SEQ_OFFSET:VAL_OFFSET], seq)
}

func (pkg *PKG) GetLen () uint32 {
	return binary.BigEndian.Uint32 (pkg.buf[LEN_OFFSET:SEQ_OFFSET])
}

func (pkg *PKG) SetLen (l int) {
	binary.BigEndian.PutUint32 (pkg.buf[LEN_OFFSET:SEQ_OFFSET], uint32 (l))
}

func (pkg *PKG) SetBufLen (l int) {
	pkg.buflen = l
	pkg.buf = pkg.buf[:pkg.buflen]
}

func (pkg *PKG) GetVal () []byte {
	return pkg.buf[VAL_OFFSET:]
}

func (pkg *PKG) SetVal (val []byte) {
	copy (pkg.buf[VAL_OFFSET:], val)
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
	return pkg.buf[:]
}

func (pkg *PKG) GetBufLen () int {
	return pkg.buflen
}

func (pkg *PKG) GetLastAck () uint32 {
	return binary.BigEndian.Uint32 (pkg.buf[PROTO_VAL_OFFSET:PROTO_VAL_OFFSET + 4])
}

func (pkg *PKG) Reset () {
	pkg.buf = pkg.buf[0:VAL_OFFSET]
	pkg.buflen = PKG_HEADER_LEN
}

type ProtoPKG struct {
}
/*
func (pkg *PKG) String () string {
	buf := make ([]byte, 1024)
	
	return string (buf)
}
*/
