package isecnet

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestOnlineCommandsRequireProtocolEvidence(t *testing.T) {
	client := NewClient("192.168.1.50", 9009, "123456", time.Second)
	commands := []struct {
		name string
		run  func() error
	}{
		{name: "arm stay", run: func() error {
			_, err := client.ArmPartition(0, "stay")
			return err
		}},
		{name: "disarm partition 1", run: func() error {
			_, err := client.DisarmPartition(1)
			return err
		}},
		{name: "clock", run: func() error {
			_, err := client.SetPanelDateTime(time.Now())
			return err
		}},
		{name: "alarm memory", run: func() error {
			_, err := client.ClearAlarmMemory()
			return err
		}},
		{name: "pgm", run: func() error {
			_, err := client.SetPGM(1, true)
			return err
		}},
	}

	for _, command := range commands {
		t.Run(command.name, func(t *testing.T) {
			if err := command.run(); !errors.Is(err, ErrOnlineCommandUnsupported) {
				t.Fatalf("error = %v, want ErrOnlineCommandUnsupported", err)
			}
		})
	}
}

func TestDisarmPartitionUsesCapturedCommandAndVerifiesStatus(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	errc := make(chan error, 1)
	go serveDisarmPanelFixture(listener, errc)

	host, rawPort, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(host, port, "123456", time.Second)

	status, err := client.DisarmPartition(0)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
	if status.State != "DISARMED" {
		t.Fatalf("state = %s, want DISARMED", status.State)
	}
	if len(status.Partitions) != 1 || status.Partitions[0].Armed {
		t.Fatalf("partition was not verified disarmed: %+v", status.Partitions)
	}
}

func TestArmPartitionUsesCapturedCommandAndVerifiesStatus(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	errc := make(chan error, 1)
	go serveArmPanelFixture(listener, errc)

	host, rawPort, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(host, port, "123456", time.Second)

	status, err := client.ArmPartition(0, "away")
	if err != nil {
		t.Fatal(err)
	}
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
	if status.State != "ARMED" {
		t.Fatalf("state = %s, want ARMED", status.State)
	}
	if len(status.Partitions) != 1 || !status.Partitions[0].Armed {
		t.Fatalf("partition was not verified armed: %+v", status.Partitions)
	}
}

