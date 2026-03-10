package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type Node struct {
	number int
	x, y   int
}

type Graph struct {
	nodes  []Node
	matrix [][]int
	cycle  []int
}

type Man struct {
	genotype []int
	fitness  int
}

func (g *Graph) addNode(nodeData string) {
	slicesData := strings.Split(nodeData, " ")

	nodeNumber, errNum := strconv.Atoi(slicesData[0])
	if errNum != nil {
		fmt.Println(errNum)
		return
	}

	xPos, errX := strconv.Atoi(slicesData[1])
	if errX != nil {
		fmt.Println(errX)
		return
	}

	yPos, errY := strconv.Atoi(slicesData[2])
	if errY != nil {
		fmt.Println(errY)
		return
	}

	g.nodes = append(g.nodes, Node{number: (nodeNumber - 1), x: xPos, y: yPos})
}

func (g *Graph) initMatrix() {
	g.matrix = make([][]int, len(g.nodes))
	for i := 0; i < len(g.nodes); i++ {
		g.matrix[i] = make([]int, len(g.nodes))
	}

	for y := 0; y < len(g.matrix); y++ {
		for x := 0; x < y; x++ {
			g.matrix[y][x] = 0
			g.matrix[x][y] = calcDistance(g.nodes[x], g.nodes[y])
		}
	}
}

func (g Graph) printMatrix() {
	for y := 0; y < len(g.matrix); y++ {
		fmt.Println(g.matrix[y])
	}
}

func (g Graph) getWeigth(nodeX, nodeY int) int {
	if nodeX < nodeY {
		return g.matrix[nodeX][nodeY]
	}
	return g.matrix[nodeY][nodeX]
}

func (g Graph) getEdge(nodeNumX, nodeNumY int) int {
	if nodeNumX > nodeNumY {
		return g.matrix[nodeNumX][nodeNumY]
	}
	return g.matrix[nodeNumY][nodeNumX]
}

func (g *Graph) modifyEdge(nodeX, nodeY int, create bool) {
	if nodeX == nodeY {
		return
	}

	val := 0
	if create {
		val = 1
	}

	if nodeX > nodeY {
		g.matrix[nodeX][nodeY] = val
	} else {
		g.matrix[nodeY][nodeX] = val
	}
}

func (g *Graph) createEdge(nodeX, nodeY int) {
	g.modifyEdge(nodeX, nodeY, true)
}

func (g *Graph) destroyEdge(nodeX, nodeY int) {
	g.modifyEdge(nodeX, nodeY, false)
}

func (g *Graph) createEdgesByCycle(cycle []int) {
	for v := 0; v < len(cycle); v++ {
		g.createEdge(cycle[v], cycle[(v+1)%len(cycle)])
	}
}

func (g Graph) exportEdgesOfNode(node int) string {
	edges := ""
	for x := 0; x < len(g.matrix[0]); x++ {
		if node == x {
			continue
		} else if g.getEdge(x, node) == 1 {
			edges += strconv.Itoa(x) + " "
		} else if g.getEdge(x, node) > 1 {
			fmt.Println("Coś jest nie tak!!!")
		}
	}
	return edges
}

func (g Graph) exportData() string {
	data := ""
	for v := 0; v < len(g.nodes); v++ {
		data += strconv.Itoa(g.nodes[v].number) + " " + strconv.Itoa(g.nodes[v].x) + " " + strconv.Itoa(g.nodes[v].y) + " "
		data += g.exportEdgesOfNode(g.nodes[v].number)
		data += "\n"
	}
	return data
}

func (g Graph) updateMatrix(parent []int) {
	for v := 1; v < len(g.nodes); v++ {
		g.createEdge(v, parent[v])
	}
}

func (g Graph) calcCycleWeight(cycle []int) int {
	sum := 0
	for v := 0; v < len(cycle); v++ {
		sum += g.getWeigth(cycle[v], cycle[(v+1)%len(cycle)])
	}
	return sum
}

func (g *Graph) clearEdges() {
	for y := 0; y < len(g.matrix); y++ {
		for x := 0; x < y; x++ {
			g.destroyEdge(y, x)
		}
	}
}

