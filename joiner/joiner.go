package main

import (
	"fmt"
	"os"

	"github.com/zrygan/pokemonbattler/helper"
	"github.com/zrygan/pokemonbattler/messages"
)

func main() {
	// a general variable for Message struct
	var msg messages.Message

	_, conn := helper.JoinTo(os.Args)
	defer conn.Close()

	// send a HandshakeRequest to the Host (if there is?)
	msg = messages.MakeHandshakeRequest()
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

		// buffer content
		bc := buf[:n]
		msg = *messages.DeserializeMessage(bc)

		helper.VerboseEventLog(
			"A message was RECEIVED",
			&helper.LogOptions{
				MT: msg.MessageType,
				MP: fmt.Sprint(*msg.MessageParams),
				MS: addr.String(),
			},
		)

		// check if this is a HandshakeResponse
		// then create a BattleSetup once done
		if msg.MessageType == messages.HandshakeResponse {
		}
	}
}
