package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
)

func lookForMatch() map[string]string {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// broadcast
	bAddr := net.UDPAddr{
		IP:   net.IPv4bcast, // 255.255.255.255
		Port: 50000,         // discovery port
	}

	// send the broadcast message
	msg := messages.MakeJoiningMMB()
	_, err = conn.WriteToUDP(msg.SerializeMessage(), &bAddr)
	if err != nil {
		panic(err)
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
				"Found a HOST, received a "+messages.MMB_HOSTING+" message",
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

func selectMatch(hosts map[string]string) peer.PeerDescriptor {
	// once all joinable are found, ask which one to join/Handshake
	// stores [ip, port] of invited
	var hostDets []string = nil
	hostName := ""
	for hostDets == nil {
		hostName = netio.PRLine("Send an invite to...")
		hostVal, ok := hosts[hostName]
		if ok {
			hostDets = strings.Split(hostVal, " ")
		} else {
			fmt.Println("Player name not found.")
		}
	}

	return peer.MakeRemotePD(hostName, hostDets[0], hostDets[1])
}

func handshake(self peer.PeerDescriptor, host peer.PeerDescriptor) int {
	// send a HandshakeRequest to the Host
	msg := messages.MakeHandshakeRequest(self)

	netio.VerboseEventLog(
		"Inviting "+host.Name+", sent a "+messages.HandshakeRequest+" message",
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
				"The match is accepted, received a "+messages.HandshakeResponse+" message from "+host.Name,
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
	}
}

func main() {
	self := peer.MakePDFromLogin("joiner")
	defer self.Conn.Close()

	availableHosts := lookForMatch()
	matchedHost := selectMatch(availableHosts)
	handshake(self, matchedHost)
}
