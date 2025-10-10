// Package game defines the core game structures and types for Pokemon Battler.
// It contains user, game state, pokemon, and communication mode definitions.
package game

import (
	"github.com/zrygan/pokemonbattler/peer"
	"github.com/zrygan/pokemonbattler/poke"
)

// Game represents the state of a Pokemon battle.
// Contains the random seed and communication mode for the battle.
type Game struct {
	Host              *Player               // Host's peer descriptor
	Joiner            *Player               // Joiner's peer descriptor
	Spectators        []peer.PeerDescriptor // List of spectator peer descriptors
	Seed              int                   // Random seed for synchronized RNG
	CommunicationMode string                // P2P (P) or broadcast (B) mode
}

// Player represents a player in the Pokemon battle game.
type Player struct {
	Peer               peer.PeerDescriptor // Network connection information
	PokemonStruct      poke.Pokemon        // Player's pokemon
	SpecialAttackUses  int8                // Special attack boost allocation
	SpecialDefenseUses int8                // Special defense boost allocation
}

const (
	P2P       string = "P" // Direct peer-to-peer communication
	Broadcast string = "B" // Broadcast to local network
)
