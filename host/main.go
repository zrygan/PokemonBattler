package main

import (
	"fmt"
	"os"

	"github.com/zrygan/pokemonbattler/helper"
	"github.com/zrygan/pokemonbattler/messages"
)

func main() {
	_, conn := helper.HostTo(os.Args)
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, _ := conn.ReadFromUDP(buf)
		msg := buf[:n]
		des := messages.DeserializeMessage(msg)

		fmt.Println("Received from joiner:", des.MessageType, "from", remoteAddr)
	}
}
