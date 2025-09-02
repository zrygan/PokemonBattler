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

	_, conn := helper.HostTo(os.Args)
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		bc := buf[:n]
		msg = *messages.DeserializeMessage(bc)

		helper.VerboseEventLog(
			"A message was RECEIVED",
			&helper.LogOptions{
				MT: msg.MessageType,
				MP: fmt.Sprint(msg.MessageParams),
				MS: addr.String(),
			},
		)

		if msg.MessageType == messages.HandshakeRequest {
			// then send back a HandshakeResponse
			msg = messages.MakeHandshakeResponse()
			conn.WriteToUDP(msg.SerializeMessage(), addr)

			helper.VerboseEventLog(
				"A message was SENT",
				&helper.LogOptions{
					MT: msg.MessageType,
					MP: fmt.Sprint(*msg.MessageParams),
				},
			)

			// create a BattleSetup once done
		}

	}
}
