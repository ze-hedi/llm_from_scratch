package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
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

// Attach replaces the current process with `tmux attach -t <name>`.
// Nothing runs in Go after this call succeeds.
func Attach(name string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found in PATH: %w", err)
	}
	return syscall.Exec(tmuxPath, []string{"tmux", "attach-session", "-t", name}, os.Environ())
}
