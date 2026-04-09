# ChatBot TUI

A beautiful terminal-based chatbot interface built with Go, Bubble Tea, and Cobra.

## Features

- **Interactive TUI**: Beautiful terminal user interface with smooth scrolling
- **Real-time Chat**: Instant responses with an intelligent mock chatbot
- **Seamless Navigation**: Press Ctrl+A to switch between Chat and Extensions (state preserved!)
- **Keyboard Controls**: Fully keyboard-driven interface
- **Responsive Design**: Adapts to terminal window size
- **Clean Architecture**: Well-organized Go project structure
- **Extensions**: Modular extension system with built-in Tamagotchi game

## Prerequisites

- Go 1.18 or higher
- Terminal with UTF-8 support

## Installation

```bash
# Clone the repository
git clone <your-repo-url>
cd cli_go

# Install dependencies
go mod download

# Build the application
go build -o chatbot-tui .
```

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

### Keyboard Controls

**Global Navigation:**
- **Ctrl+A**: Switch between Chat and Extensions browser (state is preserved!)
- **Esc**: Return to Chat (from any extension)

**Chat Mode:**
- **Enter**: Send message
- **Alt+Enter**: New line
- **Ctrl+N**: Toggle sidebar
- **Ctrl+C** or **Esc**: Quit the application
- **Arrow Keys**: Navigate through chat history

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

### Quick Navigation Workflow

1. Start chatting: `./chatbot-tui chat`
2. Press **Ctrl+A** to browse extensions while keeping your chat history
3. Select and launch an extension (e.g., Tamagotchi)
4. Press **Ctrl+A** to return to your chat - all messages preserved!
5. Switch back and forth as much as you want - everything stays in memory

### Slash Commands

The chat interface supports slash commands for special actions. Simply type the command in the input field and press Enter.

**Available Commands:**
- `/exit` or `/quit` - Exit the application gracefully

More commands can be added in the future!

## Project Structure

```
.
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── chat.go            # Chat command
│   ├── settings.go        # Settings command
│   ├── extensions.go      # Extensions browser
│   └── tamagotchi.go      # Tamagotchi command
├── internal/
│   ├── coordinator/       # Main navigation coordinator
│   │   └── model.go       # Manages view switching
│   ├── commands/          # Slash command handlers
│   │   └── handler.go     # Command processing
│   ├── tui/               # Chat TUI implementation
│   │   ├── model.go       # Bubble Tea model
│   │   └── styles.go      # UI styling
│   ├── settings/          # Settings management
│   │   ├── config.go      # Configuration
│   │   ├── model.go       # Settings model
│   │   └── styles.go      # Settings styles
│   └── extensions/        # Extensions browser
│       ├── config.go      # Extensions config loader
│       ├── model.go       # Browser model
│       └── styles.go      # Browser styles
├── extensions/            # Modular extensions
│   └── tamagotchi/        # Tamagotchi game extension
│       ├── pet/           # Pet logic
│       │   └── pet.go     # Pet implementation
│       ├── tui/           # Tamagotchi TUI
│       │   ├── model.go   # Game model
│       │   └── styles.go  # Game styles
│       ├── choose.go      # Pet selection
│       └── README.md      # Extension docs
├── pkg/
│   └── chatbot/           # Chatbot logic
│       └── bot.go         # Bot implementation
├── extensions.json        # Extensions registry
├── cli_models.json        # AI models configuration
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

## Development

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

- [ ] Integrate with real AI APIs (OpenAI, Anthropic, etc.)
- [ ] Persistent chat history
- [ ] Multiple conversation threads
- [ ] Configuration file support
- [ ] Custom themes
- [ ] Export conversations
- [ ] Typing indicators
- [ ] Message timestamps

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - feel free to use this project for learning or commercial purposes.
