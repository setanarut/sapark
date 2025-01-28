package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
	"golang.org/x/image/colornames"
)

const (
	Objects           = 1000
	MaxPossibleObject = 10000
	intervalsCap      = MaxPossibleObject * 2
	IncrementObject   = 100 // Press KeyA
	ScreenWidth       = 900
	ScreenHeight      = 800
)

type Rect struct {
	X, Y, W, H float64
}

type Velocity struct {
	X, Y float64
}

type Collision struct {
	IsColliding bool
}

type Game struct {
	world        ecs.World
	filter       generic.Filter3[Rect, Velocity, Collision]
	mapObject    generic.Map3[Rect, Velocity, Collision]
	mapCollision generic.Map1[Collision]
	w, h         float64
	intervals    []Interval // interval pool for SAP
	pool         sync.Pool
	activeList   []ecs.Entity // Pre-allocated active list
	activeLen    int          // Current length of active list
}

// Interval structure represents intervals used for the SAP (Sweep and Prune) algorithm
type Interval struct {
	// Xaxis represents the position of the rectangle's edge on the x-axis
	Xaxis float64
	// IsLeftEdge indicates if this is the left edge (true) or right edge (false) of the rectangle
	IsLeftEdge bool
	// Entity holds the reference to the entity this interval belongs to
	Entity ecs.Entity
}

func main() {
	g := NewGame()

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("SAP Collision Demo")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func NewGame() *Game {
	g := &Game{
		w:          float64(ScreenWidth),
		h:          float64(ScreenHeight),
		world:      ecs.NewWorld(),
		filter:     *generic.NewFilter3[Rect, Velocity, Collision](),
		activeList: make([]ecs.Entity, MaxPossibleObject), // Maximum possible active entities
		intervals:  make([]Interval, 0, intervalsCap),     // Initial capacity for all intervals
	}
	g.mapObject = generic.NewMap3[Rect, Velocity, Collision](&g.world)
	g.mapCollision = generic.NewMap1[Collision](&g.world)
	g.pool = sync.Pool{
		New: func() interface{} {
			return &Interval{}
		},
	}

	// Create entities
	q := g.mapObject.NewBatchQ(Objects)

	for q.Next() {
		r, v, c := q.Get()
		r.W = 2 + rand.Float64()*18
		r.H = 2 + rand.Float64()*18
		r.X = rand.Float64() * (g.w - r.W)
		r.Y = rand.Float64() * (g.h - r.H)
		v.X = -1 + rand.Float64()*2
		v.Y = -1 + rand.Float64()*2
		c.IsColliding = false
	}

	return g
}

func (g *Game) Update() error {
	// Add new entities when 'A' key is pressed
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		// Create a batch of  new entities
		q := g.mapObject.NewBatchQ(IncrementObject)
		for q.Next() {
			r, v, c := q.Get()
			// Set entity dimensions
			r.W = 2 + rand.Float64()*18
			r.H = 2 + rand.Float64()*18
			// Random position within screen bounds
			r.X = rand.Float64() * (g.w - r.W)
			r.Y = rand.Float64() * (g.h - r.H)
			// Random velocity between -1 and 1
			v.X = -1 + rand.Float64()*2
			v.Y = -1 + rand.Float64()*2
			// Initialize collision state
			c.IsColliding = false
		}
	}

	// Reset SAP (Sweep and Prune) data structures
	g.intervals = g.intervals[:0]
	g.activeLen = 0

	// Update positions and build SAP intervals
	q := g.filter.Query(&g.world)
	for q.Next() {
		rect, vel, coll := q.Get()

		// Apply velocity to position
		rect.X += vel.X
		rect.Y += vel.Y

		// Handle screen boundary collisions
		handleScreenBoundaryCollision(rect, vel, g.w, g.h)

		// Reset collision state for new frame
		coll.IsColliding = false
		e := q.Entity()

		// Add entity bounds to SAP intervals
		g.intervals = append(g.intervals,
			Interval{rect.X, true, e},
			Interval{rect.X + rect.W, false, e})
	}

	// Sort intervals once
	sort.Slice(g.intervals, func(i, j int) bool {
		return g.intervals[i].Xaxis < g.intervals[j].Xaxis
	})

	// Sweep phase with pre-allocated active list
	for _, interval := range g.intervals {
		if interval.IsLeftEdge {
			// Check collisions with current active entities
			r1, v1, c1 := g.mapObject.Get(interval.Entity)

			for i := 0; i < g.activeLen; i++ {
				e2 := g.activeList[i]
				r2, v2, c2 := g.mapObject.Get(e2)

				// AABB overlap test
				if r1.Y < r2.Y+r2.H && r1.Y+r1.H > r2.Y {
					c1.IsColliding = true
					c2.IsColliding = true

					// Separate objects
					overlapX := math.Min(r1.X+r1.W, r2.X+r2.W) - math.Max(r1.X, r2.X)
					overlapY := math.Min(r1.Y+r1.H, r2.Y+r2.H) - math.Max(r1.Y, r2.Y)

					// Determine separation direction
					if overlapX < overlapY {
						// Separate on X axis
						if r1.X < r2.X {
							r1.X -= overlapX / 2
							r2.X += overlapX / 2
						} else {
							r1.X += overlapX / 2
							r2.X -= overlapX / 2
						}
						// Simple velocity exchange
						v1.X, v2.X = v2.X, v1.X
					} else {
						// Separate on Y axis
						if r1.Y < r2.Y {
							r1.Y -= overlapY / 2
							r2.Y += overlapY / 2
						} else {
							r1.Y += overlapY / 2
							r2.Y -= overlapY / 2
						}
						// Simple velocity exchange
						v1.Y, v2.Y = v2.Y, v1.Y
					}
				}
			}

			// Add to active list
			if g.activeLen < len(g.activeList) {
				g.activeList[g.activeLen] = interval.Entity
				g.activeLen++
			}
		} else {
			// Remove from active list
			for i := 0; i < g.activeLen; i++ {
				if g.activeList[i] == interval.Entity {
					g.activeList[i] = g.activeList[g.activeLen-1]
					g.activeLen--
					break
				}
			}
		}
	}

	return nil
}

// handleScreenBoundaryCollision handles collisions with the screen boundaries
func handleScreenBoundaryCollision(rect *Rect, vel *Velocity, screenWidth, screenHeight float64) {
	// Handle horizontal bounds
	if rect.X <= 0 {
		rect.X = 0
		vel.X = math.Abs(vel.X)
	} else if rect.X+rect.W >= screenWidth {
		rect.X = screenWidth - rect.W
		vel.X = -math.Abs(vel.X)
	}

	// Handle vertical bounds
	if rect.Y <= 0 {
		rect.Y = 0
		vel.Y = math.Abs(vel.Y)
	} else if rect.Y+rect.H >= screenHeight {
		rect.Y = screenHeight - rect.H
		vel.Y = -math.Abs(vel.Y)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Gray{30})
	q := g.filter.Query(&g.world)
	for q.Next() {
		r, _, c := q.Get()
		clr := colornames.Green
		if c.IsColliding {
			clr = colornames.Red
		}
		vector.DrawFilledRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), clr, false)
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %0.2f\nTPS: %0.2f\nEntities: %v",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		g.world.Stats().Entities),
		10, 10)
}

func (g *Game) Layout(w, h int) (int, int) {
	return int(g.w), int(g.h)
}
