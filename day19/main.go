package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
)

// we could make this a map, but using a struct instead, i.e. a value type,
// means we can index the memo by it directly without needing to flatten any
// values
type Resources struct {
	ore      int
	clay     int
	obsidian int
	geodes   int
}

type Resource int

const (
	Ore Resource = iota
	Clay
	Obsidian
	Geodes
)

func (r *Resources) Plus(r2 Resources) {
	r.ore += r2.ore
	r.clay += r2.clay
	r.obsidian += r2.obsidian
	r.geodes += r2.geodes
}

func (r *Resources) Subtract(r2 Resources) {
	r.ore -= r2.ore
	r.clay -= r2.clay
	r.obsidian -= r2.obsidian
	r.geodes -= r2.geodes
}

// GreaterThanOrEqual - i.e. a.Gteq(b) implies every field of a is >= every field of b
func (r *Resources) Gteq(r2 Resources) bool {
	return r.ore >= r2.ore && r.clay >= r2.clay && r.obsidian >= r2.obsidian && r.geodes >= r2.geodes
}

type Blueprint struct {
	number          int
	oreBotCost      Resources
	clayBotCost     Resources
	obsidianBotCost Resources
	geodeBotCost    Resources
}

type State struct {
	// Blueprint used to allow multiple states to exist on one memo without hitting conflicts
	blueprint     *Blueprint
	timeRemaining int
	resources     Resources
	bots          Resources
}

type Memo map[State]int

func max(ints ...int) int {
	max := math.MinInt
	for _, i := range ints {
		if i > max {
			max = i
		}
	}
	return max
}

func parseInput(r io.Reader) ([]*Blueprint, error) {
	scanner := bufio.NewScanner(r)
	blueprints := []*Blueprint{}
	for scanner.Scan() {
		line := scanner.Text()
		var blueprintNumber, oreCost, clayCost, obsidianCostOre, obsidianCostClay, geodeCostOre, geodeCostObsidian int
		parsed, err := fmt.Sscanf(line, "Blueprint %d: Each ore robot costs %d ore. Each clay robot costs %d ore. Each obsidian robot costs %d ore and %d clay. Each geode robot costs %d ore and %d obsidian",
			&blueprintNumber, &oreCost, &clayCost, &obsidianCostOre, &obsidianCostClay, &geodeCostOre, &geodeCostObsidian)

		if err != nil {
			return nil, fmt.Errorf("failed to parse line '%v': %v", line, err)
		}

		if parsed != 7 {
			return nil, fmt.Errorf("failed to parse line '%v': expected 7 arguments, got %d", line, parsed)
		}

		blueprints = append(blueprints, &Blueprint{
			number:          blueprintNumber,
			oreBotCost:      Resources{ore: oreCost},
			clayBotCost:     Resources{ore: clayCost},
			obsidianBotCost: Resources{ore: obsidianCostOre, clay: obsidianCostClay},
			geodeBotCost:    Resources{ore: geodeCostOre, obsidian: geodeCostObsidian},
		})
	}
	return blueprints, nil
}

func (s *State) StateAfterBuilding(botCost Resources, botType Resource) State {
	nextResources := s.resources
	nextResources.Plus(s.bots)
	nextResources.Subtract(botCost)

	nextBots := s.bots
	switch botType {
	case Ore:
		nextBots.ore++
	case Clay:
		nextBots.clay++
	case Obsidian:
		nextBots.obsidian++
	case Geodes:
		nextBots.geodes++
	default:
		panic(fmt.Sprintf("invalid resource %d", botType))
	}

	return State{
		blueprint:     s.blueprint,
		timeRemaining: s.timeRemaining - 1,
		resources:     nextResources,
		bots:          nextBots,
	}
}

