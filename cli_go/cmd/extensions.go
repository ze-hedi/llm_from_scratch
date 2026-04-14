package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yourusername/chatbot-tui/internal/extensions"
)

var extensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "Browse and launch available extensions",
	Long:  `Display a list of all available extensions and launch the selected one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show extensions selection UI
		model := extensions.NewModel()
		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
		)

		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running extensions browser: %v\n", err)
			return err
		}

		// Check if user selected an extension
		if extModel, ok := finalModel.(extensions.Model); ok {
			if selectedExt := extModel.SelectedExtension(); selectedExt != nil {
				fmt.Printf("\n🚀 Launching %s...\n\n", selectedExt.Name)

				// Find and execute the corresponding command
				switch selectedExt.Command {
				case "tamagotchi":
					return tamagotchiCmd.RunE(cmd, args)
				case "dino":
					return dinoCmd.RunE(cmd, args)
				default:
					fmt.Printf("Extension command '%s' not found.\n", selectedExt.Command)
					return nil
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(extensionsCmd)
}
