package tmuxui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/chatbot-tui/pkg/tmux"
)

const (
	stepPaneCount = iota
	stepLayout
	stepSession
	stepConfirm
)

var (
	availableLayouts = []tmux.Layout{
		tmux.LayoutEvenHorizontal,
		tmux.LayoutEvenVertical,
		tmux.LayoutTiled,
	}

	layoutLabels = map[tmux.Layout]string{
		tmux.LayoutEvenHorizontal: "even-horizontal",
		tmux.LayoutEvenVertical:   "even-vertical",
		tmux.LayoutTiled:          "tiled",
	}
)

// Result holds the confirmed choices from the TUI.
type Result struct {
	PaneCount   int
	Layout      tmux.Layout
	SessionName string
}

// Model is the Bubble Tea model for the tmux session setup wizard.
type Model struct {
	step         int
	paneCount    int
	layoutIdx    int
	sessionInput textinput.Model
	width        int
	height       int
	result       *Result
	quitting     bool
}

func NewModel(defaultSession string) Model {
	ti := textinput.New()
	ti.Placeholder = "dev"
	ti.SetValue(defaultSession)
	ti.CharLimit = 32
	ti.Prompt = ""

	return Model{
		step:         stepPaneCount,
		paneCount:    3,
		layoutIdx:    0,
		sessionInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "shift+tab":
			if m.step > 0 {
				m.step--
				if m.step == stepSession {
					return m, m.sessionInput.Focus()
				}
				m.sessionInput.Blur()
			}
			return m, nil

		case "tab", "enter":
			if m.step == stepConfirm {
				name := m.sessionInput.Value()
				if name == "" {
					name = "dev"
				}
				m.result = &Result{
					PaneCount:   m.paneCount,
					Layout:      availableLayouts[m.layoutIdx],
					SessionName: name,
				}
				return m, tea.Quit
			}
			m.step++
			if m.step == stepSession {
				return m, m.sessionInput.Focus()
			}
			m.sessionInput.Blur()
			return m, nil
		}

		// Per-step key handling (only if enter/tab not consumed above).
		switch m.step {
		case stepPaneCount:
			switch msg.String() {
			case "left", "h", "-":
				if m.paneCount > 1 {
					m.paneCount--
				}
			case "right", "l", "+":
				if m.paneCount < 9 {
					m.paneCount++
				}
			}
		case stepLayout:
			switch msg.String() {
			case "left", "h":
				m.layoutIdx = (m.layoutIdx - 1 + len(availableLayouts)) % len(availableLayouts)
			case "right", "l":
				m.layoutIdx = (m.layoutIdx + 1) % len(availableLayouts)
			}
		case stepSession:
			var cmd tea.Cmd
			m.sessionInput, cmd = m.sessionInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	title := titleStyle.Render("New tmux session")

	rows := strings.Join([]string{
		m.row(stepPaneCount, "Panes", m.viewPaneCount()),
		m.row(stepLayout, "Layout", m.viewLayout()),
		m.row(stepSession, "Session", m.viewSession()),
		m.row(stepConfirm, "Launch", m.viewConfirm()),
	}, "\n")

	form := formBoxStyle.Render(rows)

	preview := previewTitleStyle.Render("Preview") + "\n" +
		previewStyle.Render(renderPreview(m.paneCount, availableLayouts[m.layoutIdx]))

	hint := hintStyle.Render("←/→: change   Tab/Enter: next   Shift+Tab: back   Esc: cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		form,
		preview,
		hint,
	)

	return lipgloss.NewStyle().Padding(1, 3).Render(content)
}

// row renders one labelled field with active/inactive styling.
func (m Model) row(step int, label, value string) string {
	if m.step == step {
		return activeLabelStyle.Render(label) + "  " + value
	}
	return labelStyle.Render(label) + "  " + value
}

func (m Model) viewPaneCount() string {
	if m.step == stepPaneCount {
		return arrowStyle.Render("← ") +
			activeValueStyle.Render(fmt.Sprintf("%d", m.paneCount)) +
			arrowStyle.Render(" →")
	}
	return mutedStyle.Render(fmt.Sprintf("%d", m.paneCount))
}

func (m Model) viewLayout() string {
	name := layoutLabels[availableLayouts[m.layoutIdx]]
	if m.step == stepLayout {
		return arrowStyle.Render("← ") +
			activeValueStyle.Render(name) +
			arrowStyle.Render(" →")
	}
	return mutedStyle.Render(name)
}

func (m Model) viewSession() string {
	if m.step == stepSession {
		return m.sessionInput.View()
	}
	val := m.sessionInput.Value()
	if val == "" {
		val = "dev"
	}
	return mutedStyle.Render(val)
}

func (m Model) viewConfirm() string {
	if m.step == stepConfirm {
		return confirmActiveStyle.Render("Press Enter to launch")
	}
	return confirmMutedStyle.Render("Press Enter to launch")
}

// Result returns the confirmed result, or nil if cancelled.
func (m Model) Result() *Result {
	return m.result
}

// Run starts the Bubble Tea TUI and returns the user's choices, or nil if cancelled.
func Run(defaultSession string) (*Result, error) {
	p := tea.NewProgram(NewModel(defaultSession))
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	return final.(Model).Result(), nil
}

// ── ASCII preview renderer ─────────────────────────────────────────────────

func renderPreview(n int, layout tmux.Layout) string {
	if n < 1 {
		n = 1
	}
	switch layout {
	case tmux.LayoutEvenVertical:
		return renderVertical(n)
	case tmux.LayoutTiled:
		return renderTiled(n)
	default:
		return renderHorizontal(n)
	}
}

func renderHorizontal(n int) string {
	cell := "──────"
	top := "┌" + strings.Repeat(cell+"┬", n-1) + cell + "┐"
	mid := "│" + strings.Repeat("      │", n)
	bot := "└" + strings.Repeat(cell+"┴", n-1) + cell + "┘"
	return top + "\n" + mid + "\n" + mid + "\n" + bot
}

func renderVertical(n int) string {
	var b strings.Builder
	b.WriteString("┌──────────────────────┐\n")
	for i := 0; i < n; i++ {
		b.WriteString("│                      │\n")
		if i < n-1 {
			b.WriteString("├──────────────────────┤\n")
		}
	}
	b.WriteString("└──────────────────────┘")
	return b.String()
}

func renderTiled(n int) string {
	cols := 2
	if n == 1 {
		cols = 1
	}
	rows := (n + cols - 1) / cols
	w := 11

	var b strings.Builder

	topRow := "┌" + strings.Repeat(strings.Repeat("─", w)+"┬", cols-1) + strings.Repeat("─", w) + "┐\n"
	midRow := "│" + strings.Repeat(strings.Repeat(" ", w)+"│", cols) + "\n"
	divRow := "├" + strings.Repeat(strings.Repeat("─", w)+"┼", cols-1) + strings.Repeat("─", w) + "┤\n"
	botRow := "└" + strings.Repeat(strings.Repeat("─", w)+"┴", cols-1) + strings.Repeat("─", w) + "┘"

	b.WriteString(topRow)
	for r := 0; r < rows; r++ {
		b.WriteString(midRow)
		if r < rows-1 {
			b.WriteString(divRow)
		}
	}
	b.WriteString(botRow)

	return b.String()
}
