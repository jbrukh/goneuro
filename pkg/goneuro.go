//target:goneuro
package goneuro

// 
// A stream parser for the NeuroSky MindBand.
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

type RawSignalListener func(byte, byte)
type BlinkStrengthListener func(byte)
type MeditationListener func(byte)
type AttentionListener func(byte)
type SignalStrengthListener func(byte)
type EEGPowerListener func(int, int, int, int, int, int, int, int)

// ThinkGearListener will listen to different
// messages from the ThinkGearDevice
type ThinkGearListener struct {
    RawSignal RawSignalListener
    BlinkStrength BlinkStrengthListener
    Attention AttentionListener
    Meditation MeditationListener
    SignalStrength SignalStrengthListener
    EEGPower EEGPowerListener
}

// parses the TG byte stream
func ThinkGearRead(reader *bufio.Reader, listener *ThinkGearListener) {
    var row int
    // function that reads the stream
    // one byte at a time
    next := func() byte {
        b, err := reader.ReadByte()
        if err != nil {
            println("error reading stream:", err)
        }
        fmt.Fprintf(os.Stdout, "%v\t:%v\n", row, b)
        if row > 0 {
            row++
        }
        return b
    }

    for {
        fmt.Println("---------------------------")
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

        row = 1

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
        row = 0
        if checksum != stated {
            println("checksum has failed: ", checksum, "expected: ", stated)
            continue
        }
        parsePayload(payload, listener)
    }
}

func parsePayload(payload []byte, listener *ThinkGearListener) {
    inx := 0
    var codeLevel int
    if len(payload) > 4 {
        fmt.Println("long")
    }
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



func PrintSignal(first, second byte) {
    value := uint16(first)<<8 | uint16(second)
    fmt.Println("raw: ", int(value))
}
