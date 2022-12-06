package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

type Stack[T any] []T

type Crate rune

func (s *Stack[T]) Size() int {
	return len(*s)
}

func (s *Stack[T]) Push(val T) {
	(*s) = append((*s), val)
}

// this will panic if the stack is empty - fine for this problem
// but in a general case it's hard to error if you don't have a value of type T to hand!
func (s *Stack[T]) Pop() T {
	final := len(*s) - 1
	val := (*s)[final]
	(*s) = (*s)[:final]

	return val
}

func (s *Stack[T]) Peek() T {
	final := len(*s) - 1
	return (*s)[final]
}

type Move struct {
	count       int
	source      int
	destination int
}

func readInput(r io.Reader) ([]Stack[Crate], []Move, error) {
	scanner := bufio.NewScanner(r)

	crateMatch, err := regexp.Compile(`^(?:(?:\[.\]|   ) ?)+$`)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile crate regex: %v", err)
	}

	state := make([]Stack[Crate], 0)
	// parse the stacks of crates
	for scanner.Scan() {
		// when we see a blank line, we should parse moves instead
		line := scanner.Text()
		if line == "" {
			break
		}

		lineRunes := []rune(line)

		// on the first move, we need to initialise state
		if len(state) == 0 {
			nCrates := (len(lineRunes) + 1) / 4
			for i := 0; i < nCrates; i++ {
				state = append(state, make(Stack[Crate], 0))
			}
		}

		// skip any lines with an invalid format
		// (in particular the numbering at the bottom)
		if !crateMatch.MatchString(line) {
			continue
		}

		for i := 0; i < len(state); i++ {
			crate := lineRunes[4*i+1]

			if crate != ' ' {
				fmt.Printf("adding %c to %v\n", crate, state[i])
				state[i].Push(Crate(crate))
			}
		}
	}

	// we built the state by pushing them on in the opposite order, we need to now reverse them
	// (better this way to avoid lots of memory allocations)
	for i, stack := range state {
		newStack := make(Stack[Crate], 0, stack.Size())

		for j := stack.Size() - 1; j >= 0; j-- {
			newStack.Push(stack[j])
		}

		state[i] = newStack
	}

	// parse the moves
	matchMoves, err := regexp.Compile(`move (\d+) from (\d+) to (\d+)`)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile regex: %v", err)
	}

	moves := make([]Move, 0)
	for scanner.Scan() {
		vals := matchMoves.FindStringSubmatch(scanner.Text())

		intVals := make([]int, 0, len(vals))

		for i, val := range vals {
			// first match is the full string
			if i == 0 {
				continue
			}
			intVal, err := strconv.ParseInt(val, 10, 32)

			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse val %d: %v", i, err)
			}

			intVals = append(intVals, int(intVal))
		}

		// the input values are 1-indexed, 0-index them for running
		move := Move{intVals[0], intVals[1] - 1, intVals[2] - 1}
		moves = append(moves, move)
	}

	return state, moves, nil
}

func cloneCrates(crates []Stack[Crate]) []Stack[Crate] {
	out := make([]Stack[Crate], 0, len(crates))

	for _, stack := range crates {
		newStack := make(Stack[Crate], stack.Size())
		copy(newStack, stack)
		out = append(out, newStack)
	}

	return out
}

func part1(crates []Stack[Crate], moves []Move) []Crate {
	for _, m := range moves {
		for i := 0; i < m.count; i++ {
			val := crates[m.source].Pop()
			crates[m.destination].Push(val)
		}
	}

	out := make([]Crate, 0, len(crates))
	for _, s := range crates {
		out = append(out, s.Peek())
	}

	return out
}

func part2(crates []Stack[Crate], moves []Move) []Crate {
	for _, m := range moves {
		sourceCrates := crates[m.source]
		cutPoint := len(sourceCrates) - m.count

		toMove := sourceCrates[cutPoint:]

		//fmt.Printf("moving %v from %v to %v\n", toMove, crates[m.source], crates[m.destination])

		crates[m.source] = sourceCrates[:cutPoint]

		crates[m.destination] = append(crates[m.destination], toMove...)

		//fmt.Printf("crates: %v\n", crates)
	}

	out := make([]Crate, 0, len(crates))
	for _, s := range crates {
		out = append(out, s.Peek())
	}

	return out
}

func run() error {
	crates, moves, err := readInput(os.Stdin)

	if err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	fmt.Println("Crates: ", crates)
	fmt.Println("Moves: ", moves)

	part1 := part1(cloneCrates(crates), moves)

	fmt.Println("Part 1:", string(part1))

	part2 := part2(cloneCrates(crates), moves)

	fmt.Println("Part 2:", string(part2))

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve:", err)
		os.Exit(1)
	}
}
