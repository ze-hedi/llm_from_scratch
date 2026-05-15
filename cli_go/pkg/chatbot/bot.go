package chatbot

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/chatbot-tui/pkg/runtime"
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

type StreamThinkingMsg struct {
	Text string
}

type StreamToolStartMsg struct {
	Name string
	Args string
}

type StreamToolEndMsg struct {
	Name    string
	Result  string
	IsError bool
}

type AgentInitializedMsg struct {
	AgentID   string
	SessionID string
	Name      string
}

type Bot struct {
	name         string
	random       *rand.Rand
	SystemPrompt string

	// Runtime integration
	client     *runtime.Client
	sessionID  string
	agentID    string
	modelID    string
	sseChannel <-chan runtime.SSEEvent
}

func NewBot(client *runtime.Client) *Bot {
	return &Bot{
		name:         "ChatBot",
		random:       rand.New(rand.NewSource(time.Now().UnixNano())),
		SystemPrompt: "You are a helpful AI assistant. Be friendly and conversational.",
		client:       client,
	}
}

// Client returns the runtime client (used by slash commands).
func (b *Bot) Client() *runtime.Client {
	return b.client
}

// SessionID returns the active session ID.
func (b *Bot) SessionID() string {
	return b.sessionID
}

// AgentID returns the active agent ID.
func (b *Bot) AgentID() string {
	return b.agentID
}

// InitAgent creates a new agent session on the runtime server.
func (b *Bot) InitAgent(modelID, systemPrompt string) tea.Cmd {
	if b.client == nil {
		return nil
	}
	return func() tea.Msg {
		req := runtime.RunRequest{
			Agent: runtime.AgentData{
				ID:    "tui-chat-default",
				Name:  "TUI Assistant",
				Model: modelID,
			},
		}
		if systemPrompt != "" {
			req.Files = []runtime.FilePayload{
				{Type: "soul", Content: systemPrompt},
			}
		}

		resp, err := b.client.Run(req)
		if err != nil {
			return StreamErrorMsg{Err: fmt.Errorf("init agent: %w", err)}
		}

		b.sessionID = resp.SessionID
		b.agentID = resp.AgentID
		return AgentInitializedMsg{
			AgentID:   resp.AgentID,
			SessionID: resp.SessionID,
			Name:      resp.Name,
		}
	}
}

// SetActiveAgent sets the bot to use an existing agent session.
func (b *Bot) SetActiveAgent(agentID, sessionID, name string) {
	b.agentID = agentID
	b.sessionID = sessionID
}

// GetResponseStream returns a tea.Cmd that streams a response from the runtime.
// Falls back to pattern matching if no client or session is available.
func (b *Bot) GetResponseStream(input string) tea.Cmd {
	if b.client == nil || b.sessionID == "" {
		return func() tea.Msg {
			return StreamChunkMsg{Chunk: b.GetResponse(input)}
		}
	}

	return func() tea.Msg {
		ch, err := b.client.ChatStream(b.sessionID, input)
		if err != nil {
			return StreamErrorMsg{Err: fmt.Errorf("chat stream: %w", err)}
		}
		b.sseChannel = ch
		return b.readNextSSEEvent()
	}
}

// ContinueStream returns a tea.Cmd that reads the next SSE event from the active stream.
func (b *Bot) ContinueStream() tea.Cmd {
	if b.sseChannel == nil {
		return nil
	}
	return func() tea.Msg {
		return b.readNextSSEEvent()
	}
}

func (b *Bot) readNextSSEEvent() tea.Msg {
	event, ok := <-b.sseChannel
	if !ok {
		b.sseChannel = nil
		return StreamDoneMsg{}
	}

	switch event.Type {
	case "delta":
		return StreamChunkMsg{Chunk: event.Text}
	case "thinking":
		return StreamThinkingMsg{Text: event.Text}
	case "tool_start":
		return StreamToolStartMsg{Name: event.Name, Args: string(event.Args)}
	case "tool_end":
		return StreamToolEndMsg{Name: event.Name, Result: event.Result, IsError: event.IsError}
	case "done":
		b.sseChannel = nil
		return StreamDoneMsg{}
	case "error":
		b.sseChannel = nil
		return StreamErrorMsg{Err: fmt.Errorf("agent error: %s", event.Message)}
	default:
		return StreamChunkMsg{Chunk: event.Text}
	}
}

// GetResponse returns a pattern-matched response (offline fallback).
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
