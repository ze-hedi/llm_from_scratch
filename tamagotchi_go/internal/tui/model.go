package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/tamagotchi-tui/pkg/tamagotchi"
)

type tickMsg time.Time

type Model struct {
	pet      *tamagotchi.Pet
	textarea textarea.Model
	width    int
	height   int
	messages []string
	err      error
}

func NewModel() Model {
	ta := textarea.New()
	ta.Placeholder = "Enter command: feed, play, heal, status, or quit"
	ta.Focus()
	ta.Prompt = "➤ "
	ta.CharLimit = 100
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	// Apply grey background
	ta.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("235"))

	return Model{
		pet:      tamagotchi.NewPet("Mochi"),
		textarea: ta,
		messages: []string{"Welcome! Take care of Mochi! 🐾"},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			input := strings.ToLower(strings.TrimSpace(m.textarea.Value()))
			if input == "" {
				break
			}

			var response string
			switch input {
			case "feed":
				response = m.pet.Feed()
			case "play":
				response = m.pet.Play()
			case "heal":
				response = m.pet.Heal()
			case "status":
				response = m.pet.GetStatus()
			case "quit", "exit":
				return m, tea.Quit
			case "help":
				response = "Commands: feed, play, heal, status, quit"
			default:
				response = fmt.Sprintf("Unknown command '%s'. Try: feed, play, heal, status", input)
			}

			m.messages = append(m.messages, fmt.Sprintf("You: %s", input))
			m.messages = append(m.messages, response)

			// Keep only last 10 messages
			if len(m.messages) > 10 {
				m.messages = m.messages[len(m.messages)-10:]
			}

			m.textarea.Reset()
		}

	case tickMsg:
		m.pet.Update()
		return m, tickCmd()

	case error:
		m.err = msg
		return m, nil
	}

	return m, cmd
}

func (m Model) View() string {
	header := m.renderHeader()
	pet := m.renderPet()
	messages := m.renderMessages()
	input := m.renderInput()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		pet,
		messages,
		input,
	)
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("🐾 Tamagotchi Game")
	subtitle := subtitleStyle.Render("Keep Mochi alive! Commands: feed, play, heal, status")
	line := strings.Repeat("─", max(0, m.width-2))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(line),
	)
}

func (m Model) renderPet() string {
	petArt := m.pet.GetASCII()
	status := m.pet.GetStatus()

	petBox := petStyle.Render(petArt)
	statusBar := statusStyle.Render(status)

	line := strings.Repeat("─", max(0, m.width-2))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		petBox,
		statusBar,
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(line),
	)
}

func (m Model) renderMessages() string {
	if len(m.messages) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\nRecent Activity:\n"))

	for _, msg := range m.messages {
		sb.WriteString(messageStyle.Render("  " + msg + "\n"))
	}

	return sb.String()
}

func (m Model) renderInput() string {
	line := strings.Repeat("─", max(0, m.width-2))
	info := infoStyle.Render("Enter: Send | Esc/Ctrl+C: Quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(line),
		m.textarea.View(),
		info,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
