package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
)

type Comparer interface {
	// return 1 if in order, 0 if equal, and -1 if not in order
	Compare(c Comparer) int
}

type Integer int

type List []Comparer

func (i Integer) Compare(c Comparer) int {
	switch c := c.(type) {
	case Integer:
		if i < c {
			return 1
		} else if i == c {
			return 0
		} else {
			return -1
		}
	case List:
		return List([]Comparer{i}).Compare(c)
	default:
		panic("tried to compare integer to comparer of unknown type")
	}
}

func (l List) Compare(c Comparer) int {
	switch c := c.(type) {
	case Integer:
		return l.Compare(List([]Comparer{c}))
	case List:
		for i, v := range l {
			// left hand is longer than right hand -> not in order
			if i >= len(c) {
				return -1
			}

			comparison := v.Compare(c[i])
			if comparison != 0 {
				return comparison
			}
		}
		// left shorter than right, so right order
		if len(l) < len(c) {
			return 1
		}
		// convince yourself they're the same length and have the same values
		return 0
	default:
		panic("tried to compare list to comparer of unknown type")
	}
}

// I know that I could just parse these as json, but it's been years since I last implemented a recursive parser
// just cos I wanted

func parseList(s string, startIndex int) (Comparer, int, error) {
	if s[startIndex] != '[' {
		return nil, 0, fmt.Errorf("attempted to parse list starting with non-[ character")
	}

	i := startIndex + 1
	list := []Comparer{}
	for i < len(s) {
		c := s[i]

		// end of list -> return the value
		if c == ']' {
			return List(list), i + 1, nil
		}

		// break between values -> just skip the comma
		if c == ',' {
			i++
			continue
		}

		comparer, nextIndex, err := parseComparer(s, i)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse list: %v", err)
		}

		list = append(list, comparer)
		i = nextIndex
	}

	return nil, 0, fmt.Errorf("unclosed list")
}

func parseInteger(s string, startIndex int) (Comparer, int, error) {
	acc := 0

	i := startIndex
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || '9' < c {
			return Integer(acc), i, nil
		}
		acc *= 10
		acc += int(c - '0')
	}
	return Integer(acc), i, nil
}

func parseComparer(s string, startIndex int) (Comparer, int, error) {
	c := s[startIndex]
	// parse a list
	if c == '[' {
		list, nextIndex, err := parseList(s, startIndex)
		if err != nil {
			return nil, 0, fmt.Errorf("failed parsing comparer: %v", err)
		}
		return list, nextIndex, nil
	}

	if '0' <= c && c <= '9' {
		integer, nextIndex, err := parseInteger(s, startIndex)
		if err != nil {
			return nil, 0, fmt.Errorf("failed parsing comparer: %v", err)
		}
		return integer, nextIndex, nil
	}

	return nil, 0, fmt.Errorf("failed to parse comparer: unknown character %c", c)
}

func parseInput(r io.Reader) ([][]Comparer, error) {
	scanner := bufio.NewScanner(r)
	pairs := [][]Comparer{}
	for scanner.Scan() {
		line1 := scanner.Text()
		//ignore blank lines
		if line1 == "" {
			continue
		}

		c1, nextIndex, err := parseComparer(line1, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to parse comparer on first line of pair: %v", err)
		}

		if nextIndex != len(line1) {
			return nil, fmt.Errorf("first line of pair not consumed: expected %d characters, got %d", len(line1), nextIndex)
		}

		if !scanner.Scan() {
			return nil, fmt.Errorf("attempted to parse pair with no second part")
		}

		line2 := scanner.Text()

		c2, nextIndex, err := parseComparer(line2, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to parse comparer on second line of pair: %v", err)
		}

		if nextIndex != len(line2) {
			return nil, fmt.Errorf("first line of pair not consumed: expected %d characters, got %d", len(line2), nextIndex)
		}

		pairs = append(pairs, []Comparer{c1, c2})
	}
	return pairs, nil
}

func part1(pairs [][]Comparer) int {
	sum := 0
	for i, pair := range pairs {
		val := pair[0].Compare(pair[1])
		if val == 1 {
			sum += i + 1
		}
	}
	return sum
}

func part2(pairs [][]Comparer) int {
	flat := make([]Comparer, 0, len(pairs)*2+2)
	for _, v := range pairs {
		flat = append(flat, v...)
	}

	pac2, pac2end, err := parseComparer("[[2]]", 0)
	if err != nil || pac2end != 5 {
		panic("failed to parse [[2]]")
	}

	pac6, pac6end, err := parseComparer("[[6]]", 0)
	if err != nil || pac6end != 5 {
		panic("failed to parse [[6]]")
	}

	flat = append(flat, pac2, pac6)
	sort.Slice(flat, func(p, q int) bool {
		return flat[p].Compare(flat[q]) == 1
	})

	pos2, pos6 := 0, 0
	for i, v := range flat {
		if v.Compare(pac2) == 0 {
			pos2 = i + 1
		} else if v.Compare(pac6) == 0 {
			pos6 = i + 1
		}

		if pos2 != 0 && pos6 != 0 {
			break
		}
	}

	return pos2 * pos6
}

func run() error {
	input, err := os.Open("input.txt")
	if err != nil {
		return fmt.Errorf("failed to open input: %v", err)
	}
	defer input.Close()

	pairs, err := parseInput(input)
	if err != nil {
		return fmt.Errorf("failed to parse input: %v", err)
	}

	fmt.Println("Part 1:", part1(pairs))
	fmt.Println("Part 2:", part2(pairs))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve day 13:", err)
		os.Exit(1)
	}
}
