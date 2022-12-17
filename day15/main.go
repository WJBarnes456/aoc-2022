package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
)

type SensorBeacon struct {
	sensorX int
	sensorY int
	beaconX int
	beaconY int
}

type Position struct {
	x int
	y int
}

type Range struct {
	start int
	end   int
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *SensorBeacon) Distance(pointX int, pointY int) int {
	return Abs(s.sensorX-pointX) + Abs(s.sensorY-pointY)
}

func (s *SensorBeacon) DistanceToBeacon() int {
	return s.Distance(s.beaconX, s.beaconY)
}

// Unoccupied returns a range of X values which must be unoccupied for a line at y = yLine
func (s *SensorBeacon) Unoccupied(yLine int) *Range {
	distanceToBeacon := s.DistanceToBeacon()
	distanceToLine := s.Distance(s.sensorX, yLine)

	// no unoccupied space if the line is further away than the beacon
	if distanceToLine > distanceToBeacon {
		return nil
	}

	// line is the same distance or closer than the beacon
	diff := distanceToBeacon - distanceToLine
	return &Range{s.sensorX - diff, s.sensorX + diff}
}

func parseSensorBeacon(s string) (*SensorBeacon, error) {
	var sensorX, sensorY, beaconX, beaconY int
	_, err := fmt.Sscanf(s, "Sensor at x=%d, y=%d: closest beacon is at x=%d, y=%d", &sensorX, &sensorY, &beaconX, &beaconY)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sensorbeacon line: %v", err)
	}
	return &SensorBeacon{sensorX, sensorY, beaconX, beaconY}, nil
}

func parseInput(r io.Reader) ([]SensorBeacon, error) {
	scanner := bufio.NewScanner(r)
	sbs := []SensorBeacon{}
	for scanner.Scan() {
		line := scanner.Text()
		sb, err := parseSensorBeacon(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sb line: %v", err)
		}
		sbs = append(sbs, *sb)
	}
	return sbs, nil
}

// Gets a de-duplicated list of all beacons
func getBeacons(sbs []SensorBeacon) []Position {
	beaconSet := map[Position]struct{}{}
	for _, sb := range sbs {
		beaconSet[Position{sb.beaconX, sb.beaconY}] = struct{}{}
	}

	beacons := make([]Position, 0, len(beaconSet))
	for beacon := range beaconSet {
		beacons = append(beacons, beacon)
	}
	return beacons
}

func findBlocked(sbs []SensorBeacon, lineY int) []Range {
	// Get all the unoccupied areas on that line
	ranges := make([]Range, 0, len(sbs))
	for _, sb := range sbs {
		sbRange := sb.Unoccupied(lineY)
		if sbRange != nil {
			ranges = append(ranges, *sbRange)
		}
	}

	// Combine the unoccupied areas
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].start < ranges[j].start })

	newRanges := make([]Range, 0, len(ranges))
	var curRange *Range
	for _, r := range ranges {
		if curRange == nil {
			curRange = &Range{r.start, r.end}
			continue
		}

		// combine overlapping ranges
		// nb. because we sorted, r.start >= curRange.start is implicitly true
		if r.start <= curRange.end {
			curRange.end = max(curRange.end, r.end)
			continue
		}

		// new range does not overlap, so make that the current one
		newRanges = append(newRanges, *curRange)
		curRange = &Range{r.start, r.end}
	}

	if curRange != nil {
		newRanges = append(newRanges, *curRange)
	}

	return newRanges
}

func part1(sbs []SensorBeacon) int {
	ranges := findBlocked(sbs, 2000000)
	// total up the length of ranges
	total := 0
	for _, r := range ranges {
		total += r.end - r.start + 1
	}

	// subtract any beacons which are actually on that line
	beacons := getBeacons(sbs)
	for _, beacon := range beacons {
		if beacon.y == 2000000 {
			total--
		}
	}
	return total
}

func part2(sbs []SensorBeacon) int {
	result := make(chan int)
	for y := 0; y <= 4000000; y++ {
		go func(lineY int) {
			blocked := findBlocked(sbs, lineY)
			for _, r := range blocked {
				if r.start <= 0 && r.end >= 4000000 {
					return
				}
			}
			// this is where the beacon must be
			x := blocked[0].end + 1
			fmt.Println("Found on y=", lineY, ":", blocked)
			result <- 4000000*x + lineY
		}(y)
	}
	return <-result
}

func run() error {
	file, err := os.Open("input.txt")
	if err != nil {
		return fmt.Errorf("failed to open input file")
	}
	defer file.Close()

	data, err := parseInput(file)
	if err != nil {
		return fmt.Errorf("failed to parse input: %v", err)
	}

	fmt.Println("Part 1:", part1(data))
	fmt.Println("Part 2:", part2(data))

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to solve day 15: %v", err)
		os.Exit(1)
	}
}
