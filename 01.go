package main

import (
	"fmt"
	"log"
	"github.com/duhaitao/multicast/rmcast"
	//"encoding/binary"
)

func HandlePackage (npkg *rmcast.PKG) {
	fmt.Println ("HandlePackage rcv seq: ", npkg.GetSeq ())
}

func main() {
	maddr, err := rmcast.NewMAddr ("eth1", "230.1.1.1", 12345)
	if err != nil {
		log.Fatal("NewMaddr err: ", err)
	}

	client := rmcast.NewClient (maddr)
	client.RegisterHandlePackage (HandlePackage)
	client.Run ()
}
