package main

import (
    "goneuro"
    "fmt"
)

const SERIAL_PORT = "/dev/tty.MindBand"

func main() {
    listener := &goneuro.ThinkGearListener{
        RawSignal: func(a, b byte) {
            fmt.Println(int16(a)<<8|int16(b))
        },
    }
    goneuro.Connect(SERIAL_PORT, listener)
}



