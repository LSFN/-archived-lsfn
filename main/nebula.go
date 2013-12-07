package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/LSFN/lsfn/nebula"
)

func main() {
	server := nebula.NewStarshipServer()
	go server.Listen()
	if server.Listening() {
		return
	} else {
		fmt.Println("Listening for new connections")
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
