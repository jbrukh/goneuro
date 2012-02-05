//target:goneuro
package goneuro

// 
// A stream parser for the ThinkGear protocol
// for NeuroSky consumer EEG devices.
//
// Author: Jake Brukhman <jbrukh@gmail.com>
//

import (
    "bufio"
    "fmt"
    "os"
)

//
// Data:
//
//  POOR_SIGNAL      (0-200, 0 == best)
//  ATTENTION        (0-100)
//  MEDITATION       (0-100)
//  RAW Wave         16-bit value (2 bytes)
//  ASIC_EEG_POWER   8 3-byte unsigned ints, for: delta, theta, low-alpha, hi-alpha, low-beta, hi-beta, low-gamma, mid-gamma
//  Blink Strength   (1-255)

const MAX_PAYLOAD_LENGTH = 169

// protocol symbols
const (
    SYNC = 0xAA
    EXCODE = 0x55
)

// payload CODE values
const (
    CODE_POOR_SIGNAL    = 0x02
    CODE_ATTENTION      = 0x04
    CODE_MEDITATION     = 0x05
    CODE_BLINK_STRENGTH = 0x16
    CODE_RAW_VALUE      = 0x80 // 128
    CODE_EEG_POWER      = 0x83
)

// RawSignalListener listens for the raw data
// signal that comes in at 512Hz; the format
// of the signal is two bytes that need to
// be concatenated as follows:
//
//    int16(first)<<8 | int16(second)
//
type RawSignalListener func(byte, byte)

// BlinkStrengthListener listens for the blinl
// signal and returns a value in the range
// 1-255 corresponding to the blink; this
// value is sampled on demand and appears in
// the 1Hz messages
type BlinkStrengthListener func(byte)

// MediationListener listens to the meditation
// value, which is a value between 0-100 and
// is sampled at 1Hz
type MeditationListener func(byte)

// same as Meditation value
type AttentionListener func(byte)

// SignalStrengthListener listens for signal
// strength, with 0 being the best and 200
// being no signal; sampled at 1 Hz
type SignalStrengthListener func(byte)

// EEGPowerListener is sampled at 1 Hz and
// provides 
type EEGPowerListener func(int, int, int, int, int, int, int, int)

// ThinkGearListener will listen to different
// messages from the device; you can instantiate
// only those listeners that you wish to trigger,
// for example as follows:
//
//   l := ThinkGearListener{
//      RawSignal: func(a, b byte) {
//          ...
//      },
//   }
//
type ThinkGearListener struct {
    RawSignal RawSignalListener
    BlinkStrength BlinkStrengthListener
    Attention AttentionListener
    Meditation MeditationListener
    SignalStrength SignalStrengthListener
    EEGPower EEGPowerListener
}

// Connect to the device over the serial port
// and start parsing data; the serial port
// is typically a string of the form
// 
//   /dev/tty.MindBand
//
// or whatever you set up in your systems Bluetooth
// options for the device. Note that the various
// portions of the ThinkGearListener will be triggered
// synchronously to parsing, so it may be desirable
// in certain situations for the user to throw data
// onto channels for serial, asynchronous processing.
//
// This method will return a send-only channel
// for the purposes of ceasing the connection. In
// order to close the connection, send true to
// the disconnect channel.
func Connect(serialPort string, listener *ThinkGearListener) (disconnect chan<- bool, err os.Error) {
    device, err := os.Open(serialPort)
    if err != nil {
        str := fmt.Sprintf("device problem: %s", err)
        return nil, os.NewError(str)
    }
    println("connected: ", serialPort)

    // create the disconnect channel
    ch := make(chan bool)

    // go and process this this stream asynchronously
    // until the user sends a signal to disconnect
    go thinkGearParse(device, listener, ch)

    disconnect = ch // cast to send-only
    return
}


// thinkGearParse parses the TG byte stream
func thinkGearParse(device *os.File, listener *ThinkGearListener, disconnect <-chan bool) {
    //var row int

    reader := bufio.NewReader(device)
    defer device.Close()

    // function that reads the stream
    // one byte at a time
    next := func() byte {
        b, err := reader.ReadByte()
        if err != nil {
            fmt.Fprintln(os.Stderr, "error reading stream:", err)
            os.Exit(1)
        }
        //fmt.Fprintf(os.Stderr, "%v\t:%v\n", row, b)
        //if row > 0 {
        //    row++
        //}
        return b
    }

    for {
        // check for exit
        select {
            case v, ok := (<-disconnect):
                println("v, ok:", v, ok)
                if ok && v == true {
                    println("disconnecting from device")
                    return
                }
            default:
        }

        //fmt.Fprintln(os.Stderr, "---------------------------")
        // sync up
        if next() != SYNC || next() != SYNC {
            continue
        }
        var pLength byte    // payload length
        syncLength:         // using a label makes code 2 lines shorter :)
            pLength = next()
            if (pLength == SYNC) {
               goto syncLength
            }
        if pLength > MAX_PAYLOAD_LENGTH {
            continue
        }

        //row = 1

        // read the entire payload
        payload := make([]byte, 0, pLength)
        count := int(pLength)
        var checksum byte

        // populate the payload slice
        for count > 0 {
            b := next()
            payload = append(payload, b)
            checksum += b
            count--
        }

        // and check it
        checksum = 0xFF &^ checksum

        stated := next()
        //row = 0
        if checksum != stated {
            println("checksum has failed: ", checksum, "expected: ", stated)
            continue
        }
        parsePayload(payload, listener)
    }
    println("done with parsing")
}

// parsePayload will parse the payload buffer and trigger
// the appropriate listeners in the provided listener
// object
func parsePayload(payload []byte, listener *ThinkGearListener) {
    inx := 0
    var codeLevel int
    nextRow := func(k int) {
        inx += k
        codeLevel = 0
    }
    for inx < len(payload) {
        switch payload[inx] {
            case EXCODE:
                // not used in the current protocol
                // but provided here for completeness
                codeLevel++
            case CODE_POOR_SIGNAL:
                if listener.SignalStrength != nil {
                    listener.SignalStrength(payload[inx+1])
                }
                nextRow(2)
            case CODE_ATTENTION:
                if listener.Attention != nil {
                    listener.Attention(payload[inx+1])
                }
                nextRow(2)
            case CODE_MEDITATION:
                if listener.Meditation != nil {
                    listener.Meditation(payload[inx+1])
                }
                nextRow(2)
            case CODE_BLINK_STRENGTH:
                if listener.BlinkStrength != nil {
                    listener.BlinkStrength(payload[inx+1])
                }
                nextRow(2)
            case CODE_RAW_VALUE:
                if listener.RawSignal != nil {
                    if payload[inx+1] == 2 {
                        // get the data
                        listener.RawSignal(payload[inx+2], payload[inx+3])
                    } else {
                        println("raw signal did not have 2 bytes")
                        break
                    }
                }
                nextRow(4)
            case CODE_EEG_POWER:
                if listener.EEGPower != nil {
                    if payload[inx+1] == 24 {
                        result := make([]int, 8)
                        for i := 0; i < 8; i++ {
                            p := inx+3*i+2
                            result[i] =
                              int(payload[p])<<16|int(payload[p+1])<<8|int(payload[p+2])
                        }
                        listener.EEGPower(result[0], result[1], result[2], result[3],
                                result[4], result[5], result[6], result[7])
                    }
                }
                nextRow(26) // the CODE, the VLENGTH, and 24 bytes
            default:
                fmt.Fprintln(os.Stderr, "could not process payload:", payload)
                break
        }
    }
}
