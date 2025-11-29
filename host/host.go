// Package main implements the Pokemon Battler host application.
// The host waits for joiners to connect and manages the battle setup process.
package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
)

// waitForMatch listens for incoming joiner connections and match requests.
// It handles discovery messages (MMB_JOINING) and handshake requests.
// Returns a PeerDescriptor for the accepted joiner and a slice of spectators.
func waitForMatch(self peer.PeerDescriptor) (peer.PeerDescriptor, []peer.PeerDescriptor) {
	buf := make([]byte, 1024)
	spectators := make([]peer.PeerDescriptor, 0)

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
				"PokeProtocol: Host Peer received MMB_JOINING discovery message",
				&netio.LogOptions{
					MessageParams: msg.MessageParams,
					MS:            rem.String(),
				},
			)

			res := messages.MakeHostingMMB(self)

			netio.VerboseEventLog(
				"PokeProtocol: Host Peer sent discovery response message",
				&netio.LogOptions{
					MessageParams: res.MessageParams,
				},
			)

			_, _ = self.Conn.WriteToUDP(
				res.SerializeMessage(),
				rem,
			)

		case messages.SpectatorRequest:
			netio.VerboseEventLog(
				"PokeProtocol: Host Peer received SPECTATOR_REQUEST from new spectator",
				&netio.LogOptions{
					MessageParams: msg.MessageParams,
					MS:            rem.String(),
				},
			)

			// Accept spectator automatically
			spectatorName := "Spectator" + rem.String()
			spectator := peer.MakePD(spectatorName, nil, rem)
			spectators = append(spectators, spectator)
			netio.VerboseEventLog(
				"PokeProtocol: Spectator Peer joined battle session",
				&netio.LogOptions{
					MS: rem.String(),
				},
			)

		// if somebody is asking to handshake
		// return the peer descriptor of the joiner if you accepted it
		case messages.HandshakeRequest:
			name, nameOK := (*msg.MessageParams)["name"].(string)
			if nameOK {
				netio.VerboseEventLog(
					"PokeProtocol: Host Peer received HANDSHAKE_REQUEST from Joiner Peer '"+name+"'",
					&netio.LogOptions{
						MessageParams: msg.MessageParams,
						MS:            rem.String(),
					},
				)
				isAccepted := strings.ToLower(netio.PRLine("Accept this player? [Y:default / N]: "))
				if isAccepted != "n" {
					return peer.MakePD(name, nil, rem), spectators
				} else {
					// Send rejection message to joiner
					rejectMsg := messages.MakeHandshakeRejected()
					self.Conn.WriteToUDP(rejectMsg.SerializeMessage(), rem)
					netio.VerboseEventLog("PokeProtocol: Host Peer rejected connection, sent HANDSHAKE_REJECTED to Joiner Peer", nil)
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
		"PokeProtocol: Host Peer sent HANDSHAKE_RESPONSE with seed to Joiner Peer '"+join.Name+"'",
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
	// Parse command-line flags
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging of network events")
	flag.Parse()

	// Set global verbose mode
	netio.Verbose = *verboseFlag

	self := peer.MakePDFromLogin("hostW")
	defer self.Conn.Close()

	// Main host loop - keep hosting battles
	for {
		fmt.Println("\n=== HOSTING NEW BATTLE ===")
		fmt.Println("Waiting for players to join...")

		// at the start say that somebody can join you
		joiner, spectators := waitForMatch(self)

		// when watchForMatch returns, initialize a handshake
		seed := handshake(self, joiner)

		// set the communication for a battle
		cmode := game.Host_setCMode(self, joiner)

		// create Host's player
		p := game.PlayerSetUp(self)

		// make BattleSetup and get opponent player info
		opponentPlayer := game.BattleSetup(p, joiner, cmode, spectators)

		// Start the battle with spectators
		game.RunBattle(&p, &opponentPlayer, seed, cmode, true, spectators)

		// Battle ended, clear spectators and return to main menu
		fmt.Println("\n=== BATTLE COMPLETED ===")

		// Notify spectators that battle has ended and they should look for new battles
		if len(spectators) > 0 {
			fmt.Printf("Disconnecting %d spectator(s)...\n", len(spectators))
			// Give spectators time to process GAME_OVER message
			time.Sleep(1 * time.Second)
		}

		fmt.Println("Returning to host menu...")
		time.Sleep(2 * time.Second)
	}
}
