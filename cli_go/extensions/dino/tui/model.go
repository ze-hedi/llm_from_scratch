package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yourusername/chatbot-tui/extensions/dino/game"
)

type tickMsg time.Time

const (
	FPS = 30
)

type Model struct {
	game   *game.Game
	width  int
	height int
}

func NewModel(highScore int) Model {
	return Model{
		game:   game.NewGame(highScore),
		width:  80,
		height: 20,
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/FPS, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case " ", "up", "w":
			if m.game.GameOver {
				m.game.Reset()
			} else {
				m.game.Jump()
			}
		case "down", "s":
			m.game.Duck(true)
		case "r":
			if m.game.GameOver {
				m.game.Reset()
			}
		default:
			// Release duck when any other key is pressed
			m.game.Duck(false)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		if !m.game.GameOver {
			m.game.Update()
		}
		return m, tickCmd()
	}

	return m, nil
}

func (m Model) View() string {
	if m.width < 80 || m.height < 20 {
		return "Terminal too small. Please resize to at least 80x20."
	}

	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("🦖 DINO RUNNER")
	b.WriteString(title + "\n\n")

	// Score
	scoreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(scoreStyle.Render(fmt.Sprintf("Score: %d | High Score: %d | Speed: %.1f",
		m.game.Score, m.game.HighScore, m.game.Speed)))
	b.WriteString("\n\n")

	// Render game field
	field := m.renderGameField()
	b.WriteString(field)

	// Game over message
	if m.game.GameOver {
		b.WriteString("\n")
		gameOverStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9")).
			Align(lipgloss.Center)
		b.WriteString(gameOverStyle.Render("GAME OVER!"))
		b.WriteString("\n")

		restartStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center)
		b.WriteString(restartStyle.Render("Press SPACE or R to restart"))
	}

	// Controls
	b.WriteString("\n\n")
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(controlsStyle.Render("Controls: SPACE/UP to jump | DOWN to duck | Q/ESC to quit"))

	return b.String()
}

func (m Model) renderGameField() string {
	width := 80
	height := 18

	// Create a 2D grid
	grid := make([][]string, height)
	for i := range grid {
		grid[i] = make([]string, width)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// Draw ground
	groundY := game.GroundLevel
	for x := 0; x < width; x++ {
		grid[groundY][x] = "_"
	}

	// Draw obstacles
	for _, obs := range m.game.Obstacles {
		x := int(obs.X)
		if x >= 0 && x < width {
			switch obs.Type {
			case game.Cactus:
				// Draw cactus - simple ASCII cactus
				for h := 0; h < obs.Height; h++ {
					y := obs.Y - h
					if y >= 0 && y < height {
						if x < width {
							if h == 0 {
								grid[y][x] = "┴"
							} else if h == obs.Height-1 {
								grid[y][x] = "+"
							} else {
								grid[y][x] = "|"
							}
						}
						if x+1 < width && h == 1 {
							grid[y][x+1] = "-"
						}
					}
				}
			case game.Bird:
				// Draw bird - simple ASCII bird
				for h := 0; h < obs.Height; h++ {
					y := obs.Y - h
					if y >= 0 && y < height {
						for w := 0; w < obs.Width && x+w < width; w++ {
							if m.game.FrameCount%6 < 3 {
								if w == 0 {
									grid[y][x+w] = "<"
								} else if w == 1 {
									grid[y][x+w] = "V"
								} else {
									grid[y][x+w] = ">"
								}
							} else {
								if w == 0 {
									grid[y][x+w] = "^"
								} else if w == 1 {
									grid[y][x+w] = "V"
								} else {
									grid[y][x+w] = "^"
								}
							}
						}
					}
				}
			}
		}
	}

	// Draw dino
	dinoY := int(m.game.Dino.Y)
	dinoHeight := m.game.GetDinoHeight()

	if m.game.Dino.IsDucking {
		// Ducking dino - lowered profile
		if dinoY >= 0 && dinoY < height && game.DinoX < width {
			grid[dinoY][game.DinoX] = "/"
			if game.DinoX+1 < width {
				grid[dinoY][game.DinoX+1] = "="
			}
			if game.DinoX+2 < width {
				grid[dinoY][game.DinoX+2] = "\\"
			}
		}
	} else {
		// Standing/jumping dino
		for h := 0; h < dinoHeight; h++ {
			y := dinoY - h
			if y >= 0 && y < height && game.DinoX < width {
				if h == dinoHeight-1 {
					// Head
					grid[y][game.DinoX] = "O"
					if game.DinoX+1 < width {
						grid[y][game.DinoX+1] = ">"
					}
				} else if h == dinoHeight-2 {
					// Body
					grid[y][game.DinoX] = "|"
					if game.DinoX+1 < width {
						grid[y][game.DinoX+1] = "|"
					}
				} else if h == 0 {
					// Legs
					grid[y][game.DinoX] = "/"
					if game.DinoX+1 < width {
						grid[y][game.DinoX+1] = "\\"
					}
				}
			}
		}
	}

	// Convert grid to string
	var b strings.Builder
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			b.WriteString(grid[i][j])
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) GetHighScore() int {
	return m.game.HighScore
}
