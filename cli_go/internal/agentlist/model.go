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

type OrchestratorInfo struct {
	ID          string
	Name        string
	Description string
	Model       string
	Playground  string
	SubAgents   []runtime.SubAgentEntry
}

// AgentSelectedMsg is emitted when the user selects an agent to create a new session.
type AgentSelectedMsg struct {
	AgentID     string
	Name        string
	Description string
	Model       string
}

// OrchestratorSelectedMsg is emitted when the user selects an orchestrator.
type OrchestratorSelectedMsg struct {
	OrchestratorID string
	Name           string
	Model          string
	Playground     string
	SubAgents      []runtime.SubAgentEntry
}

// SessionSelectedMsg is emitted when the user selects an existing session.
type SessionSelectedMsg struct {
	SessionID string
}

type tabMode int

const (
	agentsTab tabMode = iota
	orchestratorsTab
)

type agentsLoadedMsg struct {
	agents []AgentInfo
}

type orchestratorsLoadedMsg struct {
	orchestrators []OrchestratorInfo
}

type loadErrorMsg struct {
	err error
}

// listItem represents one row in the flat navigation list.
type listItem struct {
	kind      string // "agent", "orchestrator", or "session"
	agentIdx  int
	orchIdx   int
	sessionID string
}

type Model struct {
	client        *runtime.Client
	agents        []AgentInfo
	orchestrators []OrchestratorInfo
	openSessions  []*chatbot.Session
	flatItems     []listItem
	selectedIndex int
	tab           tabMode
	width         int
	height        int
	err           error
	agentsLoaded  bool
	orchsLoaded   bool
	loading       bool
}

func NewModel(client *runtime.Client, sessions []*chatbot.Session) Model {
	return Model{
		client:        client,
		agents:        []AgentInfo{},
		orchestrators: []OrchestratorInfo{},
		openSessions:  sessions,
		selectedIndex: 0,
		tab:           agentsTab,
		loading:       true,
	}
}

func (m *Model) buildFlatItems() {
	m.flatItems = nil
	if m.tab == agentsTab {
		for i, agent := range m.agents {
			m.flatItems = append(m.flatItems, listItem{kind: "agent", agentIdx: i})
			for _, s := range m.openSessions {
				if s.AgentID == agent.ID {
					m.flatItems = append(m.flatItems, listItem{kind: "session", agentIdx: i, sessionID: s.SessionID})
				}
			}
		}
	} else {
		for i, orch := range m.orchestrators {
			m.flatItems = append(m.flatItems, listItem{kind: "orchestrator", orchIdx: i})
			for _, s := range m.openSessions {
				if s.AgentID == orch.ID {
					m.flatItems = append(m.flatItems, listItem{kind: "session", orchIdx: i, sessionID: s.SessionID})
				}
			}
		}
	}
	if m.selectedIndex >= len(m.flatItems) {
		m.selectedIndex = 0
	}
}