func (m *Memo) maxGeodes(s State, bestSoFar *int) int {
	// look up in memo if present
	if val, exists := (*m)[s]; exists {
		return val
	}

	// base geodes is the number we will make in the remaining time
	score := s.resources.geodes + s.bots.geodes*s.timeRemaining

	// if out of time, no opportunity to do anything more
	if s.timeRemaining == 0 {
		return score
	}

	// if there's no way to exceed the best we've seen so far, no need to continue
	// best case scenario, we build another geode bot every turn, so result is timeRemaining + timeRemaining-1 + ... + 1
	// i.e. t(t+1)/2
	if score+(s.timeRemaining*(s.timeRemaining+1))/2 < *bestSoFar {
		return score
	}

	// if you can build a geode bot, it is always the best thing to do to maximise geodes
	if s.resources.Gteq(s.blueprint.geodeBotCost) {
		nextState := s.StateAfterBuilding(s.blueprint.geodeBotCost, Geodes)

		nextScore := m.maxGeodes(nextState, bestSoFar)
		(*m)[s] = nextScore
		return nextScore
	}

	// if you can't build everything on one turn, consider building an ore bot
	if s.bots.ore < max(s.blueprint.oreBotCost.ore, s.blueprint.clayBotCost.ore, s.blueprint.obsidianBotCost.ore, s.blueprint.geodeBotCost.ore) && s.resources.Gteq(s.blueprint.oreBotCost) {
		nextState := s.StateAfterBuilding(s.blueprint.oreBotCost, Ore)

		nextScore := m.maxGeodes(nextState, bestSoFar)
		if nextScore > score {
			score = nextScore
		}

		if nextScore > *bestSoFar {
			*bestSoFar = nextScore
		}
	}

	// if you can't build an obsidian bot every turn, consider building a clay bot
	if s.bots.clay < max(s.blueprint.obsidianBotCost.clay) && s.resources.Gteq(s.blueprint.clayBotCost) {
		nextState := s.StateAfterBuilding(s.blueprint.clayBotCost, Clay)

		nextScore := m.maxGeodes(nextState, bestSoFar)
		if nextScore > score {
			score = nextScore
		}

		if nextScore > *bestSoFar {
			*bestSoFar = nextScore
		}
	}

	// if you can't build a geode bot every turn, consider building an obsidian bot
	if s.bots.obsidian < max(s.blueprint.geodeBotCost.obsidian) && s.resources.Gteq(s.blueprint.obsidianBotCost) {
		nextState := s.StateAfterBuilding(s.blueprint.obsidianBotCost, Obsidian)

		nextScore := m.maxGeodes(nextState, bestSoFar)
		if nextScore > score {
			score = nextScore
		}

		if nextScore > *bestSoFar {
			*bestSoFar = nextScore
		}
	}

	// consider doing nothing and accumulating resources
	nextResources := s.resources
	nextResources.Plus(s.bots)
	nextScore := m.maxGeodes(State{
		blueprint:     s.blueprint,
		resources:     nextResources,
		bots:          s.bots,
		timeRemaining: s.timeRemaining - 1,
	}, bestSoFar)
	if nextScore > score {
		score = nextScore
	}

	if nextScore > *bestSoFar {
		*bestSoFar = nextScore
	}

	(*m)[s] = score
	return score
}

func (b *Blueprint) maxGeodes(startState State) int {
	memo := make(Memo)
	best := 0
	return memo.maxGeodes(startState, &best)
}

func (b *Blueprint) qualityScore(startState State) int {
	return b.number * b.maxGeodes(startState)
}

func part1_worker(blueprints <-chan *Blueprint, scores chan<- int) {
	for b := range blueprints {
		startState := State{
			blueprint:     b,
			bots:          Resources{ore: 1},
			timeRemaining: 24,
		}
		scores <- b.qualityScore(startState)
	}
}

func part1(blueprints []*Blueprint) int {
	scores := make(chan int, len(blueprints))
	defer close(scores)

	jobs := make(chan *Blueprint, len(blueprints))
	defer close(jobs)

	for i := 0; i < 8; i++ {
		go part1_worker(jobs, scores)
	}

	for _, blueprint := range blueprints {
		jobs <- blueprint
	}

	sum := 0
	for range blueprints {
		sum += <-scores
	}

	return sum
}

func part2(blueprints []*Blueprint) int {
	total := 1
	for _, blueprint := range blueprints[:3] {
		startState := State{
			blueprint:     blueprint,
			timeRemaining: 32,
			bots:          Resources{ore: 1},
		}
		total *= blueprint.maxGeodes(startState)
	}
	return total
}

func run() error {
	input, err := os.Open("input.txt")
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}

	defer input.Close()

	blueprints, err := parseInput(input)
	if err != nil {
		return fmt.Errorf("failed to parse input: %v", err)
	}
	fmt.Println("blueprints:", blueprints)

	fmt.Println("part 1:", part1(blueprints))

	fmt.Println("part 2:", part2(blueprints))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to solve day19:", err)
		os.Exit(1)
	}
}
