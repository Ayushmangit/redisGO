package main

func HandleInput(executor *Executor, input string) (*Response, error) {
	command, err := ParseCommand(input)
	if err != nil {
		return nil, err
	}
	return executor.Execute(command)
}
