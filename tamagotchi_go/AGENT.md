# Tamagotchi TUI - Agent Documentation

## Project Overview

A terminal-based virtual pet game (Tamagotchi) built with Go, using the Bubble Tea TUI framework and Cobra CLI library. Features real-time stat degradation, ASCII art animations, and interactive pet care mechanics.

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
│                    cmd/play.go                               │
│              (Play Subcommand)                               │
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
│  │  - pet: Tamagotchi instance                    │         │
│  │  - textarea: Single-line command input         │         │
│  │  - messages: Activity log                      │         │
│  │  - width/height: Window dimensions             │         │
│  └────────────────────────────────────────────────┘         │
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │ Init()                                         │         │
│  │  - Starts textarea blinking                    │         │
│  │  - Starts 1-second tick timer                  │         │
│  └────────────────────────────────────────────────┘         │
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │ Update(msg) → (Model, Cmd)                     │         │
│  │  - WindowSizeMsg: Resize textarea              │         │
│  │  - KeyMsg(Enter): Process command              │         │
│  │  - KeyMsg(Esc/Ctrl+C): Quit                    │         │
│  │  - tickMsg: Update pet stats every 1s          │         │
│  └────────────────────────────────────────────────┘         │
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │ View() → string                                │         │
│  │  - Renders header                              │         │
│  │  - Renders pet ASCII art                       │         │
│  │  - Renders status bar                          │         │
│  │  - Renders recent messages                     │         │
│  │  - Renders input field                         │         │
│  └────────────────────────────────────────────────┘         │
└──────────────────────┬───────────────────┬──────────────────┘
                       │                   │
                       ▼                   ▼
        ┌──────────────────────┐  ┌──────────────────────┐
        │ internal/tui/        │  │ pkg/tamagotchi/      │
        │ styles.go            │  │ pet.go               │
        │                      │  │                      │
        │ - titleStyle         │  │ Pet struct           │
        │ - subtitleStyle      │  │ - Name               │
        │ - petStyle           │  │ - Age                │
        │ - statusStyle        │  │ - Hunger (0-100)     │
        │ - messageStyle       │  │ - Happiness (0-100)  │
        │ - infoStyle          │  │ - Health (0-100)     │
        │                      │  │ - LastUpdate (time)  │
        │ Lipgloss styling     │  │ - random (RNG)       │
        │ configuration        │  │                      │
        │                      │  │ Mood enum            │
        │                      │  │ - Happy              │
        │                      │  │ - Neutral            │
        │                      │  │ - Sad                │
        │                      │  │ - Sick               │
        │                      │  │ - Dead               │
        │                      │  │                      │
        │                      │  │ Methods:             │
        │                      │  │ - Update()           │
        │                      │  │ - Feed()             │
        │                      │  │ - Play()             │
        │                      │  │ - Heal()             │
        │                      │  │ - GetMood()          │
        │                      │  │ - GetASCII()         │
        │                      │  │ - GetStatus()        │
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
- Parent to `play.go` command

---

#### `cmd/play.go`
**Purpose**: Implements the `play` subcommand  
**Responsibilities**:
- Creates new TUI model
- Initializes Bubble Tea program with options
- Handles program errors

**Key Components**:
- `playCmd`: Cobra command for playing the game
- `RunE`: Command execution function

**Bubble Tea Options**:
- `tea.WithAltScreen()`: Uses alternate screen buffer
- `tea.WithMouseCellMotion()`: Enables mouse support

**Interactions**:
```
playCmd → tui.NewModel() → tea.NewProgram() → p.Run()
```

---

### 📁 **internal/tui/** (Presentation Layer)

#### `internal/tui/model.go`
**Purpose**: Core TUI logic following Bubble Tea's Elm Architecture  
**Responsibilities**:
- Game state management
- Command processing
- Real-time pet updates
- UI rendering

**Custom Message Types**:
```go
type tickMsg time.Time  // 1-second timer for pet updates
```

**Model Structure**:
```go
type Model struct {
    pet      *tamagotchi.Pet  // The virtual pet
    textarea textarea.Model   // Command input
    width    int              // Terminal width
    height   int              // Terminal height
    messages []string         // Activity log (max 10)
    err      error            // Error state
}
```

**Key Functions**:

1. **`NewModel() Model`**
   - Creates initial model instance
   - Initializes pet named "Mochi"
   - Configures single-line textarea for commands
   - Sets placeholder: "Enter command: feed, play, heal, status, or quit"
   - Applies grey background
   - Adds welcome message

