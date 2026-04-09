package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/chatbot-tui/internal/commands"
	"github.com/yourusername/chatbot-tui/internal/settings"
	"github.com/yourusername/chatbot-tui/pkg/chatbot"
)

type Model struct {
	viewport       viewport.Model
	textarea       textarea.Model
	messages       []chatbot.Message
	bot            *chatbot.Bot
	cmdHandler     *commands.Handler
	width          int
	height         int
	ready          bool
	err            error
	inputTokens    int
	outputTokens   int
	sidebarVisible bool
	currentModel   *settings.Model
	maxTokens      int
}

func NewModel() Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 2000
	ta.SetWidth(80)
	ta.SetHeight(1) // Start with one line
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(true) // Allow multi-line input
	// Dynamically expand as user types, max 10 lines
	ta.MaxHeight = 10

	// Apply uniform grey background to entire textarea
	ta.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("235"))
	ta.BlurredStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))

	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Load selected model
	currentModel, _ := settings.LoadSelectedModel()
	maxTokens := 200000 // default
	if currentModel != nil {
		maxTokens = currentModel.MaxTokens
	}

	return Model{
		textarea:       ta,
		viewport:       vp,
		messages:       []chatbot.Message{},
		bot:            chatbot.NewBot(),
		cmdHandler:     commands.NewHandler(),
		ready:          false,
		inputTokens:    0,
		outputTokens:   0,
		sidebarVisible: true,
		currentModel:   currentModel,
		maxTokens:      maxTokens,
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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Account for sidebar width (30 chars + 4 for borders/padding)
		sidebarWidth := 0
		if m.sidebarVisible {
			sidebarWidth = 34
		}
		mainWidth := msg.Width - sidebarWidth

		if !m.ready {
			m.viewport = viewport.New(mainWidth, msg.Height-10)
			m.viewport.YPosition = 0
			m.ready = true
		} else {
			m.viewport.Width = mainWidth
			m.viewport.Height = msg.Height - 10
		}
		m.textarea.SetWidth(mainWidth - 4)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyCtrlN:
			// Toggle sidebar visibility
			m.sidebarVisible = !m.sidebarVisible

			// Recalculate layout with new sidebar state
			sidebarWidth := 0
			if m.sidebarVisible {
				sidebarWidth = 34
			}
			mainWidth := m.width - sidebarWidth

			m.viewport.Width = mainWidth
			m.textarea.SetWidth(mainWidth - 4)

			// Re-render messages with new width
			m.viewport.SetContent(m.renderMessages())

			return m, nil

		case tea.KeyEnter:
			// Plain Enter (without Alt) sends the message
			// Alt+Enter is handled by the textarea for new lines
			if !msg.Alt {
				userInput := strings.TrimSpace(m.textarea.Value())
				if userInput == "" {
					// Let textarea handle it if empty
					break
				}

				// Check if it's a command
				cmdResult := m.cmdHandler.Process(userInput)

				if cmdResult.IsCommand {
					// It's a command - handle it
					if cmdResult.ErrorMessage != "" {
						// Command error
						m.messages = append(m.messages, chatbot.Message{
							Role:    chatbot.RoleUser,
							Content: userInput,
						})
						m.messages = append(m.messages, chatbot.Message{
							Role:    chatbot.RoleBot,
							Content: "❌ " + cmdResult.ErrorMessage,
						})
					} else if cmdResult.Message != "" {
						// Command success with message
						m.messages = append(m.messages, chatbot.Message{
							Role:    chatbot.RoleUser,
							Content: userInput,
						})
						m.messages = append(m.messages, chatbot.Message{
							Role:    chatbot.RoleBot,
							Content: cmdResult.Message,
						})
					}

					// Update viewport
					m.viewport.SetContent(m.renderMessages())
					m.viewport.GotoBottom()

					// Clear textarea
					m.textarea.Reset()

					// Check if we should quit
					if cmdResult.ShouldQuit {
						return m, tea.Quit
					}

					return m, nil
				}

				// Not a command - normal chat message
				// Add user message
				m.messages = append(m.messages, chatbot.Message{
					Role:    chatbot.RoleUser,
					Content: userInput,
				})

				// Track input tokens
				m.inputTokens += estimateTokens(userInput)

				// Get bot response
				response := m.bot.GetResponse(userInput)
				m.messages = append(m.messages, chatbot.Message{
					Role:    chatbot.RoleBot,
					Content: response,
				})

				// Track output tokens
				m.outputTokens += estimateTokens(response)

				// Update viewport
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()

				// Clear textarea
				m.textarea.Reset()

				// Don't let the textarea process this Enter
				m.viewport, vpCmd = m.viewport.Update(msg)
				return m, vpCmd
			}
		}

	case error:
		m.err = msg
		return m, nil
	}

	// Let textarea and viewport handle all other messages
	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Calculate widths
	sidebarWidth := 0
	if m.sidebarVisible {
		sidebarWidth = 34 // 30 + 4 for borders/padding
	}
	mainWidth := m.width - sidebarWidth

	header := m.renderHeader(mainWidth)
	content := m.viewport.View()
	footer := m.renderFooter()

	// Main content area (left side) with fixed width
	mainContent := lipgloss.NewStyle().
		Width(mainWidth).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			content,
			footer,
		))

	// If sidebar is visible, join it with main content
	if m.sidebarVisible {
		sidebar := m.renderSidebar()
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			mainContent,
			sidebar,
		)
	}

	return mainContent
}

