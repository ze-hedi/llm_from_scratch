# ChatBot TUI

A beautiful terminal-based chatbot interface built with Go, Bubble Tea, and Cobra.

## Features

- **Interactive TUI**: Beautiful terminal user interface with smooth scrolling
- **Real-time Chat**: Instant responses with an intelligent mock chatbot
- **Seamless Navigation**: Press Ctrl+A to switch between Chat and Extensions (state preserved!)
- **Agent Management**: Browse and select agents with Ctrl+G, view detailed agent information and tools
- **Smart Sidebar**: Token usage sidebar auto-shows in full screen mode (≥120 columns)
- **Keyboard Controls**: Fully keyboard-driven interface
- **Responsive Design**: Adapts to terminal window size
- **Clean Architecture**: Well-organized Go project structure
- **Extensions**: Modular extension system with built-in games (Tamagotchi & Dino Runner)
- **Web Terminal**: Browser-based terminal with support for up to 10 simultaneous shells

## Prerequisites

- Go 1.18 or higher
- Terminal with UTF-8 support
- Node.js and npm (for web terminal server)

## Quick Start

```bash
# Clone the repository
git clone <your-repo-url>
cd cli_go

# Install dependencies
go mod download

# Build the application
go build -o chatbot-tui .

# Start chatting
./chatbot-tui chat

# Or browse extensions
./chatbot-tui extensions
```

## Available Commands

| Command | Description |
|---------|-------------|
| `./chatbot-tui chat` | Start interactive chat session |
| `./chatbot-tui extensions` | Browse and launch extensions |
| `./chatbot-tui tamagotchi` | Play Tamagotchi virtual pet game |
| `./chatbot-tui tamagotchi choose` | Choose a new pet |
| `./chatbot-tui tamagotchi reset` | Reset your pet |
| `./chatbot-tui dino` | Play Dino Runner game |
| `./chatbot-tui dino reset` | Reset high score |

## Usage

### Start a chat session

```bash
./chatbot-tui chat
```

### Browse Extensions

```bash
# View and launch available extensions
./chatbot-tui extensions
```

### Play Tamagotchi

```bash
# Start the game (choose a pet if first time)
./chatbot-tui tamagotchi

# Choose a new pet
./chatbot-tui tamagotchi choose

# Reset your pet
./chatbot-tui tamagotchi reset
```

### Play Dino Runner

```bash
# Start the game
./chatbot-tui dino

# Reset your high score
./chatbot-tui dino reset
```

### Web Terminal Server

Launch a browser-based terminal with support for multiple simultaneous shell sessions:

```bash
# Install dependencies (first time only)
npm install

# Start the web terminal server
node server.js
```

Then open your browser to `http://localhost:3000`

**Features:**
- **Multiple Terminals**: Click "Launch Terminal" up to 10 times to create separate shell instances
- **Independent Sessions**: Each terminal runs its own shell process
- **Real-time Updates**: Live terminal output with full xterm.js support
- **Easy Management**: Close individual terminals with the close button
- **Session Counter**: Button shows current active terminals (e.g., "Launch Terminal (3/10)")
- **Auto-disable**: Button automatically disables when max limit (10) is reached
- **Clean Cleanup**: Properly closes WebSocket connections and cleans up resources

