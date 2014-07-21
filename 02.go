package main

import (
	// "encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
	"math"
	"github.com/duhaitao/multicast/rmcast"
)

type ReceiverInfo struct {
	last_ack_seq uint32
	id           int
	last_seen    time.Time
	addr		 net.UDPAddr
}

func main() {
	/// conn, err := net.Dial ("udp", "224.0.0.1:12345")
	/// conn, err := net.Dial("udp", "230.1.1.1:12345")
	fmt.Println ("dial to 127.0.0.1:12345")
	UdpAddr, err := net.ResolveUDPAddr ("udp", "127.0.0.1:12345")
	if err != nil {
		log.Fatal ("ResolveUDPAddr err: ", err)
	}
	conn, err := net.DialUDP("udp", nil, UdpAddr)
	if err != nil {
		log.Fatal("dial err: ", err)
	}
	defer conn.Close ()

	pkgcache := rmcast.NewPkgCache (1024)
/*
	// recv go routine
	go func () {
		buf := make ([]byte, 4096)
		for {
			nread, peer_addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal("ReadFromUDP err")
			}

			// rcv ack or nak pkg

		}
	} ()
*/
var (
	cwin_size uint32 = 1 // init congestion windown is 1, like tcp
	//cwin_left uint32 = 1
	//cwin_right uint32 = 2

	swin_size uint32 = 10 // init slide window size 
	swin_left uint32 = 1
	//swin_right uint32 = 11
)

	receiver_map := make (map[int]*ReceiverInfo) // key: id, value: last_ack_seq

	wchan := make (chan *rmcast.PKG, 1024)
	rchan := make (chan *rmcast.PKG, 1024)

	squeue := rmcast.NewSqueue ()

	go func () {
		for {
			select {
				case pkg := <-wchan:
					squeue.Enque (pkg)
					seq := pkg.GetSeq ()
					// first judge slide window
					if seq < swin_left {
						log.Println ("swin_left: ", swin_left, "first unack seq: ", seq)
						// timer move the swin_left, and deque all acked pkg
						log.Fatal ("impossiable")
					}

					if seq < swin_left + swin_size {
						// in slide window, can be send
						// further more, check cwin
						if seq < cwin_size + swin_left {
							// seq locate in slide window and congestion window can be sent
							conn.Write (pkg.GetBuf ())
						}
					}
				case pkg := <-rchan:
					pkg_type := pkg.GetType ()
					switch pkg_type {
					case rmcast.TYPE_PROTOCOL:
						proto_type := pkg.GetProtoType ()

					}

				case <-time.After (time.Microsecond * 200):
					var min_ack_seq uint32 = math.MaxUint32
					// every 100 us, to adjust slide window 
					for id, receiver_info := range receiver_map {
						if time.Since (receiver_info.last_seen) >= time.Second * 10 {
							// 10s, no ack or nak from receiver, we think the receiver is down
							delete (receiver_map, id)
						}

						if min_ack_seq > receiver_info.last_ack_seq {
							min_ack_seq = receiver_info.last_ack_seq
						}
					}

					if min_ack_seq != math.MaxUint32 &&
						min_ack_seq > swin_left {
						swin_left = min_ack_seq

						// remove acked pkg
						for {
							first_pkg := squeue.First ()
							if first_pkg.GetSeq () >= swin_left {
								break
							}

							pkg := squeue.Deque ()
							pkgcache.Put (pkg)
						}
					}
			} // end select
		}
	} ()

	go func () {
		buf := make([]byte, 8192)
		for {
			nread, peer_addr, err := conn.ReadFromUDP (buf)
			if err != nil {
				log.Fatal("ReadFromUDP err: ", err)
			}

			var rcv_pkg *rmcast.PKG
			rcv_pkg = pkgcache.Get ()
			rcv_pkg.Reset ()
			rcv_pkg.SetBuf (buf[:nread])
			rcv_pkg.Addr = *peer_addr
			rchan<- rcv_pkg
		}
	} ()

	var seq uint32
	snd_pkg := pkgcache.Get ()
	for {
		seq++

		snd_pkg.SetType (rmcast.TYPE_DATA)
		snd_pkg.SetLen (10)
		snd_pkg.SetSeq (seq)
		snd_pkg.SetVal ([]byte("0123456789"))

		//fmt.Println(string(buf))
		_, err = conn.Write(snd_pkg.GetBuf ())
		fmt.Println ("seq: ", snd_pkg.GetSeq (), "type: ", snd_pkg.GetType (),
			"val: ", string (snd_pkg.GetVal ()))

		snd_pkg.Reset ()
	//	time.Sleep (time.Second)
	}
	conn.Close()
}
