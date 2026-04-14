package coordinator

import (
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
)

// ViewType represents the current active view
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

// SwitchToExtensionsMsg signals to show extensions browser
type SwitchToExtensionsMsg struct{}

// SwitchToChatMsg signals to return to chat
type SwitchToChatMsg struct{}

// LaunchExtensionMsg signals to launch a specific extension
type LaunchExtensionMsg struct {
	ExtensionID string
}

// Model is the coordinator that manages different views
type Model struct {
	currentView       ViewType
	chatModel         tui.Model
	extensionsModel   extensions.Model
	tamagotchiModel   tamagotchiTui.Model
	dinoModel         dinoTui.Model
	settingsModel     settings.SettingsModel
	systemPromptModel systemprompt.SystemPromptModel
	agentListModel    agentlist.Model
	width             int
	height            int
}

// NewModel creates a new coordinator model starting with chat
func NewModel() Model {
	return Model{
		currentView:     ChatView,
		chatModel:       tui.NewModel(),
		extensionsModel: extensions.NewModel(),
	}
}

func (m Model) Init() tea.Cmd {
	// Initialize the chat view
	return m.chatModel.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propagate to all models
		var tmpModel tea.Model
		tmpModel, _ = m.chatModel.Update(msg)
		m.chatModel = tmpModel.(tui.Model)
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
		// Global Ctrl+G handling for Agent List
		if msg.String() == "ctrl+g" {
			if m.currentView == ChatView {
				// Switch to agent list view
				m.currentView = AgentListView
				m.agentListModel = agentlist.NewModel(m.chatModel.GetBot())
				return m, m.agentListModel.Init()
			} else if m.currentView == AgentListView {
				// Switch back to chat
				m.currentView = ChatView
				return m, nil
			}
		}

		// Global Ctrl+A handling
		if msg.String() == "ctrl+a" {
			switch m.currentView {
			case ChatView:
				// Switch to extensions browser
				m.currentView = ExtensionsView
				m.extensionsModel = extensions.NewModel()
				return m, m.extensionsModel.Init()
			case ExtensionsView, TamagotchiView, DinoView:
				// Switch back to chat
				// Save dino high score if coming from dino game
				if m.currentView == DinoView {
					game.SaveHighScore(m.dinoModel.GetHighScore())
				}
				m.currentView = ChatView
				return m, nil
			}
		}

		// Global Ctrl+Y handling for Settings
		if msg.String() == "ctrl+y" {
			if m.currentView == ChatView {
				// Switch to settings
				m.currentView = SettingsView
				m.settingsModel = settings.NewSettingsModel()
				return m, m.settingsModel.Init()
			} else if m.currentView == SettingsView {
				// Switch back to chat
				m.currentView = ChatView
				return m, nil
			}
		}

		// Global Ctrl+S handling for System Prompt
		if msg.String() == "ctrl+s" {
			if m.currentView == ChatView {
				// Switch to system prompt editor
				m.currentView = SystemPromptView
				currentPrompt := m.chatModel.GetBot().SystemPrompt
				m.systemPromptModel = systemprompt.NewSystemPromptModel(currentPrompt)
				return m, m.systemPromptModel.Init()
			} else if m.currentView == SystemPromptView {
				// Switch back to chat
				m.currentView = ChatView
				return m, nil
			}
		}

		// Handle Ctrl+C and Esc globally
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// For Esc, only quit from chat view, otherwise go back to chat
		if msg.String() == "esc" {
			if m.currentView == ChatView {
				return m, tea.Quit
			} else {
				// Return to chat (cancel any changes)
				m.currentView = ChatView
				return m, nil
			}
		}
	}

	// Route messages to the appropriate view
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

		// Check if an extension was selected
		if selectedExt := m.extensionsModel.SelectedExtension(); selectedExt != nil {
			switch selectedExt.Command {
			case "tamagotchi":
				// Load or create tamagotchi
				p, err := pet.LoadPet()
				if err != nil {
					// If no pet, we should show the choose screen first
					// For now, create a default pet
					p = pet.NewPet("Mochi", pet.PetTypeCat)
				}
				m.tamagotchiModel = tamagotchiTui.NewModelWithPet(p)
				m.currentView = TamagotchiView
				return m, m.tamagotchiModel.Init()
			case "dino":
				// Load high score and create dino game
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

		// Save pet state periodically
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			// Save after each command
			if p := m.tamagotchiModel.GetPet(); p != nil {
				pet.SavePet(p)
			}
		}
		return m, cmd

	case DinoView:
		var newModel tea.Model
		newModel, cmd = m.dinoModel.Update(msg)
		m.dinoModel = newModel.(dinoTui.Model)

		// Save high score when exiting (on quit)
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

		// Check if user confirmed selection (pressed Enter)
		if m.settingsModel.Confirmed() {
			// Return to chat after confirming model selection
			m.currentView = ChatView
			// Reload model settings without losing chat history
			m.chatModel.ReloadModelSettings()
			return m, nil
		}
		return m, cmd

	case SystemPromptView:
		var newModel tea.Model
		newModel, cmd = m.systemPromptModel.Update(msg)
		m.systemPromptModel = newModel.(systemprompt.SystemPromptModel)

		// Check if user confirmed (pressed Ctrl+S)
		if m.systemPromptModel.Confirmed() {
			// Update the bot's system prompt
			m.chatModel.GetBot().SystemPrompt = m.systemPromptModel.GetPrompt()
			// Return to chat
			m.currentView = ChatView
			return m, nil
		}
		return m, cmd

	case AgentListView:
		var newModel tea.Model
		newModel, cmd = m.agentListModel.Update(msg)
		m.agentListModel = newModel.(agentlist.Model)

		// Check if user wants to go back (pressed Esc, q, or Ctrl+G)
		// This is handled by the agent list model returning tea.Quit
		// which we intercept here to switch back to chat
		if _, ok := msg.(tea.KeyMsg); ok {
			if keyMsg := msg.(tea.KeyMsg); keyMsg.String() == "esc" || keyMsg.String() == "q" || keyMsg.String() == "ctrl+g" {
				m.currentView = ChatView
				return m, nil
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

// GetChatModel returns the chat model (for saving state, etc.)
func (m Model) GetChatModel() tui.Model {
	return m.chatModel
}
