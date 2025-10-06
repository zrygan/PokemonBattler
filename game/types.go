// Package game defines the core game structures and types for Pokemon Battler.
// It contains user, game state, pokemon, and communication mode definitions.
package game

import "github.com/zrygan/pokemonbattler/peer"

// User represents a player in the Pokemon battle game.
// Contains peer connection info, game state, pokemon, and user type.
type User struct {
	PD            peer.PeerDescriptor // Network connection information
	GameStruct    *Game               // Current game state
	PokemonStruct *Pokemon            // Player's pokemon
	UserType      UserTypeEnum        // Whether host or joiner
}

// Game represents the state of a Pokemon battle.
// Contains the random seed and communication mode for the battle.
type Game struct {
	Seed              int                   // Random seed for synchronized RNG
	CommunicationMode CommunicationModeEnum // P2P or broadcast mode
}

// Pokemon represents a player's pokemon with stats and boosts.
type Pokemon struct {
	PokemonName      string     // Name of the pokemon (TODO: convert to hashmap)
	StatBoostsStruct StatBoosts // Applied stat modifications
}

// StatBoosts represents special attack and defense modifications.
// The sum of both boosts must equal 10.
type StatBoosts struct {
	SpecialAttackUses  int8 // Special attack boost allocation
	SpecialDefenseUses int8 // Special defense boost allocation
}

// UserTypeEnum distinguishes between host and joiner players.
type UserTypeEnum int

const (
	HostUser   UserTypeEnum = iota // Player hosting the game
	JoinerUser                     // Player joining the game
)

// CommunicationModeEnum defines how game communication is handled.
type CommunicationModeEnum int

const (
	P2P       CommunicationModeEnum = iota // Direct peer-to-peer communication
	Broadcast                              // Broadcast to local network
)
