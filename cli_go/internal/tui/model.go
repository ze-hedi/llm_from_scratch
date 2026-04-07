package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/chatbot-tui/pkg/chatbot"
)

type Model struct {
	viewport viewport.Model
	textarea textarea.Model
	messages []chatbot.Message
	bot      *chatbot.Bot
	width    int
	height   int
	ready    bool
	err      error
}

func NewModel() Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 500
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	// Apply uniform grey background to entire textarea
	ta.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("235"))
	ta.BlurredStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))

	vp := viewport.New(80, 20)
	vp.SetContent("")

	return Model{
		textarea: ta,
		viewport: vp,
		messages: []chatbot.Message{},
		bot:      chatbot.NewBot(),
		ready:    false,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-10)
			m.viewport.YPosition = 0
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 10
		}
		m.textarea.SetWidth(msg.Width - 4)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			userInput := strings.TrimSpace(m.textarea.Value())
			if userInput == "" {
				break
			}

			// Add user message
			m.messages = append(m.messages, chatbot.Message{
				Role:    chatbot.RoleUser,
				Content: userInput,
			})

			// Get bot response
			response := m.bot.GetResponse(userInput)
			m.messages = append(m.messages, chatbot.Message{
				Role:    chatbot.RoleBot,
				Content: response,
			})

			// Update viewport
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()

			// Clear textarea
			m.textarea.Reset()
		}

	case error:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	header := m.renderHeader()
	content := m.viewport.View()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("🤖 ChatBot TUI")
	subtitle := subtitleStyle.Render("Press Ctrl+C or Esc to quit")

	line := strings.Repeat("─", max(0, m.width-2))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(line),
	)
}

func (m Model) renderMessages() string {
	var sb strings.Builder

	for i, msg := range m.messages {
		if i > 0 {
			sb.WriteString("\n\n")
		}

		if msg.Role == chatbot.RoleUser {
			sb.WriteString(userMessageStyle.Render("You: " + msg.Content))
		} else {
			sb.WriteString(botMessageStyle.Render("Bot: " + msg.Content))
		}
	}

	return sb.String()
}

func (m Model) renderFooter() string {
	info := infoStyle.Render(
		"Enter: Send | Esc: Quit | Messages: " +
			lipgloss.NewStyle().Bold(true).Render(string(rune(len(m.messages)))),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
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
