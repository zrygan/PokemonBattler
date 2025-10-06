package main

import (
	"fmt"
	"net"
	"time"

	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
)

func lookForJoinables(udp *net.UDPAddr) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// broadcast
	bAddr := net.UDPAddr{
		IP:   net.IPv4bcast,      // 255.255.255.255
		Port: game.DiscoveryPort, // 50000
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
			hostName := fmt.Sprint((*msg.MessageParams)["hostName"])
			hostPort := fmt.Sprint((*msg.MessageParams)["port"])
			hostIP := fmt.Sprint((*msg.MessageParams)["ip"])

			netio.VerboseEventLog(
				"Found a HOST, asked to join.",
				&netio.LogOptions{
					Name: hostName,
					Port: hostPort,
					IP:   hostIP,
				},
			)

			hostDets := fmt.Sprintf("%s %s", hostIP, hostPort)

			discoveredHosts[hostDets] = hostName
		}
	}

	fmt.Println("Discovered Hosts")
	for details, name := range discoveredHosts {
		fmt.Printf("%s @ %s\n", name, details)
	}
}

func joinTo() game.PeerDescriptor {
	name, port, ip := netio.Login()

	addr, err := net.ResolveUDPAddr("udp", ip+":"+port)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	netio.VerboseEventLog(
		"A new JOINER connected.",
		&netio.LogOptions{
			Port: port,
			IP:   ip,
		},
	)

	return game.PeerDescriptor{
		Name: name,
		Conn: conn,
		Addr: addr,
	}
}

func handshakeHost(pd game.PeerDescriptor) {
	// send a HandshakeRequest to the Host (if there is?)
	msg := messages.MakeHandshakeRequest()
	pd.Conn.Write(msg.SerializeMessage())

	netio.VerboseEventLog(
		"A message was SENT",
		&netio.LogOptions{
			MT: msg.MessageType,
		},
	)

	buf := make([]byte, 1024)
	for {
		n, addr, err := pd.Conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		// buffer content
		bc := buf[:n]
		msg := *messages.DeserializeMessage(bc)

		netio.VerboseEventLog(
			"A message was RECEIVED",
			&netio.LogOptions{
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

func main() {
	pd := joinTo()
	defer pd.Conn.Close()

	lookForJoinables(pd.Addr)
}
