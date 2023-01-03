package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
)

// representing the file as a linked list within an array means we can iterate
// over the nodes in the order they appear (the array) while altering their
// positions (the linked list)
type Node struct {
	value int
	prev  *Node
	next  *Node
}

func parseInput(r io.Reader) ([]*Node, error) {
	// first pass: just get the integers
	ints := []int{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		val, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line as int: %v", err)
		}
		ints = append(ints, int(val))
	}

	// second pass: build the linked list
	var first, prev *Node
	nodes := []*Node{}
	for _, val := range ints {
		node := &Node{value: val, prev: nil, next: nil}

		if prev == nil {
			first = node
		} else {
			node.prev = prev
			prev.next = node
		}

		nodes = append(nodes, node)
		prev = node
	}

	// the final value of prev is the last value, so connect them up
	if first != nil {
		first.prev = prev
		prev.next = first
	}

	return nodes, nil
}

func clone(nodes []*Node) []*Node {
	origFirst := nodes[0]
	newNodes := make([]*Node, 0, len(nodes))

	first := &Node{
		origFirst.value,
		nil,
		nil,
	}
	newNodes = append(newNodes, first)

	prev := first
	for cur := origFirst.next; cur != origFirst; cur = cur.next {
		node := &Node{
			cur.value,
			prev,
			nil,
		}
		prev.next = node
		newNodes = append(newNodes, node)
		prev = node
	}

	// connect the head and tail
	prev.next = first
	first.prev = prev

	return newNodes

}

func Abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func mix(nodes []*Node) {
	for _, node := range nodes {
		val := node.value

		// taking the value modulo len(nodes) - 1 means we never pass around the array a full circle
		// (noting that the elements minus the current value is precisely len(nodes) - 1 in length)
		shift := Abs(val) % (len(nodes) - 1)

		oldPrev, oldNext := node.prev, node.next
		newPrev, newNext := node.prev, node.next

		// splice the node out
		oldPrev.next = oldNext
		oldNext.prev = oldPrev

		// follow the circular list
		for shift != 0 {
			if val < 0 {
				newPrev, newNext = newPrev.prev, newPrev
			} else {
				newPrev, newNext = newNext, newNext.next
			}
			shift--
		}

		// add the node in the new location
		node.prev = newPrev
		newPrev.next = node

		node.next = newNext
		newNext.prev = node
	}
}

func getCoordSum(nodes []*Node) (int, error) {
	//find 0 in the linked list
	cur := nodes[0]
	for cur.value != 0 {
		cur = cur.next
		if cur == nodes[0] {
			return 0, fmt.Errorf("attempted to get coord sum from list with no 0 element")
		}
	}

	//then traverse the linked list the correct number of elements
	var first, second, third int
	for i := 0; i < 3000; i++ {
		cur = cur.next
		if i == 999 {
			first = cur.value
		} else if i == 1999 {
			second = cur.value
		} else if i == 2999 {
			third = cur.value
		}
	}
	return first + second + third, nil

}

func part1(nodes []*Node) (int, error) {
	mix(nodes)
	return getCoordSum(nodes)
}

const DECRYPTION_KEY = 811589153

func part2(nodes []*Node) (int, error) {
	for _, node := range nodes {
		node.value *= DECRYPTION_KEY
	}

	for i := 0; i < 10; i++ {
		mix(nodes)
	}

	return getCoordSum(nodes)
}

func run() error {
	input, err := os.Open("input.txt")
	if err != nil {
		return fmt.Errorf("failed to open input.txt: %v", err)
	}

	defer input.Close()

	nodes, err := parseInput(input)
	if err != nil {
		return fmt.Errorf("failed to parse input file: %v", err)
	}

	fmt.Println(nodes)
	clone := clone(nodes)

	part1, err := part1(nodes)
	if err != nil {
		return fmt.Errorf("failed to solve part1: %v", err)
	}
	fmt.Println("Part 1:", part1)

	part2, err := part2(clone)
	if err != nil {
		return fmt.Errorf("failed to solve part2: %v", err)
	}
	fmt.Println("Part 2:", part2)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve day20:", err)
		os.Exit(1)
	}
}
