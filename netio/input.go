package netio

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ShowMenu(texts ...string) {
	for _, line := range texts {
		fmt.Printf("%s", line)
	}
}

// PRLine is Print-Read Line, prints instructions and takes input
func PRLine(instruction string) string {
	fmt.Println(instruction)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("$ ")
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// RLine is Read-only Line, takes input. Uses PRLine but with empty string as input
func RLine() string {
	return PRLine("")
}
