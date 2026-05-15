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
	agentName      string
	mdRenderer     *glamour.TermRenderer
	lastKeyTime    time.Time
}

func NewModel(bot *chatbot.Bot) Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 2000
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(true)

	ta.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("235"))
	ta.BlurredStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))

	vp := viewport.New(80, 20)
	vp.SetContent("")

	currentModel, _ := settings.LoadSelectedModel()
	maxTokens := 200000
	if currentModel != nil {
		maxTokens = currentModel.MaxTokens
	}

	mdRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	cmdHandler := commands.NewHandler()

	cmdHandler.RegisterCommand("status", func(args []string) commands.CommandResult {
		s := bot.Sessions.ActiveSession()
		if s == nil {
			return commands.CommandResult{IsCommand: true, Message: "No active session"}
		}
		info := fmt.Sprintf("Agent: %s\nSession: %s", s.AgentID, s.SessionID)
		return commands.CommandResult{IsCommand: true, Message: info}
	})
	cmdHandler.RegisterCommand("abort", func(args []string) commands.CommandResult {
		s := bot.Sessions.ActiveSession()
		if s == nil || bot.Client() == nil {
			return commands.CommandResult{IsCommand: true, ErrorMessage: "No active session"}
		}
		if err := bot.Client().AbortAgent(s.SessionID); err != nil {
			return commands.CommandResult{IsCommand: true, ErrorMessage: err.Error()}
		}
		return commands.CommandResult{IsCommand: true, Message: "Session aborted"}
	})

	return Model{
		textarea:       ta,
		viewport:       vp,
		bot:            bot,
		cmdHandler:     cmdHandler,
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

func (m *Model) ReloadModelSettings() {
	currentModel, _ := settings.LoadSelectedModel()
	maxTokens := 200000
	if currentModel != nil {
		maxTokens = currentModel.MaxTokens
	}
	m.currentModel = currentModel
	m.maxTokens = maxTokens
}

func (m Model) isFullScreen() bool {
	return m.width >= 120
}

func (m *Model) SetAgentName(name string) {
	m.agentName = name
}

func (m *Model) SetActiveSession(sessionID string) {
	m.bot.Sessions.SetActive(sessionID)
	s := m.bot.Sessions.ActiveSession()
	if s != nil {
		m.agentName = s.AgentName
	}
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
}

func (m Model) activeMessages() []chatbot.Message {
	s := m.bot.Sessions.ActiveSession()
	if s == nil {
		return nil
	}
	return s.Messages
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

		canShowSidebar := m.isFullScreen()
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

		mdWidth := mainWidth - 8
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
		if msg.Type != tea.KeyEnter {
			m.lastKeyTime = time.Now()
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyCtrlN:
			if !m.isFullScreen() {
				return m, nil
			}
			m.sidebarVisible = !m.sidebarVisible
			sidebarWidth := 0
			if m.sidebarVisible {
				sidebarWidth = 34
			}
			mainWidth := m.width - sidebarWidth
			m.viewport.Width = mainWidth
			m.textarea.SetWidth(mainWidth - 4)
			m.viewport.SetContent(m.renderMessages())
			return m, nil

		case tea.KeyEnter:
			if msg.Alt {
				break
			}
			if time.Since(m.lastKeyTime) < 50*time.Millisecond {
				break
			}

			userInput := strings.TrimSpace(m.textarea.Value())
			if userInput == "" {
				break
			}

			cmdResult := m.cmdHandler.Process(userInput)
			if cmdResult.IsCommand {
				s := m.bot.Sessions.ActiveSession()
				if s != nil {
					if cmdResult.ErrorMessage != "" {
						s.AddUserMessage(userInput)
						s.Messages = append(s.Messages, chatbot.Message{Role: chatbot.RoleBot, Content: "❌ " + cmdResult.ErrorMessage})
					} else if cmdResult.Message != "" {
						s.AddUserMessage(userInput)
						s.Messages = append(s.Messages, chatbot.Message{Role: chatbot.RoleBot, Content: cmdResult.Message})
					}
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
			s := m.bot.Sessions.ActiveSession()
			if s == nil {
				break
			}
			s.AddUserMessage(userInput)
			m.inputTokens += estimateTokens(userInput)
			s.StartStreaming(nil) // adds "Thinking..." placeholder

			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			m.textarea.Reset()

			return m, m.bot.StartStream(s.SessionID, userInput)
		}

	// Session-tagged stream events — these are routed here by coordinator
	// only for the active session's viewport update
	case chatbot.SessionStreamChunkMsg:
		m.outputTokens += estimateTokens(msg.Chunk)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case chatbot.SessionStreamThinkingMsg:
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case chatbot.SessionStreamToolStartMsg:
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case chatbot.SessionStreamToolEndMsg:
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case chatbot.SessionStreamDoneMsg:
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case chatbot.SessionStreamErrorMsg:
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.resizeTextarea()

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	messages := m.activeMessages()
	if len(messages) == 0 {
		return m.renderInitialView()
	}

	canShowSidebar := m.isFullScreen() && m.sidebarVisible
	sidebarWidth := 0
	if canShowSidebar {
		sidebarWidth = 34
	}
	mainWidth := m.width - sidebarWidth

	header := m.renderHeader(mainWidth)
	content := m.viewport.View()
	footer := m.renderFooter()

	mainContent := lipgloss.NewStyle().
		Width(mainWidth).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			content,
			footer,
		))

	if canShowSidebar {
		sidebar := m.renderSidebar()
		return lipgloss.JoinHorizontal(lipgloss.Top, mainContent, sidebar)
	}

	return mainContent
}

func (m Model) renderInitialView() string {
	titleText := titleStyle.Render("🤖 " + m.displayTitle())
	welcomeMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Start a conversation...")
	inputArea := m.textarea.View()
	info := infoStyle.Render("Enter: Send | Alt+Enter: New Line | Esc: Quit")

	centeredContent := lipgloss.JoinVertical(
		lipgloss.Center, titleText, "", "", welcomeMsg, "", "", inputArea, info,
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, centeredContent)
}

func (m Model) renderHeader(width int) string {
	titleText := titleStyle.Render("🤖 " + m.displayTitle())
	title := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(titleText)
	line := strings.Repeat("─", max(0, width-2))
	return lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(line))
}


