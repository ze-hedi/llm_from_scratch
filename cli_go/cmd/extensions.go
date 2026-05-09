package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yourusername/chatbot-tui/internal/extensions"
)

func newExtensionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "extensions",
		Short: "Browse and launch available extensions",
		Long:  `Display a list of all available extensions and launch the selected one.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if extModel, ok := finalModel.(extensions.Model); ok {
				if selectedExt := extModel.SelectedExtension(); selectedExt != nil {
					fmt.Printf("\n🚀 Launching %s...\n\n", selectedExt.Name)

					switch selectedExt.Command {
					case "tamagotchi":
						return runTamagotchi(cmd, nil)
					case "dino":
						return runDino(cmd, nil)
					default:
						fmt.Printf("Extension command '%s' not found.\n", selectedExt.Command)
						return nil
					}
				}
			}

			return nil
		},
	}
}
