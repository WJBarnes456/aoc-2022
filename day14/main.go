package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// Sparse array to keep track of what space is filled
// and keep track of the lowest point of the world, as you can't hit anything below there

type Material int

const (
	Rock Material = iota
	Sand
)

type World struct {
	// Indexed by y, then x (for ease of determining the lowest point)
	filled     map[int]map[int]Material
	lowestRock int
}

// Gets the Y value of the lowest point (i.e. highest Y) in the world
func (w *World) lowestPoint() int {
	return w.lowestRock
}

func (w *World) fillLine(startX int, startY int, endX int, endY int) error {
	if startX == endX {
		// vertical line
		x := startX
		var lowY, highY int
		if startY < endY {
			lowY, highY = startY, endY
		} else {
			lowY, highY = endY, startY
		}
		if highY > w.lowestRock {
			w.lowestRock = highY
		}

		for y := lowY; y <= highY; y++ {
			if w.filled[y] == nil {
				w.filled[y] = map[int]Material{}
			}
			w.filled[y][x] = Rock
		}
	} else if startY == endY {
		// horizontal line
		y := startY
		if y > w.lowestRock {
			w.lowestRock = y
		}
		var lowX, highX int
		if startX < endX {
			lowX, highX = startX, endX
		} else {
			lowX, highX = endX, startX
		}
		if w.filled[y] == nil {
			w.filled[y] = map[int]Material{}
		}
		for x := lowX; x <= highX; x++ {
			w.filled[y][x] = Rock
		}
	} else {
		// neither horizontal nor vertical, so not valid
		return fmt.Errorf("tried to draw non-horizontal, non-vertical line from (%d,%d) to (%d,%d)", startX, startY, endX, endY)
	}
	return nil
}

// Adds sand to the world, returning whether sand was actually added
func (w *World) addSand(floor bool) bool {
	sandX := 500
	sandY := 0

	if w.filled[sandY] != nil {
		_, sourceBlocked := w.filled[sandY][sandX]
		if sourceBlocked {
			return false
		}
	}

	lowest := w.lowestPoint()
	for sandY < lowest+2 {
		// kind of nasty, but it should work - only check if you're not currently trying to place on the floor
		if !(floor && sandY == lowest+1) {
			nextY := sandY + 1

			// if next row doesn't exist, it's empty, so move down
			nextRow := w.filled[nextY]
			if nextRow == nil {
				sandY = nextY
				continue
			}

			_, belowFull := nextRow[sandX]
			if !belowFull {
				sandY = nextY
				continue
			}

			// below is not free, so try down and left
			_, leftFull := nextRow[sandX-1]
			if !leftFull {
				sandX, sandY = sandX-1, nextY
				continue
			}

			// below and below-left are not free, so try down and right
			_, rightFull := nextRow[sandX+1]
			if !rightFull {
				sandX, sandY = sandX+1, nextY
				continue
			}
		}

		// none are free, so place
		if w.filled[sandY] == nil {
			w.filled[sandY] = map[int]Material{}
		}

		_, alreadyFilled := w.filled[sandX][sandY]
		if alreadyFilled {
			panic("attempted to place sand on a full square")
		}

		w.filled[sandY][sandX] = Sand
		return true
	}
	return false
}

func (w *World) Clone() *World {
	newWorld := map[int]map[int]Material{}
	for y, row := range w.filled {
		newRow := make(map[int]Material, len(row))
		for x, val := range row {
			newRow[x] = val
		}
		newWorld[y] = newRow
	}
	return &World{newWorld, w.lowestRock}
}

func parseInput(r io.Reader) (*World, error) {
	scanner := bufio.NewScanner(r)
	world := World{map[int]map[int]Material{}, math.MinInt}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " -> ")
		var prevX, prevY *int
		for _, part := range parts {
			var x, y int
			_, err := fmt.Sscanf(part, "%d,%d", &x, &y)
			if err != nil {
				return nil, fmt.Errorf("failed to parse value %s: %v", part, err)
			}

			if prevX != nil && prevY != nil {
				world.fillLine(*prevX, *prevY, x, y)
			}
			prevX, prevY = &x, &y
		}
	}
	return &world, nil
}

func part1(w *World) int {
	count := 0
	for w.addSand(false) {
		count += 1
	}
	return count
}

func part2(w *World) int {
	count := 0
	for w.addSand(true) {
		count += 1
	}
	return count
}

func run() error {
	file, err := os.Open("input.txt")
	if err != nil {
		return fmt.Errorf("failed to open input: %v", err)
	}
	defer file.Close()

	world, err := parseInput(file)
	if err != nil {
		return fmt.Errorf("failed to parse world: %v", err)
	}

	fmt.Println(world.filled)

	fmt.Println("Part 1:", part1(world.Clone()))

	fmt.Println("Part 2:", part2(world.Clone()))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve day 14:", err)
		os.Exit(1)
	}
}
