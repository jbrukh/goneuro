package main

import (
	"goneuro"
	"fmt"
	"os"
	"time"
)

const SERIAL_PORT = "/dev/tty.MindBand"

func signalHandler(disconnect chan<- bool, ch chan byte) {
	println("sleeping 10 seconds...")
	time.Sleep(2 * 1e9)
	println("disconnecting...")
	disconnect <- true
	println("closing ch")
	close(ch)
}

func main() {
	// collect meditation on a channel
	ch := make(chan byte, 512)
	listener := &goneuro.ThinkGearListener{
		Meditation: func(b byte) {
			ch <- b
		},
	}

	// open the device
	disconnect, err := goneuro.Connect(SERIAL_PORT, listener)
	if err != nil {
		println("couldn't connect to device")
		os.Exit(1)
	}

	// listen for Ctrl-C
	go signalHandler(disconnect, ch)

	// wait for and print values indefinitely 
	for {
		b, ok := <-ch
		if !ok {
			println("will die in 2")
			time.Sleep(1e9 * 2)
			break // we are done
		}
		fmt.Println("Meditation: ", b)
	}
}