func (g *Graph) makeRandomCycle() []int {
	nodesList := make([]int, len(g.matrix))
	for v := 0; v < len(g.matrix); v++ {
		nodesList[v] = v
	}

	rand.Shuffle(
		len(nodesList),
		func(i, j int) { nodesList[i], nodesList[j] = nodesList[j], nodesList[i] },
	)

	g.clearEdges()

	g.createEdgesByCycle(nodesList)
	g.cycle = nodesList

	return nodesList
}

func insertWithSort(population []Man, child Man) []Man {
	newPopulation := make([]Man, len(population)+1)

	if len(population) == 0 {
		newPopulation[0] = child
		return newPopulation
	}

	childAdded := false
	newPopulationIdx := 0

	for oldPopulationIdx := 0; oldPopulationIdx < len(population); oldPopulationIdx++ {
		if !childAdded && child.fitness < population[oldPopulationIdx].fitness {
			newPopulation[newPopulationIdx] = child
			newPopulationIdx++
			childAdded = true
		}
		newPopulation[newPopulationIdx] = population[oldPopulationIdx]
		newPopulationIdx++
	}

	if !childAdded {
		newPopulation[newPopulationIdx] = child
	}

	return newPopulation
}

func insertWithoutSort(population []Man, child Man) []Man {
	if len(population) == 0 {
		return append(population, child)
	}

	newPopulation := make([]Man, len(population))
	copy(newPopulation, population)

	index := rand.Intn(len(newPopulation))

	newPopulation = append(newPopulation[:index+1], newPopulation[index:]...)
	newPopulation[index] = child

	return newPopulation
}

func chooseParents(population []Man, chance int) []Man {
	rand1 := rand.Intn(len(population))
	rand2 := rand1
	for rand1 == rand2 {
		rand2 = rand.Intn(len(population))
	}

	return []Man{population[rand1], population[rand2]}
}

func (g Graph) makeChilds(parents []Man, mutationFrequency int) []Man {
	// Ordered Crossover
	genotypeLength := len(parents[0].genotype)
	childs := make([]Man, 2)

	childs[0] = Man{
		genotype: make([]int, genotypeLength),
		fitness:  0,
	}

	childs[1] = Man{
		genotype: make([]int, genotypeLength),
		fitness:  0,
	}

	randomPoint0 := rand.Intn(genotypeLength-10) + 10
	randomPoint1 := rand.Intn(genotypeLength-randomPoint0) + randomPoint0

	// fmt.Println(randomPoint0, " ", randomPoint1)

	copy(childs[0].genotype[randomPoint0:randomPoint1], parents[1].genotype[randomPoint0:randomPoint1])
	// fmt.Println(childs[0])
	copy(childs[1].genotype[randomPoint0:randomPoint1], parents[0].genotype[randomPoint0:randomPoint1])

	childMiddlePart := childs[0].genotype[randomPoint0:randomPoint1]
	parentGeneIdx := randomPoint1
	childGeneIdx := randomPoint1
	for i := 0; i < genotypeLength; i++ {

		if !numberInside(parents[0].genotype[parentGeneIdx], childMiddlePart) {
			childs[0].genotype[childGeneIdx] = parents[0].genotype[parentGeneIdx]
			childGeneIdx = (childGeneIdx + 1) % genotypeLength
		}

		parentGeneIdx = (parentGeneIdx + 1) % genotypeLength
	}

	childMiddlePart = childs[1].genotype[randomPoint0:randomPoint1]
	parentGeneIdx = randomPoint1
	childGeneIdx = randomPoint1
	for i := 0; i < genotypeLength; i++ {

		if !numberInside(parents[1].genotype[parentGeneIdx], childMiddlePart) {
			childs[1].genotype[childGeneIdx] = parents[1].genotype[parentGeneIdx]
			childGeneIdx = (childGeneIdx + 1) % genotypeLength
		}

		parentGeneIdx = (parentGeneIdx + 1) % genotypeLength
	}

	for _, child := range childs {
		if rand.Intn(mutationFrequency) == 0 { // 1 na mutationFrequency przypadków mutacji
			randPoint1 := rand.Intn(len(child.genotype)-10) + 10
			randPoint2 := rand.Intn(len(child.genotype)-randPoint1) + randPoint1
			child.genotype = reverse(child.genotype, randPoint1, randPoint2)
		}
	}

	childs[0].fitness = g.calcCycleWeight(childs[0].genotype)
	childs[1].fitness = g.calcCycleWeight(childs[1].genotype)

	return childs
}

