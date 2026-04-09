package tamagotchi

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/chatbot-tui/extensions/tamagotchi/pet"
)

type choice struct {
	name    string
	petType pet.PetType
	ascii   string
}

type choiceModel struct {
	choices  []choice
	cursor   int
	selected bool
}

func NewChoiceModel() choiceModel {
	// Create temporary pets to get their ASCII art
	cat := pet.NewPet("Mochi", pet.PetTypeCat)
	turtle := pet.NewPet("Lucy", pet.PetTypeTurtle)
	octopus := pet.NewPet("Ottopus", pet.PetTypeOctopus)

	return choiceModel{
		choices: []choice{
			{name: "Mochi", petType: pet.PetTypeCat, ascii: cat.GetASCII()},
			{name: "Lucy", petType: pet.PetTypeTurtle, ascii: turtle.GetASCII()},
			{name: "Ottopus", petType: pet.PetTypeOctopus, ascii: octopus.GetASCII()},
		},
		cursor: 0,
	}
}

func (m choiceModel) Init() tea.Cmd {
	return nil
}

func (m choiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter":
			m.selected = true
			chosen := m.choices[m.cursor]
			p := pet.NewPet(chosen.name, chosen.petType)

			if err := pet.SavePet(p); err != nil {
				fmt.Printf("Error saving pet: %v\n", err)
				return m, tea.Quit
			}

			fmt.Printf("\nYou've chosen %s!\n", chosen.name)
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m choiceModel) View() string {
	if m.selected {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(30)

	selectedBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(30).
		Bold(true)

	s := titleStyle.Render("🐾 Choose Your Tamagotchi") + "\n\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Use ↑/↓ or j/k to navigate, Enter to select, q to quit") + "\n\n"

	for i, choice := range m.choices {
		cursor := "  "
		style := boxStyle
		if m.cursor == i {
			cursor = "→ "
			style = selectedBoxStyle
		}

		petInfo := fmt.Sprintf("%s\n%s\nType: %s", choice.name, choice.ascii, choice.petType)
		s += cursor + style.Render(petInfo) + "\n"
	}

	return s
}

func RunChooseUI() error {
	p := tea.NewProgram(NewChoiceModel())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
