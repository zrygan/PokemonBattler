package helper

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

func ReadLine() string {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n$ ")
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
