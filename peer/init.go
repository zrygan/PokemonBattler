package peer

import (
	"strconv"

	"github.com/zrygan/pokemonbattler/netio"
)

func Login() PeerDescriptor {
	pd := MakePDBase(netio.Login())

	netio.VerboseEventLog(
		"A new JOINER connected.",
		&netio.LogOptions{
			Port: strconv.Itoa(pd.Addr.Port),
			IP:   string(pd.Addr.IP),
		},
	)

	return pd
}
