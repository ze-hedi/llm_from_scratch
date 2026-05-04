package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run fires a tmux subcommand and returns combined output + any error.
func Run(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux %s: %w — %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// SessionExists returns true if a session with that name is already running.
func SessionExists(name string) bool {
	_, err := Run("has-session", "-t", name)
	return err == nil
}

// Attach attaches the current terminal to the named tmux session.
// It blocks until the user exits or detaches, then returns.
func Attach(name string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
