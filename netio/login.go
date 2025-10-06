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

func Login() (string, string, string) {
	fmt.Println("Welcome to PokeBattler")
	fmt.Println("(c) Zhean Ganituen /zrygan/, 2025")
	fmt.Println()

	fmt.Println("What is your trainer name?")
	name := ReadLine()

	fmt.Println("Choose your port?")
	port := ReadLine()
	if port == "" {
		port = "50000"
	}

	fmt.Println("Choose your IP?")
	ip := ReadLine()
	if ip == "" {
		ip = getLocalIP()
	}

	return name, port, ip
}
