# ChatBot TUI - Agent Documentation

## Project Overview

A terminal-based chatbot application built with Go, using the Bubble Tea TUI framework and Cobra CLI library. The application follows clean architecture principles with clear separation of concerns.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                              │
│                    (Entry Point)                             │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    cmd/root.go                               │
│              (Cobra Root Command)                            │
│  - Defines CLI structure                                     │
│  - Error handling                                            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    cmd/chat.go                               │
│              (Chat Subcommand)                               │
│  - Initializes TUI model                                     │
│  - Launches Bubble Tea program                               │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              internal/tui/model.go                           │
│           (Bubble Tea MVC Pattern)                           │
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │ Model (State)                                  │         │
│  │  - viewport: Scrollable message display        │         │
│  │  - textarea: Multi-line input field            │         │
│  │  - messages: Chat history                      │         │
│  │  - bot: Chatbot instance                       │         │
│  │  - width/height: Window dimensions             │         │
│  └────────────────────────────────────────────────┘         │
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │ Init()                                         │         │
│  │  - Starts textarea blinking cursor             │         │
│  └────────────────────────────────────────────────┘         │
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │ Update(msg) → (Model, Cmd)                     │         │
│  │  - WindowSizeMsg: Resize viewport/textarea     │         │
│  │  - KeyMsg(Enter): Send message, get response   │         │
│  │  - KeyMsg(Esc/Ctrl+C): Quit                    │         │
│  │  - Updates textarea and viewport               │         │
│  └────────────────────────────────────────────────┘         │
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │ View() → string                                │         │
│  │  - Renders header                              │         │
│  │  - Renders viewport (messages)                 │         │
│  │  - Renders footer (textarea + info)            │         │
│  └────────────────────────────────────────────────┘         │
└──────────────────────┬───────────────────┬──────────────────┘
                       │                   │
                       ▼                   ▼
        ┌──────────────────────┐  ┌──────────────────────┐
        │ internal/tui/        │  │ pkg/chatbot/         │
        │ styles.go            │  │ bot.go               │
        │                      │  │                      │
        │ - titleStyle         │  │ Message struct       │
        │ - subtitleStyle      │  │ - Role (user/bot)    │
        │ - userMessageStyle   │  │ - Content (text)     │
        │ - botMessageStyle    │  │                      │
        │ - infoStyle          │  │ Bot struct           │
        │                      │  │ - name               │
        │ Lipgloss styling     │  │ - random (RNG)       │
        │ configuration        │  │                      │
        │                      │  │ GetResponse(input)   │
        │                      │  │ - Pattern matching   │
        │                      │  │ - Mock responses     │
        └──────────────────────┘  └──────────────────────┘
