//target:tester
package main

import (
    "fmt"
    "bufio"
    "os"
    "gothink"
)

const SERIAL_PORT = "/dev/tty.MindBand"
func main() {
    mindBand, e := os.Open(SERIAL_PORT)
    if e != nil {
        fmt.Println("error:", e)
        os.Exit(1)
    }
    println("connected!")

    reader := bufio.NewReader(mindBand)
    gothink.TGRead(reader)
}
