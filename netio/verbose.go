package netio

import (
	"fmt"
)

const verbose bool = true

type LogOptions struct {
	// Port
	Port string

	// IP
	IP string

	// MessageSent
	MessageString string

	// MessageType
	MT string

	// MessageParameters
	MP string

	// MessageSender (address)
	MS string
}

func VerboseEventLog(
	message string,
	opts *LogOptions,
) {
	if verbose {
		fmt.Print("🔴 LOG :: ", message)

		if opts != nil {
			fmt.Println()
			if opts.Port != "" {
				fmt.Printf("\t> Port: %s\n", opts.Port)
			}
			if opts.IP != "" {
				fmt.Printf("\t> Addr: %s\n", opts.IP)
			}
			if opts.MT != "" {
				fmt.Printf("\t> MessageType: %s\n", opts.MT)
			}
			if opts.MP != "" {
				fmt.Printf("\t> MessageParams: %s\n", opts.MP)
			}
			if opts.MS != "" {
				fmt.Printf("\t> MessageSender: %s\n", opts.MS)
			}
		}
	}

	fmt.Println()
}