**Usage:**
1. Click "Launch Terminal" to create a new shell
2. Each terminal is numbered (Terminal #1, Terminal #2, etc.)
3. Use the close button on any terminal to remove it
4. Launch up to 10 terminals simultaneously
5. When a terminal disconnects, it's automatically removed from the list

### Agent List Browser

Browse and manage AI agents directly from the chat interface by pressing **Ctrl+G**:

```bash
# In chat mode, press Ctrl+G to open agent list
```

**Features:**
- **Browse Agents**: View all available agents from the server with descriptions
- **Agent Details**: Select an agent to view comprehensive information in a popup
- **Dual Forms**: Toggle between Main Agent details and Tools using F2
- **Scrollable Content**: Navigate long agent details with up/down arrow keys
- **Responsive UI**: Adapts to terminal size changes

**Main Agent Form displays:**
- **Soul**: Agent's core personality/purpose
- **Agent Instructions**: Detailed instructions for the agent
- **Agent Pattern**: The agent type/pattern being used
- **Human in Loop**: Whether human intervention is required
- **Max Iterations**: Maximum number of iterations allowed

**Tools Form displays:**
- **Tool List**: All available tools for the selected agent
- **Tool Details**: Name and description for each tool

**Navigation:**
1. Press **Ctrl+G** from chat to open agent list
2. Use **↑/↓** or **j/k** to browse agents
3. Press **Enter** to view agent details in a green popup
4. Press **F2** to toggle between "Main Agent" and "Tools" forms
5. Use **↑/↓** or **j/k** to scroll through long content
6. Press **Ctrl+G** or **Esc** to return to chat

### Keyboard Controls

**Global Navigation:**
- **Ctrl+A**: Switch between Chat and Extensions browser (state is preserved!)
- **Ctrl+G**: Open agent list browser / Return to chat
- **Ctrl+Y**: Open model settings / Return to chat
- **Esc**: Return to Chat (from any extension)

**Chat Mode:**
- **Enter**: Send message
- **Alt+Enter**: New line
- **Ctrl+N**: Toggle sidebar (only in full screen mode, ≥120 columns)
- **Ctrl+C** or **Esc**: Quit the application
- **Arrow Keys**: Navigate through chat history

**Agent List Browser:**
- **↑/↓** or **j/k**: Navigate through agents
- **Enter**: Select agent and view details
- **Ctrl+G** or **Esc**: Return to chat
- **F2**: Toggle between Main Agent and Tools forms (when viewing agent details)
- **↑/↓** or **j/k** (in popup): Scroll through agent details

**Settings Mode:**
- **↑/↓** or **j/k**: Navigate models
- **Enter**: Select model and return to chat
- **Esc** or **Ctrl+Y**: Back to chat without changes

**Slash Commands (in Chat):**
- **/exit** or **/quit**: Exit the application

**Extensions Browser:**
- **↑/↓** or **j/k**: Navigate through extensions
- **Enter**: Launch selected extension
- **Ctrl+A** or **Esc**: Back to Chat

**Tamagotchi Mode:**
- Type commands: `feed`, `play`, `heal`, `status`, `quit`
- **Enter**: Send command
- **Ctrl+A** or **Esc**: Back to Chat

**Dino Runner Mode:**
- **SPACE/UP/W**: Jump over obstacles
- **DOWN/S**: Duck under birds
- **R/SPACE** (when game over): Restart game
- **Q/Esc/Ctrl+A**: Back to Chat (auto-saves high score)

### Quick Navigation Workflow

1. Start chatting: `./chatbot-tui chat`
2. Press **Ctrl+A** to browse extensions while keeping your chat history
3. Select and launch an extension (e.g., Tamagotchi or Dino Runner)
4. Press **Ctrl+A** to return to your chat - all messages preserved!
5. Switch back and forth as much as you want - everything stays in memory

### Slash Commands

The chat interface supports slash commands for special actions. Simply type the command in the input field and press Enter.

**Available Commands:**
- `/exit` or `/quit` - Exit the application gracefully

More commands can be added in the future!

### Sidebar Behavior

The token usage sidebar is **intelligent and responsive**:
- **Auto-shows** when your terminal is in full screen mode (≥120 columns wide)
- **Auto-hides** when terminal is smaller to maximize chat space
- **Toggle with Ctrl+N** (only available in full screen mode)
- Shows real-time token tracking, usage percentage, and current AI model

**Tip:** Maximize your terminal window to see the sidebar!

## Project Structure

```
.
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── chat.go            # Chat command
│   ├── settings.go        # Settings command
│   ├── extensions.go      # Extensions browser
│   ├── tamagotchi.go      # Tamagotchi command
│   └── dino.go            # Dino Runner command
├── internal/
│   ├── coordinator/       # Main navigation coordinator
│   │   └── model.go       # Manages view switching
│   ├── commands/          # Slash command handlers
│   │   └── handler.go     # Command processing
│   ├── tui/               # Chat TUI implementation
│   │   ├── model.go       # Bubble Tea model
│   │   └── styles.go      # UI styling
│   ├── agentlist/         # Agent list browser
│   │   ├── model.go       # Agent browser model
│   │   └── styles.go      # Agent browser styles
│   ├── settings/          # Settings management
│   │   ├── config.go      # Configuration
│   │   ├── model.go       # Settings model
│   │   └── styles.go      # Settings styles
│   └── extensions/        # Extensions browser
│       ├── config.go      # Extensions config loader
│       ├── model.go       # Browser model
│       └── styles.go      # Browser styles
├── extensions/            # Modular extensions
│   ├── tamagotchi/        # Tamagotchi game extension
│   │   ├── pet/           # Pet logic
│   │   │   └── pet.go     # Pet implementation
│   │   ├── tui/           # Tamagotchi TUI
│   │   │   ├── model.go   # Game model
│   │   │   └── styles.go  # Game styles
│   │   ├── choose.go      # Pet selection
│   │   └── README.md      # Extension docs
│   └── dino/              # Dino Runner game extension
│       ├── game/          # Game logic
│       │   ├── game.go    # Physics & collision
│       │   └── storage.go # High score persistence
│       ├── tui/           # Dino TUI
│       │   └── model.go   # Game rendering
│       └── README.md      # Extension docs
├── pkg/
│   └── chatbot/           # Chatbot logic
│       └── bot.go         # Bot implementation
├── extensions.json        # Extensions registry
├── cli_models.json        # AI models configuration
├── index.html             # Web terminal frontend
├── server.js              # Web terminal server (Node.js)
├── package.json           # Node.js dependencies
├── main.go                # Entry point
└── README.md              # This file
```

## Architecture

### Cobra CLI

The application uses Cobra for command-line interface management:
- `root.go`: Defines the base command and global flags
- `chat.go`: Implements the chat subcommand

### Bubble Tea TUI

The TUI is built with the Elm Architecture pattern:
- **Model**: Represents application state
- **Update**: Handles events and updates state
- **View**: Renders the UI

### Components

- **Viewport**: Scrollable message history
- **Textarea**: Multi-line input field
- **Lipgloss**: Beautiful styling and layout

## Chatbot Features

The mock chatbot includes:
- Greeting responses
- Small talk capabilities
- Joke telling
- Context-aware responses
- Question handling

## Extensions

The application features a modular extension system that allows you to launch mini-applications from within the chat interface.

### Available Extensions

#### 🐾 Tamagotchi
A virtual pet game where you care for your digital companion!

**Features:**
- Choose from different pet types (Cat, Dog, Dragon)
- Feed, play, and heal your pet
- Watch stats like hunger, happiness, and health
- Persistent pet state across sessions

**Commands:**
- `feed` - Feed your pet
- `play` - Play with your pet
- `heal` - Heal your pet
- `status` - Check pet stats
- `quit` - Exit the game

#### 🦖 Dino Runner
Chrome's classic T-Rex endless runner game, recreated for your terminal!

**Features:**
- Jump and duck mechanics with realistic physics
- Progressive difficulty (speed increases over time)
- Two obstacle types: Cacti and Birds
- High score tracking with persistence
- ASCII art rendering for compatibility
- Smooth 30 FPS animation

**Gameplay:**
- Avoid cacti on the ground by jumping
- Duck under flying birds
- Score increases based on distance traveled
- Game speed gradually increases
- Beat your high score!

**Technical Details:**
- Gravity-based physics system
- Real-time collision detection
- Procedural obstacle generation
- Persistent storage in `~/.cli_go/dino/save.json`

## Development

### Adding New Extensions

To create a new extension, follow this pattern:

1. **Add to extensions.json:**
```json
{
  "id": "myext",
  "name": "My Extension",
  "description": "Description of my extension",
  "command": "myext",
  "icon": "🎯"
}
```

2. **Create extension structure:**
```bash
mkdir -p extensions/myext/{core,tui}
```

3. **Implement core logic** (`extensions/myext/core/logic.go`)
4. **Implement TUI** (`extensions/myext/tui/model.go`) using Bubble Tea
5. **Create Cobra command** (`cmd/myext.go`)
6. **Hook into extensions browser** (`cmd/extensions.go`)
7. **Integrate with coordinator** (`internal/coordinator/model.go`) for Ctrl+A switching

See `extensions/dino/` for a complete example.

### Adding New Commands

```go
// In cmd/newcommand.go
var newCmd = &cobra.Command{
    Use:   "new",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
        return nil
    },
}

func init() {
    rootCmd.AddCommand(newCmd)
}
```

### Customizing the Bot

Edit `pkg/chatbot/bot.go` to add new response patterns:

```go
case strings.Contains(input, "your-pattern"):
    return "Your response"
```

### Styling

Modify `internal/tui/styles.go` to customize colors and formatting:

```go
var customStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("123")).
    Bold(true)
```

## Future Enhancements

### ChatBot TUI
- [ ] Integrate with real AI APIs (OpenAI, Anthropic, etc.)
- [ ] Persistent chat history
- [ ] Multiple conversation threads
- [ ] Configuration file support
- [ ] Custom themes
- [ ] Export conversations
- [ ] Typing indicators
- [ ] Message timestamps

### Extensions
- [x] Tamagotchi virtual pet game
- [x] Dino Runner endless runner game
- [ ] Snake game
- [ ] Tetris
- [ ] Pomodoro timer
- [ ] File browser
- [ ] System monitor
- [ ] Note-taking app

### Web Terminal
- [x] Multiple terminal instances (up to 10)
- [ ] Terminal session persistence
- [ ] Customizable terminal themes
- [ ] File upload/download support
- [ ] Terminal sharing and collaboration
- [ ] Authentication and security

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - feel free to use this project for learning or commercial purposes.
