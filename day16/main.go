package main

import (
	"bufio"
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
	for i, valve := range valves {
		newValves[i] = valve
	}
	sort.Strings(newValves)

	return strings.Join(newValves, "")
}

func (_ *Memo) getState(occupiedValves []string, openValves map[string]*Valve, timeRemaining int) State {
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
		// ignore all valves with a flow rate less than 0
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

func (m *Memo) score(valves map[string]*Valve, occupiedValves []string, openValves map[string]*Valve, timeRemaining int) int {
	state := m.getState(occupiedValves, openValves, timeRemaining)
	value, alreadyCalculated := (*m)[state]
	if alreadyCalculated {
		return value
	}

	if timeRemaining == 0 {
		return 0
	}

	// once all valves are open, no further action is useful
	if allValvesOpen(valves, openValves) {
		return timeRemaining * totalScore(openValves)
	}

	roundScore := totalScore(openValves)

	bestScore := 0

	currentValve := occupiedValves[0]
	_, opened := openValves[currentValve]
	if valves[currentValve].flowRate > 0 && !opened {
		potentialOpenValves := make(map[string]*Valve, len(openValves)+1)
		for name, valve := range openValves {
			potentialOpenValves[name] = valve
		}
		potentialOpenValves[currentValve] = valves[currentValve]
		bestScore = max(bestScore, m.score(valves, []string{currentValve}, potentialOpenValves, timeRemaining-1))
	}

	for _, neighbour := range valves[currentValve].neighbours {
		bestScore = max(bestScore, m.score(valves, []string{neighbour.name}, openValves, timeRemaining-1))
	}

	value = roundScore + bestScore
	(*m)[state] = value

	return value
}

func part1(valves map[string]*Valve) int {
	memo := Memo(map[State]int{})
	return memo.score(valves, []string{"AA"}, map[string]*Valve{}, 30)
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

	fmt.Println("Part 1:", part1(valves))

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve day16:", err)
		os.Exit(1)
	}
}