func (m Model) Init() tea.Cmd {
	if m.client == nil {
		return nil
	}
	client := m.client
	return tea.Batch(
		func() tea.Msg {
			agentDataList, err := client.ListAgents()
			if err != nil {
				return loadErrorMsg{err: err}
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
		},
		func() tea.Msg {
			orchDataList, err := client.ListOrchestrators()
			if err != nil {
				return loadErrorMsg{err: err}
			}
			var orchs []OrchestratorInfo
			for _, od := range orchDataList {
				orchs = append(orchs, OrchestratorInfo{
					ID:          od.ID,
					Name:        od.Name,
					Description: od.Description,
					Model:       od.Model,
					Playground:  od.Playground,
					SubAgents:   od.SubAgents,
				})
			}
			return orchestratorsLoadedMsg{orchestrators: orchs}
		},
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case agentsLoadedMsg:
		m.agents = msg.agents
		m.agentsLoaded = true
		if m.orchsLoaded {
			m.loading = false
		}
		m.err = nil
		m.buildFlatItems()

	case orchestratorsLoadedMsg:
		m.orchestrators = msg.orchestrators
		m.orchsLoaded = true
		if m.agentsLoaded {
			m.loading = false
		}
		m.err = nil
		m.buildFlatItems()

	case loadErrorMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "ctrl+g":
			return m, tea.Quit

		case "tab":
			if !m.loading {
				if m.tab == agentsTab {
					m.tab = orchestratorsTab
				} else {
					m.tab = agentsTab
				}
				m.selectedIndex = 0
				m.buildFlatItems()
			}

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
				switch item.kind {
				case "agent":
					agent := m.agents[item.agentIdx]
					return m, func() tea.Msg {
						return AgentSelectedMsg{
							AgentID:     agent.ID,
							Name:        agent.Name,
							Description: agent.Description,
							Model:       agent.Model,
						}
					}
				case "orchestrator":
					orch := m.orchestrators[item.orchIdx]
					return m, func() tea.Msg {
						return OrchestratorSelectedMsg{
							OrchestratorID: orch.ID,
							Name:           orch.Name,
							Model:          orch.Model,
							Playground:     orch.Playground,
							SubAgents:      orch.SubAgents,
						}
					}
				case "session":
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
		content := loadingStyle.Render("⏳ Fetching agents & orchestrators...\n\nPress Ctrl+G or Esc to return to chat.")
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(2, 4).Render(content)
	}

	if m.err != nil {
		content := errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress Ctrl+G or Esc to return to chat.", m.err))
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(2, 4).Render(content)
	}

	// Tab bar
	agentsLabel := "  Agents  "
	orchLabel := "  Orchestrators  "
	activeTabStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).
		Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(lipgloss.Color("86"))
	inactiveTabStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	var tabBar string
	if m.tab == agentsTab {
		tabBar = activeTabStyle.Render(agentsLabel) + "  " + inactiveTabStyle.Render(orchLabel)
	} else {
		tabBar = inactiveTabStyle.Render(agentsLabel) + "  " + activeTabStyle.Render(orchLabel)
	}

	subtitle := subtitleStyle.Render("Tab to switch | Enter to select | Enter on session to resume")

	if len(m.flatItems) == 0 {
		label := "agents"
		if m.tab == orchestratorsTab {
			label = "orchestrators"
		}
		content := lipgloss.JoinVertical(lipgloss.Left,
			tabBar, subtitle, "",
			errorStyle.Render(fmt.Sprintf("No %s available.", label)),
		)
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(2, 4).Render(content)
	}

	var listContent string
	modelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	sessionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(4)
	streamingBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	subAgentCountStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("213"))

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
		} else if item.kind == "orchestrator" {
			orch := m.orchestrators[item.orchIdx]
			cursor := "  "
			subCount := subAgentCountStyle.Render(fmt.Sprintf("  %d sub-agents", len(orch.SubAgents)))
			if isSelected {
				cursor = "▶ "
				nameStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213")).
					Render(fmt.Sprintf("%s%s", cursor, orch.Name))
				desc := descriptionStyle.Render(fmt.Sprintf("  %s", orch.Description))
				model := modelStyle.Render(fmt.Sprintf("  %s", orch.Model))
				orchContent := nameStyled + "\n" + desc + "\n" + model + "\n" + subCount
				bordered := lipgloss.NewStyle().Padding(0, 1).
					Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("213")).
					Render(orchContent)
				listContent += bordered + "\n"
			} else {
				name := fmt.Sprintf("%s%s", cursor, orch.Name)
				desc := fmt.Sprintf("  %s", orch.Description)
				model := fmt.Sprintf("  %s", orch.Model)
				listContent += agentItemStyle.Render(name) + "\n"
				listContent += descriptionStyle.Render(desc) + "\n"
				listContent += modelStyle.Render(model) + "\n"
				listContent += subCount + "\n"
			}
		} else if item.kind == "session" {
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
		tabBar, subtitle, "",
		dividerStyle.Render(strings.Repeat("─", 53)), "",
		listContent,
	)

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(2, 4).Render(content)
}
