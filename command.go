package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Ayushmangit/redisGO/protocol"
	"github.com/Ayushmangit/redisGO/store"
)

var (
	ErrInvalidCommandLength = errors.New("invalid command length")
	ErrUnknownCommand       = errors.New("unknown command error")
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
func ValueToCommand(value protocol.Value) (Command, error) {
	if value.Type != protocol.Array {
		return Command{}, fmt.Errorf("invalid value needs an array")
	}
	if len(value.Array) == 0 {
		return Command{}, fmt.Errorf("invalid length of the array")
	}

	for _, elm := range value.Array {
		if elm.Type != protocol.BulkString || elm.Num == -1 {
			return Command{}, fmt.Errorf("invalid command entry")
		}
	}
	args := make([]string, 0)
	for _, elm := range value.Array {
		args = append(args, elm.Str)
	}
	command := Command{}
	operationNormalized := strings.ToUpper(args[0])

	switch operationNormalized {
	case "GET":
		if len(args) != 2 {
			return Command{}, ErrInvalidCommandLength
		}
		if len(args[1]) == 0 {
			return Command{}, fmt.Errorf("invalid key length")
		}
		command.OP = operationNormalized
		command.Key = args[1]

	case "SET":
		if len(args) != 3 {
			return Command{}, ErrInvalidCommandLength
		}
		if len(args[1]) == 0 {
			return Command{}, fmt.Errorf("invalid key length")
		}
		command.OP = operationNormalized
		command.Key = args[1]
		command.Value = args[2]
	case "INCR":
		if len(args) != 2 {
			return Command{}, ErrInvalidCommandLength
		}
		if len(args[1]) == 0 {
			return Command{}, fmt.Errorf("invalid key length")
		}
		command.OP = operationNormalized
		command.Key = args[1]
	default:
		return Command{}, ErrUnknownCommand
	}

	return command, nil
}
