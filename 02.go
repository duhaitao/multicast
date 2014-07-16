package main

import (
	// "encoding/binary"
	"fmt"
	"log"
	"net"
	//"time"
	"github.com/duhaitao/multicast/rmcast"
)

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
	/*   +------------------------+
	 *   | type | len | seq | val |
	 *   +------------------------+
	 *     2B     4B    4B
	 *
	 *   type:
	 *     1 -- data pkg
	 *     2 -- ack
	 *     3 -- nak
	 *
	 *   len:
	 *     val len
	 *
	 *   seq:
	 *     send sequence, increase monotonicly by 1
	 *
	 *   if type == ack:
	 *     no val, receiver must ack first pkg, then ack per millsecond normally, sender
	 *     can assure reciever is alive by ack
	 *   
	 *   if type == nak:
	 *     val format:
	 *   +-----------------------------+
	 *   | seq | len | ... | seq | len |
	 *   +-----------------------------+
	 *     4B    4B          4B     4B
	 *
	 *     seq: 
	 * 		first seq of lost range, len is length of range. every receiver hole occupy a
	 *      <seq, len> pair
	 */
	// cwin_siz := 1 // init congestion windown is 1, like tcp
	// swin_siz := 10 // init slide window size 
	var seq uint32
	snd_pkg := rmcast.NewPKG ()
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
