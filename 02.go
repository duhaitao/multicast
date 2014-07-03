package main

import (
	"net"
	"log"
	"fmt"
	"bytes"
)

func main () {
	conn, err := net.Dial ("udp", "127.0.0.1:12345")
	if err != nil {
		log.Fatal ("dial err")
	}

	buf := make ([]byte, 1024)
	buffer := bytes.NewBuffer (buf)

	/*   +------------------------+
	 *   | type | len | val       |
	 *   +------------------------+
	 *     1B     1B  
	 */
	var pkttype byte = 1
	buffer.WriteByte (pkttype)
	var length byte = 10
	buffer.WriteByte (length)
	_, ret := buffer.WriteString ("0123456789")
	if ret != nil {
		log.Fatal ("WriteString err")
	}

	fmt.Println (string(buffer.Bytes ()))
	b := buffer.Bytes ()
	_, err = conn.Write (b[:12])
	conn.Close ()
}
