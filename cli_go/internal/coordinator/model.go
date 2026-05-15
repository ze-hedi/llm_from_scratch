package coordinator

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/chatbot-tui/extensions/dino/game"
	dinoTui "github.com/yourusername/chatbot-tui/extensions/dino/tui"
	"github.com/yourusername/chatbot-tui/extensions/tamagotchi/pet"
	tamagotchiTui "github.com/yourusername/chatbot-tui/extensions/tamagotchi/tui"
	"github.com/yourusername/chatbot-tui/internal/agentlist"
	"github.com/yourusername/chatbot-tui/internal/extensions"
	"github.com/yourusername/chatbot-tui/internal/settings"
	"github.com/yourusername/chatbot-tui/internal/systemprompt"
	"github.com/yourusername/chatbot-tui/internal/tui"
	"github.com/yourusername/chatbot-tui/pkg/chatbot"
	"github.com/yourusername/chatbot-tui/pkg/runtime"
)

type ViewType int

const (
	ChatView ViewType = iota
	ExtensionsView
	TamagotchiView
	DinoView
	SettingsView
	SystemPromptView
	AgentListView
)

type SwitchToExtensionsMsg struct{}
type SwitchToChatMsg struct{}
type LaunchExtensionMsg struct{ ExtensionID string }

type sessionCreatedMsg struct {
	SessionID string
	AgentID   string
	Name      string
}

type sessionErrorMsg struct {
	Err error
}

type Model struct {
	currentView       ViewType
	chatModel         tui.Model
	chatReady         bool
	extensionsModel   extensions.Model
	tamagotchiModel   tamagotchiTui.Model
	dinoModel         dinoTui.Model
	settingsModel     settings.SettingsModel
	systemPromptModel systemprompt.SystemPromptModel
	agentListModel    agentlist.Model
	client            *runtime.Client
	bot               *chatbot.Bot
	width             int
	height            int
}

func NewModel() Model {
	client := runtime.NewClient("http://localhost:5000", "http://localhost:4000")
	bot := chatbot.NewBot(client)
	return Model{
		currentView:     AgentListView,
		agentListModel:  agentlist.NewModel(client, nil),
		extensionsModel: extensions.NewModel(),
		client:          client,
		bot:             bot,
	}
}