2. **`Init() tea.Cmd`**
   - Returns batched commands:
     - `textarea.Blink`: Cursor blinking
     - `tickCmd()`: 1-second timer

3. **`tickCmd() tea.Cmd`**
   - Creates recurring 1-second tick
   - Returns `tickMsg` with current time

4. **`Update(msg tea.Msg) (tea.Model, tea.Cmd)`**
   - **Event Handling**:
     - `tea.WindowSizeMsg`: Resizes textarea
     - `tea.KeyMsg`:
       - `Ctrl+C / Esc`: Quit application
       - `Enter`: Process command
         1. Parse command (lowercase, trimmed)
         2. Execute command:
            - `feed` → `pet.Feed()`
            - `play` → `pet.Play()`
            - `heal` → `pet.Heal()`
            - `status` → `pet.GetStatus()`
            - `quit/exit` → Quit
            - `help` → Show commands
            - Unknown → Error message
         3. Add user input to messages
         4. Add response to messages
         5. Keep only last 10 messages
         6. Clear textarea
     - `tickMsg`: 
       - Call `pet.Update()` (stat degradation)
       - Return new `tickCmd()` for next tick
     - `error`: Store error state

5. **`View() string`**
   - Renders complete game UI
   - Layout: header + pet + messages + input
   - Uses `lipgloss.JoinVertical()`

6. **`renderHeader() string`**
   - Title: "🐾 Tamagotchi Game"
   - Subtitle: Command list
   - Horizontal separator

7. **`renderPet() string`**
   - ASCII art from `pet.GetASCII()`
   - Status bar from `pet.GetStatus()`
   - Visual representation of pet's mood
   - Horizontal separator

8. **`renderMessages() string`**
   - Shows "Recent Activity:" header
   - Lists last 10 messages
   - Activity log of interactions

9. **`renderInput() string`**
   - Horizontal separator
   - Textarea input field
   - Info bar with controls

**Game Loop**:
```
Every 1 second:
  tickMsg → Update() → pet.Update() → tickCmd()
  
User Input:
  Command → Update() → pet.Method() → Add to messages
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
titleStyle      // Pink, bold, padded title
subtitleStyle   // Grey, subtle subtitle
petStyle        // Yellow/cream colored ASCII art
statusStyle     // Green, bold status bar
messageStyle    // Grey activity messages
infoStyle       // Grey info text
```

**Color Palette**:
- `205`: Pink/Magenta (title)
- `241`/`240`: Grey (subtitle, info, messages)
- `229`: Yellow/Cream (pet ASCII art)
- `86`: Green (status bar)
- `235`: Dark grey (textarea background)

---

### 📁 **pkg/tamagotchi/** (Game Logic Layer)

#### `pkg/tamagotchi/pet.go`
**Purpose**: Core Tamagotchi game logic  
**Responsibilities**:
- Pet state management
- Stat calculations
- Mood determination
- ASCII art rendering
- Time-based degradation

**Types**:

1. **`Mood` (string)**
   ```go
   const (
       MoodHappy   Mood = "happy"
       MoodNeutral Mood = "neutral"
       MoodSad     Mood = "sad"
       MoodSick    Mood = "sick"
       MoodDead    Mood = "dead"
   )
   ```

2. **`Pet` struct**
   ```go
   type Pet struct {
       Name       string      // Pet's name
       Age        int         // Age (not currently used)
       Hunger     int         // 0-100, higher = more hungry
       Happiness  int         // 0-100, higher = happier
       Health     int         // 0-100, higher = healthier
       LastUpdate time.Time   // Last stat update time
       random     *rand.Rand  // RNG for varied responses
   }
   ```

**Key Functions**:

1. **`NewPet(name string) *Pet`**
   - Creates new pet with given name
   - Initial stats:
     - Hunger: 30%
     - Happiness: 80%
     - Health: 100%
   - Sets `LastUpdate` to current time
   - Initializes RNG

2. **`Update()`** ⏰
   - **Purpose**: Time-based stat degradation
   - **Trigger**: Called every second by `tickMsg`
   - **Logic**:
     ```
     elapsed = now - LastUpdate
     intervals = elapsed / 10 seconds
     
     For each 10-second interval:
       Hunger += 5%  (max 100)
       Happiness -= 3%  (min 0)
       
       If Hunger > 80% OR Happiness < 20%:
         Health -= 2%  (min 0)
     
     LastUpdate = now
     ```
   - **Result**: Pet deteriorates over time if neglected

