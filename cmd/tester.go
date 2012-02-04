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

func main() {
    listener := &goneuro.ThinkGearListener{
        RawSignal: func(a, b byte) {
            os.Stdout.Write([]byte{a,b})
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
