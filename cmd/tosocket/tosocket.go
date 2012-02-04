//target:tosocket
package main

import (
    "fmt"
//    "net"
    "os"
    "bufio"
)

func main() {
    reader, err := bufio.NewReaderSize(os.Stdin, 512)
    if err != nil {
        println("could not read stdin")
    }

    for {
        b, err := reader.ReadByte()
        if err != nil {
            break
        }
        fmt.Println(b)
    }
}
