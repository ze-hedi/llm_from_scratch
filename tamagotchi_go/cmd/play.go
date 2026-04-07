package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yourusername/tamagotchi-tui/internal/tui"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Start playing with your Tamagotchi",
	Long:  `Launch the TUI and start taking care of your virtual pet!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		model := tui.NewModel()
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
	rootCmd.AddCommand(playCmd)
}
