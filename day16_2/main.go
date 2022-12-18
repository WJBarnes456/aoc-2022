package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

// A re-implementation of day16, with a couple of key optimisations:
// - optimise part1 by excluding empty paths
// - optimise part2 by re-using part1 rather than generalising it
// (I ruined the challenge for myself a little bit by looking up how to do it, but I think it was a good learning experience, and I'd already spent so long on it - I've got other things I'd like to do!)

type Valve struct {
	name       string
	flowRate   int
	neighbours []*Valve
}

type Node struct {
	name     string
	flowRate int
	edges    []Edge
}

type Edge struct {
	timeCost int
	dest     *Node
}

type Graph struct {
	nodes []*Node
	start *Node
}

type Path []string

type Valves map[string]*Valve

type State struct {
	currentNode   string
	openValves    string
	timeRemaining int
}

func linearise(currentNode *Node, openValves map[string]struct{}, timeRemaining int) State {
	openValveSlice := make([]string, 0, len(openValves))
	for name := range openValves {
		openValveSlice = append(openValveSlice, name)
	}
	sort.Strings(openValveSlice)
	linearisedValves := strings.Join(openValveSlice, "")

	return State{
		currentNode:   currentNode.name,
		openValves:    linearisedValves,
		timeRemaining: timeRemaining,
	}
}

type Memo map[State]int

func (m *Memo) score(g Graph, currentNode *Node, openValves map[string]struct{}, timeRemaining int) int {
	state := linearise(currentNode, openValves, timeRemaining)
	if value, alreadyComputed := (*m)[state]; alreadyComputed {
		return value
	}

	nodeScore := timeRemaining * currentNode.flowRate
	newOpenValves := make(map[string]struct{}, len(openValves)+1)
	for valve := range openValves {
		newOpenValves[valve] = struct{}{}
	}
	newOpenValves[currentNode.name] = struct{}{}

	bestScore := 0
	for _, edge := range currentNode.edges {
		if _, alreadyVisited := openValves[edge.dest.name]; !alreadyVisited {
			if edge.timeCost < timeRemaining {
				score := m.score(g, edge.dest, newOpenValves, timeRemaining-edge.timeCost-1)
				if score > bestScore {
					bestScore = score
				}
			}
		}
	}

	value := nodeScore + bestScore
	(*m)[state] = value
	return value
}

func (g Graph) part1() int {
	memo := Memo{}
	return memo.score(g, g.start, map[string]struct{}{}, 30)
}

func (g Graph) part2() int {
	memo := Memo{}
	dividedNodes := make([]*Node, 0, len(g.nodes))
	for _, node := range g.nodes {
		if node != g.start {
			dividedNodes = append(dividedNodes, node)
		}
	}

	// This came from the subreddit, I'm not sure I would've come up with this on my own
	// you can break part2 down into part1 by considering the valves as being divided between you and the elephant
	// and finding the optimal division.
	// this is memoised on the same memo (!!), because the situations are otherwise the same!
	best := 0
	for _, division := range generateAllDivisions(dividedNodes) {
		youBlocked := map[string]struct{}{}
		for _, name := range division[0] {
			youBlocked[name] = struct{}{}
		}

		elBlocked := map[string]struct{}{}
		for _, name := range division[1] {
			elBlocked[name] = struct{}{}
		}

		score := memo.score(g, g.start, youBlocked, 26) + memo.score(g, g.start, elBlocked, 26)
		if score > best {
			best = score
		}
	}
	return best
}

// Generates all divisions of a list of nodes
// e.g. given [A,B], it will return [[[], [A,B]], [[A], [B]], [[B], [A]], [[A,B], []]
func generateAllDivisions(nodes []*Node) [][][]string {
	if len(nodes) == 0 {
		return [][][]string{{{}, {}}}
	}
	divisions := make([][][]string, 0, 1<<len(nodes))
	currentNode := nodes[0]
	childDivs := generateAllDivisions(nodes[1:])
	for _, division := range childDivs {
		choiceA := make([]string, len(division[0]), len(division[0])+1)
		copy(choiceA, division[0])
		choiceA = append(choiceA, currentNode.name)

		choiceB := make([]string, len(division[1]), len(division[1])+1)
		copy(choiceB, division[1])
		choiceB = append(choiceB, currentNode.name)

		divisions = append(divisions, [][]string{division[0], choiceB})
		divisions = append(divisions, [][]string{choiceA, division[1]})
	}
	return divisions
}

