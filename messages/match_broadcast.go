package messages

import "net"

func MakeJoiningMMB() Message {
	return Message{
		MessageType:   MMB_JOINING,
		MessageParams: nil, // Request has no Params
	}
}

func MakeHostingMMB(hostName string, udp *net.UDPAddr) Message {
	params := map[string]any{
		"hostName": hostName,
		"ip":       udp.IP,
		"port":     udp.Port,
	}

	return Message{
		MessageType:   MMB_HOSTING,
		MessageParams: &params,
	}
}
