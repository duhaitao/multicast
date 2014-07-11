package main

import (
	 "fmt"
	"log"
	"net"
	"container/list"
	"time"
	"github.com/duhaitao/multicast/rmcast"
	"encoding/binary"
)

func HandlePackage (npkg *rmcast.PKG) {
	fmt.Println ("rcv seq: ", npkg.GetSeq ())
}

// input go routine, which just read data from socket
// and copy to pkg, then send pkg to rchan
func(conn *UDPConn, pkgcache *PkgCache, rchan chan *rmcast.PKG) {
	for {
		nread, peer_addr, err = conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal("ReadFromUDP err")
		}

		var rcv_pkg *rmcast.PKG
		if pkglist.Len () > 0 {
			elem := pkglist.Front ()
			rcv_pkg = elem.Value.(*rmcast.PKG)
			pkglist.Remove (elem)
			// fmt.Println ("after pop pkglist len: ", pkglist.Len ())
		} else {
			/// rcv_pkg = new (rmcast.PKG)
			rcv_pkg = rmcast.NewPKG ()
		}
		rcv_pkg := pkgcache.Get ()
		rcv_pkg.SetBuf (buf[:nread])
		rchan<- rcv_pkg
	}
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
		newpkg = rmcast.NewPKG ()
		pkglist.PushBack (newpkg)
	}

	// receiver use rqueue to hold the unordered pkg
	/// lost_pkg_list := list.New ()
	rqueue := rmcast.NewRqueue ()

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

	// write to peer, ack and nak
	go func () {
		for rcv_pkg := range wchan {
			conn.WriteToUDP (rcv_pkg.GetBuf (), peer_addr)
			pkglist.PushBack (rcv_pkg)
		}
	} ()

	lost_seq_array := make ([]uint32, 0, 256)
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
					/// ack_pkg = new (rmcast.PKG)
					ack_pkg = rmcast.NewPKG ()
				}

				ack_pkg.SetType (rmcast.TYPE_ACK)
				ack_pkg.SetLen (0)
				// ack stand for next seq I will receive, like tcp
				ack_pkg.SetSeq (seq + 1)

				fmt.Println ("send ack")
				now = time.Now ()
				wchan<- ack_pkg
			}

			// duplicate pkg, ignore it
			fmt.Println ("seq : last_seq", seq, last_seq)
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
				// pkg lost, send nak immediately
				if pkglist.Len () > 0 {
					elem := pkglist.Front ()
					nak_pkg = elem.Value.(*rmcast.PKG)
					pkglist.Remove (elem)
			//		fmt.Println ("after pop pkglist len: ", pkglist.Len ())
				} else {
					/// nak_pkg = new (rmcast.PKG)
					nak_pkg = rmcast.NewPKG ()
				}

				// enque this unordered pkg to lost_pkg_list
				lost_pkg_list.PushBack (rcv_pkg)

				nak_pkg.SetType (rmcast.TYPE_NAK)
				nak_pkg.SetSeq (last_seq)

				var lost_seq_count uint32
				lost_seq_array = append (lost_seq_array, last_seq + 1)
				if lost_pkg_list.Front() != nil {
					next_pkg := lost_pkg_list.Front().Value.(*rmcast.PKG)
					lost_seq_array = append (lost_seq_array, next_pkg.GetSeq () - last_seq - 1)
					lost_seq_count++
				} else {
					log.Fatal ("lost_pkg_list must not empty")
				}
				// nak should carray hole info
				for e := lost_pkg_list.Front(); e != nil; e = e.Next() {
					unordered_pkg := e.Value.(*rmcast.PKG)
					if e.Next () != nil {
						next_unordered_pkg := e.Next ().Value.(*rmcast.PKG)
						if unordered_pkg.GetSeq () + 1 != next_unordered_pkg.GetSeq () {
							// found a hole
							lost_seq_array = append (lost_seq_array, unordered_pkg.GetSeq () + 1)
							lost_seq_array = append (lost_seq_array,
								next_unordered_pkg.GetSeq () - 1 - unordered_pkg.GetSeq ())

							fmt.Println (unordered_pkg.GetSeq () + 1, next_unordered_pkg.GetSeq () - 1, lost_seq_count)
							lost_seq_count++
							if lost_seq_count == 255 {
								break
							}
						}
					}
				}

				// lost seq hole in lost_seq_array
				// nak_pkg.SetVal ([]byte (lost_seq_array[:lost_seq_count]))
				val_buf := nak_pkg.GetBuf ()
				var i uint32
				for i = 0; i < lost_seq_count; i++ {
					var tmp [8]byte
					lost_seq := lost_seq_array[2 * i]
					binary.BigEndian.PutUint32 (tmp[:4], lost_seq)
					lost_len := lost_seq_array[2 * i + 1]
					binary.BigEndian.PutUint32 (tmp[4:], lost_len)

					copy (val_buf[10 + 4 * i:10 + 4 * (i + 1)], tmp[:])

					/// fmt.Println ("lost seq, len lost_seq_count", lost_seq, lost_len, lost_seq_count)
					/// binary.BigEndian.PutUint32 (val_buf[4 * i:4 * (i + 1)], lost_seq_array[i])
				}

				binary.BigEndian.PutUint32 (val_buf[4:8], lost_seq_count * 8)

				fmt.Println ("send nak pkg: ", last_seq)
				wchan<- nak_pkg

				// unordered pkg push to lost_pkg_list, and recv next
				continue
			}

			last_seq = seq
			/// fmt.Println (length, seq)
			pkglist.PushBack (rcv_pkg)
		}
	}
}
