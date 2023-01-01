package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
)

// Sparse array made implementing part 1 simpler
// I'm not convinced it's useful for part 2
type Grid struct {
	occupancy map[int]map[int]map[int]struct{}
	minX      int
	minY      int
	minZ      int
	maxX      int
	maxY      int
	maxZ      int
}

func (g *Grid) IsOccupied(x int, y int, z int) bool {
	xSlice, exists := g.occupancy[x]
	if !exists {
		return false
	}

	ySlice, exists := xSlice[y]
	if !exists {
		return false
	}

	_, zExists := ySlice[z]

	return zExists
}

func (g *Grid) ExposedSides(x int, y int, z int) int {
	occupancy := []bool{
		g.IsOccupied(x-1, y, z),
		g.IsOccupied(x+1, y, z),
		g.IsOccupied(x, y+1, z),
		g.IsOccupied(x, y-1, z),
		g.IsOccupied(x, y, z+1),
		g.IsOccupied(x, y, z-1),
	}

	exposed := 0
	for _, v := range occupancy {
		if !v {
			exposed++
		}
	}
	return exposed
}

func (g *Grid) Place(x int, y int, z int) {
	if _, exists := g.occupancy[x]; !exists {
		g.occupancy[x] = make(map[int]map[int]struct{})
	}

	if _, exists := g.occupancy[x][y]; !exists {
		g.occupancy[x][y] = make(map[int]struct{})
	}

	g.occupancy[x][y][z] = struct{}{}

	// maintain the mins/maxes

	if x > g.maxX {
		g.maxX = x
	}

	if x < g.minX {
		g.minX = x
	}

	if y > g.maxY {
		g.maxY = y
	}

	if y < g.minY {
		g.minY = y
	}

	if z > g.maxZ {
		g.maxZ = z
	}

	if z < g.minZ {
		g.minZ = z
	}

}

func parseInput(r io.Reader) (*Grid, error) {
	grid := Grid{occupancy: make(map[int]map[int]map[int]struct{}),
		maxX: math.MinInt, maxY: math.MinInt, maxZ: math.MinInt, minX: math.MaxInt, minY: math.MaxInt, minZ: math.MaxInt}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		var x, y, z int
		parsed, err := fmt.Sscanf(text, "%d,%d,%d", &x, &y, &z)
		if err != nil {
			return nil, fmt.Errorf("error parsing %v: %v", text, err)
		}

		if parsed != 3 {
			return nil, fmt.Errorf("%v doesn't match format", text)
		}

		grid.Place(x, y, z)
	}

	return &grid, nil
}

func part1(g *Grid) int {
	surfaceArea := 0
	for x, xSlice := range g.occupancy {
		for y, ySlice := range xSlice {
			for z := range ySlice {
				surfaceArea += g.ExposedSides(x, y, z)
			}
		}
	}
	return surfaceArea
}

func (steamGrid *Grid) part2_attemptPlace(g *Grid, x int, y int, z int) bool {
	// don't place outside the bounding box
	if x < steamGrid.minX || x > steamGrid.maxX {
		return false
	}

	if y < steamGrid.minY || y > steamGrid.maxY {
		return false
	}

	if z < steamGrid.minZ || z > steamGrid.maxZ {
		return false
	}

	// don't place if the square is occupied in the lava droplet
	if g.IsOccupied(x, y, z) {
		return false
	}

	// no change in placing if the square is occupied in the steam box
	if steamGrid.IsOccupied(x, y, z) {
		return false
	}

	// within the bounding box, and not occupied, so a valid free space
	steamGrid.Place(x, y, z)
	return true
}

func part2(g *Grid) int {
	// This isn't the most intelligent algorithm (the running time depends on
	// the size of the droplet!), but it's fine for these purposes
	changed := true
	steamGrid := Grid{occupancy: make(map[int]map[int]map[int]struct{}),
		minX: g.minX - 1, minY: g.minY - 1, minZ: g.minZ - 1, maxX: g.maxX + 1, maxY: g.maxY + 1, maxZ: g.maxZ + 1}

	// this square is outside the cuboid, and the bounding box is one larger
	// than each direction, so it'll propagate all the way around the outside
	steamGrid.Place(steamGrid.minX, steamGrid.minY, steamGrid.minZ)

	// until the enclosing cuboid stops changing, proliferate the steam to any
	// exposed air
	for changed {
		changed = false
		for x, xSlice := range steamGrid.occupancy {
			for y, ySlice := range xSlice {
				for z := range ySlice {
					placed := []bool{
						steamGrid.part2_attemptPlace(g, x-1, y, z),
						steamGrid.part2_attemptPlace(g, x+1, y, z),
						steamGrid.part2_attemptPlace(g, x, y+1, z),
						steamGrid.part2_attemptPlace(g, x, y-1, z),
						steamGrid.part2_attemptPlace(g, x, y, z+1),
						steamGrid.part2_attemptPlace(g, x, y, z-1),
					}

					for _, v := range placed {
						changed = changed || v
					}
				}
			}
		}
	}

	// now invert the steam grid (I'm pretty sure this isn't needed, but just
	// taking the surface area of the steam grid and subtracting the external
	// surface didn't seem to work)
	newGrid := Grid{occupancy: map[int]map[int]map[int]struct{}{}}
	for x := steamGrid.minX; x <= steamGrid.maxX; x++ {
		for y := steamGrid.minY; y <= steamGrid.maxY; y++ {
			for z := steamGrid.minZ; z <= steamGrid.maxZ; z++ {
				if !steamGrid.IsOccupied(x, y, z) {
					newGrid.Place(x, y, z)
				}
			}
		}
	}
	return part1(&newGrid)
}

func run() error {
	input, err := os.Open("input.txt")
	if err != nil {
		return fmt.Errorf("failed to open input file")
	}
	defer input.Close()

	grid, err := parseInput(input)
	if err != nil {
		return fmt.Errorf("failed to parse input: %v", err)
	}

	fmt.Println("Part 1:", part1(grid))
	fmt.Println("Part 2:", part2(grid))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to solve day 18: %v", err)
		os.Exit(1)
	}
}
