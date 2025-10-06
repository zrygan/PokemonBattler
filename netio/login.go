package netio

import (
	"fmt"
	"net"
)

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func Login(userType string) (string, string, string, bool) {
	fmt.Println("Welcome to PokeBattler")
	fmt.Println("(c) Zhean Ganituen /zrygan/, 2025")
	fmt.Println()

	name := PRLine("What is your trainer name?")

	port := PRLine("Choose your port?")
	if port == "" {
		port = "0" // make OS assign a random port
	}
	if userType == "host" {
		port = "50000" // discovery port
	}

	ip := PRLine("Choose your IP?")
	if ip == "" {
		ip = getLocalIP()
	}

	// false here is the remote checker in PD constructor functions
	return name, ip, port, true
}
