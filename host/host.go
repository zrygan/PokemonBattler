// Package main implements the Pokemon Battler host application.
// The host waits for joiners to connect and manages the battle setup process.
package main

import (
	"strings"

	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
)

// waitForMatch listens for incoming joiner connections and match requests.
// It handles discovery messages (MMB_JOINING) and handshake requests.
// Returns a PeerDescriptor for the accepted joiner.
func waitForMatch(self peer.PeerDescriptor) peer.PeerDescriptor {
	buf := make([]byte, 1024)

	for {
		n, rem, err := self.Conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		msg := messages.DeserializeMessage(buf[:n])

		// send a message if somebody wants to join
		switch msg.MessageType {
		case messages.MMB_JOINING:
			netio.VerboseEventLog(
				"Found a JOINER, received a "+messages.MMB_JOINING+" message",
				nil,
			)

			res := messages.MakeHostingMMB(self)

			netio.VerboseEventLog(
				"Found a JOINER, sent a "+messages.MMB_HOSTING+" message",
				&netio.LogOptions{
					MessageParams: res.MessageParams,
				},
			)

			_, _ = self.Conn.WriteToUDP(
				res.SerializeMessage(),
				rem,
			)

		// if somebody is asking to handshake
		// return the peer descriptor of the joiner if you accepted it
		case messages.HandshakeRequest:
			name, nameOK := (*msg.MessageParams)["name"].(string)
			if nameOK {
				netio.VerboseEventLog(
					"Match found, received a "+messages.HandshakeRequest+" message from "+name,
					&netio.LogOptions{
						MessageParams: msg.MessageParams,
						MS:            rem.String(),
					},
				)

				isAccepted := strings.ToLower(netio.PRLine("Accept this player? [Y:default / N]: "))
				if isAccepted != "n" {
					return peer.MakePD(name, nil, rem)
				}
			}
		}
	}
}

// handshake sends a handshake response to the joiner and returns the battle seed.
// The seed is used to synchronize random number generation between host and joiner.
func handshake(self peer.PeerDescriptor, join peer.PeerDescriptor) int {
	msg := messages.MakeHandshakeResponse()
	self.Conn.WriteToUDP(msg.SerializeMessage(), join.Addr)

	netio.VerboseEventLog(
		"The match is accepted, sent a "+messages.HandshakeResponse+" message to "+join.Name,
		&netio.LogOptions{
			MessageParams: msg.MessageParams,
		},
	)

	val, ok := (*msg.MessageParams)["seed"].(int)
	if !ok {
		panic("seed not found in handshake response")
	}
	return val
}

// main is the entry point for the host application.
// It initializes the host, waits for a joiner, performs handshake, and starts the battle.
func main() {
	self := peer.MakePDFromLogin("hostW")
	defer self.Conn.Close()

	// at the start say that somebody can join you
	joiner := waitForMatch(self)

	// when watchForMatch returns, initialize a handshake
	handshake(self, joiner)

	// set the communication for a battle
	game.Host_setCommMode()

	// create Host's player
	game.PlayerSetUp(self)
}
