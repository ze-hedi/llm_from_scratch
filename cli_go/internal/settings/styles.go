package settings

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

	modelItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedModelItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86")).
				Background(lipgloss.Color("236"))

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)

	tokenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Padding(2, 4)
)
