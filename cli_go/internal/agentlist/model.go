package agentlist

import (
	"fmt"
	"strings"
	"time"

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
	popupScroll   int    // Scroll offset for popup content
	activeForm    string // Current active form: "main" or "tools"
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
		m.popupScroll = 0
		m.activeForm = "main"

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
			if m.showPopup {
				// Scroll popup content up
				if m.popupScroll > 0 {
					m.popupScroll--
				}
			} else if !m.loading && len(m.agents) > 0 && m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.showPopup {
				// Scroll popup content down
				m.popupScroll++
			} else if !m.loading && len(m.agents) > 0 && m.selectedIndex < len(m.agents)-1 {
				m.selectedIndex++
			}

		case "enter":
			// Call set_agent endpoint when Enter is pressed on an agent (only if popup not shown)
			if !m.showPopup && !m.loading && len(m.agents) > 0 {
				agentName := m.agents[m.selectedIndex].Name
				m.showPopup = true
				m.popupMessage = fmt.Sprintf("loading %s ...", agentName)
				m.agentResponse = nil
				m.popupScroll = 0     // Reset scroll when showing new agent
				m.activeForm = "main" // Start with main form
				// Call the set_agent endpoint
				return m, m.bot.SetAgent(agentName)
			}

		case "f2":
			// Toggle between main and tools forms circularly when popup is shown
			if m.showPopup {
				if m.activeForm == "main" || m.activeForm == "" {
					m.activeForm = "tools"
				} else {
					m.activeForm = "main"
				}
				m.popupScroll = 0 // Reset scroll when switching forms
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
		// Get window width with default fallback
		windowWidth := m.width
		if windowWidth == 0 {
			windowWidth = 120 // Default width if not yet initialized
		}

		// Render just the popup message in the green box with width constraint
		responsivePopupStyle := popupStyle.Copy().Width(windowWidth - 20)
		popup := responsivePopupStyle.Render(m.popupMessage)

		// If we have a response, render main_agent fields
		var centeredContent string
		if m.agentResponse != nil {
			// Extract main_agent from response
			var agentData map[string]interface{}

			// Try to get main_agent
			if mainAgent, ok := m.agentResponse["main_agent"].(map[string]interface{}); ok {
				agentData = mainAgent
			} else {
				// Fallback: use the entire response if main_agent doesn't exist
				agentData = m.agentResponse
			}

			// Extract agent_name for header
			agentName := ""
			if name, ok := m.agentResponse["agent_name"]; ok && name != nil {
				agentName = fmt.Sprintf("%v", name)
			}

			// Determine active form (default to "main" if not set)
			activeForm := m.activeForm
			if activeForm == "" {
				activeForm = "main"
			}

			// Create tab navigation indicators
			activeTabStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color("86")).
				Bold(true).
				Padding(0, 2)

			inactiveTabStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 2)

			mainTab := inactiveTabStyle.Render("Main Agent")
			toolsTab := inactiveTabStyle.Render("Tools")

			if activeForm == "main" {
				mainTab = activeTabStyle.Render("Main Agent")
			} else if activeForm == "tools" {
				toolsTab = activeTabStyle.Render("Tools")
			}

			navigationHint := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true).
				Render("(F2 to toggle)")

			tabs := lipgloss.JoinHorizontal(lipgloss.Left, mainTab, "  ", toolsTab, "  ", navigationHint)

			// Render content based on active form
			var formContent string
			var header string

			if activeForm == "tools" {
				// Render Tools form
				headerStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("86")).
					Bold(true).
					Underline(true).
					Width(windowWidth - 20)
				header = headerStyle.Render("Tools")
				formContent = m.renderToolsForm(agentData, windowWidth)
			} else {
				// Render Main Agent form (default)
				headerStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("86")).
					Bold(true).
					Underline(true).
					Width(windowWidth - 20)
				header = headerStyle.Render(fmt.Sprintf("Main Agent: %s", agentName))
				formContent = m.renderAgentFields(agentData, windowWidth)
			}

			// Apply scrolling to form content
			scrolledContent := m.applyScrolling(formContent)

			// Combine popup, tabs, header, and form content vertically
			centeredContent = lipgloss.JoinVertical(
				lipgloss.Center,
				popup,
				"",
				tabs,
				"",
				header,
				"",
				scrolledContent,
			)
		} else {
			centeredContent = popup
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

// renderAgentFields renders the main_agent attributes as a form
func (m Model) renderAgentFields(agentData map[string]interface{}, width int) string {
	// Calculate content width with some padding
	contentWidth := width - 20
	if contentWidth < 40 {
		contentWidth = 40
	}

	fieldLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Width(contentWidth)

	fieldValueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(contentWidth)

	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(contentWidth)

	// Check if there's any data
	if len(agentData) == 0 {
		noDataStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		return containerStyle.Render(noDataStyle.Render("No data available"))
	}

	// Extract main_agent_card from agentData
	var mainAgentCard map[string]interface{}
	if card, ok := agentData["main_agent_card"].(map[string]interface{}); ok {
		mainAgentCard = card
	} else {
		noDataStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		return containerStyle.Render(noDataStyle.Render("main_agent_card not found"))
	}

	var fields []string

	// 1. Soul from main_agent_card["soul"]
	soul := ""
	if val, ok := mainAgentCard["soul"]; ok && val != nil {
		if str, ok := val.(string); ok {
			soul = str
		} else {
			soul = fmt.Sprintf("%v", val)
		}
	}

	soulField := lipgloss.JoinVertical(
		lipgloss.Left,
		fieldLabelStyle.Render("Soul:"),
		fieldValueStyle.Render("  "+soul),
	)
	fields = append(fields, soulField)

	// 2. Agent instructions from main_agent_card["agent_instructions"]
	instructions := ""
	if val, ok := mainAgentCard["agent_instructions"]; ok && val != nil {
		if str, ok := val.(string); ok {
			instructions = str
		} else {
			instructions = fmt.Sprintf("%v", val)
		}
	}

	instructionsField := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		fieldLabelStyle.Render("Agent instructions:"),
		fieldValueStyle.Render("  "+instructions),
	)
	fields = append(fields, instructionsField)

	// 3. Agent pattern from main_agent_card["agent_type"]
	agentType := ""
	if val, ok := mainAgentCard["agent_type"]; ok && val != nil {
		if str, ok := val.(string); ok {
			agentType = str
		} else {
			agentType = fmt.Sprintf("%v", val)
		}
	}

	agentTypeField := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		fieldLabelStyle.Render("Agent pattern:"),
		fieldValueStyle.Render("  "+agentType),
	)
	fields = append(fields, agentTypeField)

	// 4. Human in loop from main_agent_card["human_in_loop"] (default to false if null)
	humanInLoop := false
	if val, ok := mainAgentCard["human_in_loop"]; ok && val != nil {
		if b, ok := val.(bool); ok {
			humanInLoop = b
		}
	}

	humanInLoopField := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		fieldLabelStyle.Render("Human in loop:"),
		fieldValueStyle.Render(fmt.Sprintf("  %v", humanInLoop)),
	)
	fields = append(fields, humanInLoopField)

	// 5. Max iterations from main_agent_card["max_iterations"]
	maxIterations := "undefined"
	if val, ok := mainAgentCard["max_iterations"]; ok && val != nil {
		maxIterations = fmt.Sprintf("%v", val)
	}

	maxIterationsField := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		fieldLabelStyle.Render("Max iterations:"),
		fieldValueStyle.Render("  "+maxIterations),
	)
	fields = append(fields, maxIterationsField)

	// Join all fields vertically
	fieldsContent := lipgloss.JoinVertical(lipgloss.Left, fields...)

	return containerStyle.Render(fieldsContent)
}