3. **`Feed() string`** 🍖
   - **Purpose**: Reduce hunger
   - **Logic**:
     ```
     If Hunger < 20%:
       Return "not hungry" message
     
     Hunger -= 30%  (min 0)
     Happiness += 10%  (max 100)
     
     Return random success message
     ```
   - **Messages**: 
     - "*nom nom nom* 🍖"
     - "Yummy! Thank you! 😋"
     - "*happily munching* ✨"
     - "That was delicious! 🌟"

4. **`Play() string`** 🎾
   - **Purpose**: Increase happiness
   - **Logic**:
     ```
     If Health < 30%:
       Return "too sick to play"
     
     If Hunger > 70%:
       Return "too hungry to play"
     
     Happiness += 20%  (max 100)
     Hunger += 10%  (max 100)
     
     Return random success message
     ```
   - **Side Effect**: Playing increases hunger!
   - **Messages**:
     - "*bounces around excitedly* 🎾"
     - "Wheee! This is fun! 🎉"
     - "*plays happily* ⭐"
     - "Yay! Let's play more! 🎈"

5. **`Heal() string`** 💊
   - **Purpose**: Restore health
   - **Logic**:
     ```
     If Health > 80%:
       Return "already healthy" message
     
     Health += 30%  (max 100)
     
     Return random success message
     ```
   - **Messages**:
     - "*feels much better* 💊"
     - "Thank you! I'm feeling better now! 💚"
     - "*health restored* ✨"
     - "All better! 🌈"

6. **`GetMood() Mood`**
   - **Purpose**: Determine current mood based on stats
   - **Priority Logic**:
     ```
     If Health == 0:
       Return MoodDead
     
     If Health < 30 OR Hunger > 80:
       Return MoodSick
     
     If Happiness > 60 AND Hunger < 50:
       Return MoodHappy
     
     If Happiness < 30 OR Hunger > 60:
       Return MoodSad
     
     Otherwise:
       Return MoodNeutral
     ```

7. **`GetASCII() string`**
   - **Purpose**: Return ASCII art based on mood
   - **ASCII Art**:

   **Happy** (`(ﾟ､ ｡７`):
   ```
      ／l、
     （ﾟ､ ｡７
      l  ~ヽ
      じしf_,)ノ
   ```

   **Neutral** (`(･ω･)`):
   ```
      ／l、
     （ ･ω･）
      l  ~ヽ
      じしf_,)ノ
   ```

   **Sad** (`(˘︹˘)`):
   ```
      ／l、
     （ ˘︹˘ ）
      l  ~ヽ
      じしf_,)ノ
   ```

   **Sick** (`(×_×)`):
   ```
      ／l、
     （ ×_× ）
      l  ~ヽ
      じしf_,)ノ
   ```

   **Dead** (`(X_X)`):
   ```
      ／l、
     （ X_X ）
      l  ~ヽ
      じしf_,)ノ
       R.I.P
   ```

8. **`GetStatus() string`**
   - **Purpose**: Format status bar string
   - **Format**:
     ```
     📊 [Name]'s Status | Health: [H]% | Hunger: [Hu]% | Happiness: [Ha]% | Age: [A] | Mood: [M]
     ```
   - **Example**:
     ```
     📊 Mochi's Status | Health: 85% | Hunger: 45% | Happiness: 70% | Age: 0 | Mood: happy
     ```

9. **`randomChoice(options []string) string`**
   - Helper function for varied responses
   - Randomly selects from array

**Helper Functions**:
```go
func min(a, b int) int  // Returns minimum
func max(a, b int) int  // Returns maximum
```

---

## Game Mechanics

### Stat System

#### Hunger (0-100)
- **Starting**: 30%
- **Increase**: +5% per 10 seconds
- **Feed**: -30%
- **Play**: +10%
- **Critical**: >80% damages health

#### Happiness (0-100)
- **Starting**: 80%
- **Decrease**: -3% per 10 seconds
- **Feed**: +10%
- **Play**: +20%
- **Critical**: <20% damages health

#### Health (0-100)
- **Starting**: 100%
- **Damage**: -2% per 10 seconds when critical
- **Heal**: +30%
- **Death**: Health = 0

### Mood Determination

