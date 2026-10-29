package main

import (
	"fmt"
	"strings"
)

func ParseCommand(input string) (Command, error) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return Command{}, fmt.Errorf("ERR: invalid input")
	}
	input = strings.Join(fields, " ")

	parsedInput := strings.Split(input, " ")
	command := strings.ToUpper(parsedInput[0])

	if command != "INCR" && command != "GET" && command != "SET" {
		return Command{}, fmt.Errorf("ERR:invalid command name")
	}

	res := Command{}
	if command == "SET" {
		if len(parsedInput) != 3 {
			return Command{}, fmt.Errorf("ERR: invalid SET needs 3 tokens")
		}
		res = Command{
			OP:    command,
			Key:   parsedInput[1],
			Value: parsedInput[2],
		}

	}
	if command == "INCR" || command == "GET" {
		if len(parsedInput) != 2 {
			return Command{}, fmt.Errorf("ERR: invalid Command needs 2 tokens")
		}
		res = Command{
			OP:  command,
			Key: parsedInput[1],
		}
	}
	return res, nil
}
