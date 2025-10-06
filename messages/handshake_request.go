package messages

import "github.com/zrygan/pokemonbattler/peer"

// MakeHandshakeRequest creates a handshake request message from a peer descriptor.
// The message includes the peer's name, IP address, and port for identification.
func MakeHandshakeRequest(pd peer.PeerDescriptor) Message {
	param := map[string]any{
		"name": pd.Name,
		"ip":   pd.Addr.IP,
		"port": pd.Addr.Port,
	}
	return Message{
		MessageType:   HandshakeRequest,
		MessageParams: &param, // Request has no Params
	}
}
