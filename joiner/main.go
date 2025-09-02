package main

import (
	"fmt"
	"os"

	"github.com/zrygan/pokemonbattler/helper"
)

func main() {
	_, conn := helper.JoinTo(os.Args)
	defer conn.Close()

	msg := []byte("hello via udp")
	conn.Write(msg)
	msg = []byte("okay bye now!")
	conn.Write(msg)

	buf := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buf)
	fmt.Println("Server replied:", string(buf[:n]))
}
