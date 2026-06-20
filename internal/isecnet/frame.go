package isecnet

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	clientID = []byte{0x8f, 0xe0}
	panelID  = []byte{0x00, 0x00}
)

type Frame struct {
	Dst     uint16
	Src     uint16
	Command uint16
	Payload []byte
}

func encodeFrame(command uint16, payload []byte) []byte {
	length := uint16(len(payload) + 2)
	frame := make([]byte, 0, 9+len(payload))
	frame = append(frame, panelID...)
	frame = append(frame, clientID...)
	frame = binary.BigEndian.AppendUint16(frame, length)
	frame = binary.BigEndian.AppendUint16(frame, command)
	frame = append(frame, payload...)
	frame = append(frame, checksum(frame))
	return frame
}

func decodeFrame(data []byte) (Frame, error) {
	if len(data) < 9 {
		return Frame{}, errors.New("frame too short")
	}
	if checksum(data[:len(data)-1]) != data[len(data)-1] {
		return Frame{}, errors.New("invalid checksum")
	}
	length := int(binary.BigEndian.Uint16(data[4:6]))
	if length < 2 {
		return Frame{}, fmt.Errorf("invalid frame length %d", length)
	}
	expected := 6 + length + 1
	if len(data) < expected {
		return Frame{}, fmt.Errorf("truncated frame: got %d want %d", len(data), expected)
	}
	payloadLen := length - 2
	return Frame{
		Dst:     binary.BigEndian.Uint16(data[0:2]),
		Src:     binary.BigEndian.Uint16(data[2:4]),
		Command: binary.BigEndian.Uint16(data[6:8]),
		Payload: append([]byte(nil), data[8:8+payloadLen]...),
	}, nil
}

func checksum(data []byte) byte {
	var out byte
	for _, b := range data {
		out ^= b
	}
	return out ^ 0xff
}

func encodePassword(password string) ([]byte, error) {
	if len(password) != 4 && len(password) != 6 {
		return nil, errors.New("password must have 4 or 6 digits")
	}

	out := make([]byte, 0, 6)
	if len(password) == 4 {
		out = append(out, 0x0a, 0x0a)
	}
	for _, r := range password {
		if r < '0' || r > '9' {
			return nil, errors.New("password must contain only digits")
		}
		digit := byte(r - '0')
		if digit == 0 {
			digit = 0x0a
		}
		out = append(out, digit)
	}
	return out, nil
}
