package main

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

var ip = "localhost"
var port = "2222"

var address = ip + ":" + port
var origin = "http://" + ip + "/"

var inputA = "ws://" + address + "/inputA"
var inputB = "ws://" + address + "/inputB"

func createInputHandler(inputName string) func(*websocket.Conn) {
	inputHandler := func(ws *websocket.Conn) {

		for {
			var message string
			err := websocket.Message.Receive(ws, &message)
			if err != nil {
				fmt.Printf("Received error '%v'\n", err)
			} else {
				fmt.Printf("Received message '%s' at input [%s]\n", message, inputName)
			}
		}
	}
	return inputHandler
}

func outputSender(ws *websocket.Conn, outChannel chan string, done chan bool) {
	count := 0
	max := 5

	for {

		message := <-outChannel

		//message := "Don't kill the messenger"
		websocket.Message.Send(ws, message)
		//time.Sleep(10 * time.Millisecond)
		fmt.Printf("    Send message '%s' to server\n", message)

		count++
		if count > max {
			done <- true
		}
	}
}

//           *CELL1*					   *CELL2*
//       +------------+					+------------+
//	  -->o /inA  outA o---->  ....  --->o /inA  outA o---->
//	     |            | 				|            |
//	  -->o /inB  outB o---->  ....  --->o /inB  outB o---->
//       +------------+					+------------+
//         [IP]:[PORT]		webSocket	  [IP]:[PORT]
//
//    server        client			 server        client
//

func main() {

	done := make(chan bool)

	outChannel := make(chan string)

	createServer := func() {

		onInputA := createInputHandler("inputA")
		onInputB := createInputHandler("inputB")

		// Register Handlers
		http.Handle("/inputA", websocket.Handler(onInputA))
		http.Handle("/inputB", websocket.Handler(onInputB))

		// Listen to clients connecting
		err := http.ListenAndServe(address, nil)

		// Show Error
		if err != nil {
			panic("ListenAndServe: " + err.Error())
		}

	}

	createClient := func(url string) {
		ws, err := websocket.Dial(url, "", origin)

		if err != nil {
			fmt.Println(err)
			return
		}

		outputSender(ws, outChannel, done)
	}

	go createServer()       // with two input channels [A,B]
	go createClient(inputA) // Connect to server input channel A
	go createClient(inputB) // Connect to server input channel B

	// Send some data out
	go func() {
		for {
			outChannel <- fmt.Sprintf("%s", time.Now().Format(time.RFC850))
			time.Sleep(500 * time.Millisecond)
		}
	}()

	<-done
	fmt.Println("Done ....")
}