```

---

## File Structure & Responsibilities

### 📁 **Root Level**

#### `main.go`
**Purpose**: Application entry point  
**Responsibilities**:
- Imports the `cmd` package
- Calls `cmd.Execute()` to start the CLI

**Code Flow**:
```go
main() → cmd.Execute()
```

---

### 📁 **cmd/** (Command Layer)

#### `cmd/root.go`
**Purpose**: Defines the root Cobra command  
**Responsibilities**:
- Sets up CLI metadata (name, description)
- Error handling for command execution
- Disables default completion command

**Key Components**:
- `rootCmd`: Cobra command definition
- `Execute()`: Runs the root command
- `init()`: Configures completion options

**Interactions**:
- Called by `main.go`
- Parent to `chat.go` command

---

#### `cmd/chat.go`
**Purpose**: Implements the `chat` subcommand  
**Responsibilities**:
- Creates new TUI model
- Initializes Bubble Tea program with options
- Handles program errors

**Key Components**:
- `chatCmd`: Cobra command for chat
- `RunE`: Command execution function

**Bubble Tea Options**:
- `tea.WithAltScreen()`: Uses alternate screen buffer
- `tea.WithMouseCellMotion()`: Enables mouse support

**Interactions**:
```
chatCmd → tui.NewModel() → tea.NewProgram() → p.Run()
```

---

### 📁 **internal/tui/** (Presentation Layer)

#### `internal/tui/model.go`
**Purpose**: Core TUI logic following Bubble Tea's Elm Architecture  
**Responsibilities**:
- State management
- Event handling
- UI rendering

**Model Structure**:
```go
type Model struct {
    viewport viewport.Model  // Scrollable chat display
    textarea textarea.Model  // User input field
    messages []chatbot.Message  // Chat history
    bot      *chatbot.Bot    // Chatbot instance
    width    int             // Terminal width
    height   int             // Terminal height
    ready    bool            // Initialization flag
    err      error           // Error state
}
```

**Key Functions**:

1. **`NewModel() Model`**
   - Creates initial model instance
   - Configures textarea (placeholder, prompt, size, styling)
   - Applies grey background to entire textarea
   - Creates viewport
   - Initializes bot

2. **`Init() tea.Cmd`**
   - Returns `textarea.Blink` command (cursor blinking)

3. **`Update(msg tea.Msg) (tea.Model, tea.Cmd)`**
   - **Event Handling**:
     - `tea.WindowSizeMsg`: Resizes viewport and textarea
     - `tea.KeyMsg`:
       - `Ctrl+C / Esc`: Quit application
       - `Enter`: Process user input
         1. Get user input
         2. Add to messages as user message
         3. Get bot response
         4. Add to messages as bot message
         5. Update viewport content
         6. Scroll to bottom
         7. Clear textarea
     - `error`: Store error state
   
   - **Component Updates**: Updates textarea and viewport with their messages

4. **`View() string`**
   - Renders complete UI
   - Combines: header + viewport + footer
   - Uses `lipgloss.JoinVertical()` for layout

5. **`renderHeader() string`**
   - Title: "🤖 ChatBot TUI"
   - Subtitle: Instructions
   - Horizontal line separator

6. **`renderMessages() string`**
   - Iterates through message history
   - Applies different styles for user vs bot
   - Format: "You: ..." or "Bot: ..."

7. **`renderFooter() string`**
   - Shows textarea input
   - Info bar with controls and message count

**Interactions**:
```
User Input → Update() → bot.GetResponse() → Update messages → View()
```

---

#### `internal/tui/styles.go`
**Purpose**: Centralized styling configuration  
**Responsibilities**:
- Define all Lipgloss styles
- Color scheme management
- Consistent visual appearance

**Styles Defined**:
```go
titleStyle       // Pink, bold, padded title
subtitleStyle    // Grey, subtle subtitle
userMessageStyle // Green, bold user messages
botMessageStyle  // Purple, bold bot messages
infoStyle        // Grey info text
```

**Color Palette**:
- `205`: Pink/Magenta (titles, accents)
- `241`/`240`: Grey (subtitles, info)
- `86`: Green (user messages)
- `212`: Purple (bot messages)
- `235`: Dark grey (textarea background)

---

### 📁 **pkg/chatbot/** (Business Logic Layer)

#### `pkg/chatbot/bot.go`
**Purpose**: Chatbot logic and response generation  
**Responsibilities**:
- Message structure definition
- Pattern-matching conversation logic
- Mock response generation

**Types**:

1. **`Role` (string)**
   - `RoleUser = "user"`
   - `RoleBot = "bot"`

2. **`Message` struct**
   ```go
   type Message struct {
       Role    Role
       Content string
   }
   ```

3. **`Bot` struct**
   ```go
   type Bot struct {
       name   string      // Bot name
       random *rand.Rand  // RNG for varied responses
   }
   ```

**Key Functions**:

1. **`NewBot() *Bot`**
   - Creates bot instance
   - Initializes random generator with current timestamp

2. **`GetResponse(input string) string`**
   - **Pattern Matching Logic**:
     - Converts input to lowercase
     - Uses `switch` with `strings.Contains()`
   
   - **Response Patterns**:
     - Greetings: "hello", "hi" → Welcome messages
     - Status: "how are you" → Status responses
     - Farewells: "bye", "goodbye" → Goodbye messages
     - Help: "help" → Usage instructions
     - Identity: "name" → Bot name
     - Weather: "weather" → Humorous deflection
     - Entertainment: "joke" → Programming jokes
     - Gratitude: "thank" → Polite acknowledgment
     - Questions: "?" → Thoughtful responses
     - Default: Generic conversational responses

3. **`randomChoice(options []string) string`**
   - Selects random string from array
   - Provides response variation

4. **`generateThought(input string) string`**
   - Creates contextual thoughtful responses
   - Used for question handling

**Response Strategy**:
- Multiple response variations per pattern
- Random selection for natural conversation feel
- Fallback to generic responses for unknown input

---

## Data Flow

### Complete Request-Response Cycle

```
1. User types message in textarea
   ↓
2. User presses Enter
   ↓
3. KeyMsg(Enter) event → Update()
   ↓
4. Extract input: m.textarea.Value()
   ↓
5. Create user message: 
   messages.append(Message{Role: RoleUser, Content: input})
   ↓
6. Call bot: response = m.bot.GetResponse(input)
   ├─→ Pattern matching in bot.go
   ├─→ Random response selection
   └─→ Return response string
   ↓
7. Create bot message:
   messages.append(Message{Role: RoleBot, Content: response})
   ↓
8. Update viewport:
   m.viewport.SetContent(m.renderMessages())
   m.viewport.GotoBottom()
   ↓
9. Clear textarea: m.textarea.Reset()
   ↓
10. View() re-renders UI
   ↓
