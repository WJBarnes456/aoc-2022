package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

type Elves [][]int

func getElves(r io.Reader) ([][]int, error) {
	scanner := bufio.NewScanner(r)

	elves := make([][]int, 0)
	current_elf := make([]int, 0)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			elves = append(elves, current_elf)
			current_elf = make([]int, 0)
		} else {
			val64, err := strconv.ParseInt(scanner.Text(), 10, 32)

			if err != nil {
				return nil, fmt.Errorf("failed to parse calories: %w", err)
			}

			val := int(val64)

			current_elf = append(current_elf, val)
		}
	}

	elves = append(elves, current_elf)

	return elves, nil
}

func getSums(elves Elves) []int {
	elves_sums := make([]int, 0, len(elves))

	for _, elf := range elves {
		sum := 0
		for _, calories := range elf {
			sum += calories
		}
		elves_sums = append(elves_sums, sum)
	}

	return elves_sums
}

func part1(elves Elves) int {
	return topN(elves, 1)
}

func part2(elves Elves) int {
	return topN(elves, 3)
}

func topN(elves Elves, top_n int) int {
	elves_sums := getSums(elves)
	num_elves := len(elves_sums)

	sort.Ints(elves_sums)

	sum := 0
	for _, value := range elves_sums[num_elves-top_n:] {
		sum += value
	}

	return sum
}

func main() {
	elves, err := getElves(os.Stdin)
	if err != nil {
		fmt.Println("Error encountered", err)
		os.Exit(1)
	}

	part1 := part1(elves)

	fmt.Println("Part 1 solution:", part1)

	part2 := part2(elves)

	fmt.Println("Part 2 solution", part2)
}