// applyScrolling applies scroll offset to content and adds scroll indicators
func (m Model) applyScrolling(content string) string {
	// Split content into lines
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// Calculate visible window height (reserve some space for popup message and indicators)
	// Reserve minimal space for indicators and borders
	height := m.height
	if height == 0 {
		// Use a reasonable default if height hasn't been set yet
		height = 50
	}
	maxVisibleLines := height - 4
	if maxVisibleLines < 40 {
		maxVisibleLines = 40
	}

	// Clamp scroll offset
	scrollOffset := m.popupScroll
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	maxScroll := totalLines - maxVisibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scrollOffset > maxScroll {
		scrollOffset = maxScroll
	}

	// Extract visible lines
	startLine := scrollOffset
	endLine := startLine + maxVisibleLines
	if endLine > totalLines {
		endLine = totalLines
	}

	visibleLines := lines[startLine:endLine]

	// Pad with empty lines to fill the entire visible space
	// This ensures old content is completely cleared when switching forms
	actualVisibleLines := len(visibleLines)
	if actualVisibleLines < maxVisibleLines {
		// Add empty lines to fill the space
		for i := actualVisibleLines; i < maxVisibleLines; i++ {
			visibleLines = append(visibleLines, "")
		}
	}

	visibleContent := strings.Join(visibleLines, "\n")

	// Add scroll indicators
	scrollIndicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	var indicators []string
	if scrollOffset > 0 {
		indicators = append(indicators, scrollIndicatorStyle.Render("▲ Scroll up (↑/k)"))
	}
	if endLine < totalLines {
		indicators = append(indicators, scrollIndicatorStyle.Render("▼ Scroll down (↓/j)"))
	}

	if len(indicators) > 0 {
		indicatorText := strings.Join(indicators, "  ")
		return lipgloss.JoinVertical(
			lipgloss.Left,
			visibleContent,
			"",
			indicatorText,
		)
	}

	return visibleContent
}