```
Health == 0           → Dead 💀
Health < 30 || Hunger > 80 → Sick 🤒
Happiness > 60 && Hunger < 50 → Happy 😊
Happiness < 30 || Hunger > 60 → Sad 😢
Otherwise             → Neutral 😐
```

### Time-Based Degradation

**Every 10 seconds**:
1. Hunger increases by 5%
2. Happiness decreases by 3%
3. If critical (Hunger >80% OR Happiness <20%):
   - Health decreases by 2%

**Death Condition**:
- When Health reaches 0%, pet dies
- ASCII changes to dead state
- Game over (pet cannot be revived)

---

## Data Flow

### Game Loop (Every Second)

```
1. tickCmd() fires
   ↓
2. tickMsg → Update()
   ↓
3. pet.Update() checks elapsed time
   ↓
4. If 10+ seconds elapsed:
   - Hunger += 5%
   - Happiness -= 3%
   - If critical: Health -= 2%
   ↓
5. View() re-renders with new stats
   ↓
6. Return tickCmd() for next iteration
```

### Command Processing

```
1. User types command in textarea
   ↓
2. User presses Enter
   ↓
3. KeyMsg(Enter) → Update()
   ↓
4. Parse input: strings.ToLower(strings.TrimSpace())
   ↓
5. Switch on command:
   - feed → pet.Feed()
   - play → pet.Play()
   - heal → pet.Heal()
   - status → pet.GetStatus()
   - quit → tea.Quit
   - help → Show commands
   - default → Error message
   ↓
6. Add to messages:
   messages.append("You: " + input)
   messages.append(response)
   ↓
7. Keep last 10 messages
   ↓
8. Clear textarea
   ↓
9. View() re-renders
```

### Pet State Update

```
Command → Update() → pet.Method()
                         ↓
                    Check conditions
                         ↓
                    Modify stats
                         ↓
                    Return message
                         ↓
                    Display to user
```

---

## Component Interactions

### Bubble Tea Timer System

**tickMsg Pattern**:
```go
// Custom message type
type tickMsg time.Time

// Create tick command
func tickCmd() tea.Cmd {
    return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

// Handle in Update()
case tickMsg:
    m.pet.Update()
    return m, tickCmd()  // Schedule next tick
```

**Flow**:
```
Init() → tickCmd() → Wait 1s → tickMsg → Update() → pet.Update() → tickCmd() → ...
```

### Textarea Component

**Configuration**:
```go
ta.Placeholder = "Enter command: feed, play, heal, status, or quit"
ta.Prompt = "➤ "        // Command prompt
ta.CharLimit = 100      // Max chars
ta.SetHeight(1)         // Single line
ta.KeyMap.InsertNewline.SetEnabled(false)  // No newlines
```

**Grey Background**:
```go
ta.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("235"))
ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("235"))
```

---

## Extension Points

### Adding New Commands

**In `internal/tui/model.go`**:
```go
case "newcommand":
    response = "Command response"
```

### Adding New Pet Actions

**In `pkg/tamagotchi/pet.go`**:
```go
func (p *Pet) NewAction() string {
    p.Update()
    
    // Validation logic
    if someCondition {
        return "Cannot do this"
    }
    
    // Modify stats
    p.SomeStat += value
    
    // Return message
    return p.randomChoice([]string{
        "Message 1",
        "Message 2",
    })
}
```

### Adding New Moods

1. **Add mood constant**:
   ```go
   MoodExcited Mood = "excited"
   ```

2. **Update `GetMood()`**:
   ```go
   case someCondition:
       return MoodExcited
   ```

3. **Add ASCII art**:
   ```go
   case MoodExcited:
       return `excited ASCII art`
   ```

### Persistence (Save/Load)

**Add to `pkg/tamagotchi/pet.go`**:
```go
import "encoding/json"

func (p *Pet) Save(filename string) error {
    data, _ := json.Marshal(p)
    return os.WriteFile(filename, data, 0644)
}

func LoadPet(filename string) (*Pet, error) {
    data, _ := os.ReadFile(filename)
    var pet Pet
    json.Unmarshal(data, &pet)
    return &pet, nil
}
```

**Add commands**:
- `save` → `pet.Save("mochi.json")`
- `load` → Load pet state

---

## Key Design Patterns

### 1. **Elm Architecture** (Bubble Tea)
- **Model**: Game state (pet, messages, UI)
- **Update**: Pure state transitions
- **View**: Pure rendering

