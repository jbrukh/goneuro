//target:tester
package main

import (
    "bufio"
    "os"
    "goneuro"
    //"flag"
    "net"
    "fmt"
)

//const SERIAL_PORT = "/dev/tty.MindWaveMobile-SSPDev"
const SERIAL_PORT = "/dev/tty.MindBand"
const BUF_SIZE = 512
const PORT = "9999"

func main() {
    connect(goneuro.PrintSignal)
}

func waitForTcp() {
    println("listening...")
    l, err := net.Listen("tcp", "localhost:"+PORT)
    if err != nil {
        fmt.Fprintf(os.Stderr, "could not listen: %v\n", err)
        os.Exit(1)
    }
    // wait for connection
    conn, err := l.Accept()
    if err != nil {
        println("could not accept:", err)
    }

    println("got tcp connection...")

    socketWriter := func (first, second byte) {
        println("writing to socket: ", first, second)
        conn.Write([]byte{first, second})
    }

    connect(socketWriter)

}

func connect(consumer goneuro.RawSignalConsumer) {
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
    goneuro.TGRead(reader, consumer)
    mindBand.Close()
}
