package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yourusername/chatbot-tui/extensions/tamagotchi"
	"github.com/yourusername/chatbot-tui/extensions/tamagotchi/pet"
	"github.com/yourusername/chatbot-tui/extensions/tamagotchi/tui"
)

var tamagotchiCmd = &cobra.Command{
	Use:   "tamagotchi",
	Short: "Tamagotchi virtual pet game",
	Long:  `Play with your virtual Tamagotchi pet. Choose, care for, and interact with your digital companion!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Try to load existing pet
		p, err := pet.LoadPet()
		if err != nil {
			// If no pet exists, show the choose screen
			fmt.Println("No pet found. Let's choose one!")
			if err := tamagotchi.RunChooseUI(); err != nil {
				return err
			}
			// Load the newly chosen pet
			p, err = pet.LoadPet()
			if err != nil {
				return fmt.Errorf("failed to load pet after selection: %w", err)
			}
		}

		// Start the game
		model := tui.NewModelWithPet(p)
		program := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := program.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
			return err
		}

		// Save the pet state after exiting
		if err := pet.SavePet(p); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving pet: %v\n", err)
		}

		return nil
	},
}

var tamagotchiChooseCmd = &cobra.Command{
	Use:   "choose",
	Short: "Choose a new Tamagotchi pet",
	Long:  `Select a new virtual pet to care for. This will replace your current pet if you have one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tamagotchi.RunChooseUI()
	},
}

var tamagotchiResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset your Tamagotchi",
	Long:  `Delete your current pet and start fresh.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		filePath := homeDir + "/tamagotchi.json"
		if err := os.Remove(filePath); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No pet to reset.")
				return nil
			}
			return fmt.Errorf("failed to remove pet file: %w", err)
		}

		fmt.Println("Your pet has been reset. Run 'tamagotchi' to choose a new one!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tamagotchiCmd)
	tamagotchiCmd.AddCommand(tamagotchiChooseCmd)
	tamagotchiCmd.AddCommand(tamagotchiResetCmd)
}
