package agentlist

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/chatbot-tui/pkg/chatbot"
	"github.com/yourusername/chatbot-tui/pkg/runtime"
)

type AgentInfo struct {
	ID          string
	Name        string
	Description string
	Model       string
}

// AgentSelectedMsg is emitted when the user selects an agent to create a new session.
type AgentSelectedMsg struct {
	AgentID     string
	Name        string
	Description string
	Model       string
}

// SessionSelectedMsg is emitted when the user selects an existing session.
type SessionSelectedMsg struct {
	SessionID string
}

type agentsLoadedMsg struct {
	agents []AgentInfo
}

type agentsErrorMsg struct {
	err error
}

// listItem represents one row in the flat navigation list.
type listItem struct {
	kind      string // "agent" or "session"
	agentIdx  int
	sessionID string
}

type Model struct {
	client        *runtime.Client
	agents        []AgentInfo
	openSessions  []*chatbot.Session
	flatItems     []listItem
	selectedIndex int
	width         int
	height        int
	err           error
	loading       bool
}

func NewModel(client *runtime.Client, sessions []*chatbot.Session) Model {
	return Model{
		client:       client,
		agents:       []AgentInfo{},
		openSessions: sessions,
		selectedIndex: 0,
		loading:      true,
	}
}

func (m *Model) buildFlatItems() {
	m.flatItems = nil
	for i, agent := range m.agents {
		m.flatItems = append(m.flatItems, listItem{kind: "agent", agentIdx: i})
		// Add open sessions for this agent
		for _, s := range m.openSessions {
			if s.AgentID == agent.ID {
				m.flatItems = append(m.flatItems, listItem{kind: "session", agentIdx: i, sessionID: s.SessionID})
			}
		}
	}
}

func (m Model) Init() tea.Cmd {
	if m.client == nil {
		return nil
	}
	return func() tea.Msg {
		agentDataList, err := m.client.ListAgents()
		if err != nil {
			return agentsErrorMsg{err: err}
		}

		var agents []AgentInfo
		for _, ad := range agentDataList {
			agents = append(agents, AgentInfo{
				ID:          ad.ID,
				Name:        ad.Name,
				Description: ad.Description,
				Model:       ad.Model,
			})
		}

		return agentsLoadedMsg{agents: agents}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case agentsLoadedMsg:
		m.agents = msg.agents
		m.loading = false
		m.err = nil
		m.buildFlatItems()

	case agentsErrorMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "ctrl+g":
			return m, tea.Quit

		case "up", "k":
			if !m.loading && m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if !m.loading && m.selectedIndex < len(m.flatItems)-1 {
				m.selectedIndex++
			}

		case "enter":
			if !m.loading && len(m.flatItems) > 0 && m.selectedIndex < len(m.flatItems) {
				item := m.flatItems[m.selectedIndex]
				if item.kind == "agent" {
					agent := m.agents[item.agentIdx]
					return m, func() tea.Msg {
						return AgentSelectedMsg{
							AgentID:     agent.ID,
							Name:        agent.Name,
							Description: agent.Description,
							Model:       agent.Model,
						}
					}
				} else if item.kind == "session" {
					sid := item.sessionID
					return m, func() tea.Msg {
						return SessionSelectedMsg{SessionID: sid}
					}
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.loading {
		content := loadingStyle.Render("⏳ Fetching agents from server...\n\nPress Ctrl+G or Esc to return to chat.")
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(2, 4).Render(content)
	}

	if m.err != nil {
		content := errorStyle.Render(fmt.Sprintf("Error loading agents: %v\n\nPress Ctrl+G or Esc to return to chat.", m.err))
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(2, 4).Render(content)
	}

	if len(m.flatItems) == 0 {
		content := errorStyle.Render("No agents available.\n\nPress Ctrl+G or Esc to return to chat.")
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(2, 4).Render(content)
	}

	header := titleStyle.Render("🤖 Available Agents")
	subtitle := subtitleStyle.Render("Enter on agent = new session | Enter on session = switch to it")

	var listContent string
	modelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	sessionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(4)
	sessionActiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).PaddingLeft(4)
	streamingBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)

	for idx, item := range m.flatItems {
		isSelected := idx == m.selectedIndex

		if item.kind == "agent" {
			agent := m.agents[item.agentIdx]
			cursor := "  "
			if isSelected {
				cursor = "▶ "
				nameStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).
					Render(fmt.Sprintf("%s%s", cursor, agent.Name))
				desc := descriptionStyle.Render(fmt.Sprintf("  %s", agent.Description))
				model := modelStyle.Render(fmt.Sprintf("  %s", agent.Model))
				agentContent := nameStyled + "\n" + desc + "\n" + model
				bordered := lipgloss.NewStyle().Padding(0, 1).
					Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("86")).
					Render(agentContent)
				listContent += bordered + "\n"
			} else {
				name := fmt.Sprintf("%s%s", cursor, agent.Name)
				desc := fmt.Sprintf("  %s", agent.Description)
				model := fmt.Sprintf("  %s", agent.Model)
				listContent += agentItemStyle.Render(name) + "\n"
				listContent += descriptionStyle.Render(desc) + "\n"
				listContent += modelStyle.Render(model) + "\n"
			}
		} else if item.kind == "session" {
			// Find the session
			var session *chatbot.Session
			for _, s := range m.openSessions {
				if s.SessionID == item.sessionID {
					session = s
					break
				}
			}
			if session == nil {
				continue
			}

			msgCount := len(session.Messages)
			label := fmt.Sprintf("💬 Session (%d msgs)", msgCount)
			if session.IsStreaming {
				label += " " + streamingBadge.Render("● streaming")
			}

			if isSelected {
				bordered := lipgloss.NewStyle().Padding(0, 1).MarginLeft(4).
					Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).
					Render(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("▶ " + label))
				listContent += bordered + "\n"
			} else {
				listContent += sessionStyle.Render("  "+label) + "\n"
			}
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header, subtitle, "",
		dividerStyle.Render(strings.Repeat("─", 53)), "",
		listContent,
	)

	_ = sessionActiveStyle // reserved for future use

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(2, 4).Render(content)
}
