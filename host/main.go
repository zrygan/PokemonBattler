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
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		bc := buf[:n]
		des := messages.DeserializeMessage(bc)
		fmt.Println("Received: ", des.MessageType, des.MessageParams)

		if des.MessageType == messages.HandshakeRequest {
			// then send back a HandshakeResponse
			msg := messages.MakeHandshakeResponse()
			conn.WriteToUDP(msg.SerializeMessage(), addr)
			fmt.Println("Sent: ", msg.MessageType, msg.MessageParams)
		}

	}
}
