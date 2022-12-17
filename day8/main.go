package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Direction int

const (
	North = iota
	East
	South
	West
)

type Tree struct {
	height      int
	visibleFrom map[Direction]struct{}
}

func parseInput(r io.Reader) ([][]int, error) {
	heights := [][]int{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		row := []int{}
		for _, c := range line {
			if c < '0' || c > '9' {
				return nil, fmt.Errorf("invalid height %c in grid", c)
			}
			row = append(row, int(c-'0'))
		}
		heights = append(heights, row)
	}

	return heights, nil
}

func setVisibility(trees [][]Tree, tallestTrees []*Tree, i int, j int, d Direction) error {
	var staticVariable int

	switch d {
	case North:
		staticVariable = j
	case West:
		staticVariable = i
	case South:
		staticVariable = j
	case East:
		staticVariable = i
	default:
		return fmt.Errorf("setting visibility from unknown direction %d", d)
	}

	tree := &trees[i][j]

	tallestTree := tallestTrees[staticVariable]
	//fmt.Println("tree", tree, "tallest tree", tallestTree)
	if tallestTree == nil || tree.height > tallestTree.height {
		tree.visibleFrom[d] = struct{}{}
		tallestTrees[staticVariable] = tree
	}
	//fmt.Println(tallestTrees)
	return nil
}

func getVisibility(heights [][]int) ([][]Tree, error) {
	trees := make([][]Tree, 0, len(heights))

	// build some rows of trees instead

	gridHeight := len(heights)
	gridWidth := len(heights[0])
	for _, heightRow := range heights {
		if len(heightRow) != gridWidth {
			return nil, fmt.Errorf("ragged row in grid: expected length %d, but got %d", gridWidth, len(heightRow))
		}

		treeRow := make([]Tree, 0, len(heightRow))
		for _, h := range heightRow {
			treeRow = append(treeRow, Tree{h, map[Direction]struct{}{}})
		}
		trees = append(trees, treeRow)
	}

	// do 4 passes over the whole grid, considering just north, east, south and west in each pass
	// I'm sure you can deduplicate this a bit more, but seems fine by me for now

	// north pass
	northTallestTrees := make([]*Tree, gridWidth)
	for i := 0; i < gridHeight; i++ {
		for j := 0; j < gridWidth; j++ {
			setVisibility(trees, northTallestTrees, i, j, North)
		}
	}

	// east pass
	eastTallestTrees := make([]*Tree, gridHeight)
	for j := gridWidth - 1; j >= 0; j-- {
		for i := 0; i < gridHeight; i++ {
			setVisibility(trees, eastTallestTrees, i, j, East)
		}
	}

	// south pass
	southTallestTrees := make([]*Tree, gridHeight)
	for i := gridHeight - 1; i >= 0; i-- {
		for j := 0; j < gridWidth; j++ {
			setVisibility(trees, southTallestTrees, i, j, South)
		}
	}

	// west pass
	westTallestTrees := make([]*Tree, gridHeight)
	for j := 0; j < gridWidth; j++ {
		for i := 0; i < gridHeight; i++ {
			setVisibility(trees, westTallestTrees, i, j, West)
		}
	}

	return trees, nil
}

func part1(trees [][]Tree) int {
	total := 0
	for _, row := range trees {
		for _, tree := range row {
			if len(tree.visibleFrom) > 0 {
				total += 1
			}
		}
	}
	return total
}

func part2(trees [][]Tree) int {
	bestScore := 0
	gridHeight, gridWidth := len(trees), len(trees[0])

	for i, row := range trees {
		for j, tree := range row {
			// ignore edges
			if i == 0 || j == 0 || i == gridHeight-1 || j == gridWidth-1 {
				continue
			}

			// do a pass in each direction
			var north, east, south, west int
			for north = 1; north < i; north++ {
				northTree := trees[i-north][j]
				if northTree.height >= tree.height {
					break
				}
			}

			for east = 1; east < gridWidth-j-1; east++ {
				eastTree := trees[i][j+east]
				if eastTree.height >= tree.height {
					break
				}
			}

			for south = 1; south < gridHeight-i-1; south++ {
				southTree := trees[i+south][j]
				if southTree.height >= tree.height {
					break
				}
			}

			for west = 1; west < j; west++ {
				westTree := trees[i][j-west]
				if westTree.height >= tree.height {
					break
				}
			}

			score := north * east * south * west
			fmt.Println(tree, north, east, south, west)
			if score > bestScore {
				bestScore = score
			}
		}
	}
	return bestScore
}

func run() error {
	grid, err := parseInput(os.Stdin)

	if err != nil {
		return fmt.Errorf("failed to parse input: %v", err)
	}

	fmt.Println(grid)

	trees, err := getVisibility(grid)

	if err != nil {
		return fmt.Errorf("failed to get visibility: %v", err)
	}

	for _, row := range trees {
		fmt.Println(row)
	}

	fmt.Println("Part 1:", part1(trees))

	fmt.Println("Part 2:", part2(trees))

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve day 8:", err)
		os.Exit(1)
	}
}
