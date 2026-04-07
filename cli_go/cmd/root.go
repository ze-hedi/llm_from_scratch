package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "chatbot-tui",
	Short: "A beautiful TUI chatbot application",
	Long: `A terminal-based chatbot interface built with Bubble Tea.
Interact with an AI chatbot directly from your terminal with a beautiful,
responsive user interface.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
