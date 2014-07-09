package main

import (
	"encoding/binary"
	//"fmt"
	"log"
	"net"
	"time"
)

func main() {
	/// conn, err := net.Dial ("udp", "224.0.0.1:12345")
	/// conn, err := net.Dial("udp", "230.1.1.1:12345")
	UdpAddr, err := net.ResolveUDPAddr ("udp", "127.0.0.1:12345")
	if err != nil {
		log.Fatal ("ResolveUDPAddr err")
	}
	conn, err := net.DialUDP("udp", nil, UdpAddr)
	if err != nil {
		log.Fatal("dial err")
	}
	defer conn.Close ()

	// recv go routine
	go func () {
		buf := make ([]byte, 4096)
		for {
			nread, peer_addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal("ReadFromUDP err")
			}

			// rcv nak pkg

		}
	} ()

	buf := make([]byte, 1024)
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
	cwin_siz := 1 // init congestion windown is 1, like tcp
	swin_siz := 10 // init slide window size 
	var seq uint32
	tbegin := time.Now ()
	for {
		var pkttype uint16 = 1
		binary.BigEndian.PutUint16(buf[:2], pkttype)
		var length uint32 = 10
		binary.BigEndian.PutUint32(buf[2:6], length)

		seq++

		binary.BigEndian.PutUint32(buf[6:10], seq)
		copy(buf[10:], []byte("0123456789"))

		//fmt.Println(string(buf))
		_, err = conn.Write(buf[:20])

		if time.Since (tbegin) > 1000000 { // great 1ms
			tbegin = time.Now ()
			// overtime, send ack

		}
	}
	conn.Close()
}
