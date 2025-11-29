// Package main implements the Pokemon Battler joiner application.
// The joiner discovers available hosts and connects to one for a battle.
package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
)

// lookForMatch broadcasts a discovery message to find available hosts on the network.
// It listens for 3 seconds and returns a map of discovered hosts with their connection details.
// The returned map keys are host names, values are "ip port" strings.
func lookForMatch() map[string]string {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Broadcast to multiple ports to discover hosts (50000-50010)
	// This allows discovery of hosts that auto-incremented to different ports
	msg := messages.MakeJoiningMMB()
	msgBytes := msg.SerializeMessage()

	for port := 50000; port <= 50010; port++ {
		bAddr := net.UDPAddr{
			IP:   net.IPv4bcast, // 255.255.255.255
			Port: port,
		}
		conn.WriteToUDP(msgBytes, &bAddr)
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	// make a map of discovered hosts
	discoveredHosts := make(map[string]string)

	for {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
			// panic(err)
		}

		msg := messages.DeserializeMessage(buf[:n])
		if msg.MessageType == messages.MMB_HOSTING {
			hostName := fmt.Sprint((*msg.MessageParams)["name"])
			hostPort := fmt.Sprint((*msg.MessageParams)["port"])
			hostIP := fmt.Sprint((*msg.MessageParams)["ip"])

			netio.VerboseEventLog(
				"PokeProtocol: Joiner Peer discovered available Host Peer '"+hostName+"'",
				&netio.LogOptions{
					Name: hostName,
					Port: hostPort,
					IP:   hostIP,
				},
			)

			hostDets := fmt.Sprintf("%s %s", hostIP, hostPort)

			discoveredHosts[hostName] = hostDets
		}
	}

	fmt.Println("Discovered Hosts")
	for name, details := range discoveredHosts {
		fmt.Printf("\t%s @ %s\n", name, details)
	}
	fmt.Println()

	return discoveredHosts
}

// selectMatch prompts the user to choose a host from the discovered hosts.
// It validates the selection and returns a PeerDescriptor for the chosen host.
// Returns nil if user requests to restart discovery with /R command.
func selectMatch(hosts map[string]string) *peer.PeerDescriptor {
	// once all joinable are found, ask which one to join/Handshake
	// stores [ip, port] of invited
	var hostDets []string = nil
	hostName := ""
	for hostDets == nil {
		hostName = netio.PRLine("Send an invite to... (or type /R to search again)")

		// Check for restart command (case-insensitive)
		if strings.ToUpper(hostName) == "/R" {
			return nil
		}

		// Try exact match first, then case-insensitive
		hostVal, ok := hosts[hostName]
		if !ok {
			// Try case-insensitive search
			for key, val := range hosts {
				if strings.EqualFold(key, hostName) {
					hostVal = val
					ok = true
					break
				}
			}
		}
		if ok {
			hostDets = strings.Split(hostVal, " ")
		} else {
			fmt.Println("Player name not found.")
		}
	}

	pd := peer.MakeRemotePD(hostName, hostDets[0], hostDets[1])
	return &pd
}

// handshake sends a handshake request to the selected host and waits for a response.
// Returns the battle seed received from the host for synchronized random number generation.
func handshake(self peer.PeerDescriptor, host peer.PeerDescriptor) int {
	// send a HandshakeRequest to the Host
	msg := messages.MakeHandshakeRequest(self)

	netio.VerboseEventLog(
		"PokeProtocol: Joiner Peer sent HANDSHAKE_REQUEST to Host Peer '"+host.Name+"'",
		&netio.LogOptions{
			MessageParams: msg.MessageParams,
		},
	)

	// send the handshake to host address
	_, err := self.Conn.WriteToUDP(msg.SerializeMessage(), host.Addr)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)

	for {
		n, _, err := self.Conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		msg := messages.DeserializeMessage(buf[:n])

		if msg.MessageType == messages.HandshakeResponse {
			netio.VerboseEventLog(
				"PokeProtocol: Joiner Peer received HANDSHAKE_RESPONSE from Host Peer '"+host.Name+"'",
				&netio.LogOptions{
					MessageParams: msg.MessageParams,
				},
			)

			val, ok := (*msg.MessageParams)["seed"].(int)
			if !ok {
				panic("seed not found in handshake response")
			}

			return val
		} else if msg.MessageType == messages.HandshakeRejected {
			netio.VerboseEventLog(
				"PokeProtocol: Connection rejected by Host Peer '"+host.Name+"'",
				nil,
			)
			fmt.Println("\nHost declined your connection request.")
			return -1 // Return -1 to signal rejection
		}
	}
}

// main is the entry point for the joiner application.
// It discovers hosts, allows user selection, and initiates the handshake process.
func main() {
	// Parse command-line flags
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging of network events")
	flag.Parse()

	// Set global verbose mode
	netio.Verbose = *verboseFlag

	self := peer.MakePDFromLogin("joiner")
	defer self.Conn.Close()

	// Main joiner loop - keep looking for battles
	for {
		fmt.Println("\n=== LOOKING FOR BATTLE ===")
		fmt.Println("Searching for hosts...")

		// Allow restarting host discovery
		var host *peer.PeerDescriptor
		var seed int
		for {
			for host == nil {
				availableHosts := lookForMatch()
				host = selectMatch(availableHosts)

				if host == nil {
					fmt.Println("\n--- Restarting host discovery ---")
				}
			}

			// when selectMatch returns, initialize a handshake
			seed = handshake(self, *host)

			// Check if handshake was rejected
			if seed == -1 {
				fmt.Println("Searching for hosts again...")
				host = nil // Reset host to restart discovery
				continue
			}

			// Handshake successful, break out of loop
			break
		}

		// get the communication mode from the host
		cmode := game.Joiner_getCMode(self)

		// create joiner's player
		p := game.PlayerSetUp(self)

		// exchange BattleSetup and get opponent player info
		opponentPlayer := game.BattleSetup(p, *host, cmode, []peer.PeerDescriptor{})

		// Start the battle (joiner has no spectators)
		game.RunBattle(&p, &opponentPlayer, seed, cmode, false, []peer.PeerDescriptor{})

		// Battle ended, return to main menu
		fmt.Println("\n=== BATTLE COMPLETED ===")
		fmt.Println("Returning to joiner menu...")
		time.Sleep(2 * time.Second)
	}
}
