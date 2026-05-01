package tmuxui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 0, 1, 0)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(10)

	activeLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86")).
				Width(10)

	activeValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	arrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	previewTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(1, 0, 0, 0)

	previewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62"))

	formBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	confirmActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("46")).
				Background(lipgloss.Color("22")).
				Padding(0, 1)

	confirmMutedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 1)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			Padding(1, 0, 0, 0)

	// Paths frame styles.

	pathsFrameTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86")).
				Padding(0, 0, 1, 0)

	pathBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("34")).
			Padding(1, 2).
			Width(58)

	paneRowActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	paneRowMutedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)
