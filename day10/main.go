package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func run() error {
	xStates := []int{}

	x := 1
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		vals := strings.Split(line, " ")
		switch {
		case len(vals) == 1 && vals[0] == "noop":
			xStates = append(xStates, x)
		case len(vals) == 2 && vals[0] == "addx":
			var delta int
			fmt.Sscanf(vals[1], "%d", &delta)
			xStates = append(xStates, x, x)
			x += delta
		default:
			return fmt.Errorf("failed to parse instruction %s", line)
		}
	}

	part1 := 0
	for i := 20; i < 221; i += 40 {
		part1 += i * xStates[i-1]
	}

	fmt.Println("Part 1", part1)

	fmt.Println("Part 2:")

	// it's not efficient to draw character by character, but I want to feel like I'm using a real CRT
	for i, v := range xStates {
		pixelNo := i % 40
		diff := (pixelNo) - v
		if -2 < diff && diff < 2 {
			fmt.Print("#")
		} else {
			fmt.Print(".")
		}

		if pixelNo == 39 {
			fmt.Println()
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to solve: %v", err)
		os.Exit(1)
	}
}
