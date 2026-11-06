package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Ayushmangit/redisGO/store"
	"github.com/Ayushmangit/redisGO/store/kv"
)

func main() {
	s := kv.NewStore(1000, 1*time.Hour)

	if len(os.Args) < 2 {
		fmt.Println("Usage: redisGO [server|cli]")
		return
	}

	var err error

	switch os.Args[1] {
	case "server":
		err = runServer(s)
	case "cli":
		err = runCLI(s)
	default:
		fmt.Println("Unknown mode. Use server or cli.")
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCLI(s store.Storer) error {
	exec := NewExecutor(s, 1000)
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
			return nil
		}

		res, err := HandleInput(exec, input)
		if err != nil {
			fmt.Printf("(error) %v\n", err)
		}
		printResponse(res)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "input error: %v\n", err)
	}
	return nil
}

func runServer(s store.Storer) error {
	exec := NewExecutor(s, 1000)
	defer exec.Stop()

	ln, err := StartTCPServer(":1882")
	if err != nil {
		return err
	}

	defer ln.Close()

	return AcceptLoop(ln, exec)
}

func printResponse(res *Response) {
	if res == nil {
		return
	}

	switch res.responseType {
	case SimpleString:
		fmt.Println(res.val)

	case ErrorResponse:
		fmt.Printf("(error) %s\n", res.val)

	case Integer:
		fmt.Printf("(integer) %d\n", res.intVal)

	case BulkString:
		fmt.Printf("\"%s\"\n", res.val)

	case NullBulkString:
		fmt.Println("(nil)")

	}
}
