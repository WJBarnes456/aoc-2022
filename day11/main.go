package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Operator int

const (
	Add Operator = iota
	Multiply
	Subtract
)

// prev = true -> substitute the previous value
// prev = false -> use the value instead
type Value struct {
	prev  bool
	value int
}

func (v *Value) asInt(old int) int {
	if v.prev {
		return old
	}
	return v.value
}

type Expression struct {
	operator Operator
	a        Value
	b        Value
}

func (e *Expression) Evaluate(old int) (int, error) {
	a := e.a.asInt(old)
	b := e.b.asInt(old)

	switch e.operator {
	case Add:
		return a + b, nil
	case Multiply:
		return a * b, nil
	case Subtract:
		return a - b, nil
	default:
		return 0, fmt.Errorf("unknown expression type %v", e.operator)
	}
}

type Monkey struct {
	items            []int
	operation        Expression
	divisibilityTest int
	trueDest         int
	falseDest        int
}

// value mod the product of each divisibility test preserves divisibility by each factor
// (it helps that all the factors are prime)
func getSharedBasis(monkeys []Monkey) int {
	sharedBasis := 1
	for _, m := range monkeys {
		sharedBasis *= m.divisibilityTest
	}
	return sharedBasis
}

func (m *Monkey) turn(monkeys []Monkey, part2 bool) error {
	sharedBasis := getSharedBasis(monkeys)

	// each monkey will always end its turn with no items, so we can just iterate over
	for _, value := range m.items {
		value, err := m.operation.Evaluate(value)
		if err != nil {
			return fmt.Errorf("failed to take turn: %v", err)
		}

		if !part2 {
			value = value / 3
		}
		value = (value + sharedBasis) % sharedBasis

		divisible := (value % m.divisibilityTest) == 0

		dest := m.falseDest
		if divisible {
			dest = m.trueDest
		}

		monkeys[dest].items = append(monkeys[dest].items, value)
	}

	m.items = m.items[:0]
	return nil
}

func parseItemLine(itemLine string) ([]int, error) {
	itemParts := strings.Split(itemLine, " ")

	if itemParts[0] != "" || itemParts[1] != "" || itemParts[2] != "Starting" || itemParts[3] != "items:" {
		return nil, fmt.Errorf("tried to parse invalid item line %s", itemLine)
	}

	items := make([]int, 0, len(itemParts)-3)
	for i := 4; i < len(itemParts); i++ {
		value := itemParts[i]
		if value[len(value)-1] == ',' {
			value = value[:len(value)-1]
		}
		var intValue int
		_, err := fmt.Sscanf(value, "%d", &intValue)
		if err != nil {
			return items, fmt.Errorf("tried to parse invalid item %s", itemParts[i])
		}

		items = append(items, intValue)
	}
	return items, nil
}

func parseValue(valueStr string) (Value, error) {
	if valueStr == "old" {
		return Value{prev: true, value: 0}, nil
	}

	var val int
	_, err := fmt.Sscanf(valueStr, "%d", &val)
	if err != nil {
		return Value{}, fmt.Errorf("failed to parse value %s: %v", valueStr, err)
	}

	return Value{prev: false, value: val}, nil
}

func parseOperator(opStr string) (Operator, error) {
	switch opStr {
	case "*":
		return Multiply, nil
	case "+":
		return Add, nil
	case "-":
		return Subtract, nil
	default:
		return 0, fmt.Errorf("unknown operator %s", opStr)
	}
}

func parseOpline(opLine string) (Expression, error) {
	var aStr, opStr, bStr string
	_, err := fmt.Sscanf(opLine, "  Operation: new = %s %s %s", &aStr, &opStr, &bStr)
	if err != nil {
		return Expression{}, fmt.Errorf("failed to read operation line: %v", err)
	}

	a, err := parseValue(aStr)
	if err != nil {
		return Expression{}, fmt.Errorf("failed to parse value a: %v", err)
	}

	op, err := parseOperator(opStr)
	if err != nil {
		return Expression{}, fmt.Errorf("failed to parse operator: %v", err)
	}

	b, err := parseValue(bStr)
	if err != nil {
		return Expression{}, fmt.Errorf("failed to parse value b: %v", err)
	}

	return Expression{a: a, operator: op, b: b}, nil
}

