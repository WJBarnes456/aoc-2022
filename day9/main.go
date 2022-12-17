package main

import (
	"bufio"
	"fmt"
	"os"
)

type Location struct {
	X int
	Y int
}

func Abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func Sign(i int) int {
	switch {
	case i < 0:
		return -1
	case i == 0:
		return 0
	case i > 0:
		return 1
	default:
		panic("tried to get sign of an integer that does not compare to 0")
	}
}

// Gets the distance in number of moves between two locations
func (from *Location) Distance(to *Location) int {
	// Because diagonal moves are allowed, the distance between two points is just their larger dimension
	Xdiff := Abs(from.X - to.X)
	Ydiff := Abs(from.Y - to.Y)

	if Ydiff > Xdiff {
		return Ydiff
	}
	return Xdiff

}

// Given the location of the head, update the location of the tail
// NB: this mutates the tail, copy it if you need the old location!
func (tail *Location) StepTail(head *Location) error {
	distance := tail.Distance(head)

	if distance > 2 {
		return fmt.Errorf("no way to move %v to %v in 2 moves", head, tail)
	}

	// <2 means they're already touching or sharing a square
	if distance < 2 {
		return nil
	}

	// assert distance is exactly 2 now

	// step towards the head
	Xdiff := head.X - tail.X
	Ydiff := head.Y - tail.Y

	tail.X += Sign(Xdiff)
	tail.Y += Sign(Ydiff)

	return nil
}

type Instruction struct {
	deltaX     int
	deltaY     int
	iterations int
}

// Parses a line of the input as an instruction
func getMove(line string) (*Instruction, error) {
	var direction string
	var iterations int
	fmt.Sscanf(line, "%s %d", &direction, &iterations)

	deltaX, deltaY := 0, 0
	switch direction {
	case "R":
		deltaX = 1
	case "U":
		deltaY = 1
	case "D":
		deltaY = -1
	case "L":
		deltaX = -1
	default:
		return nil, fmt.Errorf("unknown direction %s in input", direction)
	}

	return &Instruction{deltaX, deltaY, iterations}, nil

}

func simulate(moves []*Instruction, ropeLength int) (int, error) {
	rope := make([]*Location, ropeLength)
	for i := range rope {
		rope[i] = &Location{0, 0}
	}

	tail := rope[ropeLength-1]

	visitedLocations := make(map[Location]struct{})
	visitedLocations[*tail] = struct{}{}

	for _, m := range moves {
		for i := 0; i < m.iterations; i++ {
			rope[0].X += m.deltaX
			rope[0].Y += m.deltaY

			for j := 1; j < ropeLength; j++ {
				if err := rope[j].StepTail(rope[j-1]); err != nil {
					return 0, fmt.Errorf("failed to step tail: %v", err)
				}
			}

			visitedLocations[*tail] = struct{}{}
		}
	}

	return len(visitedLocations), nil
}

func run() error {
	scanner := bufio.NewScanner(os.Stdin)

	moves := []*Instruction{}
	for scanner.Scan() {
		move, err := getMove(scanner.Text())
		if err != nil {
			return fmt.Errorf("failed to parse move: %v", err)
		}
		moves = append(moves, move)
	}

	part1, err := simulate(moves, 2)

	if err != nil {
		return fmt.Errorf("failed to solve part1: %v", err)
	}
	fmt.Println("Part 1:", part1)

	part2, err := simulate(moves, 10)
	if err != nil {
		return fmt.Errorf("failed to solve part2: %v", err)
	}
	fmt.Println("Part 2:", part2)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error in day 9: %v", err)
		os.Exit(1)
	}
}
