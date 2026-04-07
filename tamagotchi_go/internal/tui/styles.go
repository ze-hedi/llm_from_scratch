package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(1, 0, 0, 2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 0, 1, 2)

	petStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			PaddingLeft(2)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			PaddingLeft(2).
			PaddingTop(1)

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("248"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(1, 0, 0, 2)
)
