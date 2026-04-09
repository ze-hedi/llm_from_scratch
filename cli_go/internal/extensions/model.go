package extensions

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	extensions    []Extension
	selectedIndex int
	width         int
	height        int
	err           error
	selectedExt   *Extension
}

func NewModel() Model {
	extensions, err := LoadExtensions()
	if err != nil {
		return Model{
			extensions: []Extension{},
			err:        err,
		}
	}

	return Model{
		extensions:    extensions,
		selectedIndex: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.extensions)-1 {
				m.selectedIndex++
			}

		case "enter":
			// Store the selected extension and quit
			if len(m.extensions) > 0 {
				m.selectedExt = &m.extensions[m.selectedIndex]
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error loading extensions: %v\n\nPress 'q' to quit.", m.err))
	}

	if len(m.extensions) == 0 {
		return errorStyle.Render("No extensions available.\n\nPress 'q' to quit.")
	}

	// Header
	header := titleStyle.Render("🧩 Available Extensions")
	subtitle := subtitleStyle.Render("Select an extension to launch")
	instructions := instructionsStyle.Render("↑/↓ or j/k: Navigate | Enter: Launch | Ctrl+A/Esc: Back to Chat")

	// Extensions list
	var extensionsList string
	for i, ext := range m.extensions {
		cursor := "  "
		style := extensionItemStyle

		if i == m.selectedIndex {
			cursor = "▶ "
			style = selectedExtensionItemStyle
		}

		title := fmt.Sprintf("%s %s", ext.Icon, ext.Name)
		description := fmt.Sprintf("  %s", ext.Description)
		command := fmt.Sprintf("  Command: %s", ext.Command)

		extensionsList += style.Render(cursor+title) + "\n"
		extensionsList += descriptionStyle.Render(description) + "\n"
		extensionsList += commandStyle.Render(command) + "\n\n"
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		subtitle,
		"",
		instructions,
		"",
		dividerStyle.Render("─────────────────────────────────────────────────────"),
		"",
		extensionsList,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(2, 4).
		Render(content)
}

// SelectedExtension returns the selected extension if user pressed Enter
func (m Model) SelectedExtension() *Extension {
	return m.selectedExt
}
