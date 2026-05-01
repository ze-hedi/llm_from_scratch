package chatbot

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Role string

const (
	RoleUser Role = "user"
	RoleBot  Role = "bot"
)

type Message struct {
	Role    Role
	Content string
}

// Streaming message types for Bubble Tea
type StreamChunkMsg struct {
	Chunk string
}

type StreamDoneMsg struct{}

type StreamErrorMsg struct {
	Err error
}

// Agent info message types
type AgentInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type FetchAgentListMsg struct{} // Trigger to fetch agent list

type AgentListMsg struct {
	Agents []AgentInfo
}

type AgentListErrorMsg struct {
	Err error
}

// SetAgent message types
type SetAgentMsg struct {
	Response map[string]interface{}
}

type SetAgentErrorMsg struct {
	Err error
}

type Bot struct {
	name         string
	random       *rand.Rand
	SystemPrompt string
	httpClient   *http.Client
	runtimeURL   string // base URL of the otto_code runtime server
	agentId      string // ID of the currently active agent (set after /runtime/run)
	useAgent     bool   // Toggle between pattern-matching and real agent
}

func NewBot() *Bot {
	return &Bot{
		name:         "ChatBot",
		random:       rand.New(rand.NewSource(time.Now().UnixNano())),
		SystemPrompt: "You are a helpful AI assistant. Be friendly and conversational.",
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		runtimeURL: "http://localhost:5000",
		agentId:    "",
		useAgent:   true,
	}
}

func (b *Bot) GetResponse(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))

	switch {
	case strings.Contains(input, "hello") || strings.Contains(input, "hi"):
		return b.randomChoice([]string{
			"Hello! How can I help you today?",
			"Hi there! What's on your mind?",
			"Hey! Nice to meet you!",
		})

	case strings.Contains(input, "how are you"):
		return b.randomChoice([]string{
			"I'm doing great, thanks for asking! How about you?",
			"I'm functioning perfectly! How can I assist you?",
			"Fantastic! Ready to help you with anything you need.",
		})

	case strings.Contains(input, "bye") || strings.Contains(input, "goodbye"):
		return b.randomChoice([]string{
			"Goodbye! Have a great day!",
			"See you later! Feel free to come back anytime.",
			"Take care! It was nice chatting with you.",
		})

	case strings.Contains(input, "help"):
		return "I'm here to chat with you! You can ask me questions, have a conversation, or just say hello. What would you like to talk about?"

	case strings.Contains(input, "name"):
		return fmt.Sprintf("My name is %s. What's yours?", b.name)

	case strings.Contains(input, "weather"):
		return b.randomChoice([]string{
			"I don't have access to real-time weather data, but I hope it's nice where you are!",
			"As an AI, I can't check the weather, but I recommend looking outside or checking a weather app!",
		})

	case strings.Contains(input, "joke"):
		return b.randomChoice([]string{
			"Why don't programmers like nature? It has too many bugs!",
			"Why do programmers prefer dark mode? Because light attracts bugs!",
			"What's a programmer's favorite hangout place? Foo Bar!",
		})

	case strings.Contains(input, "thank"):
		return b.randomChoice([]string{
			"You're welcome! Happy to help!",
			"No problem at all!",
			"Anytime! That's what I'm here for.",
		})

	case strings.Contains(input, "?"):
		return b.randomChoice([]string{
			"That's an interesting question! Let me think... " + b.generateThought(input),
			"Great question! " + b.generateThought(input),
			"Hmm, " + b.generateThought(input),
		})

	default:
		return b.randomChoice([]string{
			"That's interesting! Tell me more.",
			"I see. What else would you like to discuss?",
			"Fascinating! Can you elaborate on that?",
			"I understand. How can I help you with that?",
			"That's a good point. What are your thoughts on it?",
		})
	}
}

func (b *Bot) randomChoice(options []string) string {
	return options[b.random.Intn(len(options))]
}

func (b *Bot) generateThought(input string) string {
	thoughts := []string{
		"I'd need more context to give you a detailed answer, but I'm here to discuss it!",
		"That depends on various factors. What specifically are you curious about?",
		"I think that's something worth exploring further. What's your perspective?",
		"Based on what you're asking, I'd say it's quite nuanced. Want to dive deeper?",
	}
	return b.randomChoice(thoughts)
}

