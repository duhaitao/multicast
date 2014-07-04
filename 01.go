package main

import (
	"net"
	"log"
	"fmt"
	"encoding/binary"
)

func main () {
	ifi, err := net.InterfaceByName ("eth0")
	if err != nil {
		log.Fatal ("eth0 err")
	}

	multicastip := net.ParseIP ("230.1.1.1")
	pUDPAddr := &net.UDPAddr {IP: multicastip, Port: 12345}
	// fmt.Println (*pUDPAddr)
	conn, err := net.ListenMulticastUDP ("udp4", ifi, pUDPAddr)
	if err != nil {
		log.Fatal ("net.ListenMulticastUDP err")
	}

	buf := make ([]byte, 4096)
	for {
		length, _, err := conn.ReadFromUDP (buf)
		if err != nil {
			log.Fatal ("ReadFromUDP err")
		}

		fmt.Println ("type: ", binary.BigEndian.Uint16 (buf[:2]))
		fmt.Println ("len: ", binary.BigEndian.Uint32 (buf[2:6]))
		fmt.Println (string(buf[6:length]))
	}
}
