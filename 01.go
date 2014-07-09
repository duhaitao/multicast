package main

import (
	 "fmt"
	"log"
	"net"
	"container/list"
	"time"
	"github.com/duhaitao/multicast/rmcast"
)

func HandlePackage (npkg *rmcast.PKG) {
	fmt.Println ("rcv seq: ", npkg.GetSeq ())
}

func main() {
	ifi, err := net.InterfaceByName("eth0")
	if err != nil {
		log.Fatal("eth0 err")
	}

	// make pkg cache
	var newpkg *rmcast.PKG
	pkglist := list.New ()
	for i := 0; i < 10000; i++ {
		newpkg = new (rmcast.PKG)
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
	rchan := make (chan *rmcast.PKG, 4096)
	wchan := make (chan *rmcast.PKG, 4096)
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
			var rcv_pkg *rmcast.PKG
			if pkglist.Len () > 0 {
				elem := pkglist.Front ()
				rcv_pkg = elem.Value.(*rmcast.PKG)
				pkglist.Remove (elem)
		//		fmt.Println ("after pop pkglist len: ", pkglist.Len ())
			} else {
				rcv_pkg = new (rmcast.PKG)
			}
			rcv_pkg.SetBuf (buf[:nread])
/*
			seq := binary.BigEndian.Uint32(buf[6:10])
			if seq%10000 == 0 {
				log.Fatal(seq)
			}
*/
			rchan<- rcv_pkg
		}
	} ()

/*
	go func () {
		for rcv_pkg := range wchan {
			conn.WriteToUDP (rcv_pkg.content[:4], peer_addr)
			pkglist.PushBack (rcv_pkg)
		}
	} ()
*/
	now := time.Now ()
	for {
		var last_seq uint32 // last continous seq and first unordered seq
		var nak_pkg, ack_pkg *rmcast.PKG
		for rcv_pkg := range rchan {
			length := rcv_pkg.GetLen ()
			seq := rcv_pkg.GetSeq ()

			/* recv first pkg, send ack immediately */
			if seq == 1 || time.Since (now) > 1000000 { // greater then 1ms, send ack
				if pkglist.Len () > 0 {
					elem := pkglist.Front ()
					ack_pkg = elem.Value.(*rmcast.PKG)
					pkglist.Remove (elem)
			//		fmt.Println ("after pop pkglist len: ", pkglist.Len ())
				} else {
					ack_pkg = new (rmcast.PKG)
				}

				ack_pkg.SetType (rmcast.TYPE_ACK)
				ack_pkg.SetLen (0)
				// ack stand for next seq I will receive, like tcp
				ack_pkg.SetSeq (seq + 1)

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
					unordered_pkg := e.Value.(*rmcast.PKG)
					if last_seq + 1 == unordered_pkg.GetSeq () {
						lost_pkg_list.Remove (e)

						HandlePackage (unordered_pkg)
						last_seq += 1
						pkglist.PushBack (unordered_pkg)
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
				unordered_pkg := e.Value.(*rmcast.PKG)
				pkt_seq := unordered_pkg.GetSeq ()
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
					nak_pkg = elem.Value.(*rmcast.PKG)
					pkglist.Remove (elem)
			//		fmt.Println ("after pop pkglist len: ", pkglist.Len ())
				} else {
					nak_pkg = new (rmcast.PKG)
				}

				// enque this unordered pkg to lost_pkg_list
				lost_pkg_list.PushBack (rcv_pkg)

				nak_pkg.SetType (rmcast.TYPE_NAK)
				nak_pkg.SetSeq (last_seq)
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
