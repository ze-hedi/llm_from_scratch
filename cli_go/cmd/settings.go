package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yourusername/chatbot-tui/internal/settings"
	"github.com/yourusername/chatbot-tui/internal/tui"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Configure chatbot settings",
	Long:  `Open the settings interface to configure model selection and other options.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		model := settings.NewSettingsModel()
		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
		)

		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running settings: %v\n", err)
			return err
		}

		// Check if user confirmed the selection
		if settingsModel, ok := finalModel.(settings.SettingsModel); ok {
			if settingsModel.Confirmed() {
				fmt.Println("✓ Model settings saved successfully!")
				fmt.Println("\nStarting chat with selected model...")

				// Launch chat after settings
				chatModel := tui.NewModel()
				chatProgram := tea.NewProgram(
					chatModel,
					tea.WithAltScreen(),
					tea.WithMouseCellMotion(),
				)

				if _, err := chatProgram.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Error running chat: %v\n", err)
					return err
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(settingsCmd)
}
