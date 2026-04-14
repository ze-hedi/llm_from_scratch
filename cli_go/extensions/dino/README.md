# 🦖 Dino Runner

A terminal-based recreation of Chrome's classic T-Rex endless runner game!

## Features

- **Jumping & Ducking**: Avoid obstacles by jumping or ducking
- **Dynamic Obstacles**: Dodge cacti on the ground and birds in the air
- **Progressive Difficulty**: Game speed increases as you progress
- **High Score Tracking**: Your best score is automatically saved
- **Smooth Animation**: 30 FPS terminal rendering

## How to Play

### Launch the Game

```bash
# From the extensions browser
chatbot-tui extensions

# Or directly
chatbot-tui dino
```

### Controls

- **SPACE** or **UP** or **W**: Jump
- **DOWN** or **S**: Duck (hold while pressing)
- **R** or **SPACE** (when game over): Restart
- **Q** or **ESC**: Quit

### Gameplay

- Your dinosaur runs automatically from left to right
- Jump over cacti (🌵) that appear on the ground
- Duck under flying birds (🦅/🦆) at various heights
- The game speed increases as your score rises
- Collision with any obstacle ends the game
- Try to beat your high score!

## Commands

### Play the Game
```bash
chatbot-tui dino
```

### Reset High Score
```bash
chatbot-tui dino reset
```

## Technical Details

### Architecture

- **Game Logic** (`game/game.go`): Physics, collision detection, obstacle spawning
- **TUI Model** (`tui/model.go`): Bubble Tea interface, rendering, input handling
- **Storage** (`game/storage.go`): High score persistence

### Physics

- Gravity: 1.2 units/frame
- Jump velocity: -8.0 units
- Ground level: Y=15
- Speed range: 3.0 to 12.0 units/frame

### Obstacles

- **Cacti**: Ground-level obstacles, 2 units wide, 3 units tall
- **Birds**: Flying obstacles at 3 different heights, appear after score 100

## File Structure

```
extensions/dino/
├── README.md           # This file
├── game/
│   ├── game.go        # Core game logic
│   └── storage.go     # High score persistence
└── tui/
    └── model.go       # Bubble Tea TUI model
```

## Score Persistence

Your high score is automatically saved to `~/.cli_go/dino/save.json` and will persist across game sessions.
