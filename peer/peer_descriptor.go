// Package peer provides networking abstractions for Pokemon Battler peers.
// It handles peer connections, addressing, and communication setup.
package peer

import (
	"net"
	"strconv"

	"github.com/zrygan/pokemonbattler/netio"
)

// PeerDescriptor represents a network peer with connection information.
// It contains the peer's name, UDP connection, and network address.
type PeerDescriptor struct {
	Name string       // Human-readable name of the peer
	Conn *net.UDPConn // UDP connection for communication
	Addr *net.UDPAddr // Network address (IP and Port)
}

// MakePDFromLogin creates a PeerDescriptor using interactive login.
// The userType parameter is used for logging and connection setup.
func MakePDFromLogin(userType string) PeerDescriptor {
	pd := MakePDBase(netio.Login(userType))

	netio.VerboseEventLog(
		"A new "+userType+" connected.",
		&netio.LogOptions{
			Port: strconv.Itoa(pd.Addr.Port),
			IP:   pd.Addr.IP.String(),
		},
	)

	return pd
}

// MakePDBase creates a PeerDescriptor with the specified parameters.
// If isLocal is true, it creates a listening UDP connection; otherwise, it creates a remote descriptor.
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

// MakeLocalPD creates a PeerDescriptor for a local peer with a listening connection.
func MakeLocalPD(name, ip, port string) PeerDescriptor {
	return MakePDBase(name, ip, port, true)
}

// MakeRemotePD creates a PeerDescriptor for a remote peer without a listening connection.
func MakeRemotePD(name, ip, port string) PeerDescriptor {
	return MakePDBase(name, ip, port, false)
}

// MakePD creates a PeerDescriptor with the given name, connection, and address.
// This is the low-level constructor used by other Make functions.
func MakePD(name string, conn *net.UDPConn, addr *net.UDPAddr) PeerDescriptor {
	return PeerDescriptor{
		Name: name,
		Conn: conn,
		Addr: addr,
	}
}
