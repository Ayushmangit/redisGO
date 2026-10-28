package main

import (
	"fmt"
	"sync"

	"github.com/Ayushmangit/redisGO/store"
)

// waitgroups use with go routines
type Command struct {
	OP    string
	Key   string
	Value string
}

func dispatch(s store.Storer, c Command) error {
	switch c.OP {
	case "GET":
		s.Get(c.Key)
	case "SET":
		return s.Set(c.Key, c.Value, 0)
	case "INCR":
		_, err := s.Incr(c.Key)
		return err
	default:
		return fmt.Errorf("Invalid command :[%v]", c.OP)
	}
	return nil
}

func RestoreOnBoot(s store.Storer, cmds []Command) {
	var wg sync.WaitGroup               // mutable variable
	errs := make(chan error, len(cmds)) // no blocking till len(cmds)
	for _, command := range cmds {
		wg.Go(func() {
			errs <- dispatch(s, command)
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
