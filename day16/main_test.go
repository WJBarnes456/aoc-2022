package main

import (
	"reflect"
	"testing"
)

func TestAllMoveCombinations(t *testing.T) {
	a := []Move{{"A", nil}}
	b_str := "B"
	ab := []Move{{"A", nil}, {b_str, &b_str}}
	if combos := allMoveCombinations([][]Move{a}); !reflect.DeepEqual(combos, [][]Move{a}) {
		t.Errorf("failed to get combinations of one move for one agent")
	}

	ab_expected := [][]Move{
		{{"A", nil}},
		{{b_str, &b_str}},
	}
	if combos := allMoveCombinations([][]Move{ab}); !reflect.DeepEqual(combos, ab_expected) {
		t.Errorf("failed to get combinations of two moves for one agent")
	}

	a_ab_expected := [][]Move{
		{{"A", nil}, {"A", nil}},
		{{"A", nil}, {b_str, &b_str}},
	}
	if combos := allMoveCombinations([][]Move{a, ab}); !reflect.DeepEqual(combos, a_ab_expected) {
		t.Errorf("failed to get combinations of (one, two) moves for two agents")
	}

	ab_ab_expected := [][]Move{
		{{"A", nil}, {"A", nil}},
		{{"A", nil}, {b_str, &b_str}},
		{{b_str, &b_str}, {"A", nil}},
		{{b_str, &b_str}, {b_str, &b_str}},
	}

	if combos := allMoveCombinations([][]Move{ab, ab}); !reflect.DeepEqual(combos, ab_ab_expected) {
		t.Errorf("failed to get combinations of (two, two) moves for 2 agents")
	}
}

func TestAddOpenValves(t *testing.T) {
	valves := map[string]*Valve{
		"AA": {
			"AA", 10, nil,
		},
		"AB": {
			"AB", 9, nil,
		},
		"AC": {
			"AC", 10, nil,
		},
		"AD": {
			"AD", 10, nil,
		},
		"AE": {
			"AE", 10, nil,
		},
	}
	expectedAB := map[string]*Valve{
		"AB": valves["AB"],
	}
	aa_str := "AA"
	ab_str := "AB"
	if newValves := addOpenValves(valves, map[string]*Valve{}, []*string{&ab_str, nil}); !reflect.DeepEqual(newValves, expectedAB) {
		t.Errorf("failed to add AB to empty values")
	}

	expected_AAAB := map[string]*Valve{
		"AA": valves["AA"],
		"AB": valves["AB"],
	}
	if newValves := addOpenValves(valves, map[string]*Valve{}, []*string{nil, &aa_str, nil, &ab_str}); !reflect.DeepEqual(newValves, expected_AAAB) {
		t.Errorf("failed to add AA and AB to empty values")
	}

	if newValves := addOpenValves(valves, expectedAB, []*string{nil, nil, &aa_str, nil, &ab_str}); !reflect.DeepEqual(newValves, expected_AAAB) {
		t.Errorf("failed to add AA and AB to AB to get AAAB")
	}

	if newValves := addOpenValves(valves, expected_AAAB, []*string{nil}); !reflect.DeepEqual(newValves, expected_AAAB) {
		t.Errorf("failed to add nothing to existing open valves")
	}
}
