package chatbot

import (
	"fmt"

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

type Bot struct {
	client   *runtime.Client
	Sessions *SessionManager
}

func NewBot(client *runtime.Client) *Bot {
	return &Bot{
		client:   client,
		Sessions: NewSessionManager(),
	}
}

// Client returns the runtime client.
func (b *Bot) Client() *runtime.Client {
	return b.client
}

// StartStream sends a message to a session and begins streaming the response.
func (b *Bot) StartStream(sessionID, input string) tea.Cmd {
	if b.client == nil {
		return nil
	}
	return func() tea.Msg {
		ch, err := b.client.ChatStream(sessionID, input)
		if err != nil {
			return SessionStreamErrorMsg{SessionID: sessionID, Err: fmt.Errorf("chat stream: %w", err)}
		}
		// Store channel on session — must be done here (inside the Cmd goroutine)
		// but session mutation is safe because Bubble Tea processes one Msg at a time
		// and this goroutine returns before the next Update call.
		s := b.Sessions.GetSession(sessionID)
		if s != nil {
			s.SSEChannel = ch
		}
		return b.readNextEvent(sessionID, ch)
	}
}

// ListenToSession returns a tea.Cmd that reads the next SSE event for a session.
func (b *Bot) ListenToSession(sessionID string) tea.Cmd {
	s := b.Sessions.GetSession(sessionID)
	if s == nil || s.SSEChannel == nil {
		return nil
	}
	ch := s.SSEChannel
	return func() tea.Msg {
		return b.readNextEvent(sessionID, ch)
	}
}

func (b *Bot) readNextEvent(sessionID string, ch <-chan runtime.SSEEvent) tea.Msg {
	event, ok := <-ch
	if !ok {
		return SessionStreamDoneMsg{SessionID: sessionID}
	}

	switch event.Type {
	case "delta":
		return SessionStreamChunkMsg{SessionID: sessionID, Chunk: event.Text}
	case "thinking":
		return SessionStreamThinkingMsg{SessionID: sessionID, Text: event.Text}
	case "tool_start":
		return SessionStreamToolStartMsg{SessionID: sessionID, Name: event.Name, Args: string(event.Args)}
	case "tool_end":
		return SessionStreamToolEndMsg{SessionID: sessionID, Name: event.Name, Result: event.Result, IsError: event.IsError}
	case "done":
		return SessionStreamDoneMsg{SessionID: sessionID}
	case "error":
		return SessionStreamChunkMsg{SessionID: sessionID, Chunk: "\n\n**Error:** " + event.Message + "\n\n"}
	default:
		return SessionStreamChunkMsg{SessionID: sessionID, Chunk: event.Text}
	}
}