func (g *Graph) initiateGA(popSize, generationsAmount, maxPopulation, choosingParentsPPB, mutationFrequency int) Man {
	iteration := 0

	siblingsCounter := 0
	parents := make([]Man, 2)

	population := make([]Man, 0)
	for i := 0; i < popSize; i++ {
		rand_cycle := g.makeRandomCycle()
		fitness := g.calcCycleWeight(rand_cycle)
		population = insertWithSort(population, Man{g.makeRandomCycle(), fitness})
	}

	fmt.Println(iteration, ":\t", population[0].fitness)

	for siblingsCounter < 100 && iteration < generationsAmount {

		parents = chooseParents(population, choosingParentsPPB)

		childs := g.makeChilds(parents, mutationFrequency) // Ordered Crossover

		for _, child := range childs {
			population = insertWithSort(population, child)
		}

		if len(population) > maxPopulation {
			population = population[:maxPopulation]
		}

		// if generationsAmount%mutationFrequency == 0 {
		// 	pop := rand.Intn(len(population))
		// 	randPoint1 := rand.Intn(len(population[pop].genotype)-10) + 10
		// 	randPoint2 := rand.Intn(len(population[pop].genotype)-randPoint1) + randPoint1
		// 	population[pop].genotype = reverse(population[pop].genotype, randPoint1, randPoint2)
		// }

		if childs[0].fitness == childs[1].fitness {
			siblingsCounter++
		} else {
			siblingsCounter = 0
		}

		iteration++
		if iteration%50000 == 0 {
			fmt.Println(iteration, ":\t", population[0].fitness)
		}

	}

	return population[0]
}

func (g *Graph) initiateTabuSearch(tabu_size, maxIter, unitTests int) []int {
	// var pathLength float64 = float64(g.calcCycleWeight(g.cycle))

	it := 0

	currentCycle := make([]int, len(g.cycle))
	copy(currentCycle, g.cycle)
	var currentLength int = g.calcCycleWeight(currentCycle)

	var bestCycle []int = make([]int, len(g.cycle))
	copy(bestCycle, currentCycle)
	var bestLength int = currentLength

	newCycle := make([]int, len(g.cycle))
	copy(newCycle, currentCycle)

	tabu_list := make([][]int, 0)
	emptyIt := 0

	fmt.Println(it, " ", currentLength)

	// fmt.Println(currentCycle)

	for emptyIt < maxIter {
		it += 1
		tempCycle, tempLength := g.findBestNeigbour(100, currentCycle, currentLength)

		if inside(tempCycle, tabu_list) {
			// fmt.Println(tempCycle)
			if it%1000 == 0 {
				fmt.Println(it, " ", currentLength)
			}
			emptyIt++
			continue
		}

		// currentCycle = tempCycle
		copy(currentCycle, tempCycle)
		currentLength = tempLength
		// fmt.Println(currentLength)

		if currentLength < bestLength {
			copy(bestCycle, currentCycle)
			bestLength = currentLength
			emptyIt = 0
		} else {
			emptyIt++
		}
		// fmt.Println(tempLength)

		tabu_list = append(tabu_list, currentCycle)

		if len(tabu_list) > tabu_size {
			tabu_list = pop(tabu_list)
		}

		if it%1000 == 0 {
			fmt.Println(it, " ", currentLength)
		}

	}
	fmt.Println(it, " ", currentLength)
	return bestCycle
}

func (g *Graph) findBestNeigbour(tries int, currentCycle []int, currLength int) ([]int, int) {
	var currCycle []int = make([]int, len(currentCycle))
	copy(currCycle, currentCycle)
	bestLength := currLength

	// fmt.Println(currCycle)

	var best []int = make([]int, len(currentCycle))
	copy(best, currCycle)

	var j int
	var i int

	for t := 0; t < tries; t++ {

		// fmt.Println(len(currCycle))

		i = rand.Intn(len(currCycle))
		j = rand.Intn(len(currCycle))

		// currCycle[i], currCycle[j] = currCycle[j], currCycle[i]
		// fmt.Println(currCycle)
		currCycle = reverse(currCycle, i, j)

		d2 := g.calcCycleWeight(currCycle)

		if d2 < bestLength {
			best = currCycle
			bestLength = d2
			// fmt.Println(best)
		}

		currCycle = reverse(currCycle, i, j)
		// currCycle[i], currCycle[j] = currCycle[j], currCycle[i]

	}
	return best, bestLength
}

