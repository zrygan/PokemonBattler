package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
)

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
		case messages.HandshakeRequest:
			name, nameOK := (*msg.MessageParams)["name"].(string)
			port, portOK := (*msg.MessageParams)["port"].(string)
			ip, ipOK := (*msg.MessageParams)["ip"].(string)

			if nameOK || portOK || ipOK {
				netio.VerboseEventLog(
					"Match found, received a "+messages.HandshakeRequest+" message from "+name,
					&netio.LogOptions{
						MT:            msg.MessageType,
						MessageParams: msg.MessageParams,
						MS:            rem.String(),
					},
				)

				return peer.MakeRemotePD(name, ip, port)
			}
		}
	}
}

func handshake(self peer.PeerDescriptor, addr *net.UDPAddr) {
	// send back a HandshakeResponse
	msg := messages.MakeHandshakeResponse()
	self.Conn.WriteToUDP(msg.SerializeMessage(), addr)

	netio.VerboseEventLog(
		"A message was SENT",
		&netio.LogOptions{
			MT:            msg.MessageType,
			MessageParams: msg.MessageParams,
		},
	)

	// create a BattleSetup once done
	netio.ShowMenu(
		"Commmunicatnetion Mode:",
		"\n\t(0) P2P: Game is directly between players",
		"\n\t(1) BROADCAST: Game is announced to local network.",
	)

	var commMode int
	for {
		line := netio.RLine()
		commMode, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("Enter 0 or 1")
			continue
		}

		if commMode == 0 || commMode == 1 {
			break
		} else {
			fmt.Println("Enter 0 or 1")
		}
	}

	netio.ShowMenu(
		"Enter your chosen Pokemon",
	)

	// Send CommunicatnetionMode to Joiner
	self.Conn.WriteToUDP([]byte(strconv.Itoa(commMode)), addr)

	//FIXME: do some checking here if the pokemon is valid
	pokemon := netio.RLine()

	netio.ShowMenu(
		"Allocate stat boosts for your special attack and defense!",
		"\nFormat is 2 integers <attack_boost> <defense_boost>",
		"\nThe two integers must sum to 10",
	)

	sb := new(game.StatBoosts)
	for {
		line := strings.Split(netio.RLine(), " ")

		// check if this is of length 2
		if len(line) != 2 {
			fmt.Println("Follow the format")
			continue
		}

		atk, err := strconv.Atoi(line[0])
		if err != nil {
			fmt.Println("Enter a integer")
			continue
		}

		def, err := strconv.Atoi(line[1])
		if err != nil {
			fmt.Println("Enter an integer")
			continue
		}

		if atk+def == 10 {
			break
		} else {
			fmt.Println("The attack and defense boost must sum to 10")
		}

		sb.SpecialAttackUses = int8(atk)
		sb.SpecialDefenseUses = int8(def)
	}

	// Then create the battle setup message
	msg = messages.MakeBattleSetup(
		game.CommunicationModeEnum(commMode),
		pokemon,
		*sb,
	)

	self.Conn.WriteToUDP(msg.SerializeMessage(), addr)

}

func main() {
	self := peer.MakePDFromLogin("hostW")
	defer self.Conn.Close()

	// at the start say that somebody can join you
	waitForMatch(self)
}
