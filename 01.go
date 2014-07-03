package main

import (
	"net"
	"log"
	"fmt"
)

func main () {
	ifi, err := net.InterfaceByName ("wlan0")
	if err != nil {
		log.Fatal ("wlan0 err")
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
		_, _, err := conn.ReadFromUDP (buf)
		if err != nil {
			log.Fatal ("ReadFromUDP err")
		}
		fmt.Println ("type: ", buf[0])
		fmt.Println ("len: ", buf[1])
		fmt.Println (buf)
	}
}
