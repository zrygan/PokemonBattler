package peer

import "github.com/zrygan/pokemonbattler/netio"

func InitPeer() PeerDescriptor {
	pd := MakePDBase(netio.Login())

	netio.VerboseEventLog(
		"A new JOINER connected.",
		&netio.LogOptions{
			Port: string(pd.Addr.Port),
			IP:   string(pd.Addr.IP),
		},
	)

	return pd
}
