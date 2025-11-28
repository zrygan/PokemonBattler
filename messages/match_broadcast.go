package messages

import (
	"github.com/zrygan/pokemonbattler/peer"
)

// MakeJoiningMMB creates a match-making broadcast message for joiners.
// This message is broadcast to discover available hosts on the network.
func MakeJoiningMMB() Message {
	return Message{
		MessageType:   MMB_JOINING,
		MessageParams: nil, // Request has no Params
	}
}

// MakeHostingMMB creates a match-making broadcast response for hosts.
// The message includes the host's connection details for joiners to connect.
func MakeHostingMMB(pd peer.PeerDescriptor) Message {
	params := map[string]any{
		"name": pd.Name,
		"ip":   pd.Addr.IP.String(), // Convert IP to string for proper serialization
		"port": pd.Addr.Port,
	}

	return Message{
		MessageType:   MMB_HOSTING,
		MessageParams: &params,
	}
}
