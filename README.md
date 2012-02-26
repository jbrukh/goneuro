## NeuroSky/ThinkGear Go Driver

Parser for the NeuroSky ThinkGear protocol used by NeuroSky devices.

Copyright (c) 2012. Jake Brukhman. (jbrukh@gmail.com).
All rights reserved.  See the LICENSE file for BSD-style
license.

------------

### Installation

You need at least the following weekly version of Go:

    go version weekly.2012-02-07 +52ba9506bd99

You can then use the 'go' command to obtain the package:

    $ go get github.com/jbrukh/goneuro

To install the package and all of the executables, use:

    $ go install -v github.com/jbrukh/goneuro/...
    $ ls $GOPATH/bin

### Documentation

See the [gopkgdocs](http://gopkgdoc.appspot.com/pkg/github.com/jbrukh/goneuro).

------------

### Notes

NeuroSky devices stream several kinds of data. First, the device
delivers "raw signal" at approximately 512 Hz and processed "black box"
data at 1 Hz.  Using `goneuro`, you can connect to both kinds of
data from the device.

The first step is pairing your device with your Bluetooth interface
and giving it a serial port. On the Mac, you can do this without any
extra hardware, as a port is built-in. Using the Bluetooth preferences,
you can set the serial port manually.  It will usually look something
like this:

    /dev/tty.JakeMindBand

First, you must obtain a `Device` and connect to it.

    d := goneuro.NewDevice("/dev/tty.JakeMindBand")

There are two ways of streaming data with `goneuro`. If you wish to
pick and choose your black box data, you will need a `ThinkGearListener`
as follows:

	listener := &goneuro.ThinkGearListener{
       Meditation: func(b byte) {
          	// do something
       },
	   Attention: func(b byte) {
	        // do something else
	   },
	}

You can then connect to the device as follows:

    if err := d.Connect(listener); err != nil {
		fmt.Println("could not connect: ", err)
		os.Exit(1)
	}
	
Alternatively, you may be interesting in only a raw data signal coming
from the device. In that case, you should create a channel for receiving
the raw data.

    data := make(chan float64, 512)

Note that the channel is asynchronous, or else you may hold up
processing.  All parsing of the data stream is done serially to calling
back listeners and placing the raw signal on the channel.

You can then connect as follows

    if err := d.ConnectRaw(data); err != nil {
	    fmt.Println("could not connect: ", err)
	    os.Exit(1)
    }

and raw signal will be delivered to the given channel, when
processing starts.

This brings us to the last and final point: in order for `goneuro`
to actually start processing the data and delivering it to listeners
and raw signal channels, you must call:

    d.Engage()

If you wish to disconnect from the device gracefully without turning
off your program, you can call:

    d.Disconnect()



}
