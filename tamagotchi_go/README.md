# Tamagotchi TUI

A virtual pet game in your terminal! Take care of Mochi, your ASCII art pet.

## Features

- **Virtual Pet "Mochi"**: Beautiful ASCII art animations
- **Dynamic Moods**: Happy, Neutral, Sad, Sick, Dead
- **Real-time Stats**: Health, Hunger, Happiness
- **Time-based Degradation**: Stats change over time - keep your pet alive!
- **Interactive Commands**: Feed, play, heal, and check status
- **Clean TUI**: Built with Bubble Tea and Cobra

## Installation

```bash
# Install dependencies
go mod download

# Build the game
go build -o tamagotchi-tui .
```

## Usage

```bash
./tamagotchi-tui play
```

## Commands

Type these commands in the game:

- **feed** - Feed Mochi when hungry (reduces hunger by 30%)
- **play** - Play with Mochi (increases happiness by 20%, increases hunger by 10%)
- **heal** - Heal Mochi when sick (restores 30% health)
- **status** - View detailed stats
- **help** - Show command list
- **quit** or **exit** - Exit the game

## Keyboard Controls

- **Enter**: Send command
- **Ctrl+C** or **Esc**: Quit

## Pet Stats

### Hunger (0-100)
- Increases by 5% every 10 seconds
- Feed when > 50% to keep Mochi happy
- Very hungry (>80%) damages health

### Happiness (0-100)
- Decreases by 3% every 10 seconds
- Play to increase happiness
- Very unhappy (<20%) damages health

### Health (0-100)
- Decreases when hungry or unhappy
- Heal to restore health
- Game over when health = 0!

## ASCII Moods

```
😊 Happy:    (ﾟ､ ｡７     Happiness > 60 and Hunger < 50
😐 Neutral:  (･ω･）     Normal state
😢 Sad:      (˘︹˘）     Happiness < 30 or Hunger > 60
🤒 Sick:     (×_×）     Health < 30 or Hunger > 80
💀 Dead:     (X_X）     Health = 0 (R.I.P)
            R.I.P
```

## Tips

- Check status regularly
- Feed before hunger reaches 70%
- Play when happiness drops below 60%
- Heal immediately when health < 50%
- Don't neglect your pet or Mochi will die!

## Project Structure

```
.
├── cmd/
│   ├── root.go         # Cobra root command
│   └── play.go         # Play command
├── internal/tui/
│   ├── model.go        # Bubble Tea TUI model
│   └── styles.go       # UI styling
├── pkg/tamagotchi/
│   └── pet.go          # Pet logic and state
├── main.go             # Entry point
└── README.md           # This file
```

## Future Enhancements

- [ ] Save/load pet state
- [ ] Multiple pets
- [ ] Pet evolution stages
- [ ] Mini-games
- [ ] Achievements
- [ ] Customizable names
- [ ] Pet aging and lifecycle
- [ ] High score system

## License

MIT License
