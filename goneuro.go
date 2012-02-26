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
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// Approx the number of data points to be
// coming in from the device per second
const WINDOW_SIZE        = 512

// MAX_PAYLOAD_LENGTH is the maximum number of
// bytes that can be contained in the payload
// message, not including SYNC, PLENGTH and
// CHECKSUM bytes.
const MAX_PAYLOAD_LENGTH = 169

// protocol symbols
const (
	SYNC   = 0xAA
	EXCODE = 0x55
)

// payload CODE values
const (
	CODE_POOR_SIGNAL    = 0x02 // (0-200, 0 == best)
	CODE_ATTENTION      = 0x04 // (0-100)
	CODE_MEDITATION     = 0x05 // (0-100)
	CODE_BLINK_STRENGTH = 0x16 // (1-255)
	CODE_RAW_VALUE      = 0x80 // 128
	CODE_EEG_POWER      = 0x83 // 8 3-byte unsigned ints, for: delta, theta, low-alpha, hi-alpha, low-beta, hi-beta, low-gamma, mid-gamma
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

// AttentionListener has same characteristics
// as MeditationListener
type AttentionListener func(byte)

// SignalStrengthListener listens for signal
// strength, with 0 being the best and 200
// being no signal; sampled at 1 Hz
type SignalStrengthListener func(byte)

// EEGPowerListener is sampled at 1 Hz and
// provides 8 integers representing the 
// different bands
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
	RawSignal      RawSignalListener
	BlinkStrength  BlinkStrengthListener
	Attention      AttentionListener
	Meditation     MeditationListener
	SignalStrength SignalStrengthListener
	EEGPower       EEGPowerListener
}

// payloadParser decides what to do with the
// payload of a checksum-verified packet
type payloadParser func(*[]byte)

// parseByteStream parses the TG byte stream
func parseByteStream(device io.ReadCloser, pparser payloadParser, conn <-chan bool) {
	reader := bufio.NewReader(device)
	defer device.Close()
	engaged := false
	
	// function that reads the stream
	// one byte at a time
	next := func() byte {
		b, err := reader.ReadByte()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading stream:", err)
			os.Exit(1)
		}
		return b
	}

	for {
		// check for exit
		select {
		case v, ok := <-conn:
			if ok && !v {
				return // disconnect when "false" sent
			} else if ok && v {
				engaged = true
				continue // engage when user sends "true"
			}
		default:
			if !engaged {
				next()
				continue // no parsing until engaged
			}
		}

		// sync up
		if next() != SYNC || next() != SYNC {
			continue
		}
		var pLength byte // payload length
	syncLength: // using a label makes code 2 lines shorter :)
		pLength = next()
		if pLength == SYNC {
			goto syncLength
		}
		if pLength > MAX_PAYLOAD_LENGTH {
			continue
		}

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
		if checksum != stated {
			println("checksum has failed: ", checksum, "expected: ", stated, "(skipping)")
			continue
		}
		if pparser != nil {
			pparser(&payload)
		}
	}
}

// fullPayloadParser delivers a payload parser with the
// given listener
func fullPayloadParser(listener *ThinkGearListener) payloadParser {
	return func(payloadPtr *[]byte) {
		parseFullPayload(payloadPtr, listener)
	}
}

// rawPayloadParser delivers a payload parser that only
// parses raw signal and ignores everything else, and then
// delivers the raw signal on a channel
func rawPayloadParser(output chan<- float64) payloadParser {
	return func(payloadPtr *[]byte) {
		parseRawPayload(payloadPtr, output)
	}
}

