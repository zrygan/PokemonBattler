package netio

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ShowMenu displays a menu or instructions to the user.
// Takes variable number of string arguments to display.
func ShowMenu(texts ...string) {
	for _, line := range texts {
		fmt.Printf("%s", line)
	}
}

// PRLine (Print-Read Line) displays an instruction and reads user input.
// Returns the trimmed input string.
func PRLine(instruction string) string {
	fmt.Println(instruction)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("$ ")
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// RLine (Read Line) reads user input without displaying an instruction.
// Uses PRLine with an empty instruction string.
func RLine() string {
	return PRLine("")
}

// StartInputListener starts a goroutine that continuously reads user input
// and sends it to the returned channel. This allows non-blocking input reading.
func StartInputListener() chan string {
	inputChan := make(chan string)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				close(inputChan)
				return
			}
			inputChan <- strings.TrimSpace(line)
		}
	}()
	return inputChan
}

// ERLine (ERror Line) displays an error message, terminates when required
// by shouldStop.
func ERLine(errorcode string, shouldStop bool) {
	err := "Error: " + errorcode
	if shouldStop {
		panic(err)
	}

	fmt.Println(err)
}
