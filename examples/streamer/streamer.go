package main

import (
	"fmt"
	"github.com/jbrukh/goneuro"
	"time"
    "flag"
    "os"
)

const DEFAULT_PORT = "/dev/tty.MindBand2"
var serialPort *string = flag.String("port", DEFAULT_PORT, "the serial port for the device")

func init() {
    flag.Parse()
}

func main() {
	data := make(chan int16)
	listener := &goneuro.ThinkGearListener{
		RawSignal: func(a, b byte) {
			data <- int16(a)<<8 | int16(b)
		},
	}

    d := goneuro.NewDevice(*serialPort)
    if err := d.Connect(listener); err != nil {
        os.Exit(1)
    }
    println("sleeping 5 seconds")
    time.Sleep(5*time.Second)
    println("engaging")
    d.Engage()
	startNanos := time.Now()
	for {
		fmt.Println(time.Now().Sub(startNanos).Nanoseconds(), <-data)
	}
}
