package main

func HandleInput(executor *Executor, input string) (string, error) {
	command, err := ParseCommand(input)
	if err != nil {
		return "", err
	}
	return executor.Execute(command)
}
