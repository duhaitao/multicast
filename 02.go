package main

import (
	// "encoding/binary"
	"fmt"
	"github.com/duhaitao/multicast/rmcast"
	"log"
	"math"
	"net"
	"time"
	"encoding/binary"
)

type ReceiverInfo struct {
	last_ack_seq uint32
	id           int
	last_seen    time.Time
	addr         net.UDPAddr
}

func main() {
	/// conn, err := net.Dial ("udp", "224.0.0.1:12345")
	/// conn, err := net.Dial("udp", "230.1.1.1:12345")
	var userid int = 1
	fmt.Println("dial to 127.0.0.1:12345")
	UdpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	if err != nil {
		log.Fatal("ResolveUDPAddr err: ", err)
	}
	conn, err := net.DialUDP("udp", nil, UdpAddr)
	if err != nil {
		log.Fatal("dial err: ", err)
	}
	defer conn.Close()

	pkgcache := rmcast.NewPkgCache(1024)
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
		//cwin_size uint32 = 100
		//cwin_left uint32 = 1
		//cwin_right uint32 = 2

		cwin_size uint32 = 100 // init slide window size
		cwin_left uint32 = 1
		//cwin_right uint32 = 11
	)

	receiver_id_map := make(map[int]*ReceiverInfo)     // key: id, value: last_ack_seq
	receiver_str_map := make(map[string]*ReceiverInfo) // key: UDPAddr.String (), value: last_ack_seq

	wchan := make(chan *rmcast.PKG, 1024)
	rchan := make(chan *rmcast.PKG, 1024)

	squeue := rmcast.NewSqueue()

	go func() {
		for {
			select {
			case pkg := <-wchan:
				squeue.Enque(pkg)
				seq := pkg.GetSeq()
				// first judge slide window
				if seq < cwin_left {
					log.Println("cwin_left: ", cwin_left, "first unack seq: ", seq)
					// timer move the cwin_left, and deque all acked pkg
					log.Fatal("impossiable")
				}

				if seq < cwin_left + cwin_size {
					// in slide window, can be send
					// further more, check cwin
					if seq < cwin_size + cwin_left {
						// seq locate in slide window and congestion window can be sent
						conn.Write(pkg.GetBuf())
					}
				}
			case pkg := <-rchan:
				pkg_type := pkg.GetType()
				switch pkg_type {
					case rmcast.TYPE_ACK: // receiver will send ack immediately when first rcv data
						receiver_info := receiver_str_map[pkg.Addr.String()]
						if receiver_info == nil {
							pReceiverInfo := &ReceiverInfo{pkg.GetSeq (), userid, time.Now(), pkg.Addr}
							receiver_str_map[pkg.Addr.String()] = pReceiverInfo
							receiver_id_map[userid] = pReceiverInfo
							userid++
						} else {
							receiver_info.last_seen = time.Now ()
							receiver_info.last_ack_seq = pkg.GetSeq ()
						}
						cwin_size++

					case rmcast.TYPE_NAK:
						receiver_info := receiver_str_map[pkg.Addr.String()]
						if receiver_info == nil {
							pReceiverInfo := &ReceiverInfo{pkg.GetSeq (), userid, time.Now(), pkg.Addr}
							receiver_str_map[pkg.Addr.String()] = pReceiverInfo
							receiver_id_map[userid] = pReceiverInfo
							userid++

							// it's possible, such as first ack is lost
							cwin_size /= 2
							if cwin_size < 100 {
								cwin_size = 100
							}
						} else {
							receiver_info.last_seen = time.Now ()
							receiver_info.last_ack_seq = pkg.GetSeq ()

							cwin_size /= 2
							if cwin_size < 100 {
								cwin_size = 100
							}
						}

						lost_pkg_count := pkg.GetLen ()
						//lost_slice := []rmcast.LostSeqInfo (pkg.GetVal ())
						lost_info_buf := pkg.GetVal ()
						var i uint32
						var j uint32
						for i = 0; i < lost_pkg_count; i++ {
							/// lost_seq := binary.BigEndian.Uint32 (lost_info_buf[i*4:i*4 + 4])
							lost_seq_count := binary.BigEndian.Uint32 (lost_info_buf[i*4 + 4: i*4 +8])

							for j = 0; j < lost_seq_count; j++ {
							}
						}
				}

			case <-time.After(time.Microsecond * 200):
				var min_ack_seq uint32 = math.MaxUint32
				// every 100 us, to adjust slide window
				delidslice := make([]int, 64, 1024)
				for id, receiver_info := range receiver_id_map {
					if time.Since(receiver_info.last_seen) >= time.Second*10 {
						// 10s, no ack or nak from receiver, we think the receiver is down
						// record id in delidslice, del them after range
						delidslice = append(delidslice, id)
					}

					if min_ack_seq > receiver_info.last_ack_seq {
						min_ack_seq = receiver_info.last_ack_seq
					}
				}

				for i := 0; i < len(delidslice); i++ {
					pReceiver := receiver_id_map[delidslice[i]]
					str := pReceiver.addr.String()
					delete(receiver_id_map, delidslice[i])
					delete(receiver_str_map, str)
				}

				if min_ack_seq != math.MaxUint32 && min_ack_seq > cwin_left {
					cwin_left = min_ack_seq

					// remove acked pkg
					for {
						first_pkg := squeue.First()
						if first_pkg.GetSeq() >= cwin_left {
							break
						}

						pkg := squeue.Deque()
						pkgcache.Put(pkg)
					}
				}
			} // end select
		}
	}()

	go func() {
		buf := make([]byte, 8192)
		for {
			nread, peer_addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal("ReadFromUDP err: ", err)
			}

			var rcv_pkg *rmcast.PKG
			rcv_pkg = pkgcache.Get()
			rcv_pkg.Reset()
			rcv_pkg.SetBuf(buf[:nread])
			rcv_pkg.Addr = *peer_addr
			rchan <- rcv_pkg
		}
	}()

	var seq uint32
	for {
		snd_pkg := pkgcache.Get()
		snd_pkg.Reset()
		seq++

		snd_pkg.SetType(rmcast.TYPE_DATA)
		snd_pkg.SetLen(10)
		snd_pkg.SetSeq(seq)
		snd_pkg.SetVal([]byte("0123456789"))

		wchan<- snd_pkg
/*
		//fmt.Println(string(buf))
		_, err = conn.Write(snd_pkg.GetBuf())
		fmt.Println("seq: ", snd_pkg.GetSeq(), "type: ", snd_pkg.GetType(),
			"val: ", string(snd_pkg.GetVal()))
*/
		//	time.Sleep (time.Second)
	}
	conn.Close()
}
