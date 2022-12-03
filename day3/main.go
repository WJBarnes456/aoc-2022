package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

type Rucksack struct {
	full   []rune
	first  []rune
	second []rune
}

// SharedItem returns the first shared item between the two compartments
func (r Rucksack) SharedItem() (rune, error) {
	set := make(map[rune]bool)

	for _, item := range r.first {
		set[item] = true
	}

	for _, item := range r.second {
		if set[item] {
			return item, nil
		}
	}

	return 0, errors.New("no shared item")
}

func readRucksacks(r io.Reader) ([]Rucksack, error) {
	scanner := bufio.NewScanner(r)

	rucksacks := make([]Rucksack, 0)
	for scanner.Scan() {
		line := scanner.Text()

		// splitting on runes is a fun way of practicing unicode support,
		// although it's not needed here
		lineRunes := []rune(line)

		if len(lineRunes)%2 != 0 {
			return nil, fmt.Errorf("line %v cannot be split in 2", line)
		}

		midpoint := len(lineRunes) / 2

		rucksack := Rucksack{lineRunes, lineRunes[:midpoint], lineRunes[midpoint:]}

		rucksacks = append(rucksacks, rucksack)
	}

	return rucksacks, nil
}

func prioritise(item rune) (int, error) {
	if 'a' <= item && item <= 'z' {
		return int(item-'a') + 1, nil
	} else if 'A' <= item && item <= 'Z' {
		return int(item-'A') + 27, nil
	} else {
		return 0, fmt.Errorf("no priority for rune %v", item)
	}
}

func part1(rucksacks []Rucksack) (int, error) {
	total := 0
	for _, rucksack := range rucksacks {
		sharedItem, err := rucksack.SharedItem()

		if err != nil {
			return total, fmt.Errorf("failed to calculate part1: %v", err)
		}

		score, err := prioritise(sharedItem)

		if err != nil {
			return total, fmt.Errorf("failed to calculate part1: %v", err)
		}

		total += score
	}
	return total, nil
}

// This only works for up to 32 rucksacks - a better solution would be to turn
// the first rucksack into a set, and test membership of each character of
// that against every subsequent rucksack
func getSharedItem(rucksacks []Rucksack) (rune, error) {
	masks := make(map[rune]int)

	// set each bit if it's present in that rucksack
	for i, rucksack := range rucksacks {
		for _, c := range rucksack.full {
			masks[c] = masks[c] | (1 << i)
		}
	}

	// the shared value is the one present in all rucksacks
	target_value := 1<<(len(rucksacks)) - 1

	for item, mask := range masks {
		if mask == target_value {
			return item, nil
		}
	}

	return 0, errors.New("no shared item in rucksacks")
}

func part2(rucksacks []Rucksack) (int, error) {
	GROUP_SIZE := 3

	total := 0
	for i := 0; i < len(rucksacks); i += GROUP_SIZE {
		group := rucksacks[i : i+GROUP_SIZE]
		item, err := getSharedItem(group)

		if err != nil {
			return total, fmt.Errorf("failed to get shared item: %v", err)
		}

		priority, err := prioritise(item)

		if err != nil {
			return total, fmt.Errorf("failed to prioritise item: %v", err)
		}

		total += priority
	}

	return total, nil
}

func run() error {
	rucksacks, err := readRucksacks(os.Stdin)

	if err != nil {
		return fmt.Errorf("failed to read rucksacks: %v", err)
	}

	//fmt.Fprintln(os.Stderr, "rucksacks:", rucksacks)

	part1, err := part1(rucksacks)

	if err != nil {
		return fmt.Errorf("failed to do part1: %v", err)
	}

	fmt.Println("Part 1:", part1)

	part2, err := part2(rucksacks)

	if err != nil {
		return fmt.Errorf("failed to do part2: %v", err)
	}

	fmt.Println("Part 2:", part2)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to solve challenge: %v", err)
		os.Exit(1)
	}
}
