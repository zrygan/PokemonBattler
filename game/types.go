package game

import "github.com/zrygan/pokemonbattler/peer"

type User struct {
	PD            peer.PeerDescriptor
	GameStruct    *Game
	PokemonStruct *Pokemon
	UserType      UserTypeEnum
}

type Game struct {
	Seed              int
	CommunicationMode CommunicationModeEnum
}

type Pokemon struct {
	PokemonName      string // FIXME: turn this into a hashmap maybe?
	StatBoostsStruct StatBoosts
}

type StatBoosts struct {
	SpecialAttackUses  int8
	SpecialDefenseUses int8
}

type UserTypeEnum int

const (
	HostUser UserTypeEnum = iota
	JoinerUser
)

type CommunicationModeEnum int

const (
	P2P CommunicationModeEnum = iota
	Broadcast
)

const DiscoveryPort = 50000
