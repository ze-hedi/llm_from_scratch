package game

import (
	"math/rand"
	"time"
)

const (
	GroundLevel   = 15 // Y position of the ground
	DinoX         = 5  // Fixed X position of the dino
	Gravity       = 1.2
	JumpVelocity  = -8.0
	InitialSpeed  = 3.0
	MaxSpeed      = 12.0
	SpeedIncrease = 0.1
)

type ObstacleType int

const (
	Cactus ObstacleType = iota
	Bird
)

type Obstacle struct {
	X      float64
	Y      int
	Type   ObstacleType
	Width  int
	Height int
}

type Dino struct {
	Y         float64
	Velocity  float64
	IsDucking bool
	IsJumping bool
}

type Game struct {
	Dino       *Dino
	Obstacles  []*Obstacle
	Score      int
	HighScore  int
	Speed      float64
	GameOver   bool
	FrameCount int
	rand       *rand.Rand
}

func NewGame(highScore int) *Game {
	return &Game{
		Dino: &Dino{
			Y:        float64(GroundLevel),
			Velocity: 0,
		},
		Obstacles:  []*Obstacle{},
		Score:      0,
		HighScore:  highScore,
		Speed:      InitialSpeed,
		GameOver:   false,
		FrameCount: 0,
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Game) Jump() {
	if !g.Dino.IsJumping && !g.Dino.IsDucking && g.Dino.Y >= float64(GroundLevel) {
		g.Dino.Velocity = JumpVelocity
		g.Dino.IsJumping = true
	}
}

func (g *Game) Duck(ducking bool) {
	if !g.Dino.IsJumping {
		g.Dino.IsDucking = ducking
	}
}

func (g *Game) Update() {
	if g.GameOver {
		return
	}

	g.FrameCount++
	g.Score = g.FrameCount / 10

	// Update high score
	if g.Score > g.HighScore {
		g.HighScore = g.Score
	}

	// Increase speed over time
	if g.Speed < MaxSpeed {
		g.Speed += SpeedIncrease * 0.01
	}

	// Update dino physics
	g.Dino.Velocity += Gravity
	g.Dino.Y += g.Dino.Velocity

	// Keep dino on the ground
	if g.Dino.Y >= float64(GroundLevel) {
		g.Dino.Y = float64(GroundLevel)
		g.Dino.Velocity = 0
		g.Dino.IsJumping = false
	}

	// Update obstacles
	for i := len(g.Obstacles) - 1; i >= 0; i-- {
		g.Obstacles[i].X -= g.Speed

		// Remove obstacles that are off screen
		if g.Obstacles[i].X < -10 {
			g.Obstacles = append(g.Obstacles[:i], g.Obstacles[i+1:]...)
		}
	}

	// Spawn new obstacles
	if len(g.Obstacles) == 0 || g.Obstacles[len(g.Obstacles)-1].X < 60 {
		if g.rand.Float64() < 0.02 { // 2% chance per frame
			g.spawnObstacle()
		}
	}

	// Check collisions
	g.checkCollisions()
}

func (g *Game) spawnObstacle() {
	obstacleType := Cactus
	if g.rand.Float64() < 0.3 && g.Score > 100 { // 30% chance for birds after score 100
		obstacleType = Bird
	}

	var obstacle *Obstacle
	switch obstacleType {
	case Cactus:
		obstacle = &Obstacle{
			X:      80,
			Y:      GroundLevel,
			Type:   Cactus,
			Width:  2,
			Height: 3,
		}
	case Bird:
		// Birds can fly at different heights
		heights := []int{GroundLevel - 2, GroundLevel - 4, GroundLevel - 6}
		obstacle = &Obstacle{
			X:      80,
			Y:      heights[g.rand.Intn(len(heights))],
			Type:   Bird,
			Width:  3,
			Height: 2,
		}
	}

	g.Obstacles = append(g.Obstacles, obstacle)
}

func (g *Game) checkCollisions() {
	dinoHeight := 4
	if g.Dino.IsDucking {
		dinoHeight = 2
	}
	dinoWidth := 3

	dinoTop := int(g.Dino.Y) - dinoHeight
	dinoBottom := int(g.Dino.Y)
	dinoLeft := DinoX
	dinoRight := DinoX + dinoWidth

	for _, obs := range g.Obstacles {
		obsLeft := int(obs.X)
		obsRight := int(obs.X) + obs.Width
		obsTop := obs.Y - obs.Height
		obsBottom := obs.Y

		// Check if bounding boxes overlap
		if dinoRight > obsLeft && dinoLeft < obsRight &&
			dinoBottom > obsTop && dinoTop < obsBottom {
			g.GameOver = true
			return
		}
	}
}

func (g *Game) Reset() {
	g.Dino = &Dino{
		Y:        float64(GroundLevel),
		Velocity: 0,
	}
	g.Obstacles = []*Obstacle{}
	g.Score = 0
	g.Speed = InitialSpeed
	g.GameOver = false
	g.FrameCount = 0
}

func (g *Game) GetDinoHeight() int {
	if g.Dino.IsDucking {
		return 2
	}
	return 4
}
