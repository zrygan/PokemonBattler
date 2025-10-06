package messages

type Message struct {
	MessageType   string
	MessageParams *map[string]any
}

const (
	// match making broadcast (MMB)
	MMB_JOINING = "FINDING_HOST"
	MMB_HOSTING = "I_AM_HOSTING"

	// handshakes
	HandshakeRequest  = "HANDSHAKE_REQUEST"
	HandshakeResponse = "HANDSHAKE_RESPONSE"
	BattleSetup       = "BATTLE_SETUP"
)
