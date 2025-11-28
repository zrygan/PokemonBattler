// Package game defines the core game structures and types for Pokemon Battler.
// It contains user, game state, pokemon, and communication mode definitions.
package game

import (
	"math/rand"

	"github.com/zrygan/pokemonbattler/game/player"
	"github.com/zrygan/pokemonbattler/peer"
)

// Game represents the state of a Pokemon battle.
// Contains the random seed and communication mode for the battle.
type Game struct {
	Host              *player.Player        // Host's peer descriptor and Pokemon
	Joiner            *player.Player        // Joiner's peer descriptor and Pokemon
	Spectators        []peer.PeerDescriptor // List of spectator peer descriptors
	Seed              int                   // Random seed for synchronized RNG
	RNG               *rand.Rand            // Seeded random number generator
	CommunicationMode string                // P2P (P) or broadcast (B) mode
	State             BattleState           // Current battle state
	CurrentTurn       string                // "host" or "joiner" - whose turn it is
	BattleLog         []string              // Log of all battle events
}

const (
	P2P       string = "P" // Direct peer-to-peer communication
	Broadcast string = "B" // Broadcast to local network
)

// NewGame creates a new Game instance with the given seed.
func NewGame(seed int, commMode string) *Game {
	return &Game{
		Seed:              seed,
		RNG:               rand.New(rand.NewSource(int64(seed))),
		CommunicationMode: commMode,
		State:             StateSetup,
		CurrentTurn:       "host", // Host always goes first
		Spectators:        make([]peer.PeerDescriptor, 0),
	}
}

// AddSpectator adds a spectator to the game.
func (g *Game) AddSpectator(spectator peer.PeerDescriptor) {
	g.Spectators = append(g.Spectators, spectator)
}

// BroadcastToSpectators sends a message to all spectators.
func (g *Game) BroadcastToSpectators(messageData []byte) {
	if g.Host == nil || g.Host.Peer.Conn == nil {
		return // No host connection available
	}

	for _, spectator := range g.Spectators {
		g.Host.Peer.Conn.WriteToUDP(messageData, spectator.Addr)
	}
}

// SendMessage sends a message according to the communication mode.
// In P2P mode, only sends directly to target.
// In Broadcast mode, sends to target and all spectators.
func (g *Game) SendMessage(messageData []byte, sender peer.PeerDescriptor, target peer.PeerDescriptor) {
	switch g.CommunicationMode {
	case "P": // P2P mode - direct send only
		sender.Conn.WriteToUDP(messageData, target.Addr)
	case "B": // Broadcast mode - send to target and broadcast to spectators
		sender.Conn.WriteToUDP(messageData, target.Addr)
		g.BroadcastToSpectators(messageData)
	default: // Default to P2P behavior
		sender.Conn.WriteToUDP(messageData, target.Addr)
	}
}
