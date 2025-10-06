package messages

import (
	"github.com/zrygan/pokemonbattler/peer"
)

func MakeJoiningMMB() Message {
	return Message{
		MessageType:   MMB_JOINING,
		MessageParams: nil, // Request has no Params
	}
}

func MakeHostingMMB(pd peer.PeerDescriptor) Message {
	params := map[string]any{
		"name": pd.Name,
		"ip":   pd.Addr.IP.String(),
		"port": pd.Addr.Port,
	}

	return Message{
		MessageType:   MMB_HOSTING,
		MessageParams: &params,
	}
}
