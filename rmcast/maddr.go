package rmcast

import (
	"net"
	"log"
)

// a multicast addr
type MAddr struct {
	Iface *net.Interface
	Addr *net.UDPAddr
}

func NewMAddr (ifname string, mcastip string, port int) (*MAddr, error) {
	maddr := new (MAddr)
	iface, err := net.InterfaceByName (ifname)
	if err != nil {
		return nil, err
	}
	maddr.Iface = iface

	ip := net.ParseIP (mcastip)
	if ip == nil {
		log.Println ("ParseIp err: ", ip)
		return nil, err // TODO, define specific error
	}
	maddr.Addr = &net.UDPAddr {IP: ip, Port: port} // ipv4 ignore zone

	return maddr, nil
}
