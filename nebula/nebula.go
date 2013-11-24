package main

import (
	"bufio"
	"fmt"
	"os"

	"lsfn/nebula"
)

func main() {
	var server *StarshipServer = new(StarshipServer)
	go server.Listen()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
