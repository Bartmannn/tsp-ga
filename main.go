package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Node struct {
	ID int
	X  int
	Y  int
}

type Candidate struct {
	Route  []int
	Length int
}

const (
	inputFileName   = "xqf131.tsp" // selected TSPLIB instance file
	populationSize  = 250          // number of candidate routes in one generation
	generations     = 1200         // maximum number of generations
	eliteCount      = 12           // number of best routes copied to the next generation
	tournamentSize  = 5            // number of candidates compared during parent selection
	mutationRate    = 0.22         // mutation chance per child, in range 0.0-1.0
	stagnationLimit = 300          // stop after this many generations without improvement
)

func main() {
	inputPath := filepath.Join("points", inputFileName)
	outputDir := filepath.Join("output", strings.TrimSuffix(inputFileName, ".tsp"))

	name, nodes, err := parseTSPLIB(inputPath)
	if err != nil {
		fail(err)
	}

	distances := buildDistanceMatrix(nodes)
	best := solveWithGA(nodes, distances)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fail(err)
	}

	routeIDs := routeToNodeIDs(best.Route, nodes)
	routePath := filepath.Join(outputDir, "route.txt")
	if err := os.WriteFile(routePath, []byte(routeToString(routeIDs)), 0o644); err != nil {
		fail(err)
	}

	fmt.Printf("Instance: %s\n", name)
	fmt.Printf("Nodes: %d\n", len(nodes))
	fmt.Printf("Best length: %d\n", best.Length)
	fmt.Printf("Route saved to: %s\n", routePath)
	fmt.Printf("Route: %s", routeToString(routeIDs))
}

func parseTSPLIB(path string) (string, []Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	nodes := make([]Node, 0)
	inSection := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "EOF" {
			break
		}
		if line == "NODE_COORD_SECTION" {
			inSection = true
			continue
		}
		if !inSection {
			if strings.HasPrefix(line, "NAME") && strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				name = strings.TrimSpace(parts[1])
			}
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			return "", nil, fmt.Errorf("invalid coordinate row: %q", line)
		}

		id, err := strconv.Atoi(fields[0])
		if err != nil {
			return "", nil, fmt.Errorf("invalid node id in %q: %w", line, err)
		}
		x, err := strconv.Atoi(fields[1])
		if err != nil {
			return "", nil, fmt.Errorf("invalid x value in %q: %w", line, err)
		}
		y, err := strconv.Atoi(fields[2])
		if err != nil {
			return "", nil, fmt.Errorf("invalid y value in %q: %w", line, err)
		}

		nodes = append(nodes, Node{ID: id, X: x, Y: y})
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}
	if len(nodes) < 3 {
		return "", nil, fmt.Errorf("instance must contain at least 3 nodes")
	}

	return name, nodes, nil
}

func buildDistanceMatrix(nodes []Node) [][]int {
	distances := make([][]int, len(nodes))
	for i := range nodes {
		distances[i] = make([]int, len(nodes))
	}
	for i := range nodes {
		for j := i + 1; j < len(nodes); j++ {
			distance := roundedEuclidean(nodes[i], nodes[j])
			distances[i][j] = distance
			distances[j][i] = distance
		}
	}
	return distances
}

func roundedEuclidean(a, b Node) int {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	return int(math.Round(math.Hypot(dx, dy)))
}

func solveWithGA(nodes []Node, distances [][]int) Candidate {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	population := make([]Candidate, populationSize)

	for i := 0; i < populationSize; i++ {
		route := rng.Perm(len(nodes))
		population[i] = Candidate{
			Route:  route,
			Length: routeLength(route, distances),
		}
	}

	best := bestCandidate(population)
	stagnation := 0

	for generation := 0; generation < generations; generation++ {
		sort.Slice(population, func(i, j int) bool {
			return population[i].Length < population[j].Length
		})

		nextPopulation := make([]Candidate, 0, populationSize)
		for i := 0; i < eliteCount && i < len(population); i++ {
			nextPopulation = append(nextPopulation, cloneCandidate(population[i]))
		}

		for len(nextPopulation) < populationSize {
			parentA := tournament(population, tournamentSize, rng)
			parentB := tournament(population, tournamentSize, rng)

			childA, childB := orderedCrossover(parentA.Route, parentB.Route, rng)
			mutate(childA, mutationRate, rng)
			mutate(childB, mutationRate, rng)

			nextPopulation = append(nextPopulation, Candidate{
				Route:  childA,
				Length: routeLength(childA, distances),
			})
			if len(nextPopulation) < populationSize {
				nextPopulation = append(nextPopulation, Candidate{
					Route:  childB,
					Length: routeLength(childB, distances),
				})
			}
		}

		population = nextPopulation
		currentBest := bestCandidate(population)
		if currentBest.Length < best.Length {
			best = currentBest
			stagnation = 0
		} else {
			stagnation++
		}

		if stagnation >= stagnationLimit {
			break
		}
	}

	return best
}

func bestCandidate(population []Candidate) Candidate {
	best := cloneCandidate(population[0])
	for i := 1; i < len(population); i++ {
		if population[i].Length < best.Length {
			best = cloneCandidate(population[i])
		}
	}
	return best
}

func cloneCandidate(candidate Candidate) Candidate {
	return Candidate{
		Route:  append([]int(nil), candidate.Route...),
		Length: candidate.Length,
	}
}

func tournament(population []Candidate, size int, rng *rand.Rand) Candidate {
	best := population[rng.Intn(len(population))]
	for i := 1; i < size; i++ {
		candidate := population[rng.Intn(len(population))]
		if candidate.Length < best.Length {
			best = candidate
		}
	}
	return cloneCandidate(best)
}

func orderedCrossover(parentA, parentB []int, rng *rand.Rand) ([]int, []int) {
	start := rng.Intn(len(parentA) - 1)
	end := start + 1 + rng.Intn(len(parentA)-start-1)
	return buildChild(parentA, parentB, start, end), buildChild(parentB, parentA, start, end)
}

func buildChild(primary, secondary []int, start, end int) []int {
	child := make([]int, len(primary))
	for i := range child {
		child[i] = -1
	}

	used := make(map[int]struct{}, len(primary))
	for i := start; i <= end; i++ {
		child[i] = primary[i]
		used[primary[i]] = struct{}{}
	}

	insertAt := (end + 1) % len(child)
	for offset := 0; offset < len(secondary); offset++ {
		gene := secondary[(end+1+offset)%len(secondary)]
		if _, exists := used[gene]; exists {
			continue
		}
		child[insertAt] = gene
		insertAt = (insertAt + 1) % len(child)
	}

	return child
}

func mutate(route []int, mutationRate float64, rng *rand.Rand) {
	if rng.Float64() >= mutationRate {
		return
	}
	start := rng.Intn(len(route) - 1)
	end := start + 1 + rng.Intn(len(route)-start-1)
	for start < end {
		route[start], route[end] = route[end], route[start]
		start++
		end--
	}
}

func routeLength(route []int, distances [][]int) int {
	length := 0
	for i := range route {
		next := (i + 1) % len(route)
		length += distances[route[i]][route[next]]
	}
	return length
}

func routeToNodeIDs(route []int, nodes []Node) []int {
	ids := make([]int, len(route))
	for i, index := range route {
		ids[i] = nodes[index].ID
	}
	return ids
}

func routeToString(route []int) string {
	var builder strings.Builder
	for i, nodeID := range route {
		if i > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(strconv.Itoa(nodeID))
	}
	builder.WriteByte('\n')
	return builder.String()
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
