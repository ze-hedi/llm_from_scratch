package settings

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SettingsModel struct {
	models        []Model
	selectedIndex int
	currentModel  *Model
	width         int
	height        int
	err           error
	confirmed     bool
}

func NewSettingsModel() SettingsModel {
	models, err := LoadAvailableModels()
	if err != nil {
		return SettingsModel{
			models: []Model{},
			err:    err,
		}
	}

	currentModel, _ := LoadSelectedModel()

	// Find the index of the current model
	selectedIndex := 0
	if currentModel != nil {
		for i, m := range models {
			if m.ID == currentModel.ID {
				selectedIndex = i
				break
			}
		}
	}

	return SettingsModel{
		models:        models,
		selectedIndex: selectedIndex,
		currentModel:  currentModel,
		confirmed:     false,
	}
}

func (m SettingsModel) Init() tea.Cmd {
	return nil
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.models)-1 {
				m.selectedIndex++
			}

		case "enter":
			// Save the selected model
			if len(m.models) > 0 {
				selectedModel := m.models[m.selectedIndex]
				if err := SaveSelectedModel(selectedModel); err != nil {
					m.err = err
					return m, nil
				}
				m.confirmed = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m SettingsModel) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to quit.", m.err))
	}

	if len(m.models) == 0 {
		return errorStyle.Render("No models available.\n\nPress 'q' to quit.")
	}

	// Header
	header := titleStyle.Render("⚙️  Model Settings")
	subtitle := subtitleStyle.Render("Select a model for your chatbot")
	instructions := instructionsStyle.Render("↑/↓ or j/k: Navigate | Enter: Confirm | q/Esc: Quit")

	// Model list
	var modelsList string
	for i, model := range m.models {
		cursor := "  "
		style := modelItemStyle

		if i == m.selectedIndex {
			cursor = "▶ "
			style = selectedModelItemStyle
		}

		// Check if this is the current model
		currentIndicator := ""
		if m.currentModel != nil && model.ID == m.currentModel.ID {
			currentIndicator = " (current)"
		}

		modelLine := fmt.Sprintf("%s%s%s", cursor, model.Name, currentIndicator)
		description := fmt.Sprintf("  %s", model.Description)
		tokens := fmt.Sprintf("  Max tokens: %d", model.MaxTokens)

		modelsList += style.Render(modelLine) + "\n"
		modelsList += descriptionStyle.Render(description) + "\n"
		modelsList += tokenStyle.Render(tokens) + "\n\n"
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
		modelsList,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(2, 4).
		Render(content)
}

// Confirmed returns true if the user confirmed their selection
func (m SettingsModel) Confirmed() bool {
	return m.confirmed
}
