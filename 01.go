package main

import (
	"encoding/binary"
	//"fmt"
	"log"
	"net"
)

func main() {
	ifi, err := net.InterfaceByName("eth0")
	if err != nil {
		log.Fatal("eth0 err")
	}

	multicastip := net.ParseIP("230.1.1.1")
	pUDPAddr := &net.UDPAddr{IP: multicastip, Port: 12345}
	// fmt.Println (*pUDPAddr)
	conn, err := net.ListenMulticastUDP("udp4", ifi, pUDPAddr)
	if err != nil {
		log.Fatal("net.ListenMulticastUDP err")
	}

	buf := make([]byte, 4096)
	for {
		_, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal("ReadFromUDP err")
		}

		//fmt.Println("type: ", binary.BigEndian.Uint16(buf[:2]))
		//fmt.Println("len: ", binary.BigEndian.Uint32(buf[2:6]))
		//fmt.Println("seq: ", binary.BigEndian.Uint32(buf[6:10]))
		//fmt.Println(string(buf[10:length]))
		seq := binary.BigEndian.Uint32(buf[6:10])
		if seq % 10000 == 0 {
			log.Fatal (seq)
		}
	}
}