func pop(slice [][]int) [][]int {
	return slice[1:]
}

func reverse(arr []int, start, end int) []int {
	newArr := make([]int, len(arr))
	copy(newArr, arr)
	for start < end {
		newArr[start], newArr[end] = newArr[end], newArr[start]
		start++
		end--
	}
	return newArr
}

func power(number, exponent int) float64 {
	return math.Pow(float64(number), float64(exponent))
}

func calcDistance(nodeX, nodeY Node) int {
	return int(math.Sqrt((power(nodeX.x-nodeY.x, 2)) + (power(nodeX.y-nodeY.y, 2))))
}

func inside(array []int, doubleArray [][]int) bool {
	if len(doubleArray) == 0 {
		return false
	}
	var wasBrake bool
	for y := 0; y < len(doubleArray); y++ {
		wasBrake = false
		for x := 0; x < len(doubleArray[0]); x++ {
			if array[x] != doubleArray[y][x] {
				wasBrake = true
				break
			}
		}
		if !wasBrake {
			return true
		}
	}
	return false
}

func numberInside(number int, arr []int) bool {
	for _, val := range arr {
		if number == val {
			return true
		}
	}
	return false
}

func save(folder, fileName, data string) {
	path := "./" + folder + "/" + strings.Split(fileName, ".")[0] + ".txt"
	f, err := os.Create(path)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(data)

	if err2 != nil {
		log.Fatal(err2)
	}
}

func getData(path string) map[int]string {
	f, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	data := make(map[int]string)

	i := 0
	for scanner.Scan() {
		data[i] = scanner.Text()
		i++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return data
}

func sum(elements []int) int {
	result := 0
	for _, val := range elements {
		result += val
	}
	return result
}

func cycleToString(cycle []int) string {
	road := ""
	for _, v := range cycle {
		road += strconv.Itoa(v) + " "
	}
	return road
}

func main() {

	files := []string{"xqf131.tsp", "xqg237.tsp", "pma343.tsp", "pka379.tsp", "bcl380.tsp", "pbl395.tsp", "pbk411.tsp", "pbn423.tsp", "pbm436.tsp", "xql662.tsp", "xit1083.tsp", "icw1483.tsp", "djc1785.tsp", "dcb2086.tsp", "pds2566.tsp"}
	var graph Graph

	which_file := 10

	for fileNameIdx := which_file; fileNameIdx < which_file+1; fileNameIdx++ {

		path := "./points/" + files[fileNameIdx]

		data := getData(path)
		graph = Graph{}

		for i := 8; i < len(data)-1; i++ {
			graph.addNode(data[i])
		}
		graph.initMatrix()

		popSize := 3000
		bestMan := graph.initiateGA(popSize, 10000000, popSize, 4, popSize*15)

		road := cycleToString(bestMan.genotype)

		// bestWeight := math.MaxInt
		// bestCycle := make([]int, len(graph.cycle))
		// var avgCycleWeight float64 = 0
		// var tries int = 5

		// for i := 0; i < tries; i++ {
		// 	graph.makeRandomCycle()

		// 	// cycle := graph.initiateAnnealing(5000, 3000, 0.999, 150.0) // 3150, 4000, 0.993, 3300.0 || 4000, 300000, 0.99, 1250.0
		// 	cycle := graph.initiateTabuSearch(1000, 150, 100) // 3150, 4000, 0.993, 3300.0 || 4000, 300000, 0.99, 1250.0

		// 	if graph.calcCycleWeight(cycle) < bestWeight {
		// 		bestCycle = cycle
		// 	}

		// 	avgCycleWeight += float64(graph.calcCycleWeight(cycle))
		// 	fmt.Println(i, " ", graph.calcCycleWeight(cycle))

		// }

		// graph.cycle = bestCycle
		// avgCycleWeight /= float64(tries)
		// fmt.Println("Average: ", avgCycleWeight)
		// road := cycleToString(graph.cycle)
		// fmt.Println(road)

		// save("ex1Data", files[fileNameIdx], graph.exportData())
		save("ex2Data", files[fileNameIdx], road)

	}
}
