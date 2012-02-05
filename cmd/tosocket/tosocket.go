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

	// write the raw signal bytes to the socket 
	println("getting connection to device...")
	handler := &goneuro.ThinkGearListener{
		RawSignal: func(a, b byte) {
		    fmt.Println("raw:", int16(a)<<8|int16(b))
			conn.Write([]byte{a,b})
		},
	}
		
	goneuro.Connect(*serialPort, handler)
}
