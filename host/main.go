package main

import (
	"fmt"
	"net"
)

func main() {
	addr := net.UDPAddr{
		Port: 5005,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, _ := net.ListenUDP("udp", &addr)
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, _ := conn.ReadFromUDP(buf)
		fmt.Printf("Got from %s: %s\n", remoteAddr, string(buf[:n]))

		conn.WriteToUDP([]byte("hello! this is my response!"), remoteAddr)
	}
}