func (m Model) renderHeader(width int) string {
	title := titleStyle.Render("🤖 ChatBot TUI")
	subtitle := subtitleStyle.Render("Ctrl+A: extensions | Ctrl+N: sidebar | Alt+Enter: new line | Esc: quit")

	line := strings.Repeat("─", max(0, width-2))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(line),
	)
}

func (m Model) renderMessages() string {
	var sb strings.Builder

	// Calculate available width for messages (account for sidebar and padding)
	sidebarWidth := 0
	if m.sidebarVisible {
		sidebarWidth = 34 // 30 + 4 for borders/padding
	}
	maxWidth := m.width - sidebarWidth - 6 // Additional padding for safety

	for i, msg := range m.messages {
		if i > 0 {
			sb.WriteString("\n\n")
		}

		if msg.Role == chatbot.RoleUser {
			// Wrap the message content with width constraint
			wrappedContent := wrapText("You: "+msg.Content, maxWidth)
			sb.WriteString(userMessageStyle.Render(wrappedContent))
		} else {
			// Wrap the message content with width constraint
			wrappedContent := wrapText("Bot: "+msg.Content, maxWidth)
			sb.WriteString(botMessageStyle.Render(wrappedContent))
		}
	}

	return sb.String()
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	var currentLine strings.Builder
	words := strings.Fields(text)

	for i, word := range words {
		// Check if adding this word would exceed the width
		if currentLine.Len()+len(word)+1 > width {
			// Write current line and start a new one
			if currentLine.Len() > 0 {
				result.WriteString(currentLine.String())
				result.WriteString("\n")
				currentLine.Reset()
			}

			// If a single word is longer than width, break it up
			if len(word) > width {
				for len(word) > width {
					result.WriteString(word[:width])
					result.WriteString("\n")
					word = word[width:]
				}
				if len(word) > 0 {
					currentLine.WriteString(word)
				}
			} else {
				currentLine.WriteString(word)
			}
		} else {
			// Add space before word (except for first word)
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}

		// Add the last line
		if i == len(words)-1 && currentLine.Len() > 0 {
			result.WriteString(currentLine.String())
		}
	}

	return result.String()
}

func (m Model) renderFooter() string {
	info := infoStyle.Render(
		"Enter: Send | Alt+Enter: New Line | /exit: Quit | Esc: Quit | Messages: " +
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

// estimateTokens provides a rough estimate of tokens in text
// Claude's tokenizer averages ~4 characters per token
func estimateTokens(text string) int {
	return len(text) / 4
}

func (m Model) renderSidebar() string {
	totalTokens := m.inputTokens + m.outputTokens
	remainingTokens := m.maxTokens - totalTokens
	percentUsed := float64(totalTokens) / float64(m.maxTokens) * 100

	title := sidebarTitleStyle.Render("Token Usage")
	divider := strings.Repeat("─", 26)

	// Get model name
	modelName := "Claude Sonnet 4.5" // default
	if m.currentModel != nil {
		modelName = m.currentModel.Name
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(divider),
		"",
		sidebarLabelStyle.Render("Input:"),
		sidebarValueStyle.Render(fmt.Sprintf("  %d tokens", m.inputTokens)),
		"",
		sidebarLabelStyle.Render("Output:"),
		sidebarValueStyle.Render(fmt.Sprintf("  %d tokens", m.outputTokens)),
		"",
		sidebarLabelStyle.Render("Total:"),
		sidebarValueStyle.Render(fmt.Sprintf("  %d tokens", totalTokens)),
		"",
		sidebarLabelStyle.Render("Remaining:"),
		sidebarValueStyle.Render(fmt.Sprintf("  %d tokens", remainingTokens)),
		"",
		sidebarLabelStyle.Render("Usage:"),
		sidebarValueStyle.Render(fmt.Sprintf("  %.2f%%", percentUsed)),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(divider),
		"",
		sidebarLabelStyle.Render("Model:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(fmt.Sprintf("  %s", modelName)),
	)

	// Apply sidebar style with proper height constraint
	return sidebarStyle.
		Height(m.height - 2).
		Render(content)
}
