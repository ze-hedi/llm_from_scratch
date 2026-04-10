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

type Bot struct {
	name         string
	random       *rand.Rand
	SystemPrompt string
	httpClient   *http.Client
	agentURL     string
	useAgent     bool // Toggle between pattern-matching and real agent
}

func NewBot() *Bot {
	return &Bot{
		name:         "ChatBot",
		random:       rand.New(rand.NewSource(time.Now().UnixNano())),
		SystemPrompt: "You are a helpful AI assistant. Be friendly and conversational.",
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for streaming
		},
		agentURL: "http://localhost:8001/agent",
		useAgent: true, // Set to false to use pattern-matching bot
	}
}

func (b *Bot) GetResponse(input string) string {
	// Note: SystemPrompt is available and would be used when integrating with a real LLM API
	// For this pattern-matching bot, the SystemPrompt is stored but not actively used in responses
	input = strings.ToLower(strings.TrimSpace(input))

	// Pattern matching for intelligent responses
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
			"Why don't programmers like nature? It has too many bugs! 🐛",
			"Why do programmers prefer dark mode? Because light attracts bugs! 💡",
			"What's a programmer's favorite hangout place? Foo Bar! 🍺",
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

// GetResponseStream returns a tea.Cmd that streams responses from the agent
func (b *Bot) GetResponseStream(input string) tea.Cmd {
	// If agent is disabled, fall back to pattern matching
	if !b.useAgent {
		return func() tea.Msg {
			response := b.GetResponse(input)
			return StreamChunkMsg{Chunk: response}
		}
	}

	return func() tea.Msg {
		// Prepare request body
		reqBody := map[string]interface{}{
			"message":       input,
			"system_prompt": b.SystemPrompt,
		}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return StreamErrorMsg{Err: fmt.Errorf("failed to marshal request: %w", err)}
		}

		// Make HTTP POST request
		req, err := http.NewRequest("POST", b.agentURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return StreamErrorMsg{Err: fmt.Errorf("failed to create request: %w", err)}
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := b.httpClient.Do(req)
		if err != nil {
			return StreamErrorMsg{Err: fmt.Errorf("failed to connect to agent: %w", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return StreamErrorMsg{Err: fmt.Errorf("agent error (status %d): %s", resp.StatusCode, string(body))}
		}

		// Stream the response
		return b.streamResponse(resp.Body)
	}
}

// streamResponse reads the streaming response and returns chunks
func (b *Bot) streamResponse(body io.Reader) tea.Msg {
	var fullResponse strings.Builder
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		chunk := scanner.Text()
		if chunk != "" {
			fullResponse.WriteString(chunk)
			fullResponse.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return StreamErrorMsg{Err: fmt.Errorf("error reading stream: %w", err)}
	}

	// For now, return the full response at once
	// TODO: Make this truly streaming by sending chunks incrementally
	return StreamChunkMsg{Chunk: fullResponse.String()}
}

// SetAgentURL updates the agent endpoint URL
func (b *Bot) SetAgentURL(url string) {
	b.agentURL = url
}

// SetUseAgent enables or disables the agent (falls back to pattern matching if disabled)
func (b *Bot) SetUseAgent(use bool) {
	b.useAgent = use
}

// IsUsingAgent returns whether the bot is using the real agent
func (b *Bot) IsUsingAgent() bool {
	return b.useAgent
}

// GetAgentList fetches the list of available agents from the server
func (b *Bot) GetAgentList() tea.Cmd {
	return func() tea.Msg {
		// Build the URL for the getagent endpoint
		getAgentURL := strings.Replace(b.agentURL, "/agent", "/getagent", 1)

		// Make HTTP GET request
		req, err := http.NewRequest("GET", getAgentURL, nil)
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

		// Parse the response
		var agents []AgentInfo
		if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
			return AgentListErrorMsg{Err: fmt.Errorf("failed to connect to server")}
		}

		return AgentListMsg{Agents: agents}
	}
}
