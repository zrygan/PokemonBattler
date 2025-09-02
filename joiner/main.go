package main

import (
	"fmt"
	"os"

	"github.com/zrygan/pokemonbattler/helper"
	"github.com/zrygan/pokemonbattler/messages"
)

func main() {
	_, conn := helper.JoinTo(os.Args)
	defer conn.Close()

	// send a HandshakeRequest to the Host (if there is?)
	msg := messages.Message{
		MessageType:   messages.HandshakeRequest,
		MessageParams: nil, // Request has no Params
	}

	conn.Write(msg.SerializeMessage())

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Read error:", err)
			continue
		}
		fmt.Println("this sould not run", string(buf[:n]), "from", remoteAddr)
	}
}
