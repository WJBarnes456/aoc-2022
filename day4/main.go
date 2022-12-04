package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

type Section struct {
	start int
	end   int
}

type Pair[T, U any] struct {
	First  T
	Second U
}

type Assignment Pair[Section, Section]

//Min gets the smaller of any two comparable types
//
// Taken from this SO answer https://stackoverflow.com/a/27516559 - I wasn't
// expecting this to be best practice.

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Returns the overlap between two sections
// If the pointer is nil, there is no overlap
func (my Section) overlap(your Section) *Section {
	var earlier, later Section

	// wlog, consider the section with the earlier start

	if my.start < your.start {
		earlier = my
		later = your
	} else {
		earlier = your
		later = my
	}

	// there is no overlap if the later starts before the earlier ends
	if later.start > earlier.end {
		return nil
	}

	overlap := Section{later.start, min(earlier.end, later.end)}

	return &overlap
}

func parseSection(s string) (Section, error) {
	parts := strings.Split(s, "-")

	if len(parts) != 2 {
		return Section{}, fmt.Errorf("section did not have two parts")
	}

	start, err := strconv.ParseInt(parts[0], 10, 32)

	if err != nil {
		return Section{}, fmt.Errorf("failed to parse start")
	}

	end, err := strconv.ParseInt(parts[1], 10, 32)

	if err != nil {
		return Section{}, fmt.Errorf("failed to parse end")
	}

	return Section{int(start), int(end)}, nil
}

func readAssignments(r io.Reader) ([]Assignment, error) {
	scanner := bufio.NewScanner(r)

	assignments := make([]Assignment, 0)
	for scanner.Scan() {
		line := scanner.Text()

		sectionsStr := strings.Split(line, ",")

		if len(sectionsStr) != 2 {
			return nil, fmt.Errorf("line did not have two comma-separated parts")
		}

		first, err := parseSection(sectionsStr[0])

		if err != nil {
			return nil, fmt.Errorf("failed to parse first section: %v", err)
		}

		second, err := parseSection(sectionsStr[1])

		if err != nil {
			return nil, fmt.Errorf("failed to parse second section: %v", err)
		}

		assignments = append(assignments, Assignment{first, second})
	}

	return assignments, nil
}

func part1(assignments []Assignment) int {
	total := 0
	for _, assignment := range assignments {
		overlap := assignment.First.overlap(assignment.Second)

		if overlap == nil {
			continue
		}

		if (*overlap) == assignment.First || (*overlap) == assignment.Second {
			total += 1
		}
	}
	return total
}

func part2(assignments []Assignment) int {
	total := 0
	for _, assignment := range assignments {
		overlap := assignment.First.overlap(assignment.Second)

		if overlap != nil {
			total += 1
		}
	}
	return total
}

func run() error {
	assignments, err := readAssignments(os.Stdin)

	if err != nil {
		return fmt.Errorf("failed to read assignments: %v", err)
	}

	part1 := part1(assignments)

	fmt.Println("Part 1:", part1)

	part2 := part2(assignments)

	fmt.Println("Part 2:", part2)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error in day 4:", err)
		os.Exit(1)
	}
}
