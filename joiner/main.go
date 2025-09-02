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
	msg := messages.MakeHandshakeRequest()
	fmt.Println("Sent: ", msg.MessageType, msg.MessageParams)
	conn.Write(msg.SerializeMessage())

	buf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		bc := buf[:n]
		des := messages.DeserializeMessage(bc)

		fmt.Println("Received: ", des.MessageType, des.MessageParams)
	}
}