func (m Model) Init() tea.Cmd {
	return m.agentListModel.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// --- Session-tagged stream events: update session, re-render if active, continue listening ---
	switch msg := msg.(type) {
	case chatbot.SessionStreamChunkMsg:
		s := m.bot.Sessions.GetSession(msg.SessionID)
		if s != nil {
			s.AppendChunk(msg.Chunk)
		}
		if m.chatReady && m.bot.Sessions.ActiveID() == msg.SessionID && m.currentView == ChatView {
			var tmpModel tea.Model
			tmpModel, _ = m.chatModel.Update(msg)
			m.chatModel = tmpModel.(tui.Model)
		}
		return m, m.bot.ListenToSession(msg.SessionID)

	case chatbot.SessionStreamThinkingMsg:
		s := m.bot.Sessions.GetSession(msg.SessionID)
		if s != nil {
			s.AppendThinking(msg.Text)
		}
		if m.chatReady && m.bot.Sessions.ActiveID() == msg.SessionID && m.currentView == ChatView {
			var tmpModel tea.Model
			tmpModel, _ = m.chatModel.Update(msg)
			m.chatModel = tmpModel.(tui.Model)
		}
		return m, m.bot.ListenToSession(msg.SessionID)

	case chatbot.SessionStreamToolStartMsg:
		s := m.bot.Sessions.GetSession(msg.SessionID)
		if s != nil {
			s.AppendToolStart(msg.Name, msg.Args)
		}
		if m.chatReady && m.bot.Sessions.ActiveID() == msg.SessionID && m.currentView == ChatView {
			var tmpModel tea.Model
			tmpModel, _ = m.chatModel.Update(msg)
			m.chatModel = tmpModel.(tui.Model)
		}
		return m, m.bot.ListenToSession(msg.SessionID)

	case chatbot.SessionStreamToolEndMsg:
		s := m.bot.Sessions.GetSession(msg.SessionID)
		if s != nil {
			s.AppendToolEnd(msg.Name, msg.Result, msg.IsError)
		}
		if m.chatReady && m.bot.Sessions.ActiveID() == msg.SessionID && m.currentView == ChatView {
			var tmpModel tea.Model
			tmpModel, _ = m.chatModel.Update(msg)
			m.chatModel = tmpModel.(tui.Model)
		}
		return m, m.bot.ListenToSession(msg.SessionID)

	case chatbot.SessionStreamDoneMsg:
		s := m.bot.Sessions.GetSession(msg.SessionID)
		if s != nil {
			s.FinishStreaming()
		}
		if m.chatReady && m.bot.Sessions.ActiveID() == msg.SessionID && m.currentView == ChatView {
			var tmpModel tea.Model
			tmpModel, _ = m.chatModel.Update(msg)
			m.chatModel = tmpModel.(tui.Model)
		}
		return m, nil // stream ended, no more listening

	case chatbot.SessionStreamErrorMsg:
		s := m.bot.Sessions.GetSession(msg.SessionID)
		if s != nil {
			if len(s.Messages) > 0 {
				s.Messages[len(s.Messages)-1].Content = fmt.Sprintf("Error: %v", msg.Err)
			}
			s.FinishStreaming()
		}
		if m.chatReady && m.bot.Sessions.ActiveID() == msg.SessionID && m.currentView == ChatView {
			var tmpModel tea.Model
			tmpModel, _ = m.chatModel.Update(msg)
			m.chatModel = tmpModel.(tui.Model)
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		var tmpModel tea.Model
		if m.chatReady {
			tmpModel, _ = m.chatModel.Update(msg)
			m.chatModel = tmpModel.(tui.Model)
		}
		tmpModel, _ = m.extensionsModel.Update(msg)
		m.extensionsModel = tmpModel.(extensions.Model)
		if m.currentView == TamagotchiView {
			tmpModel, _ = m.tamagotchiModel.Update(msg)
			m.tamagotchiModel = tmpModel.(tamagotchiTui.Model)
		}
		if m.currentView == DinoView {
			tmpModel, _ = m.dinoModel.Update(msg)
			m.dinoModel = tmpModel.(dinoTui.Model)
		}
		if m.currentView == SettingsView {
			tmpModel, _ = m.settingsModel.Update(msg)
			m.settingsModel = tmpModel.(settings.SettingsModel)
		}
		if m.currentView == SystemPromptView {
			tmpModel, _ = m.systemPromptModel.Update(msg)
			m.systemPromptModel = tmpModel.(systemprompt.SystemPromptModel)
		}
		if m.currentView == AgentListView {
			tmpModel, _ = m.agentListModel.Update(msg)
			m.agentListModel = tmpModel.(agentlist.Model)
		}

	case tea.KeyMsg:
		if msg.String() == "ctrl+g" {
			if m.currentView == ChatView {
				m.currentView = AgentListView
				m.agentListModel = agentlist.NewModel(m.client, m.bot.Sessions.AllSessions())
				return m, m.agentListModel.Init()
			} else if m.currentView == AgentListView && m.chatReady {
				m.currentView = ChatView
				return m, nil
			}
		}

		if msg.String() == "ctrl+a" {
			switch m.currentView {
			case ChatView:
				m.currentView = ExtensionsView
				m.extensionsModel = extensions.NewModel()
				return m, m.extensionsModel.Init()
			case ExtensionsView, TamagotchiView, DinoView:
				if m.currentView == DinoView {
					game.SaveHighScore(m.dinoModel.GetHighScore())
				}
				m.currentView = ChatView
				return m, nil
			}
		}

		if msg.String() == "ctrl+y" {
			if m.currentView == ChatView {
				m.currentView = SettingsView
				m.settingsModel = settings.NewSettingsModel()
				return m, m.settingsModel.Init()
			} else if m.currentView == SettingsView {
				m.currentView = ChatView
				return m, nil
			}
		}

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "esc" {
			if m.currentView == ChatView || !m.chatReady {
				return m, tea.Quit
			}
			m.currentView = ChatView
			return m, nil
		}
	}

	// Handle agent selection — create a new session
	if sel, ok := msg.(agentlist.AgentSelectedMsg); ok {
		client := m.client
		return m, func() tea.Msg {
			resp, err := client.Run(runtime.RunRequest{
				Agent: runtime.AgentData{
					ID:          sel.AgentID,
					Name:        sel.Name,
					Model:       sel.Model,
					Description: sel.Description,
				},
			})
			if err != nil {
				return sessionErrorMsg{Err: fmt.Errorf("create session: %w", err)}
			}
			return sessionCreatedMsg{
				SessionID: resp.SessionID,
				AgentID:   resp.AgentID,
				Name:      resp.Name,
			}
		}
	}

	// Handle orchestrator selection — create a new orchestrator session
	if sel, ok := msg.(agentlist.OrchestratorSelectedMsg); ok {
		client := m.client
		return m, func() tea.Msg {
			// Build the agents array with stateful flag from SubAgentEntries
			agents := make([]runtime.AgentData, len(sel.SubAgents))
			for i, sa := range sel.SubAgents {
				agents[i] = sa.Agent
				agents[i].Stateful = sa.Stateful
			}
			resp, err := client.RunOrchestrator(runtime.OrchestratorRunRequest{
				OrchestratorID: sel.OrchestratorID,
				Model:          sel.Model,
				Playground:     sel.Playground,
				SystemPrompt:   "",
				Agents:         agents,
			})
			if err != nil {
				return sessionErrorMsg{Err: fmt.Errorf("create orchestrator session: %w", err)}
			}
			return sessionCreatedMsg{
				SessionID: resp.SessionID,
				AgentID:   resp.OrchestratorID,
				Name:      sel.Name,
			}
		}
	}

	// Handle session selection — switch to existing session
	if sel, ok := msg.(agentlist.SessionSelectedMsg); ok {
		if !m.chatReady {
			m.chatModel = tui.NewModel(m.bot)
			m.chatReady = true
			if m.width > 0 && m.height > 0 {
				var tmpModel tea.Model
				tmpModel, _ = m.chatModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				m.chatModel = tmpModel.(tui.Model)
			}
		}
		m.chatModel.SetActiveSession(sel.SessionID)
		m.currentView = ChatView
		return m, m.chatModel.Init()
	}

	// Handle session created
	if created, ok := msg.(sessionCreatedMsg); ok {
		m.bot.Sessions.AddSession(created.SessionID, created.AgentID, created.Name)
		m.bot.Sessions.SetActive(created.SessionID)

		if !m.chatReady {
			m.chatModel = tui.NewModel(m.bot)
			m.chatReady = true
		}
		m.chatModel.SetActiveSession(created.SessionID)
		if m.width > 0 && m.height > 0 {
			var tmpModel tea.Model
			tmpModel, _ = m.chatModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.chatModel = tmpModel.(tui.Model)
		}
		m.currentView = ChatView
		return m, m.chatModel.Init()
	}

	if _, ok := msg.(sessionErrorMsg); ok {
		return m, nil
	}

	// Route to active view
	switch m.currentView {
	case ChatView:
		var tmpModel tea.Model
		tmpModel, cmd = m.chatModel.Update(msg)
		m.chatModel = tmpModel.(tui.Model)
		return m, cmd

	case ExtensionsView:
		var newModel tea.Model
		newModel, cmd = m.extensionsModel.Update(msg)
		m.extensionsModel = newModel.(extensions.Model)
		if selectedExt := m.extensionsModel.SelectedExtension(); selectedExt != nil {
			switch selectedExt.Command {
			case "tamagotchi":
				p, err := pet.LoadPet()
				if err != nil {
					p = pet.NewPet("Mochi", pet.PetTypeCat)
				}
				m.tamagotchiModel = tamagotchiTui.NewModelWithPet(p)
				m.currentView = TamagotchiView
				return m, m.tamagotchiModel.Init()
			case "dino":
				highScore := game.LoadHighScore()
				m.dinoModel = dinoTui.NewModel(highScore)
				m.currentView = DinoView
				return m, m.dinoModel.Init()
			}
		}
		return m, cmd

	case TamagotchiView:
		var newModel tea.Model
		newModel, cmd = m.tamagotchiModel.Update(msg)
		m.tamagotchiModel = newModel.(tamagotchiTui.Model)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			if p := m.tamagotchiModel.GetPet(); p != nil {
				pet.SavePet(p)
			}
		}
		return m, cmd

	case DinoView:
		var newModel tea.Model
		newModel, cmd = m.dinoModel.Update(msg)
		m.dinoModel = newModel.(dinoTui.Model)
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" {
				game.SaveHighScore(m.dinoModel.GetHighScore())
				m.currentView = ChatView
				return m, nil
			}
		}
		return m, cmd

	case SettingsView:
		var newModel tea.Model
		newModel, cmd = m.settingsModel.Update(msg)
		m.settingsModel = newModel.(settings.SettingsModel)
		if m.settingsModel.Confirmed() {
			m.currentView = ChatView
			m.chatModel.ReloadModelSettings()
			return m, nil
		}
		return m, cmd

	case SystemPromptView:
		var newModel tea.Model
		newModel, cmd = m.systemPromptModel.Update(msg)
		m.systemPromptModel = newModel.(systemprompt.SystemPromptModel)
		if m.systemPromptModel.Confirmed() {
			m.currentView = ChatView
			return m, nil
		}
		return m, cmd

	case AgentListView:
		var newModel tea.Model
		newModel, cmd = m.agentListModel.Update(msg)
		m.agentListModel = newModel.(agentlist.Model)
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" || keyMsg.String() == "q" || keyMsg.String() == "ctrl+g" {
				if m.chatReady {
					m.currentView = ChatView
					return m, nil
				}
				return m, tea.Quit
			}
		}
		return m, cmd

	}

	return m, cmd
}

func (m Model) View() string {
	switch m.currentView {
	case ChatView:
		return m.chatModel.View()
	case ExtensionsView:
		return m.extensionsModel.View()
	case TamagotchiView:
		return m.tamagotchiModel.View()
	case DinoView:
		return m.dinoModel.View()
	case SettingsView:
		return m.settingsModel.View()
	case SystemPromptView:
		return m.systemPromptModel.View()
	case AgentListView:
		return m.agentListModel.View()
	default:
		return "Unknown view"
	}
}

func (m Model) GetChatModel() tui.Model {
	return m.chatModel
}