11. User sees updated chat
```

---

## Component Interactions

### Bubble Tea Components

**Viewport** (`bubbles/viewport`):
- Manages scrollable message display
- Handles overflow with scrolling
- Methods used:
  - `New(width, height)`: Create viewport
  - `SetContent(string)`: Update content
  - `GotoBottom()`: Scroll to end
  - `Update(msg)`: Handle events
  - `View()`: Render visible portion

**Textarea** (`bubbles/textarea`):
- Multi-line input field
- Methods used:
  - `New()`: Create textarea
  - `SetWidth(int)`, `SetHeight(int)`: Resize
  - `Focus()`: Enable input
  - `Update(msg)`: Handle typing
  - `View()`: Render input field
  - `Value()`: Get current text
  - `Reset()`: Clear content

**Styling**:
- `FocusedStyle.Base`: Background color for entire textarea
- `FocusedStyle.CursorLine`: Background for cursor line

---

## Configuration Details

### Textarea Configuration
```go
ta.Placeholder = "Type your message..."  // Hint text
ta.Focus()                               // Auto-focus
ta.Prompt = "┃ "                        // Input prefix
ta.CharLimit = 500                       // Max characters
ta.SetWidth(80)                          // Width
ta.SetHeight(3)                          // Height (lines)
ta.ShowLineNumbers = false               // No line numbers
ta.KeyMap.InsertNewline.SetEnabled(false) // Disable newlines
```

### Grey Background Effect
```go
ta.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))
ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("235"))
```
This creates uniform grey background across all 3 lines.

---

## Extension Points

### Adding New Bot Responses

**In `pkg/chatbot/bot.go`**:
```go
case strings.Contains(input, "keyword"):
    return b.randomChoice([]string{
        "Response 1",
        "Response 2",
        "Response 3",
    })
```

### Adding New Cobra Commands

**Create `cmd/newcommand.go`**:
```go
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

### Modifying Styles

**In `internal/tui/styles.go`**:
```go
var myStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("123")).
    Background(lipgloss.Color("234")).
    Bold(true).
    Padding(1, 2)
```

### Integrating Real AI APIs

**Replace in `internal/tui/model.go`**:
```go
// Current:
response := m.bot.GetResponse(userInput)

// With API:
response, err := m.bot.CallAPI(userInput)
if err != nil {
    response = "Error: " + err.Error()
}
```

**Add to `pkg/chatbot/bot.go`**:
```go
func (b *Bot) CallAPI(input string) (string, error) {
    // HTTP request to OpenAI/Anthropic/etc.
    // Parse response
    // Return result
}
```

---

## Key Design Patterns

### 1. **Elm Architecture** (Bubble Tea)
- **Model**: Immutable state
- **Update**: Pure function for state transitions
- **View**: Pure function for rendering

### 2. **Command Pattern** (Cobra)
- Each subcommand is a separate command object
- Composable command tree

### 3. **Strategy Pattern** (Bot Responses)
- Different response strategies per input pattern
- Easy to add new patterns

### 4. **Separation of Concerns**
- `cmd/`: CLI layer
- `internal/tui/`: Presentation layer
- `pkg/chatbot/`: Business logic layer

---

## Testing Recommendations

### Unit Tests
- `pkg/chatbot/bot_test.go`: Test response patterns
- `internal/tui/model_test.go`: Test state transitions

### Integration Tests
- Test full message flow
- Test window resize handling

### Example Test
```go
func TestBotGreeting(t *testing.T) {
    bot := chatbot.NewBot()
    response := bot.GetResponse("hello")
    
    validResponses := []string{
        "Hello! How can I help you today?",
        "Hi there! What's on your mind?",
        "Hey! Nice to meet you!",
    }
    
    if !contains(validResponses, response) {
        t.Errorf("Unexpected response: %s", response)
    }
}
```

---

## Dependencies

```
github.com/charmbracelet/bubbletea  v0.24.2  // TUI framework
github.com/charmbracelet/bubbles    v0.16.1  // TUI components
github.com/charmbracelet/lipgloss   v0.9.1   // Styling
github.com/spf13/cobra              v1.8.0   // CLI framework
```

---

## Build & Run

```bash
# Build
go build -o chatbot-tui .

# Run
./chatbot-tui chat

# Development (no build)
go run . chat
```

---

## Common Issues & Solutions

### Issue: Textarea only highlights cursor line
**Solution**: Set uniform background on both `Base` and `CursorLine`:
```go
ta.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))
ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("235"))
```

### Issue: Terminal not resizing properly
**Solution**: Handle `tea.WindowSizeMsg` in `Update()`:
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.viewport.Width = msg.Width
    m.viewport.Height = msg.Height - 10
    m.textarea.SetWidth(msg.Width - 4)
```

### Issue: Chat history not scrolling
**Solution**: Call `GotoBottom()` after updating viewport:
```go
m.viewport.SetContent(m.renderMessages())
m.viewport.GotoBottom()
```

---

## Summary

**ChatBot TUI** is a well-architected terminal application that demonstrates:
- Clean separation of concerns
- Proper use of Bubble Tea framework
- Extensible chatbot logic
- Professional CLI design with Cobra
- Maintainable code structure

The codebase is production-ready and serves as an excellent foundation for adding real AI integrations, persistence, or additional features.