func TestSetZoneBypassUsesCapturedPayloadAndVerifiesStatus(t *testing.T) {
	tests := []struct {
		name           string
		zone           int
		bypassed       bool
		beforeBypassed []int
		wantPayload    []byte
	}{
		{
			name:        "bypass zone 1",
			zone:        1,
			bypassed:    true,
			wantPayload: []byte{0x00, 0x01, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04, 0x00},
		},
		{
			name:           "clear zone 1",
			zone:           1,
			bypassed:       false,
			beforeBypassed: []int{1},
			wantPayload:    []byte{0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04, 0x00},
		},
		{
			name:        "bypass zone 2",
			zone:        2,
			bypassed:    true,
			wantPayload: []byte{0x00, 0x00, 0x01, 0x01, 0x02, 0x00, 0x03, 0x00, 0x04, 0x00},
		},
		{
			name:           "clear zone 2",
			zone:           2,
			bypassed:       false,
			beforeBypassed: []int{2},
			wantPayload:    []byte{0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()

			errc := make(chan error, 1)
			go serveZoneBypassPanelFixture(listener, errc, tt.beforeBypassed, tt.wantPayload, tt.zone, tt.bypassed)

			host, rawPort, err := net.SplitHostPort(listener.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			port, err := strconv.Atoi(rawPort)
			if err != nil {
				t.Fatal(err)
			}
			client := NewClient(host, port, "123456", time.Second)

			status, err := client.SetZoneBypass(tt.zone, tt.bypassed)
			if err != nil {
				t.Fatal(err)
			}
			if err := <-errc; err != nil {
				t.Fatal(err)
			}
			if err := verifyZoneBypass(status, tt.zone, tt.bypassed); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestArmPartitionRejectsUnprovenModeAndPartition(t *testing.T) {
	client := NewClient("192.168.1.50", 9009, "123456", time.Second)

	if _, err := client.ArmPartition(1, "away"); !errors.Is(err, ErrOnlineCommandUnsupported) {
		t.Fatalf("partition error = %v, want ErrOnlineCommandUnsupported", err)
	}
	if _, err := client.ArmPartition(0, "stay"); !errors.Is(err, ErrOnlineCommandUnsupported) {
		t.Fatalf("mode error = %v, want ErrOnlineCommandUnsupported", err)
	}
}

func TestDisarmPartitionRejectsUnprovenPartition(t *testing.T) {
	client := NewClient("192.168.1.50", 9009, "123456", time.Second)

	_, err := client.DisarmPartition(1)

	if err == nil || !errors.Is(err, ErrOnlineCommandUnsupported) {
		t.Fatalf("error = %v, want ErrOnlineCommandUnsupported", err)
	}
}

func TestCheckArmPartitionResponse(t *testing.T) {
	tests := []struct {
		name    string
		frame   Frame
		wantErr string
	}{
		{
			name:  "success",
			frame: Frame{Command: cmdDisarmPartition, Payload: []byte{0xff, 0x91}},
		},
		{
			name:    "unexpected command",
			frame:   Frame{Command: cmdStatus, Payload: []byte{0xff, 0x91}},
			wantErr: "unexpected arm response command 0x0b4a",
		},
		{
			name:    "unexpected payload",
			frame:   Frame{Command: cmdDisarmPartition, Payload: []byte{0xff, 0x90}},
			wantErr: "panel rejected arm command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkArmPartitionResponse(tt.frame)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("checkArmPartitionResponse returned error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestCheckDisarmPartitionResponse(t *testing.T) {
	tests := []struct {
		name    string
		frame   Frame
		wantErr string
	}{
		{
			name:  "success",
			frame: Frame{Command: cmdDisarmPartition, Payload: []byte{0xff, 0x90}},
		},
		{
			name:    "unexpected command",
			frame:   Frame{Command: cmdStatus, Payload: []byte{0xff, 0x90}},
			wantErr: "unexpected disarm response command 0x0b4a",
		},
		{
			name:    "unexpected payload",
			frame:   Frame{Command: cmdDisarmPartition, Payload: []byte{0xff, 0x00}},
			wantErr: "panel rejected disarm command",
		},
		{
			name:    "short payload",
			frame:   Frame{Command: cmdDisarmPartition, Payload: []byte{0xff}},
			wantErr: "unexpected disarm response payload length 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkDisarmPartitionResponse(tt.frame)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("checkDisarmPartitionResponse returned error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestCheckOnlineOKResponse(t *testing.T) {
	tests := []struct {
		name    string
		frame   Frame
		wantErr string
	}{
		{
			name:  "success",
			frame: Frame{Command: cmdOnlineOK},
		},
		{
			name:    "unexpected command",
			frame:   Frame{Command: cmdStatus},
			wantErr: "unexpected zone bypass response command 0x0b4a",
		},
		{
			name:    "unexpected payload",
			frame:   Frame{Command: cmdOnlineOK, Payload: []byte{0x00}},
			wantErr: "unexpected zone bypass response payload length 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkOnlineOKResponse(tt.frame, "zone bypass")
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("checkOnlineOKResponse returned error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func serveArmPanelFixture(listener net.Listener, errc chan<- error) {
	servePartitionControlPanelFixture(listener, errc, []byte{0xff, 0x01}, []byte{0xff, 0x91}, armedStatusPayload())
}

func serveDisarmPanelFixture(listener net.Listener, errc chan<- error) {
	servePartitionControlPanelFixture(listener, errc, []byte{0xff, 0x00}, []byte{0xff, 0x90}, disarmedStatusPayload())
}

func servePartitionControlPanelFixture(listener net.Listener, errc chan<- error, requestPayload []byte, responsePayload []byte, statusPayload []byte) {
	conn, err := listener.Accept()
	if err != nil {
		errc <- err
		return
	}
	defer conn.Close()

	reader := NewClient("", 0, "", time.Second)
	if err := expectCommand(reader, conn, cmdAuth, nil); err != nil {
		errc <- err
		return
	}
	if _, err := conn.Write(encodeFrame(cmdAuth, []byte{0x00})); err != nil {
		errc <- err
		return
	}
	if err := expectCommand(reader, conn, cmdDisarmPartition, requestPayload); err != nil {
		errc <- err
		return
	}
	if _, err := conn.Write(encodeFrame(cmdDisarmPartition, responsePayload)); err != nil {
		errc <- err
		return
	}
	if err := expectCommand(reader, conn, cmdStatus, nil); err != nil {
		errc <- err
		return
	}
	if _, err := conn.Write(encodeFrame(cmdStatus, statusPayload)); err != nil {
		errc <- err
		return
	}
	if err := serveZoneNameReads(reader, conn); err != nil {
		errc <- err
		return
	}
	if err := expectCommand(reader, conn, cmdDisconnect, nil); err != nil {
		errc <- err
		return
	}
	errc <- nil
}

func serveZoneBypassPanelFixture(listener net.Listener, errc chan<- error, beforeBypassed []int, wantPayload []byte, zone int, bypassed bool) {
	conn, err := listener.Accept()
	if err != nil {
		errc <- err
		return
	}
	defer conn.Close()

	reader := NewClient("", 0, "", time.Second)
	if err := expectCommand(reader, conn, cmdAuth, nil); err != nil {
		errc <- err
		return
	}
	if _, err := conn.Write(encodeFrame(cmdAuth, []byte{0x00})); err != nil {
		errc <- err
		return
	}
	if err := expectCommand(reader, conn, cmdStatus, nil); err != nil {
		errc <- err
		return
	}
	if _, err := conn.Write(encodeFrame(cmdStatus, zoneStatusPayload(beforeBypassed...))); err != nil {
		errc <- err
		return
	}
	if err := serveZoneNameReads(reader, conn); err != nil {
		errc <- err
		return
	}
	if err := expectCommand(reader, conn, cmdZoneBypass, wantPayload); err != nil {
		errc <- err
		return
	}
	if _, err := conn.Write(encodeFrame(cmdOnlineOK, nil)); err != nil {
		errc <- err
		return
	}
	if err := expectCommand(reader, conn, cmdStatus, nil); err != nil {
		errc <- err
		return
	}
	afterBypassed := []int{}
	if bypassed {
		afterBypassed = append(afterBypassed, zone)
	}
	if _, err := conn.Write(encodeFrame(cmdStatus, zoneStatusPayload(afterBypassed...))); err != nil {
		errc <- err
		return
	}
	if err := serveZoneNameReads(reader, conn); err != nil {
		errc <- err
		return
	}
	if err := expectCommand(reader, conn, cmdDisconnect, nil); err != nil {
		errc <- err
		return
	}
	errc <- nil
}

func expectCommand(client *Client, conn net.Conn, command uint16, payload []byte) error {
	frame, err := client.readFrame(conn)
	if err != nil {
		return err
	}
	if frame.Command != command {
		return fmt.Errorf("command = 0x%04x, want 0x%04x", frame.Command, command)
	}
	if payload != nil && string(frame.Payload) != string(payload) {
		return fmt.Errorf("payload = % x, want % x", frame.Payload, payload)
	}
	return nil
}

func serveZoneNameReads(client *Client, conn net.Conn) error {
	for start := 0; start < zoneNameMaxZones; start += zoneNameBatchSize {
		if err := expectCommand(client, conn, cmdZoneNames, zoneNameRequestPayload(start)); err != nil {
			return err
		}
		if _, err := conn.Write(encodeFrame(cmdZoneNames, zoneNameFixturePayload(start))); err != nil {
			return err
		}
	}
	return nil
}

func zoneNameFixturePayload(start int) []byte {
	payload := make([]byte, 0, zoneNameBatchSize*zoneNameRecordSize)
	for i := 0; i < zoneNameBatchSize; i++ {
		index := start + i
		payload = append(payload, byte(index))
		name := fmt.Sprintf("Sensor      %02d", index+1)
		if index == 0 {
			name = "IVA DIVISA"
		}
		raw := make([]byte, zoneNameTextSize)
		copy(raw, []byte(name))
		for j := len(name); j < len(raw); j++ {
			raw[j] = ' '
		}
		payload = append(payload, raw...)
	}
	return payload
}

func armedStatusPayload() []byte {
	payload := baseStatusPayload()
	payload[20] = 0x60
	payload[21] = 0x81
	return payload
}

func disarmedStatusPayload() []byte {
	payload := baseStatusPayload()
	payload[20] = 0x00
	payload[21] = 0x80
	return payload
}

func baseStatusPayload() []byte {
	payload := make([]byte, 143)
	payload[0] = 0x8a
	payload[1], payload[2], payload[3] = 2, 0, 2
	payload[12] = 0x1f
	copy(payload[64:70], []byte{0x22, 0x06, 0x26, 0x13, 0x08, 0x39})
	payload[134] = 4
	return payload
}

func zoneStatusPayload(bypassedZones ...int) []byte {
	payload := disarmedStatusPayload()
	for _, zone := range bypassedZones {
		payload[54+(zone-1)/8] |= byte(1 << ((zone - 1) % 8))
	}
	return payload
}
