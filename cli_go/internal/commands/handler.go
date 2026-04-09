package commands

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CommandResult represents the result of executing a command
type CommandResult struct {
	Message      string
	ShouldQuit   bool
	IsCommand    bool
	ErrorMessage string
}

// Handler processes slash commands
type Handler struct {
	commands map[string]CommandFunc
}

// CommandFunc is the signature for command handler functions
type CommandFunc func(args []string) CommandResult

// NewHandler creates a new command handler with registered commands
func NewHandler() *Handler {
	h := &Handler{
		commands: make(map[string]CommandFunc),
	}

	// Register built-in commands
	h.RegisterCommand("exit", handleExit)
	h.RegisterCommand("quit", handleExit) // Alias for exit

	return h
}

// RegisterCommand registers a new command handler
func (h *Handler) RegisterCommand(name string, handler CommandFunc) {
	h.commands[name] = handler
}

// Process checks if the input is a command and processes it
func (h *Handler) Process(input string) CommandResult {
	trimmed := strings.TrimSpace(input)

	// Check if it's a command (starts with /)
	if !strings.HasPrefix(trimmed, "/") {
		return CommandResult{
			IsCommand: false,
		}
	}

	// Remove the leading slash
	trimmed = strings.TrimPrefix(trimmed, "/")

	// Split command and arguments
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return CommandResult{
			IsCommand:    true,
			ErrorMessage: "Empty command",
		}
	}

	commandName := strings.ToLower(parts[0])
	args := []string{}
	if len(parts) > 1 {
		args = parts[1:]
	}

	// Look up the command
	handler, exists := h.commands[commandName]
	if !exists {
		return CommandResult{
			IsCommand:    true,
			ErrorMessage: "Unknown command: /" + commandName,
		}
	}

	// Execute the command
	return handler(args)
}

// GetAvailableCommands returns a list of all registered commands
func (h *Handler) GetAvailableCommands() []string {
	commands := make([]string, 0, len(h.commands))
	for cmd := range h.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// handleExit handles the /exit and /quit commands
func handleExit(args []string) CommandResult {
	return CommandResult{
		IsCommand:  true,
		ShouldQuit: true,
		Message:    "Goodbye! 👋",
	}
}

// QuitMsg is sent when a command requests to quit
type QuitMsg struct{}

// ToQuitCmd converts a CommandResult to a tea.Cmd if needed
func (r CommandResult) ToQuitCmd() tea.Cmd {
	if r.ShouldQuit {
		return tea.Quit
	}
	return nil
}
