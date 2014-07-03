package main

import (
	"net"
	"time"
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
	fmt.Println (*pUDPAddr)
	_, err = net.ListenMulticastUDP ("udp4", ifi, pUDPAddr)
	if err != nil {
		log.Fatal ("net.ListenMulticastUDP err")
	}

	for {
		time.Sleep (1)
	}
}
