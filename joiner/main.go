package main

import (
	"fmt"
	"net"
)

func main() {
	serverAddr := net.UDPAddr{
		Port: 5005,
		IP:   net.ParseIP("127.0.0.1"),
	}
	conn, _ := net.DialUDP("udp", nil, &serverAddr)
	defer conn.Close()

	msg := []byte("hello via udp")
	conn.Write(msg)
	msg = []byte("okay bye now!")
	conn.Write(msg)

	buf := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buf)
	fmt.Println("Server replied:", string(buf[:n]))
}