func (m Model) renderMessages() string {
	messages := m.activeMessages()
	if messages == nil {
		return ""
	}

	var sb strings.Builder
	sidebarWidth := 0
	if m.sidebarVisible {
		sidebarWidth = 34
	}
	maxWidth := m.width - sidebarWidth - 6

	for i, msg := range messages {
		if i > 0 {
			sb.WriteString("\n\n")
		}

		if msg.Role == chatbot.RoleUser {
			wrappedContent := wrapText("You: "+msg.Content, maxWidth)
			sb.WriteString(userMessageStyle.Render(wrappedContent))
		} else {
			content := msg.Content
			var rendered string
			rest := content

			for {
				thinkIdx := strings.Index(rest, "{{THINKING}}")
				if thinkIdx < 0 {
					break
				}
				before := rest[:thinkIdx]
				if before != "" {
					if m.mdRenderer != nil {
						if md, err := m.mdRenderer.Render(before); err == nil {
							rendered += strings.TrimSpace(md) + "\n"
						} else {
							rendered += before
						}
					} else {
						rendered += before
					}
				}
				rest = rest[thinkIdx+len("{{THINKING}}"):]
				endIdx := strings.Index(rest, "{{/THINKING}}")
				if endIdx >= 0 {
					rendered += m.renderThinkingBlock(rest[:endIdx]) + "\n"
					rest = rest[endIdx+len("{{/THINKING}}"):]
				} else {
					rendered += m.renderThinkingBlock(rest)
					rest = ""
					break
				}
			}

			if rest != "" {
				if m.mdRenderer != nil {
					if md, err := m.mdRenderer.Render(rest); err == nil {
						rendered += strings.TrimSpace(md)
					} else {
						rendered += rest
					}
				} else {
					rendered += rest
				}
			}

			label := botMessageStyle.Render("Bot:")
			sb.WriteString(label + "\n" + lipgloss.NewStyle().PaddingLeft(2).Render(rendered))
		}
	}

	return sb.String()
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	var result strings.Builder
	var currentLine strings.Builder
	words := strings.Fields(text)
	for i, word := range words {
		if currentLine.Len()+len(word)+1 > width {
			if currentLine.Len() > 0 {
				result.WriteString(currentLine.String())
				result.WriteString("\n")
				currentLine.Reset()
			}
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
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
		if i == len(words)-1 && currentLine.Len() > 0 {
			result.WriteString(currentLine.String())
		}
	}
	return result.String()
}

func (m Model) renderFooter() string {
	msgCount := 0
	if msgs := m.activeMessages(); msgs != nil {
		msgCount = len(msgs)
	}
	info := infoStyle.Render(
		"Enter: Send | Alt+Enter: New Line | /exit: Quit | Esc: Quit | Messages: " +
			lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", msgCount)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, "", m.textarea.View(), info)
}

func (m Model) renderThinkingBlock(text string) string {
	width := m.width - 12
	if width < 30 {
		width = 30
	}
	bar := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Render(strings.Repeat("─", width))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).Render("💭 Thinking")
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true).PaddingLeft(1).Width(width).Render(text)
	return bar + "\n" + label + "\n" + body + "\n" + bar
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) displayTitle() string {
	if m.agentName != "" {
		return m.agentName
	}
	return "ChatBot TUI"
}

func (m *Model) textareaHeight() int {
	value := m.textarea.Value()
	if value == "" {
		return 1
	}
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

func (m *Model) resizeTextarea() {
	h := m.textareaHeight()
	m.textarea.SetHeight(h)
	overhead := 3 + 2 + h + 3
	vpHeight := m.height - overhead
	if vpHeight < 4 {
		vpHeight = 4
	}
	m.viewport.Height = vpHeight
}

func estimateTokens(text string) int {
	return len(text) / 4
}

func (m *Model) GetBot() *chatbot.Bot {
	return m.bot
}

func (m Model) renderSidebar() string {
	totalTokens := m.inputTokens + m.outputTokens
	remainingTokens := m.maxTokens - totalTokens
	percentUsed := float64(totalTokens) / float64(m.maxTokens) * 100

	title := sidebarTitleStyle.Render("📊 Token Usage")
	divider := strings.Repeat("─", 26)

	modelName := "Claude Sonnet 4.5"
	if m.currentModel != nil {
		modelName = m.currentModel.Name
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
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

	return sidebarStyle.Height(m.height - 2).Render(content)
}
