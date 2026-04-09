package extensions

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

	extensionItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(0, 1)

	selectedExtensionItemStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("86")).
					Background(lipgloss.Color("236")).
					Padding(0, 1)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Padding(2, 4)
)
