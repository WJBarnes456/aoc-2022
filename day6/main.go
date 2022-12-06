package main

import (
	"fmt"
	"io"
	"os"
)

type Packet struct {
	// Position for the start of the packet contents
	startPosition int
	// Contents of the packet
	// contents []rune
}

// this is O(n^2), but for small numbers of characters is probably better than building + maintaining a map
// (you can do it in O(n) by counting the number of times the last character is different to every previous character in the window)
func differentChars(s []rune) bool {
	for i, c1 := range s {
		if i == len(s)-1 {
			break
		}
		for _, c2 := range s[i+1:] {
			if c1 == c2 {
				return false
			}
		}
	}
	return true
}

func identifyPackets(buffer []rune, headerLength int) ([]Packet, error) {
	if len(buffer) < headerLength {
		return nil, fmt.Errorf("input of length <headerLength cannot contain any packets")
	}

	packets := []Packet{}
	start := -1
	for i := headerLength; i < len(buffer); i++ {
		potentialHeader := buffer[i-headerLength : i]

		// nothing to do if this is part of the previous packet
		if !differentChars(potentialHeader) {
			continue
		}

		// save the previous one
		if start != -1 {
			packets = append(packets, Packet{start})
		}
		start = i
	}

	if start != -1 {
		packets = append(packets, Packet{start})
	}

	return packets, nil
}

// Because "The signal is a series of seemingly-random characters that the device receives one at a time."
// I was sorely tempted to make this run online (i.e. run on stdin one byte at a time), but I resisted.
func run() error {
	buffer, err := io.ReadAll(os.Stdin)

	if err != nil {
		return fmt.Errorf("failed to read stdin to buffer: %v", err)
	}

	input := string(buffer)

	part1, err := identifyPackets([]rune(input), 4)

	if err != nil {
		return fmt.Errorf("failed to identify packets: %v", err)
	}

	fmt.Println("Part 1:", part1[0].startPosition)

	part2, err := identifyPackets([]rune(input), 14)

	if err != nil {
		return fmt.Errorf("failed to identify packets: %v", err)
	}

	fmt.Println("Part 2:", part2[0].startPosition)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error in day 6: %v", err)
		os.Exit(1)
	}
}
