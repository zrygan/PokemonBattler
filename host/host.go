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

func broadcastAsJoinable(pd peer.PeerDescriptor) {
	buf := make([]byte, 1024)

	for {
		n, rem, err := pd.Conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		msg := strings.TrimSpace(string(buf[:n]))

		// send a message if somebody wants to join
		if msg == messages.MMB_JOINING {
			netio.VerboseEventLog(
				"Found a JOINER, received a "+messages.MMB_JOINING+" message",
				nil,
			)

			res := messages.MakeHostingMMB(pd)

			netio.VerboseEventLog(
				"Found a JOINER, sent a "+messages.MMB_HOSTING+" message",
				nil,
			)

			_, _ = pd.Conn.WriteToUDP(
				res.SerializeMessage(),
				rem,
			)
		}
	}
}

func handshake(pd peer.PeerDescriptor, addr *net.UDPAddr) {
	// send back a HandshakeResponse
	msg := messages.MakeHandshakeResponse()
	pd.Conn.WriteToUDP(msg.SerializeMessage(), addr)

	netio.VerboseEventLog(
		"A message was SENT",
		&netio.LogOptions{
			MT: msg.MessageType,
			MP: fmt.Sprint(*msg.MessageParams),
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
	pd.Conn.WriteToUDP([]byte(strconv.Itoa(commMode)), addr)

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

	pd.Conn.WriteToUDP(msg.SerializeMessage(), addr)

}

func main() {
	self := peer.MakePDFromLogin("host")
	defer self.Conn.Close()

	// at the start say that somebody can join you
	broadcastAsJoinable(self)

	buf := make([]byte, 1024)
	for {
		n, addr, err := self.Conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		bc := buf[:n]
		msg := *messages.DeserializeMessage(bc)

		netio.VerboseEventLog(
			"A message was RECEIVED",
			&netio.LogOptions{
				MT: msg.MessageType,
				MP: fmt.Sprint(msg.MessageParams),
				MS: addr.String(),
			},
		)
	}
}
