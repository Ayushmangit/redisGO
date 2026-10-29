package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Ayushmangit/redisGO/store"
)

// waitgroups use with go routines
type Command struct {
	OP    string
	Key   string
	Value string
}

func dispatch(s store.Storer, c Command) (string, error) {
	switch strings.ToUpper(c.OP) {
	case "GET":
		return s.Get(c.Key)
	case "SET":
		if err := s.Set(c.Key, c.Value, 0); err != nil {
			return "", err
		}
		return "OK", nil
	case "INCR":
		n, err := s.Incr(c.Key)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(n), nil

	default:
		return "", fmt.Errorf("Invalid command :[%v]", c.OP)
	}
}

func RestoreOnBoot(s store.Storer, cmds []Command) {
	var wg sync.WaitGroup               // mutable variable
	errs := make(chan error, len(cmds)) // no blocking till len(cmds)
	for _, command := range cmds {
		wg.Go(func() {
			_, err := dispatch(s, command)
			errs <- err
		})
	}
	go func() {
		wg.Wait()
		close(errs)
	}()
	// log the errors we get from a channel
	for err := range errs {
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