func parseInput(r io.Reader) ([]Monkey, error) {
	scanner := bufio.NewScanner(r)

	monkeys := []Monkey{}

	for scanner.Scan() {
		line := scanner.Text()
		// blank line -> just a separator between monkeys
		if line == "" {
			continue
		}

		var monkeyNumber int
		fmt.Sscanf(line, "Monkey %d:", &monkeyNumber)

		if monkeyNumber != len(monkeys) {
			return monkeys, fmt.Errorf("parsing monkeys out of order: expected %d, got %d", len(monkeys), monkeyNumber)
		}

		_, itemLine := scanner.Scan(), scanner.Text()
		items, err := parseItemLine(itemLine)
		if err != nil {
			return monkeys, fmt.Errorf("failed to parse item line: %v", err)
		}

		_, operationLine := scanner.Scan(), scanner.Text()
		operation, err := parseOpline(operationLine)
		if err != nil {
			return monkeys, fmt.Errorf("failed to parse operation line: %v", err)
		}

		_, testLine := scanner.Scan(), scanner.Text()
		var divisibilityTest int
		_, err = fmt.Sscanf(testLine, "  Test: divisible by %d", &divisibilityTest)
		if err != nil {
			return monkeys, fmt.Errorf("failed to parse test line: %v", err)
		}

		_, trueLine := scanner.Scan(), scanner.Text()
		_, falseLine := scanner.Scan(), scanner.Text()
		var trueDest, falseDest int
		_, err = fmt.Sscanf(trueLine, "    If true: throw to monkey %d", &trueDest)
		if err != nil {
			return monkeys, fmt.Errorf("failed to parse true line: %v", err)
		}
		_, err = fmt.Sscanf(falseLine, "    If false: throw to monkey %d", &falseDest)
		if err != nil {
			return monkeys, fmt.Errorf("failed to parse false line: %v", err)
		}

		newMonkey := Monkey{
			items:            items,
			operation:        operation,
			divisibilityTest: divisibilityTest,
			trueDest:         trueDest,
			falseDest:        falseDest,
		}
		fmt.Println("Adding monkey", newMonkey)
		monkeys = append(monkeys, newMonkey)
	}

	return monkeys, nil
}

func Clone(monkeys []Monkey) []Monkey {
	newMonkeys := make([]Monkey, 0, len(monkeys))
	for _, m := range monkeys {
		newItems := make([]int, len(m.items))
		copy(newItems, m.items)
		m.items = newItems
		newMonkeys = append(newMonkeys, m)
	}
	return newMonkeys
}

func part1(monkeys []Monkey) (int, error) {
	inspected := make([]int, len(monkeys))
	for round := 0; round < 20; round++ {
		for i := range monkeys {
			inspected[i] += len(monkeys[i].items)
			if err := monkeys[i].turn(monkeys, false); err != nil {
				return 0, fmt.Errorf("error in round %d: %v", round+1, err)
			}
		}
	}

	sort.Ints(inspected)
	return inspected[len(inspected)-1] * inspected[len(inspected)-2], nil
}

func part2(monkeys []Monkey) (int, error) {
	inspected := make([]int, len(monkeys))
	for round := 0; round < 10000; round++ {
		for i := range monkeys {
			inspected[i] += len(monkeys[i].items)
			if err := monkeys[i].turn(monkeys, true); err != nil {
				return 0, fmt.Errorf("error in round %d: %v", round+1, err)
			}
		}
		if (round+1)%1000 == 0 {
			fmt.Println("Round", round+1, ":", inspected)
		}
	}

	sort.Ints(inspected)
	return inspected[len(inspected)-1] * inspected[len(inspected)-2], nil
}

func run() error {
	file, err := os.Open("input.txt")
	if err != nil {
		return fmt.Errorf("failed to open input.txt: %v", err)
	}
	defer file.Close()

	monkeys, err := parseInput(file)
	if err != nil {
		return fmt.Errorf("failed parsing monkeys: %v", err)
	}

	fmt.Println(monkeys)

	part1, err := part1(Clone(monkeys))
	if err != nil {
		return fmt.Errorf("failed to solve part 1: %v", err)
	}
	fmt.Println("Part 1:", part1)

	part2, err := part2(monkeys)
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
