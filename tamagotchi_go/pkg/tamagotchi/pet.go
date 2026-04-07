package tamagotchi

import (
	"fmt"
	"math/rand"
	"time"
)

type Mood string

const (
	MoodHappy   Mood = "happy"
	MoodNeutral Mood = "neutral"
	MoodSad     Mood = "sad"
	MoodSick    Mood = "sick"
	MoodDead    Mood = "dead"
)

type Pet struct {
	Name       string
	Age        int
	Hunger     int // 0-100, higher = more hungry
	Happiness  int // 0-100, higher = happier
	Health     int // 0-100, higher = healthier
	LastUpdate time.Time
	random     *rand.Rand
}

func NewPet(name string) *Pet {
	return &Pet{
		Name:       name,
		Age:        0,
		Hunger:     30,
		Happiness:  80,
		Health:     100,
		LastUpdate: time.Now(),
		random:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Update handles time-based degradation
func (p *Pet) Update() {
	now := time.Now()
	elapsed := now.Sub(p.LastUpdate)

	// Every 10 seconds, stats degrade
	intervals := int(elapsed.Seconds() / 10)
	if intervals > 0 {
		p.Hunger = min(100, p.Hunger+intervals*5)
		p.Happiness = max(0, p.Happiness-intervals*3)

		// Health decreases if very hungry or very unhappy
		if p.Hunger > 80 || p.Happiness < 20 {
			p.Health = max(0, p.Health-intervals*2)
		}

		p.LastUpdate = now
	}
}

// Feed the pet
func (p *Pet) Feed() string {
	p.Update()

	if p.Hunger < 20 {
		return fmt.Sprintf("%s is not hungry right now!", p.Name)
	}

	p.Hunger = max(0, p.Hunger-30)
	p.Happiness = min(100, p.Happiness+10)

	responses := []string{
		"*nom nom nom* 🍖",
		"Yummy! Thank you! 😋",
		"*happily munching* ✨",
		"That was delicious! 🌟",
	}

	return responses[p.random.Intn(len(responses))]
}

// Play with the pet
func (p *Pet) Play() string {
	p.Update()

	if p.Health < 30 {
		return fmt.Sprintf("%s is too sick to play...", p.Name)
	}

	if p.Hunger > 70 {
		return fmt.Sprintf("%s is too hungry to play!", p.Name)
	}

	p.Happiness = min(100, p.Happiness+20)
	p.Hunger = min(100, p.Hunger+10)

	responses := []string{
		"*bounces around excitedly* 🎾",
		"Wheee! This is fun! 🎉",
		"*plays happily* ⭐",
		"Yay! Let's play more! 🎈",
	}

	return responses[p.random.Intn(len(responses))]
}

// Heal the pet
func (p *Pet) Heal() string {
	p.Update()

	if p.Health > 80 {
		return fmt.Sprintf("%s is already healthy!", p.Name)
	}

	p.Health = min(100, p.Health+30)

	responses := []string{
		"*feels much better* 💊",
		"Thank you! I'm feeling better now! 💚",
		"*health restored* ✨",
		"All better! 🌈",
	}

	return responses[p.random.Intn(len(responses))]
}

// GetMood returns current mood based on stats
func (p *Pet) GetMood() Mood {
	p.Update()

	if p.Health == 0 {
		return MoodDead
	}

	if p.Health < 30 || p.Hunger > 80 {
		return MoodSick
	}

	if p.Happiness > 60 && p.Hunger < 50 {
		return MoodHappy
	}

	if p.Happiness < 30 || p.Hunger > 60 {
		return MoodSad
	}

	return MoodNeutral
}

// GetASCII returns ASCII art based on mood
func (p *Pet) GetASCII() string {
	mood := p.GetMood()

	switch mood {
	case MoodHappy:
		return `
   ／l、
  （ﾟ､ ｡７
   l  ~ヽ
   じしf_,)ノ
  `
	case MoodSad:
		return `
   ／l、
  （ ˘︹˘ ）
   l  ~ヽ
   じしf_,)ノ
  `
	case MoodSick:
		return `
   ／l、
  （ ×_× ）
   l  ~ヽ
   じしf_,)ノ
  `
	case MoodDead:
		return `
   ／l、
  （ X_X ）
   l  ~ヽ
   じしf_,)ノ
    R.I.P
  `
	default: // Neutral
		return `
   ／l、
  （ ･ω･）
   l  ~ヽ
   じしf_,)ノ
  `
	}
}

// GetStatus returns formatted status string
func (p *Pet) GetStatus() string {
	p.Update()

	return fmt.Sprintf(
		"📊 %s's Status | Health: %d%% | Hunger: %d%% | Happiness: %d%% | Age: %d | Mood: %s",
		p.Name,
		p.Health,
		p.Hunger,
		p.Happiness,
		p.Age,
		p.GetMood(),
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
