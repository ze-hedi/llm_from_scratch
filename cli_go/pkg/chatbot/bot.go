package chatbot

import (
	"fmt"
	"math/rand"
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

type Bot struct {
	name         string
	random       *rand.Rand
	SystemPrompt string
}

func NewBot() *Bot {
	return &Bot{
		name:         "ChatBot",
		random:       rand.New(rand.NewSource(time.Now().UnixNano())),
		SystemPrompt: "You are a helpful AI assistant. Be friendly and conversational.",
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

// GetResponseStream returns a tea.Cmd that uses pattern-matching to generate a response.
func (b *Bot) GetResponseStream(input string) tea.Cmd {
	return func() tea.Msg {
		return StreamChunkMsg{Chunk: b.GetResponse(input)}
	}
}