// initAgent calls POST /runtime/run to create a new agent session.
// It stores the returned agent ID in b.agentId for subsequent chat calls.
func (b *Bot) initAgent(agentId, model string) error {
	payload := map[string]interface{}{
		"agent": map[string]interface{}{
			"_id":         agentId,
			"name":        "TUI Assistant",
			"model":       model,
			"description": "TUI Chat assistant",
		},
	}
	if b.SystemPrompt != "" {
		payload["files"] = []map[string]string{
			{"type": "soul", "content": b.SystemPrompt},
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal run request: %w", err)
	}

	req, err := http.NewRequest("POST", b.runtimeURL+"/runtime/run", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create run request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to runtime: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("runtime error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		AgentId string `json:"agentId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse run response: %w", err)
	}

	b.agentId = result.AgentId
	return nil
}

// GetResponseStream returns a tea.Cmd that streams responses from the otto_code runtime.
// It initializes an agent session if one doesn't exist, then streams the chat response via SSE.
func (b *Bot) GetResponseStream(input string) tea.Cmd {
	if !b.useAgent {
		return func() tea.Msg {
			return StreamChunkMsg{Chunk: b.GetResponse(input)}
		}
	}

	return func() tea.Msg {
		// Initialize agent on first use
		if b.agentId == "" {
			if err := b.initAgent("tui-chat-default", "claude-sonnet-4-6"); err != nil {
				return StreamErrorMsg{Err: fmt.Errorf("failed to initialize agent: %w", err)}
			}
		}

		// POST /runtime/chat/:id
		chatURL := fmt.Sprintf("%s/runtime/chat/%s", b.runtimeURL, b.agentId)
		jsonData, err := json.Marshal(map[string]string{"message": input})
		if err != nil {
			return StreamErrorMsg{Err: fmt.Errorf("failed to marshal chat request: %w", err)}
		}

		req, err := http.NewRequest("POST", chatURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return StreamErrorMsg{Err: fmt.Errorf("failed to create chat request: %w", err)}
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		resp, err := b.httpClient.Do(req)
		if err != nil {
			return StreamErrorMsg{Err: fmt.Errorf("failed to connect to runtime: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return StreamErrorMsg{Err: fmt.Errorf("runtime error (status %d): %s", resp.StatusCode, string(body))}
		}

		return b.parseSSEResponse(resp.Body)
	}
}

// parseSSEResponse reads the SSE stream from the otto_code runtime and accumulates
// all delta text into a single StreamChunkMsg.
//
// SSE line format:  data: {"type":"delta","text":"..."}\n\n
// Possible types:   delta | tool_start | tool_end | done | error
func (b *Bot) parseSSEResponse(body io.Reader) tea.Msg {
	var fullText strings.Builder
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := scanner.Text()

		// SSE lines are either "data: <json>" or blank separators
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		raw := strings.TrimPrefix(line, "data: ")

		var event struct {
			Type    string `json:"type"`
			Text    string `json:"text"`
			Message string `json:"message"` // used by "error" events
		}
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			continue // skip malformed lines
		}

		switch event.Type {
		case "delta":
			fullText.WriteString(event.Text)
		case "done":
			// Stream complete — nothing left to read
		case "error":
			return StreamErrorMsg{Err: fmt.Errorf("agent error: %s", event.Message)}
		}
	}

	if err := scanner.Err(); err != nil {
		return StreamErrorMsg{Err: fmt.Errorf("error reading SSE stream: %w", err)}
	}

	return StreamChunkMsg{Chunk: fullText.String()}
}

// SetRuntimeURL updates the base URL of the otto_code runtime server.
// Changing the URL resets the active agent so the next message re-initializes.
func (b *Bot) SetRuntimeURL(url string) {
	b.runtimeURL = url
	b.agentId = ""
}

// SetUseAgent enables or disables the agent (falls back to pattern matching if disabled).
func (b *Bot) SetUseAgent(use bool) {
	b.useAgent = use
}

// IsUsingAgent returns whether the bot is using the real agent.
func (b *Bot) IsUsingAgent() bool {
	return b.useAgent
}

// GetAgentList fetches the list of active agent IDs from the runtime status endpoint.
func (b *Bot) GetAgentList() tea.Cmd {
	return func() tea.Msg {
		req, err := http.NewRequest("GET", b.runtimeURL+"/runtime/status", nil)
		if err != nil {
			return AgentListErrorMsg{Err: fmt.Errorf("failed to connect to server")}
		}

		resp, err := b.httpClient.Do(req)
		if err != nil {
			return AgentListErrorMsg{Err: fmt.Errorf("failed to connect to server")}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return AgentListErrorMsg{Err: fmt.Errorf("failed to connect to server")}
		}

		var status struct {
			ActiveAgents   []string `json:"activeAgents"`
			CurrentAgentId string   `json:"currentAgentId"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return AgentListErrorMsg{Err: fmt.Errorf("failed to parse server response")}
		}

		agents := make([]AgentInfo, 0, len(status.ActiveAgents))
		for _, id := range status.ActiveAgents {
			desc := ""
			if id == status.CurrentAgentId {
				desc = "(current)"
			}
			agents = append(agents, AgentInfo{Name: id, Description: desc})
		}

		return AgentListMsg{Agents: agents}
	}
}

// SetAgent switches the active agent by calling POST /runtime/run with the given agent ID.
// The model defaults to claude-sonnet-4-6; pass a non-empty model string to override.
func (b *Bot) SetAgent(agentName string) tea.Cmd {
	return func() tea.Msg {
		if err := b.initAgent(agentName, "claude-sonnet-4-6"); err != nil {
			return SetAgentErrorMsg{Err: err}
		}
		return SetAgentMsg{Response: map[string]interface{}{
			"success": true,
			"agentId": b.agentId,
		}}
	}
}
