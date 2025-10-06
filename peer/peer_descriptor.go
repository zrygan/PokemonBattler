package peer

import (
	"net"
	"strconv"

	"github.com/zrygan/pokemonbattler/netio"
)

type PeerDescriptor struct {
	Name string
	Conn *net.UDPConn
	Addr *net.UDPAddr // has Port and IP
}

func MakePDFromLogin(userType string) PeerDescriptor {
	pd := MakePDBase(netio.Login())

	netio.VerboseEventLog(
		"A new "+userType+" connected.",
		&netio.LogOptions{
			Port: strconv.Itoa(pd.Addr.Port),
			IP:   string(pd.Addr.IP),
		},
	)

	return pd
}

func MakePDBase(name, ip, port string, isLocal bool) PeerDescriptor {
	addr, err := net.ResolveUDPAddr("udp", ip+":"+port)
	if err != nil {
		panic(err)
	}

	var conn *net.UDPConn
	if isLocal {
		conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			panic(err)
		}
	}

	return MakePD(name, conn, addr)
}

func MakeLocalPD(name, ip, port string) PeerDescriptor {
	return MakePDBase(name, ip, port, true)
}

func MakeRemotePD(name, ip, port string) PeerDescriptor {
	return MakePDBase(name, ip, port, false)
}

func MakePD(name string, conn *net.UDPConn, addr *net.UDPAddr) PeerDescriptor {
	return PeerDescriptor{
		Name: name,
		Conn: conn,
		Addr: addr,
	}
}
