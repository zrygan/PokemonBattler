package helper

import "fmt"

const verbose bool = true

type LogOptions struct {
	Port string
	IP   string
}

func VerboseEventLog(
	message string,
	opts *LogOptions,
) {
	if verbose {
		fmt.Println("🔴 LOG :: ", message)

		if opts != nil {
			if opts.Port != "" {
				fmt.Printf("\t> Port: %s\n", opts.Port)
			}
			if opts.IP != "" {
				fmt.Printf("\t> Addr: %s\n", opts.IP)
			}
		}
	}
}
