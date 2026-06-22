package isecnet

import (
	"errors"
	"fmt"
	"net"
	"time"
)

var ErrOnlineCommandUnsupported = errors.New("online command requires proven AMT 8000 Pro protocol evidence")

const cmdDisarmPartition uint16 = 0x401e
const cmdZoneBypass uint16 = 0x401f
const cmdOnlineOK uint16 = 0xf0fe

func (c *Client) ArmPartition(partition int, mode string) (PanelStatus, error) {
	if partition != 0 {
		return PanelStatus{}, fmt.Errorf("arm partition %d: %w", partition, ErrOnlineCommandUnsupported)
	}
	if mode != "" && mode != "away" {
		return PanelStatus{}, fmt.Errorf("arm mode %q: %w", mode, ErrOnlineCommandUnsupported)
	}
	conn, err := c.connectAndAuth()
	if err != nil {
		return PanelStatus{}, err
	}
	defer conn.Close()
	defer func() {
		_ = c.writeFrame(conn, cmdDisconnect, nil)
	}()

	status, err := c.armPartition0OnConn(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	return status, nil
}

func (c *Client) DisarmPartition(partition int) (PanelStatus, error) {
	if partition != 0 {
		return PanelStatus{}, fmt.Errorf("disarm partition %d: %w", partition, ErrOnlineCommandUnsupported)
	}
	conn, err := c.connectAndAuth()
	if err != nil {
		return PanelStatus{}, err
	}
	defer conn.Close()
	defer func() {
		_ = c.writeFrame(conn, cmdDisconnect, nil)
	}()

	status, err := c.disarmPartition0OnConn(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	return status, nil
}

func (c *Client) SetPanelDateTime(value time.Time) (PanelStatus, error) {
	return PanelStatus{}, unsupportedOnlineCommand("sync panel clock")
}

func (c *Client) SetZoneBypass(zone int, bypassed bool) (PanelStatus, error) {
	if zone < 1 || zone > 64 {
		return PanelStatus{}, fmt.Errorf("zone %d out of range", zone)
	}
	conn, err := c.connectAndAuth()
	if err != nil {
		return PanelStatus{}, err
	}
	defer conn.Close()
	defer func() {
		_ = c.writeFrame(conn, cmdDisconnect, nil)
	}()

	status, err := c.setZoneBypassOnConn(conn, zone, bypassed)
	if err != nil {
		return PanelStatus{}, err
	}
	return status, nil
}

func (c *Client) ClearAlarmMemory() (PanelStatus, error) {
	return PanelStatus{}, unsupportedOnlineCommand("clear alarm memory")
}

func (c *Client) SetPGM(pgm int, active bool) (PanelStatus, error) {
	return PanelStatus{}, unsupportedOnlineCommand("set PGM")
}

func unsupportedOnlineCommand(action string) error {
	return fmt.Errorf("%s: %w", action, ErrOnlineCommandUnsupported)
}

func (c *Client) armPartition0OnConn(conn net.Conn) (PanelStatus, error) {
	if err := c.writeFrame(conn, cmdDisarmPartition, []byte{0xff, 0x01}); err != nil {
		return PanelStatus{}, err
	}
	frame, err := c.readFrame(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	if err := checkArmPartitionResponse(frame); err != nil {
		return PanelStatus{}, err
	}
	capture, err := c.readStatusOnConn(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	if err := verifyArmed(capture.Status, 0); err != nil {
		return PanelStatus{}, err
	}
	return capture.Status, nil
}

func (c *Client) disarmPartition0OnConn(conn net.Conn) (PanelStatus, error) {
	if err := c.writeFrame(conn, cmdDisarmPartition, []byte{0xff, 0x00}); err != nil {
		return PanelStatus{}, err
	}
	frame, err := c.readFrame(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	if err := checkDisarmPartitionResponse(frame); err != nil {
		return PanelStatus{}, err
	}
	capture, err := c.readStatusOnConn(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	if err := verifyDisarmed(capture.Status, 0); err != nil {
		return PanelStatus{}, err
	}
	return capture.Status, nil
}

func (c *Client) setZoneBypassOnConn(conn net.Conn, zone int, bypassed bool) (PanelStatus, error) {
	before, err := c.readStatusOnConn(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	payload, err := zoneBypassPayload(before.Status.Zones, zone, bypassed)
	if err != nil {
		return PanelStatus{}, err
	}
	if err := c.writeFrame(conn, cmdZoneBypass, payload); err != nil {
		return PanelStatus{}, err
	}
	frame, err := c.readFrame(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	if err := checkOnlineOKResponse(frame, "zone bypass"); err != nil {
		return PanelStatus{}, err
	}
	after, err := c.readStatusOnConn(conn)
	if err != nil {
		return PanelStatus{}, err
	}
	if err := verifyZoneBypass(after.Status, zone, bypassed); err != nil {
		return PanelStatus{}, err
	}
	return after.Status, nil
}

func checkArmPartitionResponse(frame Frame) error {
	return checkPartitionControlResponse(frame, 0x91, "arm")
}

func checkDisarmPartitionResponse(frame Frame) error {
	return checkPartitionControlResponse(frame, 0x90, "disarm")
}

func checkPartitionControlResponse(frame Frame, result byte, action string) error {
	if frame.Command != cmdDisarmPartition {
		return fmt.Errorf("unexpected %s response command 0x%04x", action, frame.Command)
	}
	if len(frame.Payload) != 2 {
		return fmt.Errorf("unexpected %s response payload length %d", action, len(frame.Payload))
	}
	if frame.Payload[0] != 0xff || frame.Payload[1] != result {
		return fmt.Errorf("panel rejected %s command: payload=% x", action, frame.Payload)
	}
	return nil
}

func checkOnlineOKResponse(frame Frame, action string) error {
	if frame.Command != cmdOnlineOK {
		return fmt.Errorf("unexpected %s response command 0x%04x", action, frame.Command)
	}
	if len(frame.Payload) != 0 {
		return fmt.Errorf("unexpected %s response payload length %d", action, len(frame.Payload))
	}
	return nil
}

func zoneBypassPayload(zones []Zone, targetZone int, bypassed bool) ([]byte, error) {
	payload := make([]byte, 0, len(zones)*2)
	found := false
	for _, zone := range zones {
		if zone.Index < 1 || zone.Index > 64 {
			continue
		}
		payload = append(payload, byte(zone.Index-1))
		nextBypassed := zone.Bypassed
		if zone.Index == targetZone {
			found = true
			nextBypassed = bypassed
		}
		if nextBypassed {
			payload = append(payload, 0x01)
		} else {
			payload = append(payload, 0x00)
		}
	}
	if !found {
		return nil, fmt.Errorf("zone %d is not enabled in current status", targetZone)
	}
	if len(payload) == 0 {
		return nil, errors.New("status does not include enabled zones")
	}
	return payload, nil
}

func verifyArmed(status PanelStatus, partition int) error {
	if status.State != "ARMED" {
		return fmt.Errorf("arm verification failed: panel state is %s", status.State)
	}
	for _, candidate := range status.Partitions {
		if candidate.Index == partition && !candidate.Armed {
			return fmt.Errorf("arm verification failed: partition %d is not armed", partition)
		}
	}
	return nil
}

func verifyDisarmed(status PanelStatus, partition int) error {
	if status.State != "DISARMED" {
		return fmt.Errorf("disarm verification failed: panel state is %s", status.State)
	}
	for _, candidate := range status.Partitions {
		if candidate.Index == partition && candidate.Armed {
			return fmt.Errorf("disarm verification failed: partition %d is still armed", partition)
		}
	}
	return nil
}

func verifyZoneBypass(status PanelStatus, zone int, bypassed bool) error {
	for _, candidate := range status.Zones {
		if candidate.Index == zone {
			if candidate.Bypassed != bypassed {
				return fmt.Errorf("zone bypass verification failed: zone %d bypassed=%t", zone, candidate.Bypassed)
			}
			return nil
		}
	}
	return fmt.Errorf("zone bypass verification failed: zone %d is not enabled", zone)
}
