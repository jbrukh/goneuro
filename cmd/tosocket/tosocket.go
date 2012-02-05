//target:tosocket
package main

import (
    "net"
    "flag"
	"os"
	"goneuro"
	"fmt"
)

var tcpPort *string = flag.String("port", "9999", "port for the socket")
var serialPort *string = flag.String("serial", "/dev/tty.MindBand", "serial port for the device")

func main() {
    flag.Parse()

    // start a socket listener
    addr := "localhost:"+*tcpPort

    for {
	    listener, err := net.Listen("tcp", addr)
	    if err != nil {
	        println("couldn't listen: ", err)
            os.Exit(1)
		}

	    // wait for an incoming connection
	    println("waiting for incoming connection...")
		conn, err := listener.Accept()
	    if err != nil {
	        println("couldn't establish connection: ", err)
	        os.Exit(1)
		}

		// write the raw signal bytes to the socket channel
		println("getting connection to device...")
		toSocket := make(chan []byte)
	    handler := &goneuro.ThinkGearListener{
			RawSignal: func(a, b byte) {
			    fmt.Println("raw:", int16(a)<<8|int16(b))
	            toSocket <- []byte{a,b}
			},
		}

		disconnect, err := goneuro.Connect(*serialPort, handler)
	    if err != nil {
	       println("couldn't connect: ", *serialPort) 
	    }

	    wait := make(chan bool)

        // write the incoming data to the socket
	    go func() {
	        for {
	            val := <-toSocket
	            _, err := conn.Write(val)
                // when tcp connection closes,
                // close the socket and start again
	            if err != nil {
	                wait <- true
                    break
	            }
	        }
	    }()

        <-wait
        println("cycling...")
	    disconnect <- true
        conn.Close()
	}
}
