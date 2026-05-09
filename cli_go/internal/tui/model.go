package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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
	isStreaming    bool // Track if we're currently receiving a stream
	mdRenderer     *glamour.TermRenderer
	lastKeyTime    time.Time // Track last keystroke for paste detection
}

func NewModel() Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 2000
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(true)

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

	mdRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

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
		mdRenderer:     mdRenderer,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// ReloadModelSettings reloads the current model settings from config
// This is useful when the user changes the model in settings
func (m *Model) ReloadModelSettings() {
	currentModel, _ := settings.LoadSelectedModel()
	maxTokens := 200000 // default
	if currentModel != nil {
		maxTokens = currentModel.MaxTokens
	}
	m.currentModel = currentModel
	m.maxTokens = maxTokens
}

// isFullScreen determines if terminal is in full screen mode
// We consider full screen when width >= 120 columns (enough for sidebar + content)
func (m Model) isFullScreen() bool {
	return m.width >= 120
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

		// Auto-hide sidebar if not in full screen mode
		// Only show sidebar when terminal is wide enough (full screen)
		canShowSidebar := m.isFullScreen()

		// Account for sidebar width (30 chars + 4 for borders/padding)
		sidebarWidth := 0
		if m.sidebarVisible && canShowSidebar {
			sidebarWidth = 34
		}
		mainWidth := msg.Width - sidebarWidth

		if !m.ready {
			m.viewport = viewport.New(mainWidth, msg.Height-8)
			m.viewport.YPosition = 0
			m.ready = true
		} else {
			m.viewport.Width = mainWidth
		}
		m.textarea.SetWidth(mainWidth - 4)
		m.resizeTextarea()

		// Recreate markdown renderer with updated width
		mdWidth := mainWidth - 8 // account for padding
		if mdWidth < 20 {
			mdWidth = 20
		}
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(mdWidth),
		)
		if err == nil {
			m.mdRenderer = r
		}

	case tea.KeyMsg:
		// Track timing for paste detection (all keys except Enter)
		if msg.Type != tea.KeyEnter {
			m.lastKeyTime = time.Now()
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyCtrlN:
			// Only allow toggling sidebar in full screen mode
			if !m.isFullScreen() {
				// Ignore Ctrl+N when not in full screen
				return m, nil
			}

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
			if msg.Alt {
				// Alt+Enter always inserts a newline
				break
			}

			// Paste detection: if Enter arrives within 50ms of last keystroke,
			// it's part of a paste — insert newline instead of sending
			if time.Since(m.lastKeyTime) < 50*time.Millisecond {
				break
			}

			userInput := strings.TrimSpace(m.textarea.Value())
			if userInput == "" {
				break
			}

			// Check if it's a command
			cmdResult := m.cmdHandler.Process(userInput)

			if cmdResult.IsCommand {
				if cmdResult.ErrorMessage != "" {
					m.messages = append(m.messages, chatbot.Message{
						Role:    chatbot.RoleUser,
						Content: userInput,
					})
					m.messages = append(m.messages, chatbot.Message{
						Role:    chatbot.RoleBot,
						Content: "❌ " + cmdResult.ErrorMessage,
					})
				} else if cmdResult.Message != "" {
					m.messages = append(m.messages, chatbot.Message{
						Role:    chatbot.RoleUser,
						Content: userInput,
					})
					m.messages = append(m.messages, chatbot.Message{
						Role:    chatbot.RoleBot,
						Content: cmdResult.Message,
					})
				}

				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
				m.textarea.Reset()

				if cmdResult.ShouldQuit {
					return m, tea.Quit
				}

				return m, nil
			}

			// Normal chat message
			m.messages = append(m.messages, chatbot.Message{
				Role:    chatbot.RoleUser,
				Content: userInput,
			})
			m.inputTokens += estimateTokens(userInput)

			m.messages = append(m.messages, chatbot.Message{
				Role:    chatbot.RoleBot,
				Content: "⏳ Thinking...",
			})

			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			m.textarea.Reset()

			m.isStreaming = true
			m.viewport, vpCmd = m.viewport.Update(msg)
			return m, tea.Batch(vpCmd, m.bot.GetResponseStream(userInput))
		}

	case chatbot.StreamChunkMsg:
		// Streaming chunk received
		if m.isStreaming && len(m.messages) > 0 {
			// Update the last message (bot response) with the chunk
			lastIdx := len(m.messages) - 1
			m.messages[lastIdx].Content = msg.Chunk

			// Track output tokens
			m.outputTokens += estimateTokens(msg.Chunk)

			// Update viewport
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()

			// Mark streaming as done
			m.isStreaming = false
		}
		return m, nil

	case chatbot.StreamErrorMsg:
		// Streaming error occurred
		if m.isStreaming && len(m.messages) > 0 {
			lastIdx := len(m.messages) - 1
			m.messages[lastIdx].Content = fmt.Sprintf("❌ Error: %v", msg.Err)
			m.isStreaming = false

			// Update viewport
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
		}
		return m, nil

	case chatbot.FetchAgentListMsg:
		// User requested agent list - show loading message and start fetch
		m.messages = append(m.messages, chatbot.Message{
			Role:    chatbot.RoleBot,
			Content: "⏳ Fetching agents from server...",
		})

		// Update viewport
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		// Start the fetch
		return m, m.bot.GetAgentList()

	case chatbot.AgentListMsg:
		// Agent list received successfully
		var content strings.Builder
		content.WriteString("📋 Available Agents:\n\n")
		for i, agent := range msg.Agents {
			content.WriteString(fmt.Sprintf("%d. %s\n", i+1, agent.Name))
			content.WriteString(fmt.Sprintf("   %s\n", agent.Description))
			if i < len(msg.Agents)-1 {
				content.WriteString("\n")
			}
		}

		// Replace the loading message with the actual agent list
		if len(m.messages) > 0 {
			lastIdx := len(m.messages) - 1
			m.messages[lastIdx].Content = content.String()
		}

		// Update viewport
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case chatbot.AgentListErrorMsg:
		// Agent list fetch failed - replace loading message with error
		if len(m.messages) > 0 {
			lastIdx := len(m.messages) - 1
			m.messages[lastIdx].Content = "❌ failed to connect to server"
		}

		// Update viewport
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	// Let textarea and viewport handle all other messages
	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	// Resize textarea to fit content
	m.resizeTextarea()

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// If no messages, show centered initial view
	if len(m.messages) == 0 {
		return m.renderInitialView()
	}

	// Only show sidebar if in full screen mode AND user wants it visible
	canShowSidebar := m.isFullScreen() && m.sidebarVisible

	// Calculate widths
	sidebarWidth := 0
	if canShowSidebar {
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

	// If sidebar is visible AND in full screen, join it with main content
	if canShowSidebar {
		sidebar := m.renderSidebar()
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			mainContent,
			sidebar,
		)
	}

	return mainContent
}

