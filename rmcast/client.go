package rmcast

import (
	"fmt"
	"log"
	"net"
	// "container/list"
	"time"
	// "github.com/duhaitao/multicast/rmcast"
	"encoding/binary"
)

type Client struct {
	rqueue *Rqueue
	last_rcv_seq uint32
	conn *net.UDPConn
	pkgcache *PkgCache
	rchan chan *PKG
	wchan chan *PKG

	// large quntity nak with same lost info, will incur server 
	// to send same data many times, so I will control nak sending
	// speed from source
	nak_snd_time time.Time
	ack_snd_time time.Time
	// flag, to inform main routine to handle Squeue
	handleSqueue bool
	// call back for upper application
	HandlePackage func (pkg *PKG)
}

func (client *Client) RegisterHandlePackage (fn func (pkg *PKG)) {
	client.HandlePackage = fn
}

func (client *Client) BindMAddr (maddr *MAddr) (*net.UDPConn, error) {
	fmt.Println ("BindMAddr")
	conn, err := net.ListenMulticastUDP ("udp4", maddr.Iface, maddr.Addr)
	if err != nil {
		log.Fatal("net.ListenMulticastUDP err")
	}

	return conn, nil
}

// input go routine, which just read data from socket
// and copy to pkg, then send pkg to rchan
func (client *Client) rcv_routine () {
	buf := make([]byte, 8192)
	for {
		nread, peer_addr, err := client.conn.ReadFromUDP (buf)
		if err != nil {
			log.Fatal("ReadFromUDP err: ", err)
		}

		var rcv_pkg *PKG
		rcv_pkg = client.pkgcache.Get ()
		rcv_pkg.Reset ()
		rcv_pkg.SetBuf (buf[:nread])
		rcv_pkg.Addr = *peer_addr
		client.rchan<- rcv_pkg
	}
}

// write to peer, ack and nak
func (client *Client ) snd_routine () {
	for rcv_pkg := range client.wchan {
		/// fmt.Println ("len: ", rcv_pkg.GetLen (),
		///	", buflen: ", len (rcv_pkg.GetBuf ()))

		client.conn.WriteToUDP (rcv_pkg.GetBuf (), &rcv_pkg.Addr)

		rcv_pkg_type := rcv_pkg.GetType ()
		switch rcv_pkg_type {
			case TYPE_NAK:
				client.nak_snd_time = time.Now ()
			case TYPE_ACK:
				client.ack_snd_time = time.Now ()
		}
		client.pkgcache.Put (rcv_pkg)
	}
}

func (client *Client) send_ack (raddr *net.UDPAddr) {
	ack_pkg := client.pkgcache.Get ()
	ack_pkg.Reset ()

	ack_pkg.SetType (TYPE_ACK)
	ack_pkg.SetSeq (client.last_rcv_seq)  // only DATA has seq

	ack_pkg.Addr = *raddr
	client.wchan <- ack_pkg
}

func (client *Client) send_nack (lostinfo []LostSeqInfo, raddr *net.UDPAddr) {
	// var lost LostSeqInfo
	nak_pkg := client.pkgcache.Get ()
	nak_pkg.Reset ()
	snd_buf := nak_pkg.GetVal ()
	var count int
	for idx, lost := range lostinfo {
		nak_pkg.SetType (TYPE_NAK)
		nak_pkg.SetSeq (client.last_rcv_seq)  // nak carry with last ordered seq, like ack
		count++
		nak_pkg.SetLen (count)

		//fmt.Println ("Begin: ", lost.Begin, ", Count: ", lost.Count)
		binary.BigEndian.PutUint32 (snd_buf[idx * 4:(idx + 1) * 4], lost.Begin)
		binary.BigEndian.PutUint32 (snd_buf[(idx + 1) * 4:(idx + 2) * 4], lost.Count)
/* test
		fmt.Println ("Begin: ", lost.Begin, "Count: ", lost.Count)
		var i uint32
		for i = 0; i < lost.Count; i++ {
			if client.rqueue.SeqInQueue (lost.Begin + i) {
				log.Fatal (lost.Begin + i, " is in rcv queue")
			}
		}
*/
	}

	if nak_pkg.GetLen () == 0 {
		log.Fatal ("must hole in rqueue, but we cann't found it")
	}

	nak_pkg.SetBufLen (count * 8 + 10)

	nak_pkg.Addr = *raddr
	client.wchan <- nak_pkg
}

