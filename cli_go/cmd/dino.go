package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yourusername/chatbot-tui/extensions/dino/game"
	"github.com/yourusername/chatbot-tui/extensions/dino/tui"
)

var dinoCmd = &cobra.Command{
	Use:   "dino",
	Short: "Dino Runner - T-Rex endless runner game",
	Long:  `Play the classic Chrome dinosaur game in your terminal! Jump over cacti and dodge birds to achieve the highest score.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load high score
		highScore := game.LoadHighScore()

		// Start the game
		model := tui.NewModel(highScore)
		program := tea.NewProgram(
			model,
			tea.WithAltScreen(),
		)

		finalModel, err := program.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
			return err
		}

		// Save high score after exiting
		if m, ok := finalModel.(tui.Model); ok {
			if err := game.SaveHighScore(m.GetHighScore()); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving high score: %v\n", err)
			}
		}

		return nil
	},
}

var dinoResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset your high score",
	Long:  `Delete your saved high score and start fresh.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := game.SaveHighScore(0); err != nil {
			return fmt.Errorf("failed to reset high score: %w", err)
		}
		fmt.Println("High score has been reset!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dinoCmd)
	dinoCmd.AddCommand(dinoResetCmd)
}
