package bot

import (
	"fmt"
)

type botCommand struct {
	Command string
	Run     CommandHandler
	Args    uint
	Hint    string
}

func RegisterCommand(command string, minArgs uint, hint string, handler CommandHandler) error {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	cmd := botCommand{
		Command: command,
		Args:    minArgs,
		Hint:    hint,
		Run:     handler,
	}

	if _, ok := commandSet[command]; ok {
		return fmt.Errorf("Command '%s' already registered", command)
	}

	commandSet[command] = cmd
	return nil
}

func SetDefaultCommand(command string) error {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	cmd, ok := commandSet[command]
	if !ok {
		return fmt.Errorf("Command '%s' not found", command)
	}

	defaultCommand = &cmd
	return nil
}
