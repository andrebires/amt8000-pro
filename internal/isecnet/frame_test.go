package isecnet

import (
	"bytes"
	"testing"
)

func TestChecksum(t *testing.T) {
	data := []byte{0x00, 0x00, 0x8f, 0xe0, 0x00, 0x02, 0x0b, 0x4a}
	if got, want := checksum(data), byte(0xd3); got != want {
		t.Fatalf("checksum = %#x, want %#x", got, want)
	}
}

func TestEncodePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     []byte
	}{
		{name: "six digits", password: "878787", want: []byte{8, 7, 8, 7, 8, 7}},
		{name: "four digits padded", password: "1234", want: []byte{0x0a, 0x0a, 1, 2, 3, 4}},
		{name: "zero as contact-id ten", password: "9090", want: []byte{0x0a, 0x0a, 9, 0x0a, 9, 0x0a}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodePassword(tt.password)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("encodePassword = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestEncodeDecodeFrame(t *testing.T) {
	raw := encodeFrame(cmdStatus, nil)
	frame, err := decodeFrame(raw)
	if err != nil {
		t.Fatal(err)
	}
	if frame.Command != cmdStatus {
		t.Fatalf("command = %#x, want %#x", frame.Command, cmdStatus)
	}
	if frame.Dst != 0x0000 || frame.Src != 0x8fe0 {
		t.Fatalf("unexpected ids dst=%#x src=%#x", frame.Dst, frame.Src)
	}
}
