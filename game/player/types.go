package player

import (
	"github.com/zrygan/pokemonbattler/peer"
	"github.com/zrygan/pokemonbattler/poke"
)

// Player represents a player in the Pokemon battle game.
type Player struct {
	Peer                   peer.PeerDescriptor // Network connection information
	PokemonStruct          poke.Pokemon        // Player's pokemon
	SpecialAttackUsesLeft  int                 // Special attack boosts remaining
	SpecialDefenseUsesLeft int                 // Special defense boosts remaining
}