// renderToolsForm renders the tools form
func (m Model) renderToolsForm(agentData map[string]interface{}, width int) string {
	// Calculate content width with some padding
	contentWidth := width - 20
	if contentWidth < 40 {
		contentWidth = 40
	}

	fieldLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Width(contentWidth)

	fieldValueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(contentWidth)

	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(contentWidth)

	// Extract main_agent_card from agentData (same as renderAgentFields)
	var mainAgentCard map[string]interface{}
	if card, ok := agentData["main_agent_card"].(map[string]interface{}); ok {
		mainAgentCard = card
	} else {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		content := messageStyle.Render("main_agent_card not found")
		return containerStyle.Render(content)
	}

	// Extract tools list from main_agent_card["tools"]
	var toolsList []interface{}
	if tools, ok := mainAgentCard["tools"]; ok && tools != nil {
		if list, ok := tools.([]interface{}); ok {
			toolsList = list
		}
	}

	// If no tools, display message
	if len(toolsList) == 0 {
		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		content := messageStyle.Render("No tools available")
		return containerStyle.Render(content)
	}

	// Render each tool
	var toolItems []string
	for i, toolInterface := range toolsList {
		tool, ok := toolInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract name
		name := ""
		if val, ok := tool["name"]; ok && val != nil {
			if str, ok := val.(string); ok {
				name = str
			} else {
				name = fmt.Sprintf("%v", val)
			}
		}

		// Extract description
		description := ""
		if val, ok := tool["description"]; ok && val != nil {
			if str, ok := val.(string); ok {
				description = str
			} else {
				description = fmt.Sprintf("%v", val)
			}
		}

		// Create tool item
		toolItem := lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			fieldLabelStyle.Render(fmt.Sprintf("Tool %d - %s:", i+1, name)),
			fieldValueStyle.Render("  "+description),
		)
		toolItems = append(toolItems, toolItem)
	}

	// Join all tool items
	toolsContent := lipgloss.JoinVertical(lipgloss.Left, toolItems...)

	return containerStyle.Render(toolsContent)
}