func (m Model) renderInitialView() string {
	// Centered title
	titleText := titleStyle.Render("🤖 ChatBot TUI")

	// Welcome message
	welcomeMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Start a conversation...")

	// Input area
	inputArea := m.textarea.View()

	// Footer info
	info := infoStyle.Render("Enter: Send | Alt+Enter: New Line | Esc: Quit")

	// Combine title, welcome, and input
	centeredContent := lipgloss.JoinVertical(
		lipgloss.Center,
		titleText,
		"",
		"",
		welcomeMsg,
		"",
		"",
		inputArea,
		info,
	)

	// Place everything in the center of the screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		centeredContent,
	)
}

func (m Model) renderHeader(width int) string {
	// Center the title
	titleText := titleStyle.Render("🤖 ChatBot TUI")
	title := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(titleText)

	line := strings.Repeat("─", max(0, width-2))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
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
			wrappedContent := wrapText("You: "+msg.Content, maxWidth)
			sb.WriteString(userMessageStyle.Render(wrappedContent))
		} else {
			rendered := msg.Content
			if m.mdRenderer != nil {
				if md, err := m.mdRenderer.Render(msg.Content); err == nil {
					rendered = strings.TrimSpace(md)
				}
			}
			label := botMessageStyle.Render("Bot:")
			sb.WriteString(label + "\n" + lipgloss.NewStyle().PaddingLeft(2).Render(rendered))
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

// textareaHeight computes how tall the textarea should be based on content,
// clamped between 1 and 10 lines.
func (m *Model) textareaHeight() int {
	value := m.textarea.Value()
	if value == "" {
		return 1
	}

	// The textarea internally wraps using its width minus the prompt width.
	// Prompt "┃ " is 2 runes wide.
	promptWidth := 2
	wrapWidth := m.textarea.Width() - promptWidth
	if wrapWidth <= 0 {
		wrapWidth = 1
	}

	total := 0
	for _, line := range strings.Split(value, "\n") {
		runeLen := len([]rune(line))
		if runeLen == 0 {
			total++
		} else {
			total += (runeLen + wrapWidth - 1) / wrapWidth
		}
	}

	if total < 1 {
		total = 1
	}
	if total > 10 {
		total = 10
	}
	return total
}

// resizeTextarea adjusts textarea height to fit content and shrinks viewport accordingly.
func (m *Model) resizeTextarea() {
	h := m.textareaHeight()
	m.textarea.SetHeight(h)

	// header(~3) + footer(info ~2) + textarea(h) + padding(~3)
	overhead := 3 + 2 + h + 3
	vpHeight := m.height - overhead
	if vpHeight < 4 {
		vpHeight = 4
	}
	m.viewport.Height = vpHeight
}

// estimateTokens provides a rough estimate of tokens in text
// Claude's tokenizer averages ~4 characters per token
func estimateTokens(text string) int {
	return len(text) / 4
}

// GetBot returns the bot instance
func (m *Model) GetBot() *chatbot.Bot {
	return m.bot
}

func (m Model) renderSidebar() string {
	totalTokens := m.inputTokens + m.outputTokens
	remainingTokens := m.maxTokens - totalTokens
	percentUsed := float64(totalTokens) / float64(m.maxTokens) * 100

	title := sidebarTitleStyle.Render("📊 Token Usage")
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
		sidebarLabelStyle.Render("🤖 Model:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(fmt.Sprintf("  %s", modelName)),
	)

	// Apply sidebar style with proper height constraint
	return sidebarStyle.
		Height(m.height - 2).
		Render(content)
}
