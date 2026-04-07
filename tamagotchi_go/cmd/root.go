package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tamagotchi-tui",
	Short: "A virtual pet Tamagotchi game in your terminal",
	Long: `Take care of your virtual pet Mochi! Feed, play, and keep your pet healthy.
Watch as your pet's mood changes with beautiful ASCII art.`,
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
