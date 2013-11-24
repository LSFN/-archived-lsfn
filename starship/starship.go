package main

import (
	"bufio"
	"fmt"
	"os"

	"starship"
)

func main() {
	var client *NebulaClient = new(NebulaClient)
	if client.Join("localhost", 39460) {
		fmt.Println("Join successful")
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
