package helper

import (
	"net"
	"strconv"
)

// The parameter arguments is
func HostTo(arguments []string) (*net.UDPAddr, *net.UDPConn) {
	if len(arguments) != 3 {
		panic("arguments must be of the form: <port> <ip>")
	}

	port, err := strconv.Atoi(arguments[1])
	if err != nil {
		panic(err)
	}

	ip := net.ParseIP(arguments[2])
	if ip == nil {
		panic("invalid IP address")
	}

	addr := &net.UDPAddr{
		Port: port,
		IP:   ip,
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	VerboseEventLog(
		"A new HOST connected.",
		&LogOptions{
			Port: arguments[1],
			IP:   arguments[2],
		},
	)

	return addr, conn
}

func JoinTo(arguments []string) (*net.UDPAddr, *net.UDPConn) {
	if len(arguments) != 3 {
		panic("arguments must be of the form: <port> <ip>")
	}

	// no need for parsing these two since we need them as strings
	port := arguments[1]
	ip := arguments[2]

	addr, err := net.ResolveUDPAddr("udp", ip+":"+port)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	VerboseEventLog(
		"A new JOINER connected.",
		&LogOptions{
			Port: port,
			IP:   ip,
		},
	)

	return addr, conn
}
