package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
)

func broadcastAsJoinable(hostName string, udp *net.UDPAddr) {

}

// The parameter arguments is
func hostTo(arguments []string) (*net.UDPAddr, *net.UDPConn) {
	if len(arguments) != 3 {
		panic("arguments must be of the form: <port> <ip>")
	}

	port, err := strconv.Atoi(arguments[1])
	if err != nil {
		panic(err)
	}

	ip := net.ParseIP(arguments[2])
	if ip == nil {
		panic("invalid IP address")
	}

	addr := &net.UDPAddr{
		Port: port,
		IP:   ip,
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	netio.VerboseEventLog(
		"A new HOST connected.",
		&netio.LogOptions{
			Port: arguments[1],
			IP:   arguments[2],
		},
	)

	return addr, conn
}

func main() {
	// a general variable for Message struct
	var msg messages.Message

	_, conn := hostTo(os.Args)
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		bc := buf[:n]
		msg = *messages.DeserializeMessage(bc)

		netio.VerboseEventLog(
			"A message was RECEIVED",
			&netio.LogOptions{
				MT: msg.MessageType,
				MP: fmt.Sprint(msg.MessageParams),
				MS: addr.String(),
			},
		)

		if msg.MessageType == messages.HandshakeRequest {
			// then send back a HandshakeResponse
			msg = messages.MakeHandshakeResponse()
			conn.WriteToUDP(msg.SerializeMessage(), addr)

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
				line := netio.ReadLine()
				commMode, err = strconv.Atoi(line)
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
			conn.WriteToUDP([]byte(strconv.Itoa(commMode)), addr)

			//FIXME: do some checking here if the pokemon is valid
			pokemon := netio.ReadLine()

			netio.ShowMenu(
				"Allocate stat boosts for your special attack and defense!",
				"\nFormat is 2 integers <attack_boost> <defense_boost>",
				"\nThe two integers must sum to 10",
			)

			sb := new(game.StatBoosts)
			for {
				line := strings.Split(netio.ReadLine(), " ")

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

			conn.WriteToUDP(msg.SerializeMessage(), addr)
		}

	}
}
