package chatbot

import (
	"fmt"
	"strings"

	"github.com/yourusername/chatbot-tui/pkg/runtime"
)

// Session holds per-session state: messages, streaming flags, SSE channel.
type Session struct {
	SessionID    string
	AgentID      string
	AgentName    string
	Messages     []Message
	IsStreaming   bool
	WasThinking  bool
	ThinkingText string
	SSEChannel   <-chan runtime.SSEEvent
}

func (s *Session) AppendChunk(chunk string) {
	if len(s.Messages) == 0 {
		return
	}
	lastIdx := len(s.Messages) - 1
	if s.Messages[lastIdx].Content == "⏳ Thinking..." {
		s.Messages[lastIdx].Content = chunk
	} else {
		s.CloseThinkingBlock()
		s.Messages[lastIdx].Content += chunk
	}
}

func (s *Session) AppendThinking(text string) {
	if len(s.Messages) == 0 {
		return
	}
	lastIdx := len(s.Messages) - 1
	if s.Messages[lastIdx].Content == "⏳ Thinking..." {
		s.Messages[lastIdx].Content = ""
	}
	if !s.WasThinking {
		s.WasThinking = true
		s.ThinkingText = ""
	}
	s.ThinkingText += text
	// Rebuild: keep everything before this thinking block, then the open block
	prefix := s.Messages[lastIdx].Content
	oldOpen := "{{THINKING}}" + s.ThinkingText[:len(s.ThinkingText)-len(text)]
	if strings.HasSuffix(prefix, oldOpen) {
		prefix = prefix[:len(prefix)-len(oldOpen)]
	}
	s.Messages[lastIdx].Content = prefix + "{{THINKING}}" + s.ThinkingText
}

func (s *Session) AppendToolStart(name, args string) {
	if len(s.Messages) == 0 {
		return
	}
	lastIdx := len(s.Messages) - 1
	if s.Messages[lastIdx].Content == "⏳ Thinking..." {
		s.Messages[lastIdx].Content = ""
	}
	s.CloseThinkingBlock()
	if args == "" {
		args = "{}"
	}
	s.Messages[lastIdx].Content += fmt.Sprintf("\n\n> 🔧 **%s**\n```json\n%s\n```\n", name, args)
}

func (s *Session) AppendToolEnd(name, result string, isError bool) {
	if len(s.Messages) == 0 {
		return
	}
	lastIdx := len(s.Messages) - 1
	if len(result) > 500 {
		result = result[:500] + "..."
	}
	status := "✅"
	if isError {
		status = "❌"
	}
	s.Messages[lastIdx].Content += fmt.Sprintf("> %s **%s** result:\n```\n%s\n```\n\n", status, name, result)
}

func (s *Session) CloseThinkingBlock() {
	if !s.WasThinking {
		return
	}
	s.WasThinking = false
	if len(s.Messages) > 0 {
		lastIdx := len(s.Messages) - 1
		open := "{{THINKING}}" + s.ThinkingText
		if strings.HasSuffix(s.Messages[lastIdx].Content, open) {
			s.Messages[lastIdx].Content = s.Messages[lastIdx].Content[:len(s.Messages[lastIdx].Content)-len(open)] + "{{THINKING}}" + s.ThinkingText + "{{/THINKING}}"
		}
	}
	s.ThinkingText = ""
}

func (s *Session) StartStreaming(ch <-chan runtime.SSEEvent) {
	s.SSEChannel = ch
	s.IsStreaming = true
	s.Messages = append(s.Messages, Message{
		Role:    RoleBot,
		Content: "⏳ Thinking...",
	})
}

func (s *Session) FinishStreaming() {
	s.IsStreaming = false
	s.WasThinking = false
	s.ThinkingText = ""
	s.SSEChannel = nil
}

func (s *Session) AddUserMessage(content string) {
	s.Messages = append(s.Messages, Message{
		Role:    RoleUser,
		Content: content,
	})
}

// SessionManager owns all sessions and tracks the active one.
type SessionManager struct {
	sessions map[string]*Session
	activeID string
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) AddSession(sessionID, agentID, agentName string) *Session {
	s := &Session{
		SessionID: sessionID,
		AgentID:   agentID,
		AgentName: agentName,
		Messages:  []Message{},
	}
	sm.sessions[sessionID] = s
	return s
}

func (sm *SessionManager) GetSession(sessionID string) *Session {
	return sm.sessions[sessionID]
}

func (sm *SessionManager) ActiveSession() *Session {
	return sm.sessions[sm.activeID]
}

func (sm *SessionManager) ActiveID() string {
	return sm.activeID
}

func (sm *SessionManager) SetActive(sessionID string) {
	sm.activeID = sessionID
}

func (sm *SessionManager) AllSessions() []*Session {
	result := make([]*Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		result = append(result, s)
	}
	return result
}

func (sm *SessionManager) SessionsForAgent(agentID string) []*Session {
	var result []*Session
	for _, s := range sm.sessions {
		if s.AgentID == agentID {
			result = append(result, s)
		}
	}
	return result
}

// Session-tagged message types

type SessionStreamChunkMsg struct {
	SessionID string
	Chunk     string
}

type SessionStreamThinkingMsg struct {
	SessionID string
	Text      string
}

type SessionStreamToolStartMsg struct {
	SessionID string
	Name      string
	Args      string
}

type SessionStreamToolEndMsg struct {
	SessionID string
	Name      string
	Result    string
	IsError   bool
}

type SessionStreamDoneMsg struct {
	SessionID string
}

type SessionStreamErrorMsg struct {
	SessionID string
	Err       error
}
