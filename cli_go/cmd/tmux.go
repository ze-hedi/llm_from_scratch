package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/chatbot-tui/internal/tmuxui"
	"github.com/yourusername/chatbot-tui/layout"
	"github.com/yourusername/chatbot-tui/pkg/tmux"
)

var tmuxCmd = &cobra.Command{
	Use:   "tmux",
	Short: "Launch a tmux session with a preset layout",
	Long: `Provisions a new tmux session with a chosen pane layout and attaches to it.

Available layouts:
  three   — one window, 3 equal horizontal panes
  dev     — two windows: "editor" (single pane) + "shells" (3 panes)
  custom  — one window with N blank panes (set with --panes)`,
}

var (
	flagSession string
	flagWorkDir string
	flagPanes   int
	flagConfig  string
)

func init() {
	rootCmd.AddCommand(tmuxCmd)

	tmuxCmd.PersistentFlags().StringVarP(&flagSession, "session", "s", "dev", "tmux session name")
	tmuxCmd.PersistentFlags().StringVarP(&flagWorkDir, "dir", "d", "", "working directory for panes (defaults to current dir)")

	tmuxCmd.AddCommand(tmuxThreeCmd)
	tmuxCmd.AddCommand(tmuxDevCmd)
	tmuxCmd.AddCommand(tmuxCustomCmd)
	tmuxCmd.AddCommand(tmuxAttachCmd)
	tmuxCmd.AddCommand(tmuxKillCmd)

	tmuxCustomCmd.Flags().IntVarP(&flagPanes, "panes", "p", 3, "number of panes to open")
	tmuxThreeCmd.Flags().StringVar(&flagConfig, "config", "", "path to a JSON config file (skips interactive setup)")
}

// ── three ──────────────────────────────────────────────────────────────────

var tmuxThreeCmd = &cobra.Command{
	Use:   "three",
	Short: "One window, N panes — interactive setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagConfig != "" {
			cfg, err := tmux.LoadConfig(flagConfig)
			if err != nil {
				return err
			}
			spec := cfg.ToSessionSpec()
			return realizeAndAttach(spec)
		}

		result, err := tmuxui.Run(flagSession)
		if err != nil {
			return err
		}
		if result == nil {
			return nil // user cancelled
		}
		panes := make([]tmux.PaneSpec, result.PaneCount)
		for i := range panes {
			panes[i] = tmux.PaneSpec{WorkDir: result.Paths[i]}
		}
		spec := layout.Custom(result.SessionName, result.Layout, panes)
		return realizeAndAttach(spec)
	},
}

// ── dev ────────────────────────────────────────────────────────────────────

var tmuxDevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Two windows: editor pane + 3 shell panes",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := resolveWorkDir()
		spec := layout.Dev(flagSession, dir)
		return realizeAndAttach(spec)
	},
}

// ── custom ─────────────────────────────────────────────────────────────────

var tmuxCustomCmd = &cobra.Command{
	Use:   "custom",
	Short: "One window with N blank panes (--panes N)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagPanes < 1 {
			return fmt.Errorf("--panes must be at least 1")
		}
		dir := resolveWorkDir()
		panes := make([]tmux.PaneSpec, flagPanes)
		for i := range panes {
			panes[i] = tmux.PaneSpec{WorkDir: dir}
		}
		spec := layout.Custom(flagSession, tmux.LayoutEvenHorizontal, panes)
		return realizeAndAttach(spec)
	},
}

// ── attach ─────────────────────────────────────────────────────────────────

var tmuxAttachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to an existing session",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tmux.SessionExists(flagSession) {
			return fmt.Errorf("session %q does not exist", flagSession)
		}
		return tmux.Attach(flagSession)
	},
}

// ── kill ───────────────────────────────────────────────────────────────────

var tmuxKillCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kill a session by name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tmux.SessionExists(flagSession) {
			return fmt.Errorf("session %q does not exist", flagSession)
		}
		_, err := tmux.Run("kill-session", "-t", flagSession)
		if err != nil {
			return err
		}
		fmt.Printf("session %q killed\n", flagSession)
		return nil
	},
}

// ── helpers ────────────────────────────────────────────────────────────────

func resolveWorkDir() string {
	if flagWorkDir != "" {
		return flagWorkDir
	}
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func realizeAndAttach(spec tmux.SessionSpec) error {
	if err := spec.Realize(); err != nil {
		return err
	}
	return tmux.Attach(spec.Name)
}