### 2. **Command Pattern** (Cobra)
- Each subcommand is separate
- Composable command structure

### 3. **State Machine** (Pet Moods)
- Mood transitions based on stats
- Clear state definitions

### 4. **Observer Pattern** (Time-based Updates)
- Timer observes time
- Pet state updates automatically

---

## Testing Recommendations

### Unit Tests

**`pkg/tamagotchi/pet_test.go`**:
```go
func TestFeedReducesHunger(t *testing.T) {
    pet := NewPet("Test")
    pet.Hunger = 50
    pet.Feed()
    
    if pet.Hunger > 20 {
        t.Errorf("Expected hunger < 20, got %d", pet.Hunger)
    }
}

func TestHealthDeathCondition(t *testing.T) {
    pet := NewPet("Test")
    pet.Health = 0
    
    if pet.GetMood() != MoodDead {
        t.Error("Pet should be dead when health = 0")
    }
}

func TestStatBounds(t *testing.T) {
    pet := NewPet("Test")
    
    // Test upper bound
    pet.Happiness = 90
    pet.Play()
    if pet.Happiness > 100 {
        t.Error("Happiness exceeded 100")
    }
    
    // Test lower bound
    pet.Hunger = 10
    pet.Feed()
    if pet.Hunger < 0 {
        t.Error("Hunger went below 0")
    }
}
```

### Integration Tests

**Test game loop**:
```go
func TestGameLoop(t *testing.T) {
    pet := NewPet("Test")
    initialHunger := pet.Hunger
    
    // Simulate 20 seconds (2 intervals)
    pet.LastUpdate = time.Now().Add(-20 * time.Second)
    pet.Update()
    
    expectedHunger := initialHunger + 10
    if pet.Hunger != expectedHunger {
        t.Errorf("Expected hunger %d, got %d", expectedHunger, pet.Hunger)
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
go build -o tamagotchi-tui .

# Run
./tamagotchi-tui play

# Development (no build)
go run . play
```

---

## Common Issues & Solutions

### Issue: Pet stats not updating
**Cause**: `tickCmd()` not being called  
**Solution**: Ensure `Init()` returns `tea.Batch(textarea.Blink, tickCmd())`

### Issue: Stats change too fast/slow
**Cause**: Degradation interval  
**Solution**: Adjust in `pet.Update()`:
```go
intervals := int(elapsed.Seconds() / 10)  // Change 10 to desired seconds
```

### Issue: Pet never dies
**Cause**: Health not reaching 0  
**Solution**: Verify critical condition logic:
```go
if p.Hunger > 80 || p.Happiness < 20 {
    p.Health = max(0, p.Health-intervals*2)
}
```

### Issue: Commands not recognized
**Cause**: Case sensitivity  
**Solution**: Convert to lowercase:
```go
input := strings.ToLower(strings.TrimSpace(m.textarea.Value()))
```

---

## Gameplay Tips

**For Players**:
- Check status every 20-30 seconds
- Feed when hunger > 50%
- Play when happiness < 70%
- Heal immediately when health < 50%
- Balance feeding and playing (playing increases hunger)

**Survival Strategy**:
```
Optimal loop:
1. Feed (if hunger > 50%)
2. Play (if happiness < 70%)
3. Heal (if health < 70%)
4. Repeat every 20-30 seconds
```

---

## Future Enhancement Ideas

### Gameplay
- [ ] Multiple pets (each with own state)
- [ ] Pet evolution (baby → adult → senior)
- [ ] Pet breeding (combine stats)
- [ ] Mini-games (earn points)
- [ ] Achievements (milestones)
- [ ] Inventory system (items)

### Technical
- [ ] Save/load pet state (JSON persistence)
- [ ] High score tracking (longest survival)
- [ ] Custom pet names (at start)
- [ ] Configurable difficulty
- [ ] Sound effects (terminal bell)
- [ ] Animated ASCII transitions

### UI
- [ ] Color-coded stats (red when critical)
- [ ] Progress bars for stats
- [ ] Pet animation cycles
- [ ] Multiple pet views (side-by-side)

---

## Summary

**Tamagotchi TUI** demonstrates:
- Advanced Bubble Tea patterns (custom messages, timers)
- Game loop implementation
- State machine design
- Time-based mechanics
- Clean architecture

The game is fully functional, engaging, and serves as an excellent foundation for learning TUI game development in Go. The codebase is well-structured for extensions and modifications.
