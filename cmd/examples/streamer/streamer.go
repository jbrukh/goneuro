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
    _, err := goneuro.Connect(SERIAL_PORT, listener)
    if err != nil {
        fmt.Println(err)
    }

    wait := make(chan bool)
    <-wait
}
