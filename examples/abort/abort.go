package main

import (
	"github.com/jbrukh/goneuro"
	"fmt"
	"os"
	"time"
    "flag"
)

const DEFAULT_PORT = "/dev/tty.MindBand2"
var serialPort *string = flag.String("port", DEFAULT_PORT, "the serial port for the device")

func init() {
    flag.Parse()
}

func signalHandler(d *goneuro.Device, data chan byte) {
	println("sleeping 10 seconds...")
	time.Sleep(10 * time.Second)
	println("disconnecting...")
	d.Disconnect()
	println("closing ch")
	close(data)
}

func main() {
	// collect meditation on a channel
	data := make(chan byte, 512)
	listener := &goneuro.ThinkGearListener{
		Meditation: func(b byte) {
			data <- b
		},
	}
    println("getting device and connecting...")
    d := goneuro.NewDevice(*serialPort)
    if err := d.Connect(listener); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    d.Engage()

	// listen for Ctrl-C
	go signalHandler(d, data)

	// wait for and print values indefinitely 
	for {
		b, ok := <-data
		if !ok {
			println("will die in 5")
			time.Sleep(1e9 * 5)
			break // we are done
		}
		fmt.Println("Meditation: ", b)
	}
}
