package commands

import (
	"context"
	"fmt"
)

// Command represents an encapsulated operation that can be executed.
// Implements the Command pattern to encapsulate daemon operations as objects.
// This allows for:
//   - Queueing and executing operations asynchronously
//   - Undoing operations (if implemented)
//   - Logging and auditing of operations
//   - Retry logic and error recovery
type Command interface {
	// Execute performs the command's operation.
	// Returns an error if the operation fails.
	Execute(ctx context.Context) error

	// GetName returns a human-readable name for this command.
	// Used for logging and debugging.
	GetName() string

	// GetDescription returns a detailed description of what this command does.
	// Used for logging and UI display.
	GetDescription() string

	// CanUndo returns true if this command supports undo operations.
	CanUndo() bool

	// Undo reverses the command's operation (if supported).
	// Only called if CanUndo() returns true.
	// Returns an error if the undo operation fails.
	Undo(ctx context.Context) error
}

// CommandExecutor manages the execution of commands.
// Provides a centralized way to execute commands with logging and error handling.
type CommandExecutor struct {
	// commandHistory tracks executed commands for potential undo operations
	commandHistory []Command
	maxHistory     int
}

// NewCommandExecutor creates a new command executor.
// maxHistory specifies how many commands to keep in history (0 = unlimited).
func NewCommandExecutor(maxHistory int) *CommandExecutor {
	return &CommandExecutor{
		commandHistory: make([]Command, 0),
		maxHistory:     maxHistory,
	}
}

// Execute runs a command and adds it to the history.
func (e *CommandExecutor) Execute(ctx context.Context, cmd Command) error {
	// Execute the command
	err := cmd.Execute(ctx)
	if err != nil {
		return fmt.Errorf("command %s failed: %w", cmd.GetName(), err)
	}

	// Add to history if command supports undo
	if cmd.CanUndo() {
		e.addToHistory(cmd)
	}

	return nil
}

// ExecuteAll executes multiple commands in sequence.
// If any command fails, execution stops and returns the error.
func (e *CommandExecutor) ExecuteAll(ctx context.Context, commands []Command) error {
	for i, cmd := range commands {
		if err := e.Execute(ctx, cmd); err != nil {
			return fmt.Errorf("command %d (%s) failed: %w", i, cmd.GetName(), err)
		}
	}
	return nil
}

// Undo reverses the most recent command (if it supports undo).
func (e *CommandExecutor) Undo(ctx context.Context) error {
	if len(e.commandHistory) == 0 {
		return fmt.Errorf("no commands to undo")
	}

	// Get most recent command
	cmd := e.commandHistory[len(e.commandHistory)-1]

	// Undo the command
	if err := cmd.Undo(ctx); err != nil {
		return fmt.Errorf("failed to undo command %s: %w", cmd.GetName(), err)
	}

	// Remove from history
	e.commandHistory = e.commandHistory[:len(e.commandHistory)-1]

	return nil
}

// ClearHistory removes all commands from the history.
func (e *CommandExecutor) ClearHistory() {
	e.commandHistory = make([]Command, 0)
}

// GetHistory returns a copy of the command history.
func (e *CommandExecutor) GetHistory() []Command {
	history := make([]Command, len(e.commandHistory))
	copy(history, e.commandHistory)
	return history
}

// addToHistory adds a command to the history, enforcing the max history size.
func (e *CommandExecutor) addToHistory(cmd Command) {
	e.commandHistory = append(e.commandHistory, cmd)

	// Enforce max history size (if set)
	if e.maxHistory > 0 && len(e.commandHistory) > e.maxHistory {
		// Remove oldest commands
		e.commandHistory = e.commandHistory[len(e.commandHistory)-e.maxHistory:]
	}
}

// BaseCommand provides a base implementation of the Command interface.
// Embed this in your command structs to get default implementations.
type BaseCommand struct {
	name        string
	description string
}

// GetName returns the command name.
func (c *BaseCommand) GetName() string {
	return c.name
}

// GetDescription returns the command description.
func (c *BaseCommand) GetDescription() string {
	return c.description
}

// CanUndo returns false by default (most commands don't support undo).
func (c *BaseCommand) CanUndo() bool {
	return false
}

// Undo returns an error by default (must be overridden if CanUndo returns true).
func (c *BaseCommand) Undo(ctx context.Context) error {
	return fmt.Errorf("command %s does not support undo", c.name)
}
