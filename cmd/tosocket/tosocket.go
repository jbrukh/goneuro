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
	    println("listening on address: ", addr)
		conn, err := listener.Accept()
	    if err != nil {
	        println("couldn't establish TCP connection: ", err)
	        os.Exit(1)
		}

		// write the raw signal bytes to the socket channel
		println("getting connection to device...")
        // this channel better be asynchronous, or else
        // the loop below is going to stop reading it,
        // the processing will block, and you won't
        // be able to disconnect
		toSocket := make(chan []byte, 1024)
	    handler := &goneuro.ThinkGearListener{
			RawSignal: func(a, b byte) {
			    //fmt.Println("raw:", int16(a)<<8|int16(b))
	            toSocket <- []byte{a,b}
			},
		}

		disconnect, err := goneuro.Connect(*serialPort, handler)
	    if err != nil {
	       println("couldn't connect: ", *serialPort)
           continue
	    }

	    for {
	        val := <-toSocket
	        _, err := conn.Write(val)
            // TODO: check bytes written

            // when tcp connection closes,
            // close the socket and start again
	        if err != nil {
                println("couldn't write to socket anymore...")
                break
	        }
	    }

	    disconnect <- true
        print("cycling... ")
	}
}
