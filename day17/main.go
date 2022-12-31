package main

import (
	"fmt"
	"os"
	"strings"
)

const CHAMBER_WIDTH = 7

type Move int

const (
	Left Move = iota
	Right
	Down
)

type ShapeClass int

const (
	HorizontalLine ShapeClass = iota
	Plus
	BackwardsL
	VerticalLine
	Square
)

type Coordinate struct {
	x int
	y int
}

// By convention, x and y are the bottom-left corner of the bounding box
// This makes placing the shape simpler
type Shape struct {
	class    ShapeClass
	position Coordinate
}

func (s *Shape) OccupiedPositions() []Coordinate {
	x := s.position.x
	y := s.position.y
	switch s.class {
	case HorizontalLine:
		return []Coordinate{{x, y}, {x + 1, y}, {x + 2, y}, {x + 3, y}}
	case Plus:
		return []Coordinate{{x + 1, y}, {x, y + 1}, {x + 1, y + 1}, {x + 2, y + 1}, {x + 1, y + 2}}
	case BackwardsL:
		return []Coordinate{{x, y}, {x + 1, y}, {x + 2, y}, {x + 2, y + 1}, {x + 2, y + 2}}
	case VerticalLine:
		return []Coordinate{{x, y}, {x, y + 1}, {x, y + 2}, {x, y + 3}}
	case Square:
		return []Coordinate{{x, y}, {x + 1, y}, {x, y + 1}, {x + 1, y + 1}}
	default:
		panic(fmt.Sprintf("invalid shape class %v", s.class))
	}
}

func (s *Shape) IsValid(c *Chamber) bool {
	for _, pos := range s.OccupiedPositions() {
		if c.IsOccupied(pos.x, pos.y) {
			return false
		}
	}
	return true
}

func (s Shape) CanMove(direction Move, c *Chamber) bool {
	switch direction {
	case Left:
		s.position.x -= 1
		return s.IsValid(c)
	case Right:
		s.position.x += 1
		return s.IsValid(c)
	case Down:
		s.position.y -= 1
		return s.IsValid(c)
	default:
		panic(fmt.Sprintf("Invalid direction %v", direction))
	}
}

func (s *Shape) ApplyMove(direction Move, c *Chamber) bool {
	if !s.CanMove(direction, c) {
		return false
	}

	switch direction {
	case Left:
		s.position.x -= 1
	case Right:
		s.position.x += 1
	case Down:
		s.position.y -= 1
	default:
		panic(fmt.Sprintf("Invalid direction %v", direction))
	}
	return true
}

type Chamber struct {
	occupancy  map[int][]bool
	jetPattern []Move
	jetIndex   int
}

type ChamberState struct {
	profile      string
	currentShape ShapeClass
	currentJet   int
}

type GameState struct {
	height int
	turn   int
}

func (c *Chamber) Profile() string {
	depths := make([]int, CHAMBER_WIDTH)
	foundDepths := make([]bool, CHAMBER_WIDTH)
	maxY := c.MaxHeight()
	for y := maxY; y >= 0; y-- {
		row := c.occupancy[y]
		for x, present := range row {
			if present && !foundDepths[x] {
				depths[x] = maxY - y
				foundDepths[x] = true
			}
		}
	}

	// if not found, it must extend to the floor
	for x, found := range foundDepths {
		if !found {
			depths[x] = maxY + 1
		}
	}

	out := make([]string, CHAMBER_WIDTH)
	for i, depth := range depths {
		out[i] = fmt.Sprint(depth)
	}
	return strings.Join(out, ",")
}

func (c *Chamber) MaxHeight() int {
	maxY := 0
	for y := range c.occupancy {
		if y > maxY {
			maxY = y
		}
	}
	return maxY
}

func (c *Chamber) IsOccupied(x int, y int) bool {
	// Walls of the chamber are occupied
	if x < 0 || x > CHAMBER_WIDTH-1 {
		return true
	}

	// Floor of the chamber is occupied
	if y < 0 {
		return true
	}

	// Row is empty
	if _, exists := c.occupancy[y]; !exists {
		return false
	}

	// Row exists, check it's non-empty
	return c.occupancy[y][x]
}

func (c *Chamber) placeShape(shape Shape) {
	for _, pos := range shape.OccupiedPositions() {
		if _, exists := c.occupancy[pos.y]; !exists {
			c.occupancy[pos.y] = make([]bool, CHAMBER_WIDTH)
		}
		c.occupancy[pos.y][pos.x] = true
	}
}

func (c *Chamber) AddRock(class ShapeClass) error {
	startY := len(c.occupancy) + 3
	shape := Shape{class, Coordinate{2, startY}}
	for i := 0; i <= startY; i++ {
		jet := c.jetPattern[c.jetIndex]
		c.jetIndex = (c.jetIndex + 1) % len(c.jetPattern)
		shape.ApplyMove(jet, c)
		movedDown := shape.ApplyMove(Down, c)
		if !movedDown {
			// could not move down, so place
			c.placeShape(shape)
			return nil
		}
	}

	return nil
}

func parseInput(input string) ([]Move, error) {
	out := make([]Move, len(input))
	for i, c := range []rune(input) {
		switch c {
		case '<':
			out[i] = Left
		case '>':
			out[i] = Right
		default:
			return nil, fmt.Errorf("invalid character %c when parsing input", c)
		}
	}
	return out, nil
}

func part1(jets []Move) int {
	chamber := Chamber{
		map[int][]bool{},
		jets,
		0,
	}

	for i := 0; i < 2022; i++ {
		shapeClass := ShapeClass(i % 5)
		chamber.AddRock(shapeClass)
	}

	return chamber.MaxHeight() + 1
}

func part2(jets []Move) int {
	chamber := Chamber{
		map[int][]bool{},
		jets,
		0,
	}

	// Notice that there's a symmetry: if we're ever in a position we've seen
	// before (i.e. the profile from the top of the board, the current position,
	// and position in the jet sequence is the same), we can iterate that number of moves from the current turn
	heightDiff := 1
	skipped := false
	memo := map[ChamberState]GameState{}

	max := 1000000000000
	for i := 0; i < max; i++ {
		shapeClass := ShapeClass(i % 5)

		chamberState := ChamberState{chamber.Profile(), shapeClass, chamber.jetIndex}

		if !skipped {
			if val, exists := memo[chamberState]; exists {
				fmt.Printf("Hit a conflict at turn %v, previously saw state at turn %v\n", i, val.turn)
				remainingTurns := max - i
				cycleLength := i - val.turn
				heightChange := chamber.MaxHeight() + heightDiff - val.height
				fmt.Printf("Cycle length %v turns, height change %v\n", cycleLength, heightChange)
				cycles := remainingTurns / cycleLength
				i += cycles * cycleLength
				heightDiff += cycles * heightChange
				skipped = true
			}

			memo[chamberState] = GameState{height: chamber.MaxHeight() + heightDiff, turn: i}
		}
		chamber.AddRock(shapeClass)
	}

	return chamber.MaxHeight() + heightDiff
}

func run() error {
	input, err := parseInput(">>><<><>><<<>><>>><<<>>><<<><<<>><>><<>>")
	if err != nil {
		return fmt.Errorf("failed to parse jet string: %v", err)
	}

	fmt.Println("Part 1:", part1(input))

	fmt.Println("Part 2:", part2(input))

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to solve day 17: %v", err)
		os.Exit(1)
	}
}
