package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Move int

const (
	Rock Move = iota
	Paper
	Scissors
)

type Result int

const (
	Loss Result = iota * 3
	Draw
	Win
)

type Round struct {
	your  Move
	their Move
}

type Interpreter interface {
	Interpret(string, Move) (Move, error)
}

func (m Move) score() int {
	return int(m) + 1
}

func (r Result) score() int {
	return int(r)
}

func (your Move) plays(their Move) Result {
	// This works because of the order: you win to the one behind, and lose to the one in front
	switch (your - their + 3) % 3 {
	case 0:
		return Draw
	case 1:
		return Win
	case 2:
		return Loss
	default:
		panic("A difference of values mod 3 was not in the range 0-2 when scoring")
	}
}

func (r Round) score() int {
	return r.your.score() + r.your.plays(r.their).score()
}

func readGuide(r io.Reader) ([][]string, error) {
	game := make([][]string, 0)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		vals := strings.Split(line, " ")

		if len(vals) > 2 {
			return game, fmt.Errorf("line with more than 2 values")
		}

		game = append(game, vals)
	}

	return game, nil
}

func interpretGuide(guide [][]string, i Interpreter) ([]Round, error) {
	game := make([]Round, 0)

	for _, vals := range guide {
		theirStr := vals[0]
		yourStr := vals[1]

		var their Move

		switch theirStr {
		case "A":
			their = Rock
		case "B":
			their = Paper
		case "C":
			their = Scissors
		default:
			return game, fmt.Errorf("their invalid move %s", theirStr)
		}

		your, err := i.Interpret(yourStr, their)

		if err != nil {
			return game, fmt.Errorf("your invalid move %s: %v", yourStr, err)
		}

		game = append(game, Round{your, their})
	}

	return game, nil
}

type Part1T int

const Part1 Part1T = iota

func (Part1T) Interpret(your string, _ Move) (Move, error) {
	switch your {
	case "X":
		return Rock, nil
	case "Y":
		return Paper, nil
	case "Z":
		return Scissors, nil
	default:
		return Rock, fmt.Errorf("%s not in X,Y,Z", your)
	}
}

type Part2T int

const Part2 Part2T = iota

func (Part2T) Interpret(outcome string, their Move) (Move, error) {
	// as with plays, this works because you win to one behind, and lose to one in front
	switch outcome {
	case "X":
		return Move((their + 2) % 3), nil
	case "Y":
		return their, nil
	case "Z":
		return Move((their + 1) % 3), nil
	default:
		return Rock, fmt.Errorf("%v not in X,Y,Z", outcome)
	}
}

func scoreGuide(guide []Round) int {
	total := 0
	for _, round := range guide {
		total += round.score()
	}

	return total
}

func main() {
	input, err := readGuide(os.Stdin)

	if err != nil {
		fmt.Errorf("failed to read guide: %v", err)
		os.Exit(1)
	}

	part1Guide, err := interpretGuide(input, Part1)

	if err != nil {
		fmt.Errorf("failed to interpret guide for part 1: %v", err)
		os.Exit(1)
	}

	fmt.Println("Part 1:", scoreGuide(part1Guide))

	part2Guide, err := interpretGuide(input, Part2)

	fmt.Println("%v", part2Guide)

	if err != nil {
		fmt.Errorf("failed to interpret guide for part 2: %v", err)
		os.Exit(1)
	}

	fmt.Println("Part 2:", scoreGuide(part2Guide))
}
