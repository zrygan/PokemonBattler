// Package netio provides network I/O utilities and user interaction functions
// for the Pokemon Battler application.
package netio

import (
	"fmt"
	"net"
)

// getLocalIP determines the local IP address by connecting to a public DNS server.
// Returns "127.0.0.1" if unable to determine the local IP.
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// Login handles user login and network configuration.
// Returns name, ip, port, and isLocal flag for peer descriptor creation.
// Special handling for "hostW" userType to use discovery port 50000.
func Login(userType string) (string, string, string, bool) {
	fmt.Println("Welcome to PokeBattler")
	fmt.Println("(c) Zhean Ganituen /zrygan/, 2025")
	fmt.Println()

	name := PRLine("What is your trainer name?")

	port := PRLine("Choose your port?")
	if port == "" {
		port = "0" // make OS assign a random port
	}
	if userType == "hostW" {
		// hostW refers to a host without a game
		// discovery port, put them on discovery port for joiners to see them
		port = "50000"
	}

	ip := PRLine("Choose your IP?")
	if ip == "" {
		ip = getLocalIP()
	}

	// false here is the remote checker in PD constructor functions
	return name, ip, port, true
}
