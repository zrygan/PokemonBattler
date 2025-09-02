package main

import (
	"fmt"
	"os"

	"github.com/zrygan/pokemonbattler/helper"
)

func main() {
	_, conn := helper.HostTo(os.Args)
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, _ := conn.ReadFromUDP(buf)
		fmt.Printf("Got from %s: %s\n", remoteAddr, string(buf[:n]))

		conn.WriteToUDP([]byte("hello! this is my response!"), remoteAddr)
	}
}
