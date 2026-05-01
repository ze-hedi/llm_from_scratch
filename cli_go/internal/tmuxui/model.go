package tmuxui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/chatbot-tui/pkg/tmux"
)

const (
	stepPaneCount = iota // 0
	stepLayout           // 1
	stepPaths            // 2
	stepSession          // 3
	stepConfirm          // 4
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
	Paths       []string // one entry per pane
}

// Model is the Bubble Tea model for the tmux session setup wizard.
type Model struct {
	step          int
	paneCount     int
	layoutIdx     int
	pathInputs    []textinput.Model
	activePathIdx int
	sessionInput  textinput.Model
	width         int
	height        int
	result        *Result
	quitting      bool
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
			if m.step == 0 {
				return m, nil
			}
			m = blurCurrentStep(m)
			m.step--
			return focusCurrentStep(m)

		case "tab", "enter":
			if m.step == stepConfirm {
				m.result = m.buildResult()
				return m, tea.Quit
			}
			m = blurCurrentStep(m)
			m.step++
			return focusCurrentStep(m)
		}

		// Per-step key handling (enter/tab already consumed above).
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

		case stepPaths:
			switch msg.String() {
			case "up":
				if m.activePathIdx > 0 {
					m.pathInputs[m.activePathIdx].Blur()
					m.activePathIdx--
					cmd := m.pathInputs[m.activePathIdx].Focus()
					return m, cmd
				}
			case "down":
				if m.activePathIdx < m.paneCount-1 {
					m.pathInputs[m.activePathIdx].Blur()
					m.activePathIdx++
					cmd := m.pathInputs[m.activePathIdx].Focus()
					return m, cmd
				}
			default:
				var cmd tea.Cmd
				m.pathInputs[m.activePathIdx], cmd = m.pathInputs[m.activePathIdx].Update(msg)
				return m, cmd
			}

		case stepSession:
			var cmd tea.Cmd
			m.sessionInput, cmd = m.sessionInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// blurCurrentStep blurs whichever inputs are live for the current step.
func blurCurrentStep(m Model) Model {
	switch m.step {
	case stepPaths:
		for i := range m.pathInputs {
			m.pathInputs[i].Blur()
		}
	case stepSession:
		m.sessionInput.Blur()
	}
	return m
}

// focusCurrentStep initialises and focuses inputs for the current step,
// returning the updated model and the focus tea.Cmd.
func focusCurrentStep(m Model) (tea.Model, tea.Cmd) {
	switch m.step {
	case stepPaths:
		if len(m.pathInputs) != m.paneCount {
			m = initPaths(m)
		}
		cmd := m.pathInputs[m.activePathIdx].Focus()
		return m, cmd
	case stepSession:
		cmd := m.sessionInput.Focus()
		return m, cmd
	}
	return m, nil
}

// initPaths (re)creates the path textinputs, preserving existing values.
func initPaths(m Model) Model {
	cwd, _ := os.Getwd()
	existing := m.pathInputs
	m.pathInputs = make([]textinput.Model, m.paneCount)
	for i := range m.pathInputs {
		ti := textinput.New()
		ti.Placeholder = cwd
		ti.CharLimit = 256
		ti.Prompt = ""
		ti.Width = 40
		if i < len(existing) {
			ti.SetValue(existing[i].Value())
		} else {
			ti.SetValue(cwd)
		}
		m.pathInputs[i] = ti
	}
	if m.activePathIdx >= m.paneCount {
		m.activePathIdx = 0
	}
	return m
}

// buildResult packages the confirmed state into a Result.
func (m Model) buildResult() *Result {
	name := m.sessionInput.Value()
	if name == "" {
		name = "dev"
	}

	paths := make([]string, m.paneCount)
	for i := range paths {
		paths[i] = m.pathInputs[i].Value()
	}

	return &Result{
		PaneCount:   m.paneCount,
		Layout:      availableLayouts[m.layoutIdx],
		SessionName: name,
		Paths:       paths,
	}
}

// ── Views ──────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if m.step == stepPaths {
		return m.viewPathsFrame()
	}
	return m.viewMainForm()
}

func (m Model) viewMainForm() string {
	title := titleStyle.Render("New tmux session")

	rows := strings.Join([]string{
		m.row(stepPaneCount, "Panes", m.viewPaneCount()),
		m.row(stepLayout, "Layout", m.viewLayout()),
		m.row(stepPaths, "Dirs", m.viewDirsSummary()),
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

func (m Model) viewPathsFrame() string {
	title := pathsFrameTitleStyle.Render("Working directories")

	var rows []string
	for i, ti := range m.pathInputs {
		var cursor, labelRendered string
		if i == m.activePathIdx {
			cursor = "▶ "
			labelRendered = paneRowActiveStyle.Render(fmt.Sprintf("Pane %-2d", i+1))
		} else {
			cursor = "  "
			labelRendered = paneRowMutedStyle.Render(fmt.Sprintf("Pane %-2d", i+1))
		}
		rows = append(rows, cursor+labelRendered+"   "+ti.View())
	}

	box := pathBoxStyle.Render(strings.Join(rows, "\n"))
	hint := hintStyle.Render("↑/↓: switch pane   Tab: next   Shift+Tab: back   Esc: cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		box,
		hint,
	)

	return lipgloss.NewStyle().Padding(1, 3).Render(content)
}

// row renders one form row with active/inactive label styling.
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

func (m Model) viewDirsSummary() string {
	if len(m.pathInputs) == 0 || m.step <= stepPaths {
		return mutedStyle.Render("—")
	}
	return mutedStyle.Render(fmt.Sprintf("%d paths set", m.paneCount))
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

// ── Public API ─────────────────────────────────────────────────────────────

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
	return final.(Model).result, nil
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
