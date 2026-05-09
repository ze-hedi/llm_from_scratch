package cmd

import (
	"github.com/spf13/cobra"
)

// NewChatRoot builds the root command for the chat binary,
// registering chat, settings, extensions, tamagotchi, and dino.
func NewChatRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "otto-chat",
		Short: "A beautiful TUI chatbot application",
		Long: `A terminal-based chatbot interface built with Bubble Tea.
Interact with an AI chatbot directly from your terminal with a beautiful,
responsive user interface.`,
	}
	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(newChatCmd())
	root.AddCommand(newSettingsCmd())
	root.AddCommand(newExtensionsCmd())
	registerTamagotchi(root)
	registerDino(root)

	return root
}

// NewTmuxRoot builds the root command for the tmux binary.
func NewTmuxRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "otto-tmux",
		Short: "Launch and manage tmux sessions with preset layouts",
		Long: `Provisions new tmux sessions with chosen pane layouts and attaches to them.

Available layouts:
  three   — one window, N panes — interactive setup
  dev     — two windows: "editor" (single pane) + "shells" (3 panes)
  custom  — one window with N blank panes (set with --panes)`,
	}
	root.CompletionOptions.DisableDefaultCmd = true

	registerTmux(root)

	return root
}
