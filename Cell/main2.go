package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/mysheep/cell"
	"github.com/mysheep/cell/brain"
)

func getCount(bits *[]bool) int {
	count := 0
	for _, bit := range *bits {
		if bit {
			count = count + 1
		}
	}
	return count
}

func getWeights(name string) ([]float64, error) {

	fileName := getFilename(name)

	//	fmt.Println("name", name)

	img, err := getImage(fileName)

	if err != nil {
		return nil, err
	}

	bits, err := getPixels(img)

	if err != nil {
		return nil, err
	}

	count := getCount(&bits)
	weights := make([]float64, len(bits))
	weight := float64(len(bits)) / float64(count)

	//
	// TODO : negative Weight values to REPRESENT NOT
	//

	for i, bit := range bits {
		if bit {
			weights[i] = weight
		}
	}

	return weights, nil

}

func getAllWeights(names []string) [][]float64 {

	wweights := make([][]float64, len(names))

	for j, name := range names {
		fmt.Println(j, ":", name)
		weights, err := getWeights(name)
		if err != nil {
			panic(fmt.Sprint("Count not get weights from file ", name))
		}
		wweights[j] = weights
	}

	return wweights
}

/*
	Create retina cells
*/
func createRetinaCells(retinaCells []*brain.EmitterCell) {
	for i, _ := range retinaCells {
		retinaCells[i] = brain.MakeEmitterCell(fmt.Sprintf("retina%2d", i))
	}
}

/*
	Create objects (recognition) cells
*/
func createObjectCells(objectCells []*brain.MultiCell, files []string, THRESHOLD int) {
	for j, _ := range objectCells {
		objectCells[j] = brain.MakeMultiCell(files[j], THRESHOLD)
	}
}

/*
	Create display cells
*/
func createDisplayCells(displayCells []*brain.DisplayCell, files []string) {
	for j, _ := range displayCells {
		displayCells[j] = brain.MakeDisplayCell(files[j])

	}
}

/*
	Connect object with display cells
*/
func connectObjectWithDisplayCells(objectCells []*brain.MultiCell, displayCells []*brain.DisplayCell) {
	for j, _ := range objectCells {
		brain.ConnectBy(objectCells[j], displayCells[j], float64(1.0))
	}
}

/*
	Connect retina with objects cells
*/
func connectRetinaWithObjectCells(retinaCells []*brain.EmitterCell, objectCells []*brain.MultiCell, wweights [][]float64) {
	for o, _ := range objectCells {

		if math.Mod(float64(o), float64(200)) == 0.0 {
			fmt.Println(fmt.Sprintf("Connect %d of %d", o, len(objectCells)))
		}

		for r, _ := range retinaCells {
			// TODO: MassConnect without append
			weight := wweights[o][r]
			brain.ConnectBy(retinaCells[r], objectCells[o], weight)
		}
	}
}

func main() {

	done := make(chan bool)
	waitUntilDone := func() { <-done }

	//
	// Setup Network
	//

	fmt.Printf("size is set to %d\n", size)

	files, err := getFiles(getFolder())

	if err != nil {
		return
	}

	var countObjects = len(files)
	const THRESHOLD = size*size - 2 // TODO:???

	fmt.Printf("%d objects found\n", countObjects)
	fmt.Printf("Cell threshold is set to %d\n", THRESHOLD)

	retinaCells := make([]*brain.EmitterCell, size*size)
	objectCells := make([]*brain.MultiCell, countObjects)
	displayCells := make([]*brain.DisplayCell, countObjects)

	allWeights := getAllWeights(files)

	createRetinaCells(retinaCells)
	createObjectCells(objectCells, files, THRESHOLD)
	createDisplayCells(displayCells, files)

	connectObjectWithDisplayCells(objectCells, displayCells)
	connectRetinaWithObjectCells(retinaCells, objectCells, allWeights)

	//
	// Console Commands
	//
	cmds := map[string]func([]string){
		"quit": func(params []string) { done <- true },
		"exit": func(params []string) { done <- true },
		"q":    func(params []string) { done <- true },
		"see": func(params []string) {
			i, err := strconv.Atoi(params[0])
			if err == nil {
				fmt.Println("Retina cells see ", "'"+files[i]+"'")
				fmt.Println("Waiting for answer ...")

				// TODO: Make func
				for j, w := range objectCells[i].Weights() {
					if w > 0 {
						retinaCells[j].EmitOne()
					}
				}

				// TODO: Reset
			}
		},
		"ws": func(params []string) {
			i, err := strconv.Atoi(params[0])
			if err == nil {
				fmt.Printf("%v\n", objectCells[i].Weights())
			}
		},
	}

	go cell.Console(cmds)

	// Wait until Done
	//
	waitUntilDone()
	//
	// Wait until Done

	fmt.Println("BYE")
}
