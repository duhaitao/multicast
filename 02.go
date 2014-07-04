package main

import (
	"encoding/binary"
	//"fmt"
	"log"
	"net"
)

func main() {
	/// conn, err := net.Dial ("udp", "224.0.0.1:12345")
	/// conn, err := net.Dial("udp", "230.1.1.1:12345")
	conn, err := net.Dial("udp", "127.0.0.1:12345")
	if err != nil {
		log.Fatal("dial err")
	}

	buf := make([]byte, 1024)

	/*   +------------------------+
	 *   | type | len | seq | val |
	 *   +------------------------+
	 *     2B     4B    4B
	 */
	var seq uint32
	for {
		var pkttype uint16 = 2
		binary.BigEndian.PutUint16(buf[:2], pkttype)
		var length uint32 = 10
		binary.BigEndian.PutUint32(buf[2:6], length)

		copy(buf[6:], []byte("0123456789"))
		seq++

		if seq % 100000 == 0 {
			seq++
		}
		binary.BigEndian.PutUint32(buf[6:10], seq)

		//fmt.Println(string(buf))
		_, err = conn.Write(buf[:20])
	}
	conn.Close()
}
