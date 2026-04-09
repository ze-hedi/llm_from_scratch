package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yourusername/chatbot-tui/internal/coordinator"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	Long:  `Launch the TUI interface and start chatting with the bot. Press Ctrl+A to browse extensions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		model := coordinator.NewModel()
		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
