package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/LSFN/lsfn/starship"
)

func main() {
	var client *starship.NebulaClient = new(starship.NebulaClient)
	if client.Join("localhost", 39461) {
		fmt.Println("Join successful")
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
