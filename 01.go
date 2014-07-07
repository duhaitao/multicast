package main

import (
	"encoding/binary"
	 "fmt"
	"log"
	"net"
	"container/list"
	"time"
)

const (
	TYPE_DATA = 1
	TYPE_ACK  = 2
	TYPE_NAK  = 3
)

type pkg struct {
	content [4096]byte
}

func HandlePackage (npkg *pkg) {
	fmt.Println ("rcv seq: ", binary.BigEndian.Uint32 (npkg.content[6:10]))
}

func main() {
	ifi, err := net.InterfaceByName("eth0")
	if err != nil {
		log.Fatal("eth0 err")
	}

	// make pkg cache
	var newpkg *pkg
	pkglist := list.New ()
	for i := 0; i < 10000; i++ {
		newpkg = new (pkg)
		pkglist.PushBack (newpkg)
	}

	// receiver use lost_pkg_list to hold the unordered pkg
	lost_pkg_list := list.New ()

	multicastip := net.ParseIP("230.1.1.1")
	pUDPAddr := &net.UDPAddr{IP: multicastip, Port: 12345}
	// fmt.Println (*pUDPAddr)
	conn, err := net.ListenMulticastUDP("udp4", ifi, pUDPAddr)
	if err != nil {
		log.Fatal("net.ListenMulticastUDP err")
	}
	defer conn.Close ()

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

	now := time.Now ()
	for {
		var last_seq uint32 // last continous seq and first unordered seq
		var nak_pkg, ack_pkg *pkg
		for rcv_pkg := range rchan {
			buf := rcv_pkg.content
			length := binary.BigEndian.Uint32(buf[2:6])
			seq := binary.BigEndian.Uint32(buf[6:10])

			/* recv first pkg, send ack immediately */
			if seq == 1 || time.Since (now) > 1000000 { // greater then 1ms, send ack
				if pkglist.Len () > 0 {
					elem := pkglist.Front ()
					ack_pkg = elem.Value.(*pkg)
					pkglist.Remove (elem)
			//		fmt.Println ("after pop pkglist len: ", pkglist.Len ())
				} else {
					ack_pkg = new (pkg)
				}

				binary.BigEndian.PutUint16 (ack_pkg.content[:2], TYPE_ACK)
				binary.BigEndian.PutUint32 (ack_pkg.content[2:6], 0)
				// ack stand for next seq I will receive, like tcp
				binary.BigEndian.PutUint32 (ack_pkg.content[6:10], seq + 1)

				now = time.Now ()
				wchan<- ack_pkg
			}

			// duplicate pkg, ignore it
			if seq <= last_seq {
				pkglist.PushBack (rcv_pkg)
				continue
			}

			if seq % 10000 == 0 {
				log.Println ("rcv goroutine seq: ", seq, length)
			}

			if seq == last_seq + 1 { // fast path
				/// first_unordered_pkg := lost_pkg_list.Front ().Value.(*pkg)
				/// first_unordered_pkg_seq := binary.BigEndian.Uint32 (first_unordered_pkg.content[6:10])
				HandlePackage (rcv_pkg)
				pkglist.PushBack (rcv_pkg)
				last_seq = seq

				// one hole is filled, will trig many or zero pkg to handle
				for e := lost_pkg_list.Front(); e != nil; e = e.Next() {
					if last_seq + 1 == binary.BigEndian.Uint32 (e.Value.(*pkg).content[6:10]) {
						lost_pkg_list.Remove (e)

						HandlePackage (e.Value.(*pkg))
						last_seq += 1
						pkglist.PushBack (e.Value.(*pkg))
					} else {
						// first seq != last_seq + 1, then break, no continue pkg
						break
					}
				}

				continue
			}

			var insertok bool
			// iterate lost_pkg_list to insert lost pkg
			for e := lost_pkg_list.Front(); e != nil; e = e.Next() {
				pkt_seq := binary.BigEndian.Uint32 (e.Value.(*pkg).content[6:10])
				if seq == pkt_seq {// duplicate pkt, ignore
					insertok = true
					break
				}

				if seq < pkt_seq {
					insertok = true 
					pkglist.InsertBefore (rcv_pkg, e)
					break
				}
			}
			// if rcv_pkt not blong to lost_pkg_list, then append it
			if insertok == false {
				pkglist.PushBack (rcv_pkg)
			}

			if last_seq != 0 && last_seq + 1 < seq {
				// pkg lost, send nak
				if pkglist.Len () > 0 {
					elem := pkglist.Front ()
					nak_pkg = elem.Value.(*pkg)
					pkglist.Remove (elem)
			//		fmt.Println ("after pop pkglist len: ", pkglist.Len ())
				} else {
					nak_pkg = new (pkg)
				}

				// enque this unordered pkg to lost_pkg_list
				lost_pkg_list.PushBack (rcv_pkg)

				binary.BigEndian.PutUint16 (nak_pkg.content[0:], TYPE_NAK)
				binary.BigEndian.PutUint32 (nak_pkg.content[0:], last_seq)
				fmt.Println ("send nak pkg: ", last_seq)
				wchan<- nak_pkg

				// unordered pkg push to lost_pkg_list, and recv next
				continue
			}

			last_seq = seq
			/// fmt.Println (length, seq)
			pkglist.PushBack (rcv_pkg)
	//		fmt.Println ("after pushback pkglist len: ", pkglist.Len ())
		}
	}
}
