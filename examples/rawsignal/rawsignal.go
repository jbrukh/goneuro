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
	data := make(chan float64, 512)

    d := goneuro.NewDevice(*serialPort)
    if err := d.ConnectRaw(data); err != nil {
        os.Exit(1)
    }
    println("engaging")
    d.Engage()
	startNanos := time.Now()
	for {
        v := <-data
		fmt.Println(time.Now().Sub(startNanos).Nanoseconds(), v)
	}
}
