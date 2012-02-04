//target:tester
package main

import (
    "bufio"
    "os"
    "goneuro"
    "fmt"
)

//const SERIAL_PORT = "/dev/tty.MindWaveMobile-SSPDev"
const SERIAL_PORT = "/dev/tty.MindBand"
const BUF_SIZE = 512
const PORT = "9999"

func main() {
    listener := &goneuro.ThinkGearListener{
        SignalStrength: func(value byte) {
            fmt.Fprintln(os.Stderr, "Signal strength:", value)
        },
        RawSignal: func(a, b byte) {
            fmt.Fprintln(os.Stderr, int16(a)<<8|int16(b))
        },
    }
    connect(listener)
}

func connect(consumer *goneuro.ThinkGearListener) {
    mindBand, e := os.Open(SERIAL_PORT)
    if e != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", e)
        os.Exit(1)
    }
    println("connected!")

    reader, e := bufio.NewReaderSize(mindBand, BUF_SIZE)
    if e != nil {
        println("error:", e)
    }
    goneuro.ThinkGearRead(reader, consumer)
    mindBand.Close()
}
