package main

import (
	"strconv"

	"github.com/mysheep/cell"
	"github.com/mysheep/playground/Cell/example1"
	"github.com/mysheep/playground/Cell/example2"
)

func main() {

	done := make(chan bool)
	waitUntilDone := func() { <-done }

	/*
		Available commands
	*/
	cmds := map[string]func([]string){
		"q": func(params []string) { done <- true },
		"ex": func(params []string) {
			if len(params) == 0 {
				return
			}
			nr, err := strconv.Atoi(params[0])
			if err != nil {
				return
			}
			if nr == 1 {
				example1.Run()
			}
			if nr == 2 {
				spec2 := example2.Spec{
					Size:           32,
					FolderTemplate: "./images/Some-Characters/%d/",
				}
				example2.Run(spec2)
			}
		},
	}

	go cell.Console(cmds)
	waitUntilDone()
}