func (client *Client) rcv_data_pkg (pkg *PKG) {
	// insert to proper place of qeue by seq
	client.rqueue.Enque (pkg)

	var ordered_count int
	for {
		first_pkg := client.rqueue.First ()
		if first_pkg == nil {
			break
		}
		first_pkg_seq := first_pkg.GetSeq ()
		if client.last_rcv_seq == 0 {
			if first_pkg_seq == 1 {
				client.HandlePackage (first_pkg)
				ordered_count++
				first_pkg = client.rqueue.Deque ()
				client.pkgcache.Put (first_pkg)
				client.last_rcv_seq = first_pkg_seq

				// first rcv seq, send ack immediately
				client.send_ack (&pkg.Addr)
				continue
			} else {
				// recv unordered pkg, send nack immediately
				lostinfo := make ([]LostSeqInfo, 1)
				lostinfo[0].Begin = 1
				lostinfo[0].Count = first_pkg_seq - 1
				client.send_nack (lostinfo, &pkg.Addr)
				break
			}
		} else {
			if client.last_rcv_seq + 1 == first_pkg_seq {
				ordered_count++
				// every 100 send a ack, server use ack to discard old pkg
				if ordered_count > 100 {
					client.send_ack (&pkg.Addr)
					ordered_count = 0
				}
				// rcv ordered seq
				client.HandlePackage (first_pkg)
				first_pkg = client.rqueue.Deque ()
				client.pkgcache.Put (first_pkg)
				client.last_rcv_seq = first_pkg_seq
			} else {
				// nak limit
				if time.Since (client.nak_snd_time) > time.Microsecond * 200 {
					// collect lost seq info
					lost_seq_info := client.rqueue.GetLostSeqInfo (client.last_rcv_seq)
					client.send_nack (lost_seq_info, &pkg.Addr)
				}
				break
			}
		}
	}

	if time.Since (client.ack_snd_time) > time.Millisecond {
		client.send_ack (&pkg.Addr)
	}
}

func (client *Client) rcv_nak_pkg (pkg *PKG) {
	log.Println ("rcv_nak_pkg")
}

func (client *Client) Run () {
	client.rchan = make (chan *PKG, 4096)
	client.wchan = make (chan *PKG, 4096)

	// rcv routine just read data from socket, 
	// and make pkg then send to rchan channel.
	go client.rcv_routine ()
	go client.snd_routine ()

	/// now := time.Now ()
	for {
		select {
			case rcv_pkg := <-client.rchan:
				{
					// enque, and then iterate the rqueue
					pkg_type := rcv_pkg.GetType ()
					// fmt.Println ("rcvpkg.pkg_type: ", pkg_type, "seq: ", rcv_pkg.GetSeq ())
					switch pkg_type {
					case TYPE_DATA:
						client.rcv_data_pkg (rcv_pkg)
						// client receive NAK from other client, to avid to send dup NAK
						// to server. Normally, one client is selected as client leader,
						// which will send NAK immediately, others will watch this NAK,
						// if the NAK contains lost seq of theirs, others will delete NAK
						// of themselves.
					case TYPE_NAK:
						client.rcv_nak_pkg (rcv_pkg)
					default:
						log.Fatal ("client just only receieve DATA or NAK pkg")
					}
				}
/*
			case <-time.After (time.Microsecond * 200):
				{
					// nak pkg may lost, so there must have a timer to trig resending nak
					if time.Since (client.nak_snd_time) > time.Microsecond * 200 {
						lost_seq_info := client.rqueue.GetLostSeqInfo (client.last_rcv_seq)
						if len (lost_seq_info) > 0 {
							client.send_nack (lost_seq_info, &pkg.Addr)
						}
					}
				}
*/
		}
	}
}

func NewClient (maddr *MAddr) *Client {
	var client = new (Client)
	conn, err := client.BindMAddr (maddr)
	if err != nil {
		log.Fatal ("BindMaddr err: ", err)
	}
	client.conn = conn
	client.pkgcache = NewPkgCache (1024)
	rqueue := NewRqueue ()
	client.rqueue = rqueue
	return client
}
