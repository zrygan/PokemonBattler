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
	conn.Write(msg.SerializeMessage())

	helper.VerboseEventLog(
		"A message was SENT",
		&helper.LogOptions{
			MT: msg.MessageType,
		},
	)

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		bc := buf[:n]
		des := messages.DeserializeMessage(bc)

		helper.VerboseEventLog(
			"A message was RECEIVED",
			&helper.LogOptions{
				MT: des.MessageType,
				MP: fmt.Sprint(*des.MessageParams),
				MS: addr.String(),
			},
		)
	}
}
