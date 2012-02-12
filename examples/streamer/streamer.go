package main

import (
	"goneuro"
	"fmt"
    "time"
)

const SERIAL_PORT = "/dev/tty.MindBand"

func main() {

    data := make(chan int16)
	listener := &goneuro.ThinkGearListener{
		RawSignal: func(a, b byte) {
			data <- int16(a)<<8 | int16(b)
		},
	}
	_, err := goneuro.Connect(SERIAL_PORT, listener)
	if err != nil {
		fmt.Println(err)
		return
	}

    startNanos := time.Nanoseconds()
    for {
        fmt.Println(time.Nanoseconds()-startNanos, <-data)
    }
}
