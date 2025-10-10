package game

import (
	"fmt"
	"strconv"

	"github.com/zrygan/pokemonbattler/game/player"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
	"github.com/zrygan/pokemonbattler/poke"
	monsters "github.com/zrygan/pokemonbattler/poke/mons"
)

func Host_setCMode(host peer.PeerDescriptor, join peer.PeerDescriptor) string {
	for {
		mode := netio.PRLine("Select a communication mode:\nP: peer-to-peer\nB: broadcast")

		switch mode {
		case P2P:
			fallthrough
		case Broadcast:
			msg := messages.GS_MakeCMode(mode)
			host.Conn.WriteToUDP(msg.SerializeMessage(), join.Addr)
			return mode
		default:
			netio.ERLine("Invalid input. Please enter P or B.", false)
		}
	}
}

func Joiner_getCMode(p peer.PeerDescriptor) string {
	buf := make([]byte, 1000)

	for {
		n, _, err := p.Conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		msg := messages.DeserializeMessage(buf[:n])

		if msg.MessageType == messages.GS_COMMMODE {
			return (*msg.MessageParams)["cmode"].(string)
		}
	}
}

func PlayerSetUp(self peer.PeerDescriptor) player.Player {
	var err error
	var ok bool

	// get pokemon name
	var poke poke.Pokemon
	for {
		pokeName := netio.PRLine("Select a pokemon: ")
		poke, ok = monsters.MONSTERS[pokeName]
		if !ok {
			netio.ERLine("Invalid pokemon. Please put a valid pokemon name", false)
		} else {
			break
		}
	}

	// allocate spatk and spdef
	var spdef int
	var spatk int
	for {
		fmt.Println("You can allocate 10 points to your special attack and special defense, use it wisely.")
		spatk, err = strconv.Atoi(netio.PRLine("Special attack allocation: "))
		if err != nil || spatk > 10 || spatk < 0 {
			netio.ERLine("Invalid input. Should be a number from 1--10", false)
			continue
		}

		spdef, err = strconv.Atoi(netio.PRLine("Special defense allocation: "))
		if err != nil || spdef > 10 || spdef < 0 {
			netio.ERLine("Invalid input. Should be a number from 1--10", false)
			continue
		}

		if spatk+spdef > 10 {
			netio.ERLine("Invalid inputs. Sum should be at most 10", false)
		} else {
			break
		}
	}

	return player.Player{
		Peer:               self,
		PokemonStruct:      poke,
		SpecialDefenseUses: int8(spatk),
		SpecialAttackUses:  int8(spdef),
	}
}

func BattleSetup(self player.Player, other peer.PeerDescriptor, cmode string) {
	// Send BATTLE_SETUP
	msg := messages.MakeBattleSetup(
		self,
		cmode,
		self.PokemonStruct.Name,
		self.SpecialAttackUses,
		self.SpecialDefenseUses,
	)

	self.Peer.Conn.WriteToUDP(msg.SerializeMessage(), other.Addr)

	netio.VerboseEventLog(
		"Sent "+messages.BattleSetup+" message to"+other.Name,
		&netio.LogOptions{
			MessageParams: msg.MessageParams,
		},
	)

	// Wait for received BATTLE_SETUP
	buf := make([]byte, 1000)
	for {
		n, addr, err := self.Peer.Conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		res := messages.DeserializeMessage(buf[:n])

		if res.MessageType == messages.BattleSetup &&
			addr.IP.Equal(other.Addr.IP) && // ensure we got it from expected user
			addr.Port == other.Addr.Port {
			netio.VerboseEventLog(
				"Received a "+messages.BattleSetup+" from "+other.Addr.String(),
				&netio.LogOptions{
					MessageParams: res.MessageParams,
				},
			)
		}
	}
}

// func Host_PBSetUp(
// 	seed int,
// 	cmode string, // always "P" or "B"
// ) Player {

// 	return Player{}
// }
