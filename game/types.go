package game

import "net"

type User struct {
	Conn          *net.UDPConn
	Addr          *net.UDPAddr
	GameStruct    *Game
	PokemonStruct *Pokemon
	UserType      UserTypeEnum
}

type Game struct {
	IP                string
	Port              string
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
