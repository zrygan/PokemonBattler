package game

import (
	"fmt"
	"strconv"

	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
	"github.com/zrygan/pokemonbattler/poke"
	monsters "github.com/zrygan/pokemonbattler/poke/mons"
)

func Host_setCommMode() string {
	for {
		mode := netio.PRLine("Select a communication mode:\nP: peer-to-peer\nB: broadcast")

		switch mode {
		case P2P:
			return "P"
		case Broadcast:
			return "B"
		default:
			netio.ERLine("Invalid input. Please enter P or B.", false)
		}
	}
}

func PlayerSetUp(self peer.PeerDescriptor) Player {
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

	netio.ShowMenu(
		"Selected pokemon: "+poke.Name,
		"Special attack: "+strconv.Itoa(spatk),
		"Special defense: "+strconv.Itoa(spdef),
	)

	return Player{
		Peer:               self,
		PokemonStruct:      poke,
		SpecialDefenseUses: int8(spatk),
		SpecialAttackUses:  int8(spdef),
	}
}

// func Host_PBSetUp(
// 	seed int,
// 	cmode string, // always "P" or "B"
// ) Player {

// 	return Player{}
// }
