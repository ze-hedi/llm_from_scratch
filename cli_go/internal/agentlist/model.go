package agentlist

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/chatbot-tui/pkg/chatbot"
)

type Model struct {
	agents        []chatbot.AgentInfo
	selectedIndex int
	width         int
	height        int
	err           error
	loading       bool
	bot           *chatbot.Bot
	showPopup     bool
	popupMessage  string
	agentResponse map[string]interface{}
	focusedOption int // 0, 1, or 2 for Option A, B, C
	popupViewport viewport.Model
	popupReady    bool
}

// HidePopupMsg is sent after a delay to hide the popup
type HidePopupMsg struct{}

func NewModel(bot *chatbot.Bot) Model {
	return Model{
		agents:        []chatbot.AgentInfo{},
		selectedIndex: 0,
		loading:       true,
		bot:           bot,
	}
}

func (m Model) Init() tea.Cmd {
	// Fetch agent list on initialization
	return m.bot.GetAgentList()
}

// hidePopupAfter returns a command that sends HidePopupMsg after a delay
func hidePopupAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return HidePopupMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case chatbot.AgentListMsg:
		// Successfully received agent list
		m.agents = msg.Agents
		m.loading = false
		m.err = nil

	case chatbot.AgentListErrorMsg:
		// Error fetching agent list
		m.err = msg.Err
		m.loading = false

	case HidePopupMsg:
		// Hide the popup after the timer expires
		m.showPopup = false
		m.popupMessage = ""
		m.agentResponse = nil

	case chatbot.SetAgentMsg:
		// Successfully received response from set_agent endpoint
		m.agentResponse = msg.Response

	case chatbot.SetAgentErrorMsg:
		// Error calling set_agent endpoint
		m.popupMessage = fmt.Sprintf("Error: %v", msg.Err)
		return m, hidePopupAfter(3 * time.Second)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "ctrl+g":
			// Return to chat - coordinator will handle this
			return m, tea.Quit

		case "up", "k":
			if !m.loading && len(m.agents) > 0 && m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if !m.loading && len(m.agents) > 0 && m.selectedIndex < len(m.agents)-1 {
				m.selectedIndex++
			}

		case "enter":
			// Call set_agent endpoint when Enter is pressed on an agent
			if !m.loading && len(m.agents) > 0 {
				agentName := m.agents[m.selectedIndex].Name
				m.showPopup = true
				m.popupMessage = fmt.Sprintf("loading %s ...", agentName)
				m.agentResponse = nil
				m.focusedOption = 0 // Reset to first option
				// Call the set_agent endpoint
				return m, m.bot.SetAgent(agentName)
			}

		case "f1":
			// Cycle through options when F1 is pressed (only when popup is shown)
			if m.showPopup {
				m.focusedOption = (m.focusedOption + 1) % 3
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.loading {
		content := loadingStyle.Render("⏳ Fetching agents from server...\n\nPress Ctrl+G or Esc to return to chat.")
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Padding(2, 4).
			Render(content)
	}

	if m.err != nil {
		content := errorStyle.Render(fmt.Sprintf("Error loading agents: %v\n\nPress Ctrl+G or Esc to return to chat.", m.err))
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Padding(2, 4).
			Render(content)
	}

	if len(m.agents) == 0 {
		content := errorStyle.Render("No agents available.\n\nPress Ctrl+G or Esc to return to chat.")
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Padding(2, 4).
			Render(content)
	}

	// Header
	header := titleStyle.Render("🤖 Available Agents")
	subtitle := subtitleStyle.Render("Browse available agents on the server")

	// Agents list
	var agentsList string
	for i, agent := range m.agents {
		cursor := "  "

		if i == m.selectedIndex {
			cursor = "▶ "
			// For selected item, apply border around the whole item but keep description styling normal
			nameStyled := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86")).
				Render(fmt.Sprintf("%s%s", cursor, agent.Name))
			description := descriptionStyle.Render(fmt.Sprintf("  %s", agent.Description))
			agentContent := nameStyled + "\n" + description

			// Wrap in border only
			bordered := lipgloss.NewStyle().
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("86")).
				Render(agentContent)
			agentsList += bordered + "\n\n"
		} else {
			// For non-selected items, use the normal style
			name := fmt.Sprintf("%s%s", cursor, agent.Name)
			description := fmt.Sprintf("  %s", agent.Description)
			agentsList += agentItemStyle.Render(name) + "\n"
			agentsList += descriptionStyle.Render(description) + "\n\n"
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		subtitle,
		"",
		dividerStyle.Render("─────────────────────────────────────────────────────"),
		"",
		agentsList,
	)

	baseView := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(2, 4).
		Render(content)

	// If popup is shown, overlay it on top of the base view
	if m.showPopup {
		// Render just the popup message in the green box
		popup := popupStyle.Render(m.popupMessage)

		// Render three option rectangles
		options := m.renderOptions()

		// If we have a response, render it separately below the popup
		var centeredContent string
		if m.agentResponse != nil {
			jsonBytes, err := json.MarshalIndent(m.agentResponse, "", "  ")
			if err == nil {
				responseStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("252")).
					Background(lipgloss.Color("235")).
					Padding(1, 2).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("240"))
				jsonResponse := responseStyle.Render(string(jsonBytes))

				// Combine popup, response, and options vertically
				centeredContent = lipgloss.JoinVertical(
					lipgloss.Center,
					popup,
					"",
					jsonResponse,
					"",
					options,
				)
			} else {
				centeredContent = lipgloss.JoinVertical(
					lipgloss.Center,
					popup,
					"",
					options,
				)
			}
		} else {
			centeredContent = lipgloss.JoinVertical(
				lipgloss.Center,
				popup,
				"",
				options,
			)
		}

		// Place the combined content in the center
		baseWithPopup := lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			centeredContent,
		)

		return baseWithPopup
	}

	return baseView
}

// renderOptions renders the three option rectangles with focus highlighting
func (m Model) renderOptions() string {
	optionLabels := []string{"Option A", "Option B", "Option C"}

	var renderedOptions []string
	for i, label := range optionLabels {
		var optionStyle lipgloss.Style

		if i == m.focusedOption {
			// Focused option - highlighted
			optionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color("86")).
				Bold(true).
				Padding(1, 3).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("86")).
				Width(20).
				Align(lipgloss.Center)
		} else {
			// Unfocused option - normal
			optionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236")).
				Padding(1, 3).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(20).
				Align(lipgloss.Center)
		}

		renderedOptions = append(renderedOptions, optionStyle.Render(label))
	}

	// Join the three rectangles horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderedOptions[0],
		" ",
		renderedOptions[1],
		" ",
		renderedOptions[2],
	)
}