// parsePayload will parse the payload buffer and trigger
// the appropriate listeners in the provided listener
// object
func parseFullPayload(payloadPtr *[]byte, listener *ThinkGearListener) {
	payload := *payloadPtr
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
					println("raw signal did not have 2 bytes (skipping)")
					break
				}
			}
			nextRow(4)
		case CODE_EEG_POWER:
			if listener.EEGPower != nil {
				if payload[inx+1] == 24 {
					result := make([]int, 8)
					for i := 0; i < 8; i++ {
						p := inx + 3*i + 2
						result[i] =
							int(payload[p])<<16 | int(payload[p+1])<<8 | int(payload[p+2])
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

// parseRawPayload will parse the payload buffer for
// raw signal only, and deliver that signal on the
// giveb channel
func parseRawPayload(payloadPtr *[]byte, output chan<- float64) {
	payload := *payloadPtr
	inx := 0
	var codeLevel int
	for inx < len(payload) {
		switch payload[inx] {
		case EXCODE:
			// not used in the current protocol
			// but provided here for completeness
			codeLevel++
		case CODE_RAW_VALUE:
			if payload[inx+1] == 2 {
			 	output <- float64(int16(payload[inx+2])<<8 | int16(payload[inx+3]))
			} else {
				println("raw signal did not have 2 bytes (skipping)")
			}
			return // no other kind of data needed
		default:
			return
		}
	}
}

// Device represents a NeuroSky/ThinkGear device
type Device struct {
	conn chan bool
	Port string
	connected bool
	lock *sync.Mutex
}

// NewDevice returns a new Device; this object
// will need to call Connect() (or ConnectRaw()) 
// and Engage() in order to start reacting to data
func NewDevice(serialPort string) *Device {
	return &Device{
		conn: make(chan bool),
		Port: serialPort,
		lock: new(sync.Mutex),
	}
}

// Egage will engage the processing of the byte
// stream. No listeners will be triggered or raw
// data will stream until this call.
// If the device is not connected, then this call
// will have no effect.
func (d *Device) Engage() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if (d.connected) {
		d.conn <- true
	}
}

// Disconnect will disconnect from the device and
// close the serial port. If the device is not
// connected, this call will have no effect.
func (d *Device) Disconnect() {
	d.lock.Lock()                // otherwise, multiple calls here can block forever
	defer d.lock.Unlock()
	if (d.connected) {
		d.conn <- false
		d.connected = false
	}
}

// Connect allows you to initiate a connection with the
// device and to provide an event handler for all the
// different types of data simultaneously.
// 
// More precisely, Connect will initialize the Bluetooth
// conection and will begin receiving data through the
// serial port. However, all this data will be quietly
// dropped until a call to Engage().
//
// If the device is already connected, then this call
// will have no effect.
// 
// The serial port is typically a string of the form
// 
//   /dev/tty.MindBand
//
// or whatever you set up in your systems Bluetooth
// options for the device. Note that the various
// portions of the ThinkGearListener will be triggered
// synchronously to parsing, so it may be desirable
// in certain situations for the user to throw data
// onto channels for serial, asynchronous processing.
// If you do use a channel, make sure that this channel
// is asynchronous, or you can still hold up processing.
func (d *Device) Connect(listener *ThinkGearListener) (err error) {
	d.lock.Lock()        				// multiple connects on the same device will block  
	defer d.lock.Unlock()

	device, err := d.connect()
	if err != nil {
		return
	}
	// start spinning the data stream on another thread 
	// and wait for Engage() call
	go parseByteStream(device, fullPayloadParser(listener), d.conn)
	return
}

// ConnectRaw will first initialize the Bluetooth
// connection and begin receiving data through the
// serial port. However, the data will not be parsed
// until a call to Engage(), at which point the
// devices raw signal will be streamed to the 
// provided channel.
//
// If the device is already connected, then this call
// will have no effect.
func (d *Device) ConnectRaw(output chan<- float64) (err error){
	d.lock.Lock()        				// multiple connects on the same device will block  
	defer d.lock.Unlock()

	device, err := d.connect()
	if err != nil {
		return
	}
	// start spinning the data stream on another thread 
	// and wait for Engage() call
	go parseByteStream(device, rawPayloadParser(output), d.conn)
	return
}

// connect will connect to the serial port and set internal
// state of the Device appropriately. This method probably
// needs to be synchronized externally.
func (d *Device) connect() (device io.ReadCloser, err error) {
	if (d.connected) {
		return nil, errors.New("device is already connected")
	}
	device, err = os.Open(d.Port)
	if err != nil {
		str := fmt.Sprintf("device problem: %s", err)
		return nil, errors.New(str)
	}
	d.connected = true
	println("connected: ", d.Port)
	return
}