// Package game defines the core game structures and types for Pokemon Battler.
// It contains user, game state, pokemon, and communication mode definitions.
package game

import (
	"github.com/zrygan/pokemonbattler/game/player"
	"github.com/zrygan/pokemonbattler/peer"
)

// Game represents the state of a Pokemon battle.
// Contains the random seed and communication mode for the battle.
type Game struct {
	Host              *player.Player        // Host's peer descriptor
	Joiner            *player.Player        // Joiner's peer descriptor
	Spectators        []peer.PeerDescriptor // List of spectator peer descriptors
	Seed              int                   // Random seed for synchronized RNG
	CommunicationMode string                // P2P (P) or broadcast (B) mode
}

const (
	P2P       string = "P" // Direct peer-to-peer communication
	Broadcast string = "B" // Broadcast to local network
)
