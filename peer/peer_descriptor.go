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
// For local connections, if the port is already in use, it automatically tries incrementing ports
// (up to 10 attempts) until a free port is found.
func MakePDBase(name, ip, port string, isLocal bool) PeerDescriptor {
	originalPort := port
	var addr *net.UDPAddr
	var conn *net.UDPConn
	var err error

	if isLocal {
		// Try up to 10 ports starting from the requested port
		maxAttempts := 10
		for attempt := 0; attempt < maxAttempts; attempt++ {
			addr, err = net.ResolveUDPAddr("udp", ip+":"+port)
			if err != nil {
				panic(err)
			}

			conn, err = net.ListenUDP("udp", addr)
			if err == nil {
				// Successfully bound to port
				// Update addr to reflect the actual bound address
				addr = conn.LocalAddr().(*net.UDPAddr)
				
				if port != originalPort {
					netio.VerboseEventLog(
						"Port "+originalPort+" was in use, automatically using port "+strconv.Itoa(addr.Port)+" instead.",
						&netio.LogOptions{
							Port: strconv.Itoa(addr.Port),
							IP:   ip,
						},
					)
				}
				break
			}

			// Port in use, try next port
			portNum, convErr := strconv.Atoi(port)
			if convErr != nil {
				// If port is "0" or invalid, let OS choose
				port = "0"
				addr, err = net.ResolveUDPAddr("udp", ip+":"+port)
				if err != nil {
					panic(err)
				}
				conn, err = net.ListenUDP("udp", addr)
				if err != nil {
					panic(err)
				}
				// Update addr to reflect the actual bound address
				addr = conn.LocalAddr().(*net.UDPAddr)
				break
			}
			port = strconv.Itoa(portNum + 1)
		}

		if conn == nil {
			panic("failed to bind to any port after " + strconv.Itoa(maxAttempts) + " attempts starting from " + originalPort)
		}
	} else {
		// Remote descriptor, no binding needed
		addr, err = net.ResolveUDPAddr("udp", ip+":"+port)
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
