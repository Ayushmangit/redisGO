package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Ayushmangit/redisGO/protocol"
	"github.com/Ayushmangit/redisGO/store"
)

// waitgroups use with go routines
type Command struct {
	OP    string
	Key   string
	Value string
}

func dispatch(s store.Storer, c Command) (*Response, error) {
	switch strings.ToUpper(c.OP) {
	case "GET":
		value, err := s.Get(c.Key)
		if err != nil {
			if errors.Is(err, store.ErrKeyNotFound) {
				return NewNullBulkString(), nil
			}
			return NewError(err.Error()), nil
		}
		return NewBulkString(value), nil

	case "SET":
		if err := s.Set(c.Key, c.Value, 0); err != nil {
			return NewError(err.Error()), nil
		}
		return NewSimpleString("OK"), nil

	case "INCR":
		n, err := s.Incr(c.Key)
		if err != nil {
			return NewError(err.Error()), nil
		}
		return NewInteger(int64(n)), nil

	default:
		return NewError(fmt.Sprintf("Invalid command: [%v]", c.OP)), nil
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

// TODO:
func ValueToCommand(value protocol.Value) (Command, error)