func (v *Valves) graphify() Graph {
	shortestPaths := getAllShortestPaths(*v)

	nodes := make(map[string]*Node, len(shortestPaths)+1)

	// first pass: turn the valves into nodes
	for usefulNodeName, _ := range shortestPaths {
		valve := (*v)[usefulNodeName]
		nodes[valve.name] = &Node{valve.name, valve.flowRate, []Edge{}}
	}

	_, startIsUseful := nodes["AA"]
	if !startIsUseful {
		nodes["AA"] = &Node{"AA", 0, []Edge{}}
	}

	// second pass: add the edges
	for destName, pathsIn := range shortestPaths {
		destNode := nodes[destName]
		for startName, path := range pathsIn {
			startNode, isUseful := nodes[startName]
			if !isUseful {
				continue
			}

			// don't loop to yourself
			if startNode == destNode {
				continue
			}

			startNode.edges = append(startNode.edges, Edge{len(path), destNode})
		}
	}

	// final pass: flatten the map, as the names are no longer important
	startNode := nodes["AA"]
	outNodes := make([]*Node, 0, len(nodes))
	for _, node := range nodes {
		outNodes = append(outNodes, node)
	}

	return Graph{
		start: startNode,
		nodes: outNodes,
	}
}

func parseInput(r io.Reader) (Valves, error) {
	scanner := bufio.NewScanner(r)
	nameToValve := map[string]*Valve{}
	nameToOtherValves := map[string][]string{}

	tunnelsMatch, err := regexp.Compile(`tunnels? leads? to valves? (.*)$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile tunnel regex: %v", err)
	}

	for scanner.Scan() {
		line := scanner.Text()
		var valveName string
		var flowRate int
		_, err := fmt.Sscanf(line, "Valve %2s has flow rate=%d;", &valveName, &flowRate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line: %v", err)
		}

		otherValvesMatch := tunnelsMatch.FindStringSubmatch(line)
		if otherValvesMatch == nil {
			return nil, fmt.Errorf("tunnelsMatch did not match line: %s", line)
		}
		otherValvesStr := otherValvesMatch[1]
		otherValves := strings.Split(otherValvesStr, ", ")
		nameToValve[valveName] = &Valve{
			valveName,
			flowRate,
			nil,
		}
		nameToOtherValves[valveName] = otherValves
	}

	// connect up the neighbours
	for _, valve := range nameToValve {
		neighbourNames := nameToOtherValves[valve.name]
		valve.neighbours = make([]*Valve, len(neighbourNames))
		for j, neighbourName := range neighbourNames {
			valve.neighbours[j] = nameToValve[neighbourName]
		}
	}

	return nameToValve, nil
}

func getShortestPaths(valves Valves, targetValve *Valve) map[string][]string {
	// just use dijkstra here
	visited := map[*Valve]struct{}{}
	shortestPaths := map[string][]string{}
	pq := PriorityQueue{}
	shortestPaths[targetValve.name] = []string{}
	heap.Push(&pq, &Item{value: targetValve, distance: 0})
	for len(pq) > 0 {
		item := heap.Pop(&pq).(*Item)

		valve := item.value
		_, alreadyVisited := visited[valve]
		if alreadyVisited {
			continue
		}

		for _, neighbour := range valve.neighbours {
			oldNeighbourPath, oldPathExists := shortestPaths[neighbour.name]
			if !oldPathExists || len(shortestPaths[neighbour.name])+1 < len(oldNeighbourPath) {
				shortestPaths[neighbour.name] = []string{valve.name}
				shortestPaths[neighbour.name] = append(shortestPaths[neighbour.name], shortestPaths[valve.name]...)
				heap.Push(&pq, &Item{value: neighbour, distance: len(shortestPaths[neighbour.name])})
			}
		}
		visited[valve] = struct{}{}
	}
	return shortestPaths
}

func getAllShortestPaths(valves Valves) map[string]map[string][]string {
	// we're only interested in shortest paths to valves with non-zero start points
	shortestPaths := map[string]map[string][]string{}
	for _, targetValve := range valves {
		if targetValve.flowRate <= 0 {
			continue
		}
		shortestPaths[targetValve.name] = getShortestPaths(valves, targetValve)
	}
	return shortestPaths
}

func run() error {
	file, err := os.Open("input.txt")
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()

	valves, err := parseInput(file)
	if err != nil {
		return fmt.Errorf("failed to parse valves: %v", err)
	}

	graph := valves.graphify()

	fmt.Println("Part 1:", graph.part1())

	fmt.Println("Part 2:", graph.part2())

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve day16:", err)
		os.Exit(1)
	}
}
