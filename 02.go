package main

import (
	"net"
	"log"
	"fmt"
	"encoding/binary"
)

func main () {
	conn, err := net.Dial ("udp", "127.0.0.1:12345")
	if err != nil {
		log.Fatal ("dial err")
	}

	buf := make ([]byte, 1024)

	/*   +------------------------+
	 *   | type | len | val       |
	 *   +------------------------+
	 *     2B     4B  
	 */
	var pkttype uint16 = 2
	binary.BigEndian.PutUint16 (buf[:2], pkttype)
	var length uint32 = 10
	binary.BigEndian.PutUint32 (buf[2:6], length)
	copy (buf[6:], []byte("0123456789"))

	fmt.Println (string(buf))
	_, err = conn.Write (buf[:16])
	conn.Close ()
}
