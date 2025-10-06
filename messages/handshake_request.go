package messages

import "github.com/zrygan/pokemonbattler/peer"

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
