package main

import (
	"encoding/binary"
	// "fmt"
	"log"
	"net"
	"container/list"
)

type pkg struct {
	content [4096]byte
}

func main() {
	ifi, err := net.InterfaceByName("eth0")
	if err != nil {
		log.Fatal("eth0 err")
	}

	var newpkg *pkg
	pkglist := list.New ()
	for i := 0; i < 10000; i++ {
		newpkg = new (pkg) 
		pkglist.PushBack (newpkg)
	}

	multicastip := net.ParseIP("230.1.1.1")
	pUDPAddr := &net.UDPAddr{IP: multicastip, Port: 12345}
	// fmt.Println (*pUDPAddr)
	conn, err := net.ListenMulticastUDP("udp4", ifi, pUDPAddr)
	if err != nil {
		log.Fatal("net.ListenMulticastUDP err")
	}

	buf := make([]byte, 4096)
	rchan := make (chan *pkg, 4096)
	wchan := make (chan *pkg, 4096)
	var peer_addr *net.UDPAddr
	var nread int
	go func() {
		for {
			nread, peer_addr, err = conn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal("ReadFromUDP err")
			}

			//fmt.Println("type: ", binary.BigEndian.Uint16(buf[:2]))
			//fmt.Println("len: ", binary.BigEndian.Uint32(buf[2:6]))
			//fmt.Println("seq: ", binary.BigEndian.Uint32(buf[6:10]))
			//fmt.Println(string(buf[10:length]))
			var rcv_pkg *pkg
			if pkglist.Len () > 0 {
				elem := pkglist.Front ()
				rcv_pkg = elem.Value.(*pkg)
				pkglist.Remove (elem)
		//		fmt.Println ("after pop pkglist len: ", pkglist.Len ())
			} else {
				rcv_pkg = new (pkg)
			}
			copy (rcv_pkg.content[0:], buf[:nread])
/*
			seq := binary.BigEndian.Uint32(buf[6:10])
			if seq%10000 == 0 {
				log.Fatal(seq)
			}
*/
			rchan<- rcv_pkg
		}
	} ()

	go func () {
		for rcv_pkg := range wchan {
			conn.WriteToUDP (rcv_pkg.content[:4], peer_addr)
			pkglist.PushBack (rcv_pkg)
		}
	} ()

	for {
		var last_seq uint32
		var nak_pkg *pkg
		for rcv_pkg := range rchan {
			buf := rcv_pkg.content
			length := binary.BigEndian.Uint32(buf[2:6])
			seq := binary.BigEndian.Uint32(buf[6:10])

			if seq % 10000 == 0 {
				log.Println ("rcv goroutine seq: ", seq, length)
			}

			if last_seq != 0 && last_seq != seq - 1 {
				// pkg lost, send nak
				if pkglist.Len () > 0 {
					elem := pkglist.Front ()
					nak_pkg = elem.Value.(*pkg)
					pkglist.Remove (elem)
			//		fmt.Println ("after pop pkglist len: ", pkglist.Len ())
				} else {
					nak_pkg = new (pkg)
				}

				binary.BigEndian.PutUint32 (nak_pkg.content[0:], last_seq)
				wchan<- nak_pkg
			}

			/// fmt.Println (length, seq)
			pkglist.PushBack (rcv_pkg)
	//		fmt.Println ("after pushback pkglist len: ", pkglist.Len ())
		}
	}
}
