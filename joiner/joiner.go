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

func lookForJoinables(udp *net.UDPAddr) map[string]string {
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
	_, err = conn.WriteToUDP([]byte(messages.MMB_JOINING), &bAddr)
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

func handshake(self peer.PeerDescriptor, host peer.PeerDescriptor) {
	// send a HandshakeRequest to the Host
	msg := messages.MakeHandshakeRequest()

	netio.VerboseEventLog(
		"Inviting "+host.Name+", sent a "+messages.HandshakeRequest+" message",
		&netio.LogOptions{
			MT: msg.MessageType,
		},
	)

	// send the handshake to host address
	_, err := self.Conn.WriteToUDP(msg.SerializeMessage(), host.Addr)
	if err != nil {
		panic(err)
	}

	// wait for host response
}

func main() {
	self := peer.MakePDFromLogin("joiner")
	defer self.Conn.Close()

	hosts := lookForJoinables(self.Addr)

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

	host := peer.MakeRemotePD(hostName, hostDets[0], hostDets[1])

	// once host is found, init a handshake
	handshake(self, host)
}
