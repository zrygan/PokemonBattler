package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/helper"
	"github.com/zrygan/pokemonbattler/messages"
)

func main() {
	// a general variable for Message struct
	var msg messages.Message

	_, conn := helper.HostTo(os.Args)
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		bc := buf[:n]
		msg = *messages.DeserializeMessage(bc)

		helper.VerboseEventLog(
			"A message was RECEIVED",
			&helper.LogOptions{
				MT: msg.MessageType,
				MP: fmt.Sprint(msg.MessageParams),
				MS: addr.String(),
			},
		)

		if msg.MessageType == messages.HandshakeRequest {
			// then send back a HandshakeResponse
			msg = messages.MakeHandshakeResponse()
			conn.WriteToUDP(msg.SerializeMessage(), addr)

			helper.VerboseEventLog(
				"A message was SENT",
				&helper.LogOptions{
					MT: msg.MessageType,
					MP: fmt.Sprint(*msg.MessageParams),
				},
			)

			// create a BattleSetup once done
			helper.ShowMenu(
				"Commmunication Mode:",
				"\n\t(0) P2P: Game is directly between players",
				"\n\t(1) BROADCAST: Game is announced to local network.",
			)

			var commMode int
			for {
				line := helper.ReadLine()
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

			helper.ShowMenu(
				"Enter your chosen Pokemon",
			)

			//FIXME: do some checking here if the pokemon is valid
			pokemon := helper.ReadLine()

			helper.ShowMenu(
				"Allocate stat boosts for your special attack and defense!",
				"\nFormat is 2 integers <attack_boost> <defense_boost>",
				"\nThe two integers must sum to 10",
			)

			sb := new(game.StatBoosts)
			for {
				line := strings.Split(helper.ReadLine(), " ")

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
