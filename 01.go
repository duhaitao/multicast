package main

import (
	//"fmt"
	"log"
	"github.com/duhaitao/multicast/rmcast"
	//"encoding/binary"
)

func HandlePackage (npkg *rmcast.PKG) {
	fmt.Println ("rcv seq: ", npkg.GetSeq ())
}

func main() {
	maddr, err : = rmcast.NewMAddr ("eth0", "230.1.1.1", 12345)
	if err != nil {
		log.Fatal("NewMaddr err: ", err)
	}

	client := rmcast.NewClient ()
	conn, err := client.BindMAddr (maddr)
	if err != nil {
		log.Fatal ("BindMaddr err: ", err)
	}

	client.RegisterHandlePackage (HandlePackage)
	client.Run ()
}
