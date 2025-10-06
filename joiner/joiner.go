package main

import (
	"fmt"
	"net"
	"os"

	"github.com/zrygan/pokemonbattler/helper"
	"github.com/zrygan/pokemonbattler/messages"
)

func lookForJoinables() {

}

func joinTo(arguments []string) (*net.UDPAddr, *net.UDPConn) {
	if len(arguments) != 3 {
		panic("arguments must be of the form: <port> <ip>")
	}

	// no need for parsing these two since we need them as strings
	port := arguments[1]
	ip := arguments[2]

	addr, err := net.ResolveUDPAddr("udp", ip+":"+port)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	helper.VerboseEventLog(
		"A new JOINER connected.",
		&helper.LogOptions{
			Port: port,
			IP:   ip,
		},
	)

	return addr, conn
}

func main() {
	// a general variable for Message struct
	var msg messages.Message

	_, conn := joinTo(os.Args)
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
		msg := *messages.DeserializeMessage(bc)

		helper.VerboseEventLog(
			"A message was RECEIVED",
			&helper.LogOptions{
				MessageString: string(msg.SerializeMessage()),
				MT:            msg.MessageType,
				MP:            fmt.Sprint(*msg.MessageParams),
				MS:            addr.String(),
			},
		)

		// check if this is a HandshakeResponse
		// then create a BattleSetup once done
		if msg.MessageType == messages.HandshakeResponse {
		}
	}
}
