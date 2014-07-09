package rmcast

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
)

const (
	TYPE_DATA = iota
	TYPE_ACK
	TYPE_NAK
)

type PKG struct {
	content [4096]byte
}

func (pkg *PKG) GetType () uint16 {
	return binary.BigEndian.Uint16 (pkg.content[:4])
}

func (pkg *PKG) SetType (t uint16) {
	binary.BigEndian.PutUint16 (pkg.content[:4], t)
}

func (pkg *PKG) GetSeq () uint32 {
	return binary.BigEndian.Uint32 (pkg.content[6:10])
}

func (pkg *PKG) SetSeq (seq uint32) {
	binary.BigEndian.PutUint32 (pkg.content[6:10], seq)
}

func (pkg *PKG) GetLen () uint32 {
	return binary.BigEndian.Uint32 (pkg.content[2:6])
}

func (pkg *PKG) SetLen (l uint32) {
	binary.BigEndian.PutUint32 (pkg.content[2:6], l)
}

func (pkg *PKG) GetVal () []byte {
	return pkg.content[10:]
}

func (pkg *PKG) SetVal (val []byte) {
	copy (pkg.content[10:], val)
}
