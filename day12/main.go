package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// encoding x and y positions directly in the nodes is a bit weird
// but it makes A* easier to program
type Node struct {
	height     int
	neighbours []*Node
	x          int
	y          int
}

type Maze struct {
	start *Node
	end   *Node
}

type Puzzle struct {
	maze   Maze
	aNodes []*Node
}

func (a *Node) canTravelTo(b *Node) bool {
	return b.height <= a.height+1
}

// I ended up just using plain Dijkstra in the end - I'm not sure why adding this doesn't work
func (a *Node) minDistanceTo(b *Node) int {
	return Abs(a.x-b.x) + Abs(a.y-b.y)
}

func buildNode(c rune, x int, y int) (Node, error) {
	if c == 'S' {
		c = 'a'
	} else if c == 'E' {
		c = 'z'
	}

	if c < 'a' || c > 'z' {
		return Node{}, fmt.Errorf("invalid node %c at %d, %d", c, x, y)
	}

	return Node{
		int(c - 'a'),
		[]*Node{},
		x, y,
	}, nil
}

func parseInput() (Puzzle, error) {
	// first pass: turn all the characters into nodes
	file, err := os.Open("input.txt")
	if err != nil {
		return Puzzle{}, fmt.Errorf("failed to read file: %v", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	nodes := [][]*Node{}
	aNodes := []*Node{}
	var start, end *Node
	for scanner.Scan() {
		line := []rune(scanner.Text())
		nodeLine := []*Node{}
		for x := 0; x < len(line); x++ {
			c := line[x]
			// convince yourself that len(nodes) is the y position of the node
			node, err := buildNode(c, x, len(nodes))

			if err != nil {
				return Puzzle{}, fmt.Errorf("failed to build node: %v", err)
			}

			if c == 'S' {
				if start != nil {
					return Puzzle{}, fmt.Errorf("attempted to build puzzle with two starts")
				}
				start = &node
			} else if c == 'E' {
				if end != nil {
					return Puzzle{}, fmt.Errorf("attempted to build puzzle with two ends")
				}
				end = &node
			} else if c == 'a' {
				aNodes = append(aNodes, &node)
			}

			nodeLine = append(nodeLine, &node)
		}
		nodes = append(nodes, nodeLine)
	}

	if start == nil {
		return Puzzle{}, fmt.Errorf("attempted to build puzzle with no start")
	}

	if end == nil {
		return Puzzle{}, fmt.Errorf("attempted to build puzzle with no end")
	}

	// second pass: connect together all of the nodes which can be travelled between
	for y := 0; y < len(nodes); y++ {
		row := nodes[y]
		for x := 0; x < len(row); x++ {
			node := nodes[y][x]

			// up
			if y > 0 {
				upNode := nodes[y-1][x]
				if node.canTravelTo(upNode) {
					node.neighbours = append(node.neighbours, upNode)
				}
			}

			// right
			if x < len(row)-1 {
				rightNode := nodes[y][x+1]
				if node.canTravelTo(rightNode) {
					node.neighbours = append(node.neighbours, rightNode)
				}
			}

			// down
			if y < len(nodes)-1 {
				downNode := nodes[y+1][x]
				if node.canTravelTo(downNode) {
					node.neighbours = append(node.neighbours, downNode)
				}
			}

			// left
			if x > 0 {
				leftNode := nodes[y][x-1]
				if node.canTravelTo(leftNode) {
					node.neighbours = append(node.neighbours, leftNode)
				}
			}
		}
	}

	return Puzzle{
		Maze{start, end},
		aNodes,
	}, nil
}

// yes, my priority queue is a linked list. problem?
type PriorityQueue struct {
	head *ScoredNode
}

type ScoredNode struct {
	node     *Node
	score    int
	prev     *ScoredNode
	next     *ScoredNode
	pathPrev *ScoredNode
}

func (p *PriorityQueue) remove(sn *ScoredNode) error {
	if sn == nil {
		return fmt.Errorf("attempted to remove nil from list")
	}

	if p.head == nil {
		return fmt.Errorf("attempted to remove from empty list")
	}

	// removing the head
	if p.head == sn {
		p.head = sn.next
		return nil
	}

	// at this point it's an inner node
	if sn.prev == nil {
		return fmt.Errorf("attempted to remove node with nil prev which was not head")
	}

	if sn.prev.next != sn || (sn.next != nil && sn.next.prev != sn) {
		return fmt.Errorf("attempted to remove node whose next.prev or prev.next was not that node")
	}

	sn.prev.next = sn.next
	if sn.next != nil {
		sn.next.prev = sn.prev
	}

	return nil
}

func (p *PriorityQueue) insert(sn ScoredNode) {
	temp := p.head
	for temp.next != nil {
		// first such node is where we want to insert
		if temp.score > sn.score {
			if temp.prev != nil {
				temp.prev.next = &sn
			}
			sn.prev = temp.prev
			sn.next = temp
			temp.prev = &sn
			return
		}
		temp = temp.next
	}

	// temp is the final node. we might need to insert before it:
	if temp.score > sn.score {
		if temp.prev != nil {
			temp.prev.next = &sn
		}
		sn.prev = temp.prev
		sn.next = temp
		temp.prev = &sn
		return
	}

	// this node is larger or equal to the last temp node, so put it at the end
	temp.next = &sn
	sn.prev = temp
	sn.next = nil
}

func (p *PriorityQueue) update(sn ScoredNode) {
	temp := p.head

	// adding to an empty list
	if temp == nil {
		p.head = &sn
		return
	}

	// is the node already in the list?
	var existing *ScoredNode
	for temp != nil {
		if temp.node == sn.node {
			existing = temp
			break
		}
		temp = temp.next
	}

	// node is present
	if existing != nil {
		// no improvement, so no change
		if existing.score <= sn.score {
			return
		}

		// the new node has a better score, so we need to remove the existing and add the new at the right position
		p.remove(existing)
	}

	p.insert(sn)
}

func solve(m Maze) ([]*Node, error) {
	startNode := &ScoredNode{
		m.start,
		0,
		nil,
		nil,
		nil,
	}
	priorityQueue := PriorityQueue{startNode}
	visited := map[*Node]struct{}{}

	path := []*Node{}

	for priorityQueue.head != nil {
		// pop from the queue
		scoredNode := priorityQueue.head
		priorityQueue.head = priorityQueue.head.next

		//fmt.Printf("visiting node %p\n", scoredNode.node)

		// if we found the end, then trace the route back and return it
		if scoredNode.node == m.end {
			revPath := []*Node{}
			temp := scoredNode
			for temp != nil {
				revPath = append(revPath, temp.node)
				temp = temp.pathPrev
			}
			return revPath, nil
		}

		// add neighbours to the queue
		for _, node := range scoredNode.node.neighbours {
			_, visited := visited[node]
			if !visited {
				newNode := ScoredNode{
					node:     node,
					score:    scoredNode.score + 1,
					pathPrev: scoredNode,
					prev:     nil,
					next:     nil,
				}
				priorityQueue.update(newNode)
			}
		}

		// visit the node
		visited[scoredNode.node] = struct{}{}
	}

	if len(path) == 0 {
		return path, fmt.Errorf("failed to find path from start to end")
	}

	return path, nil
}

func part1(p Puzzle) (int, error) {
	path, err := solve(p.maze)
	if err != nil {
		return 0, fmt.Errorf("failed to solve puzzle: %v", err)
	}

	return len(path) - 1, nil
}

func part2(p Puzzle) (int, error) {
	min := math.MaxInt

	// I am CERTAIN this can be done more efficiently by searching from the end back to the start
	// but because of how my adjacency relation works, easier to just throw compute at it :)
	for _, aNode := range p.aNodes {
		candidate, err := solve(Maze{aNode, p.maze.end})

		// some mazes will not be solveable, that's ok.
		if err != nil {
			continue
		}

		if len(candidate) < min {
			min = len(candidate)
		}
	}

	if min == math.MaxInt {
		return min, fmt.Errorf("no valid paths")
	}

	return min - 1, nil
}

func run() error {
	puzzle, err := parseInput()
	if err != nil {
		return fmt.Errorf("failed to parse input: %v", err)
	}

	part1, err := part1(puzzle)

	if err != nil {
		return fmt.Errorf("failed to solve part 1: %v", err)
	}

	fmt.Println("Part 1:", part1)

	part2, err := part2(puzzle)

	if err != nil {
		return fmt.Errorf("failed to solve part 2: %v", err)
	}

	fmt.Println("Part 2:", part2)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve:", err)
		os.Exit(1)
	}
}
