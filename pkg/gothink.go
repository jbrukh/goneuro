//target:gothink
package gothink

import "io"
import "fmt"
import "os"

//
// Data:
//
//  POOR_SIGNAL      (0-200, 0 == best)
//  ATTENTION        (0-100)
//  MEDITATION       (0-100)
//  RAW Wave         16-bit value (2 bytes)
//  ASIC_EEG_POWER   8 3-byte unsigned ints, for: delta, theta, low-alpha, hi-alpha, low-beta, hi-beta, low-gamma, mid-gamma
//  Blink Strength   (1-255)


const SYNC = 0xAA
const EXCODE = 0x55

// in the payload, if the starting byte
// is bigger than the following, then the
// data has the following format:
//
//   [CODE] [VLENGTH] [BYTES]...
//
// where VLENGTH is the number of bytes
// that follows it. This is usually two,
// for the raw signal.
const CODE_THRESH = 0x7F

// payload CODE values
const (
    CODE_POOR_SIGNAL    = 0x02
    CODE_ATTENTION      = 0x04
    CODE_MEDITATION     = 0x05
    CODE_BLINK_STRENGTH = 0x16
    CODE_RAW_VALUE      = 0x80 // 128
)

// parses the TG byte stream
func TGRead(reader io.Reader) {
    // allocate enough space for a
    // single byte
    oneByte := make([]byte, 1)
    next := func() byte {
        _, err := reader.Read(oneByte)
        if err != nil {
            println("error reading stream:", err)
        }
        fmt.Fprintf(os.Stderr, "byte: %v\n", oneByte)
        return oneByte[0]
    }
    for {
        // look for two SYNC bytes
        if next() != SYNC || next() != SYNC {
            continue
        }

        // skip all ensuing SYNC bytes
        var b byte
        for {
            b = next()
            if b != SYNC {
                break
            }
        }
        pLength := int(b)
        if (pLength > 169) { // max payload length would be exceeded
            continue // skip!
        }

        // read the entire payload
        payload := make([]byte, pLength, pLength)
        _, err := reader.Read(payload)
        if err != nil {
            println("error reading stream:", err)
        }

        // and check it
        var checksum byte = 0
        for _, v := range payload {
            checksum += v
        }
        checksum &= 0xFF
        checksum = 0xFF &^ checksum

        if checksum != uint8(next()) {
            // checksum has failed
            println("checksum has failed:", checksum)
            continue
        }

        parsePayload(payload)
    }
}

func parsePayload(payload []byte) {
    if len(payload) < 2 {
        // something is wrong
        return
    }
    // hackiness for now
    if payload[0] == CODE_RAW_VALUE && payload[1] == 2 {
        high := uint16(payload[2])<<8
        high = high | uint16(payload[3])
        sint := int16(high)
        fmt.Println(sint)
    }
    //fmt.Println(payload)
}
