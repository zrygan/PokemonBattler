package peer

import "net"

type PeerDescriptor struct {
	Name string
	Conn *net.UDPConn
	Addr *net.UDPAddr // has Port and IP
}

func MakePDBase(name string, ip string, port string) PeerDescriptor {
	addr, err := net.ResolveUDPAddr("udp", ip+":"+port)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	return MakePD(name, conn, addr)
}

func MakePD(name string, conn *net.UDPConn, addr *net.UDPAddr) PeerDescriptor {
	return PeerDescriptor{
		Name: name,
		Conn: conn,
		Addr: addr,
	}
}
