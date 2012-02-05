package main

import (
    "goneuro"
    "fmt"
    "os"
    "os/signal"
    "strings"
    "time"
)

const SERIAL_PORT = "/dev/tty.MindBand"

func signalHandler(disconnect chan<- bool) {
    for {
        sig := <-signal.Incoming
        if strings.HasPrefix(sig.String(), "SIGINT") {
            // ctrl-C
            disconnect <- true
            time.Sleep(1000000)
            os.Exit(0)
        }
    }
}

func main() {
    // collect meditation on a channel
    ch := make(chan byte)
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
    go signalHandler(disconnect)

    // wait for and print values indefinitely 
    for {
        b := <-ch
        fmt.Println("Meditation: ", b)
    }
}



