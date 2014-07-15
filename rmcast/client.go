package rmcast

import (
	"fmt"
	"log"
	"net"
	// "container/list"
	// "time"
	// "github.com/duhaitao/multicast/rmcast"
	// "encoding/binary"
)

type Client struct {
	rqueue *Rqueue
	last_seq uint32
	conn *net.UDPConn
	pkgcache *PkgCache
	rchan chan *PKG
	wchan chan *PKG

	// flag, to inform main routine to handle Squeue
	handleSqueue bool
	// call back for upper application
	HandlePackage func (pkg *PKG)
}

func (client *Client) RegisterHandlePackage (fn func (pkg *PKG)) {
	client.HandlePackage = fn
}

func (client *Client) BindMAddr (maddr *MAddr) (*net.UDPConn, error) {
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
		rcv_pkg.SetBuf (buf[:nread])
		rcv_pkg.Addr = *peer_addr
		client.rchan<- rcv_pkg
	}
}

// write to peer, ack and nak
func (client *Client ) snd_routine () {
	for rcv_pkg := range client.wchan {
		client.conn.WriteToUDP (rcv_pkg.GetBuf (), &rcv_pkg.Addr)
		client.pkgcache.Put (rcv_pkg)
	}
}

func (client *Client) send_ack (pkg *PKG) {

}

func (client *Client) send_nack (pkg *PKG) {

}

func (client *Client) rcv_data_pkg (pkg *PKG) {
	// insert to proper place of qeue by seq
	log.Println ("rcv_data_pkg")
	client.rqueue.Enque (pkg)

	for {
		first_pkg := client.rqueue.First ()
		if first_pkg == nil {
			break
		}
		first_pkg_seq := first_pkg.GetSeq ()
		if client.last_seq == 0 {
			if first_pkg_seq == 1 {
				client.HandlePackage (first_pkg)
				first_pkg = client.rqueue.Deque ()
				client.pkgcache.Put (first_pkg)
				client.last_seq = first_pkg_seq

				// first rcv seq, send ack immediately
				client.send_ack (first_pkg)
				continue
			} else {
				// recv unordered pkg, send nack immediately
				client.send_nack (first_pkg)
				break
			}
		} else {
			if client.last_seq + 1 == first_pkg_seq {
				// rcv ordered seq
				client.HandlePackage (first_pkg)
				first_pkg = client.rqueue.Deque ()
				client.pkgcache.Put (first_pkg)
				client.last_seq = first_pkg_seq
			} else {
				break
			}
		}
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
		for rcv_pkg := range client.rchan {
			// enque, and then iterate the rqueue
			pkg_type := rcv_pkg.GetType ()
			fmt.Println ("rcvpkg.pkg_type: ", pkg_type, "seq: ", rcv_pkg.GetSeq ())
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
