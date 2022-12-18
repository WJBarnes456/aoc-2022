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

type Valve struct {
	name       string
	flowRate   int
	neighbours []*Valve
}

type State struct {
	occupiedValves string
	openValves     string
	timeRemaining  int
}

type Move struct {
	position    string
	openedValve *string
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseInput(r io.Reader) (map[string]*Valve, error) {
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

type Memo map[State]int

func flattenValves(valves []string) string {
	newValves := make([]string, len(valves))
	copy(newValves, valves)

	sort.Strings(newValves)

	return strings.Join(newValves, "")
}

func (*Memo) getState(occupiedValves []string, openValves map[string]*Valve, timeRemaining int) State {
	openValveSlice := make([]string, len(openValves))
	for _, valve := range openValves {
		openValveSlice = append(openValveSlice, valve.name)
	}

	return State{
		occupiedValves: flattenValves(occupiedValves),
		openValves:     flattenValves(openValveSlice),
		timeRemaining:  timeRemaining,
	}
}

func allValvesOpen(valves map[string]*Valve, openValves map[string]*Valve) bool {
	for name, valve := range valves {
		// ignore all valves with a flow rate less than or equal to 0
		if valve.flowRate <= 0 {
			continue
		}
		_, open := openValves[name]
		if !open {
			return false
		}
	}
	return true
}

func totalScore(openValves map[string]*Valve) int {
	total := 0
	for _, valve := range openValves {
		total += valve.flowRate
	}
	return total
}

func addOpenValves(valves map[string]*Valve, openValves map[string]*Valve, valvesToOpen []*string) map[string]*Valve {
	newOpenValves := make(map[string]*Valve, len(openValves)+len(valvesToOpen))
	for name, valve := range openValves {
		newOpenValves[name] = valve
	}

	for _, name := range valvesToOpen {
		if name != nil {
			newOpenValves[*name] = valves[*name]
		}
	}
	return newOpenValves
}

// given a list of lists of moves for each agent
// return a list of every combination of moves from each of those lists
// e.g. [[A,B], [C,D]] -> [[A,C],[A,D],[B,C],[B,D]]
func allMoveCombinations(allAgentMoves [][]Move) [][]Move {
	// this is very naturally expressed in a functional style
	// I'm sure this can be optimised
	if len(allAgentMoves) == 1 {
		out := make([][]Move, 0, len(allAgentMoves[0]))
		for _, move := range allAgentMoves[0] {
			out = append(out, []Move{move})
		}
		return out
	}

	combos := allMoveCombinations(allAgentMoves[1:])
	theseMoves := allAgentMoves[0]
	out := make([][]Move, 0, len(combos)*len(theseMoves))
	for _, move := range theseMoves {
		for _, combo := range combos {
			newCombo := make([]Move, 0, len(combo)+1)
			newCombo = append(newCombo, move)
			newCombo = append(newCombo, combo...)
			out = append(out, newCombo)
		}
	}
	return out
}

func (m *Memo) score(valves map[string]*Valve, shortestPaths map[string]map[string][]string, occupiedValves []string, openValves map[string]*Valve, timeRemaining int) int {
	state := m.getState(occupiedValves, openValves, timeRemaining)
	value, alreadyCalculated := (*m)[state]
	if alreadyCalculated {
		return value
	}

	if timeRemaining == 0 {
		return 0
	}

	roundScore := totalScore(openValves)

	// once all valves are open, no further action is useful
	if allValvesOpen(valves, openValves) {
		value := timeRemaining * roundScore
		(*m)[state] = value
		return value
	}

	allAgentMoves := [][]Move{}
	for _, valveName := range occupiedValves {
		valve := valves[valveName]
		agentMoves := []Move{}
		_, opened := openValves[valveName]
		if valve.flowRate > 0 && !opened {
			// NB: can't just use &valveName as that changes as the loop progresses
			// no wonder I was ending up with non-sensical results in the multi-agent case!
			agentMoves = append(agentMoves, Move{valveName, &valve.name})
		}

		for _, neighbour := range valve.neighbours {
			// TODO only consider neighbours which are on the shortest path to a useful node
			agentMoves = append(agentMoves, Move{neighbour.name, nil})
		}
		allAgentMoves = append(allAgentMoves, agentMoves)
	}

	bestScore := 0

	combos := allMoveCombinations(allAgentMoves)
	for _, moveCombo := range combos {
		nextPositions := make([]string, len(moveCombo))
		valvesToOpen := make([]*string, len(moveCombo))
		for i, move := range moveCombo {
			nextPositions[i] = move.position
			valvesToOpen[i] = move.openedValve
		}
		newOpenValves := addOpenValves(valves, openValves, valvesToOpen)
		bestScore = max(bestScore, m.score(valves, shortestPaths, nextPositions, newOpenValves, timeRemaining-1))
	}

	value = roundScore + bestScore
	(*m)[state] = value

	return value
}

func part1(valves map[string]*Valve, shortestPaths map[string]map[string][]string) int {
	memo := Memo(map[State]int{})
	score := memo.score(valves, shortestPaths, []string{"AA"}, map[string]*Valve{}, 30)
	return score
}

func part2(valves map[string]*Valve, shortestPaths map[string]map[string][]string) int {
	memo := Memo(map[State]int{})
	return memo.score(valves, shortestPaths, []string{"AA", "AA"}, map[string]*Valve{}, 26)
}

func getShortestPaths(valves map[string]*Valve, targetValve *Valve) map[string][]string {
	// just use dijkstra here
	visited := map[*Valve]struct{}{}
	shortestPaths := map[string][]string{}
	pq := PriorityQueue{}
	shortestPaths[targetValve.name] = []string{}
	heap.Push(&pq, &Item{value: targetValve, priority: 0})
	for len(pq) > 0 {
		item := heap.Pop(&pq).(*Item)

		valve := item.value
		_, alreadyVisited := visited[valve]
		if alreadyVisited {
			continue
		}

		for _, neighbour := range valve.neighbours {
			oldNeighbourPath, oldPathExists := shortestPaths[neighbour.name]
			if !oldPathExists || len(shortestPaths[valve.name])+1 < len(oldNeighbourPath) {
				shortestPaths[neighbour.name] = []string{valve.name}
				shortestPaths[neighbour.name] = append(shortestPaths[neighbour.name], shortestPaths[valve.name]...)
				heap.Push(&pq, &Item{value: neighbour, priority: len(shortestPaths[neighbour.name])})
			}
		}
		visited[valve] = struct{}{}
	}
	return shortestPaths
}

func getAllShortestPaths(valves map[string]*Valve) map[string]map[string][]string {
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

	fmt.Println("Parsed input:")
	for _, valve := range valves {
		fmt.Println(valve)
	}

	shortestPaths := getAllShortestPaths(valves)

	fmt.Println("Part 1:", part1(valves, shortestPaths))

	fmt.Println("Part 2:", part2(valves, shortestPaths))

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve day16:", err)
		os.Exit(1)
	}
}
