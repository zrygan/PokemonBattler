package netio

import (
	"fmt"
)

// Verbose controls whether verbose logging is enabled globally.
// Can be set via command-line flags in main applications.
var Verbose bool = false

// LogOptions contains optional parameters for verbose logging.
// Used to display detailed information about network events and messages.
type LogOptions struct {
	Name string // Peer name
	Port string // Network port
	IP   string // IP address

	MessageParams any // Message parameters to display

	// MS (MessageSender) contains the sender's address
	MS string
}

// VerboseEventLog logs detailed information about network events when verbose mode is enabled.
// Takes a message string and optional LogOptions for additional context.
func VerboseEventLog(message string, opts *LogOptions) {
	if !Verbose {
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
