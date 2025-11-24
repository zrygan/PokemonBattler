// Package messages provides message types and utilities for Pokemon Battler network communication.
// It defines the protocol messages used for peer discovery, handshaking, and battle setup.
package messages

// Message represents a network message with a type and optional parameters.
// MessageParams contains key-value pairs specific to each message type.
type Message struct {
	MessageType   string          // Type identifier for the message
	MessageParams *map[string]any // Optional parameters specific to message type
}

// Message type constants for the Pokemon Battler protocol.
const (
	// Match making broadcast (MMB) message types
	MMB_JOINING = "FINDING_HOST" // Sent by joiners to discover hosts
	MMB_HOSTING = "I_AM_HOSTING" // Sent by hosts in response to discovery

	// Handshake message types
	HandshakeRequest  = "HANDSHAKE_REQUEST"  // Joiner requests to connect to host
	HandshakeResponse = "HANDSHAKE_RESPONSE" // Host accepts connection and provides seed
	HandshakeRejected = "HANDSHAKE_REJECTED" // Host rejects connection request
	SpectatorRequest  = "SPECTATOR_REQUEST"  // Spectator requests to observe battle

	// BattleSetup message types
	BattleSetup = "BATTLE_SETUP" // Battle configuration message

	// (stage) GameSetup message types
	GS_COMMMODE = "COMM_MODE"

	// Battle turn message types
	AttackAnnounce     = "ATTACK_ANNOUNCE"     // Attacker announces chosen move
	DefenseAnnounce    = "DEFENSE_ANNOUNCE"    // Defender acknowledges attack
	CalculationReport  = "CALCULATION_REPORT"  // Player reports damage calculation
	CalculationConfirm = "CALCULATION_CONFIRM" // Player confirms matching calculation
	ResolutionRequest  = "RESOLUTION_REQUEST"  // Request to resolve calculation discrepancy
	GameOver           = "GAME_OVER"           // Battle ends, declare winner

	// Chat message types
	ChatMessage = "CHAT_MESSAGE" // Chat or sticker message

	// Reliability layer message types
	ACK = "ACK" // Acknowledgement message
)
