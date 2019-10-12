package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mysheep/cell"
	"github.com/mysheep/cell/brain"
	"github.com/mysheep/cell/integer"
)

func print(ys []int) {
	for _, y := range ys {
		fmt.Print("y:", y)
	}
	fmt.Println()
}

func main() {

	//
	//  created with http://asciiflow.com/
	//
	//  in           out
	//      +------+     +---------+
	// ---->| Add1 o---->| Display |
	//      +------+     +---------+

	//          +-------+     +---------+
	// inX ---->|       | res |         |
	//          |  Add  o---->| Display |
	// inY ---->|       | res |         |
	//          +-------+     +---------+

	inX := make(chan int)
	inY := make(chan int)
	res := make(chan int)

	done := make(chan bool)
	waitUntilDone := func() { <-done }

	//
	// Setup Network
	//

	addXY := func(x, y int) int { return x + y }
	go integer.Lambda2(inX, inY, res, addXY)
	go integer.Display(res)

	//             +---------+
	// ins[0] ---->|         | out +---------+
	//             |  	     |     |		 |
	// ...         |  Aggr.  o---->| Display |
	//             |  	     |     |		 |
	// ins[n] ---->|         | out +---------+
	//             +---------+

	//
	// Aggregation
	//
	N := 5
	ins := make([]chan int, N)
	fin := make(chan int, 10)
	//agg := make(chan int, 10)
	for i := 0; i < N; i++ {
		ins[i] = make(chan int)
	}

	addFn, aggFn := integer.MakeDynAgg(&ins, fin)

	go aggFn()
	go integer.Display(fin)

	var addOneFn = func() {
		newCh := make(chan int)
		addFn(newCh)
	}

	//
	// Cell with weighted Synapses, Body and Axon
	//

	//  synapse
	//  ------> w1  / ----- \
	//             |         |  axon
	//  ------> w2 |  cell   |-------->
	//             |         |
	//  ------> w5  \ ----- /

	S := 5

	bIn := make(chan int, 100) // buffered body input for aggretion of all synopses
	sIns := make([]chan int, S)
	weights := make([]int, S)

	for j := 0; j < S; j++ {
		sIns[j] = make(chan int)
		weights[j] = rand.Intn(7)
		go brain.Synapse(weights[j], sIns[j], bIn)
	}

	A := 1
	axIn := make(chan int)
	axOuts := make([]chan int, A)

	for j := 0; j < A; j++ {
		axOuts[j] = make(chan int)
		go brain.Writer(axOuts[j])
	}
	go brain.Body(bIn, axIn)
	go brain.Axon(axIn, axOuts)

	//
	// Create two cells and connect them
	//

	fmt.Println("Setup network: 1 emitter + 2 multi cells + 1 display")

	cell1 := brain.MakeMultiCell("cell_1")
	cell2 := brain.MakeMultiCell("cell_2")

	//  13
	// -->(cell1)      (cell2)--->

	brain.Connect(cell1, cell2, 7)

	display1 := brain.MakeDisplayCell("display_1")
	brain.Connect(cell2, display1, 1)

	emitter1 := brain.MakeEmitterCell("emitter_1")
	brain.Connect(emitter1, cell1, 13)

	//  13        7
	// -->(cell1)----->(cell2)--->Display

	//
	// Exampel with 3 cells from book Manfred Spitzer
	//

	emitterA := brain.MakeEmitterCell("emitter_A")
	emitterB := brain.MakeEmitterCell("emitter_B")
	emitterC := brain.MakeEmitterCell("emitter_C")

	cellA := brain.MakeMultiCell("cell_A")
	cellB := brain.MakeMultiCell("cell_B")
	cellC := brain.MakeMultiCell("cell_C")

	displayA := brain.MakeDisplayCell("display_A")
	displayB := brain.MakeDisplayCell("display_B")
	displayC := brain.MakeDisplayCell("display_C")

	brain.Connect(emitterA, cellA, 5)
	brain.Connect(emitterA, cellB, 3)
	brain.Connect(emitterA, cellC, -3)

	brain.Connect(emitterB, cellA, -5)
	brain.Connect(emitterB, cellB, 3)
	brain.Connect(emitterB, cellC, 10)

	brain.Connect(emitterC, cellA, 5)
	brain.Connect(emitterC, cellB, 3)
	brain.Connect(emitterC, cellC, -3)

	brain.Connect(cellA, displayA, 1)
	brain.Connect(cellB, displayB, 1)
	brain.Connect(cellC, displayC, 1)

	//
	// Console Commands
	//
	cmds := map[string]func(){
		"quit": func() { done <- true },
		"exit": func() { done <- true },
		"emit": func() { inX <- 1; inY <- 2 },
		"agg": func() {
			for i := 0; i < len(ins); i++ {
				fmt.Println("send", i)
				ins[i] <- i
			}
		},
		"add": func() {
			addOneFn()
			fmt.Println("add", len(ins), "ins")
		},
		"add10": func() {
			N := 10
			for i := 0; i < N; i++ {
				addOneFn()
			}
		},
		"cell": func() {
			for ii := 0; ii < 100; ii++ {
				i := rand.Intn(S)
				sIns[i] <- i
				time.Sleep(50 * time.Millisecond)
			}
		},
		"con": func() {
			for k := 0; k < 10; k++ {
				emitter1.EmitOne()
				time.Sleep(50 * time.Millisecond)
			}
		},
		"ex1": func() {
			emitterA.EmitOne()
			emitterB.EmitOne()
			emitterC.EmitOne()
		},
	}

	go cell.Console(cmds)

	// Wait until Done
	//
	waitUntilDone()
	//
	// Wait until Done
}
