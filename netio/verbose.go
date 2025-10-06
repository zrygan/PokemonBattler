package netio

import (
	"fmt"
)

const verbose bool = true

type LogOptions struct {
	Name string
	Port string
	IP   string

	MessageParams any

	// MessageSender (address)
	MS string
}

func VerboseEventLog(message string, opts *LogOptions) {
	if !verbose {
		return
	}

	fmt.Print("🔴 LOG :: ", message)

	if opts == nil {
		fmt.Printf("\n\n")
		return
	}

	fmt.Println()

	if opts.Port != "" {
		fmt.Printf("\t> Src Port: %s\n", opts.Port)
	}
	if opts.IP != "" {
		fmt.Printf("\t> Src IP: %s\n", opts.IP)
	}

	// Handle MessageParams cleanly
	if opts.MessageParams != nil {
		fmt.Printf("\t> MessageParams:\n")

		var m map[string]any
		switch t := opts.MessageParams.(type) {
		case map[string]any:
			m = t
		case *map[string]any:
			m = *t
		default:
			fmt.Printf("\t\t<invalid type: %T>\n", t)
			return
		}

		for k, v := range m {
			fmt.Printf("\t\t%s: %v\n", k, v)
		}
	}

	if opts.MS != "" {
		fmt.Printf("\t> MessageSender: %s\n", opts.MS)
	}

	fmt.Println()
}
