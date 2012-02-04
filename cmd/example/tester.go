//target:tester
package main

import (
    "goneuro"
    "fmt"
)

const SERIAL_PORT = "/dev/tty.MindBand"

// this is a sample program that connects to
// the MindBand and displays some streaming
// data
func main() {
    listener := &goneuro.ThinkGearListener{
        RawSignal: func(a, b byte) {
            fmt.Println("raw: ", int16(a)<<8|int16(b))
        },
        EEGPower: func(delta, theta, lowAlpha, highAlpha, lowBeta, highBeta, lowGamma, midGamma int) {
            fmt.Println(delta, theta, lowAlpha, highAlpha, lowBeta, highBeta, lowGamma, midGamma)
        },
    }
   goneuro.Connect(SERIAL_PORT, listener)
}

