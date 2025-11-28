package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zrygan/pokemonbattler/game/player"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
	"github.com/zrygan/pokemonbattler/poke"
	monsters "github.com/zrygan/pokemonbattler/poke/mons"
)

func Host_setCMode(host peer.PeerDescriptor, join peer.PeerDescriptor) string {
	for {
		mode := strings.ToUpper(netio.PRLine("Select a communication mode:\nP: peer-to-peer\nB: broadcast"))

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

	// Get trainer name for profiles
	trainerName := netio.PRLine("Enter your trainer name: ")
	teamManager := poke.NewTeamManager(trainerName)

	// Option to view existing profiles
	fmt.Println("\nWould you like to view your saved Pokemon profiles? (y/n)")
	viewProfiles := netio.PRLine("> ")
	if strings.ToLower(strings.TrimSpace(viewProfiles)) == "y" {
		teamManager.ListProfiles()
	}

	// get pokemon name
	var pokemonStruct poke.Pokemon
	for {
		pokeName := netio.PRLine("Select a pokemon: ")
		// Try exact match first, then case-insensitive
		pokemonStruct, ok = monsters.MONSTERS[pokeName]
		if !ok {
			// Try case-insensitive search
			for key, mon := range monsters.MONSTERS {
				if strings.EqualFold(key, pokeName) {
					pokemonStruct = mon
					ok = true
					break
				}
			}
		}
		if !ok {
			netio.ERLine("Invalid pokemon. Please put a valid pokemon name", false)
		} else {
			break
		}
	}

	// Customize Pokemon (nickname & personality)
	profile, err := teamManager.CustomizePokemon(&pokemonStruct)
	if err != nil {
		fmt.Printf("Warning: Could not customize Pokemon: %v\n", err)
		profile = poke.NewPokemonProfile(pokemonStruct.Name)
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
		Peer:                   self,
		PokemonStruct:          pokemonStruct,
		SpecialAttackUsesLeft:  spatk,
		SpecialDefenseUsesLeft: spdef,
		Profile:                profile,
	}
}

func BattleSetup(self player.Player, other peer.PeerDescriptor, cmode string, spectators []peer.PeerDescriptor) player.Player {
	// Send BATTLE_SETUP
	msg := messages.MakeBattleSetup(
		self,
		cmode,
		self.PokemonStruct.Name,
		int8(self.SpecialAttackUsesLeft),
		int8(self.SpecialDefenseUsesLeft),
	)

	msgBytes := msg.SerializeMessage()

	// Send battle setup according to communication mode
	switch cmode {
	case "P": // P2P mode - direct send only
		self.Peer.Conn.WriteToUDP(msgBytes, other.Addr)
		// In P2P mode, explicitly broadcast to spectators
		for _, spectator := range spectators {
			self.Peer.Conn.WriteToUDP(msgBytes, spectator.Addr)
		}
	case "B": // Broadcast mode - send to target and broadcast to network
		self.Peer.Conn.WriteToUDP(msgBytes, other.Addr)
		// In broadcast mode, send to all spectators as part of network broadcast
		for _, spectator := range spectators {
			self.Peer.Conn.WriteToUDP(msgBytes, spectator.Addr)
		}
		// Note: In a full implementation, this would use actual network broadcast
	default: // Default to P2P behavior
		self.Peer.Conn.WriteToUDP(msgBytes, other.Addr)
		for _, spectator := range spectators {
			self.Peer.Conn.WriteToUDP(msgBytes, spectator.Addr)
		}
	}

	netio.VerboseEventLog(
		"PokeProtocol: Peer sent BATTLE_SETUP message to '"+other.Name+"'",
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
				"PokeProtocol: Peer received BATTLE_SETUP from '"+other.Addr.String()+"'",
				&netio.LogOptions{
					MessageParams: res.MessageParams,
				},
			)

			// Relay joiner's BATTLE_SETUP to spectators (host only)
			if len(spectators) > 0 {
				msgBytes := res.SerializeMessage()
				for _, spec := range spectators {
					self.Peer.Conn.WriteToUDP(msgBytes, spec.Addr)
				}
				netio.VerboseEventLog(
					"Relayed joiner's BATTLE_SETUP to "+strconv.Itoa(len(spectators))+" spectator(s)",
					nil,
				)
			}

			// Parse opponent's Pokemon data from the message
			params := *res.MessageParams
			opponentPokemonName := params["pokemon_name"].(string)
			specialAttackUses := params["special_attack_uses"].(int)
			specialDefenseUses := params["special_defense_uses"].(int)

			// Load opponent's Pokemon
			opponentPokemon, ok := monsters.MONSTERS[opponentPokemonName]
			if !ok {
				// Try case-insensitive
				for key, mon := range monsters.MONSTERS {
					if strings.EqualFold(key, opponentPokemonName) {
						opponentPokemon = mon
						ok = true
						break
					}
				}
			}

			if !ok {
				panic(fmt.Sprintf("Unknown Pokemon: %s", opponentPokemonName))
			}

			// Create opponent player
			opponentPlayer := player.Player{
				Peer:                   other,
				PokemonStruct:          opponentPokemon,
				SpecialAttackUsesLeft:  specialAttackUses,
				SpecialDefenseUsesLeft: specialDefenseUses,
			}

			return opponentPlayer
		}
	}
}

// func Host_PBSetUp(
// 	seed int,
// 	cmode string, // always "P" or "B"
// ) Player {

// 	return Player{}
// }
