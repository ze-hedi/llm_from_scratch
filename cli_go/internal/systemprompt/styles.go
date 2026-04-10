package systemprompt

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 0, 1, 0)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 0, 1, 0)

	instructionsStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Italic(true).
				Padding(0, 0, 1, 0)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
