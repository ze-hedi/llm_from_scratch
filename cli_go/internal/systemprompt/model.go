package systemprompt

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SystemPromptModel struct {
	textarea  textarea.Model
	width     int
	height    int
	err       error
	confirmed bool
}

func NewSystemPromptModel(currentPrompt string) SystemPromptModel {
	ta := textarea.New()
	ta.Placeholder = "Enter your system prompt here..."
	ta.Focus()
	ta.SetValue(currentPrompt)
	ta.CharLimit = 5000
	ta.ShowLineNumbers = false

	return SystemPromptModel{
		textarea:  ta,
		confirmed: false,
	}
}

func (m SystemPromptModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m SystemPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 8)
		m.textarea.SetHeight(msg.Height - 12)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "ctrl+s":
			// Save and confirm
			m.confirmed = true
			return m, nil
		}
	}

	// Update textarea
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m SystemPromptModel) View() string {
	// Header
	header := titleStyle.Render("📝 System Prompt")
	subtitle := subtitleStyle.Render("Edit the system prompt for the chatbot")
	instructions := instructionsStyle.Render("Ctrl+S: Save & Return | Esc: Cancel | Ctrl+C: Quit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		subtitle,
		"",
		instructions,
		"",
		dividerStyle.Render("─────────────────────────────────────────────────────"),
		"",
		m.textarea.View(),
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(2, 4).
		Render(content)
}

// Confirmed returns true if the user confirmed their edit
func (m SystemPromptModel) Confirmed() bool {
	return m.confirmed
}

// GetPrompt returns the edited prompt text
func (m SystemPromptModel) GetPrompt() string {
	return m.textarea.Value()
}
