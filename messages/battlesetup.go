package messages

import (
	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/peer"
)

// MakeBattleSetup creates a battle setup message with game configuration.
func MakeBattleSetup(
	host,
	joiner peer.PeerDescriptor,
	spectators []peer.PeerDescriptor,
	seed int,
	commMode string, // ensure, only "P" or "B"
) game.Game {
	// set up communication mode
	return game.Game{}
}
