package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Ayushmangit/redisGO/store/kv"
)

func main() {
	store := kv.NewStore(1000, 0)
	exec := NewExecutor(store, 1000)
	defer exec.Stop()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("redis> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		switch strings.ToUpper(input) {
		case "EXIT", "QUIT":
			return
		}

		command, err := ParseCommand(input)
		if err != nil {
			fmt.Printf("(error) PARSING: %v\n", err)
			continue
		}

		res, err := exec.Execute(command)
		if err != nil {
			fmt.Printf("(error) EXECUTE: %v\n", err)
			continue
		}

		fmt.Println(res)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "input error: %v\n", err)
	}
}
