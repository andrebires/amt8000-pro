package isecnet

import "fmt"

var states = map[byte]string{
	0: "DISARMED",
	1: "PARTIAL",
	3: "ARMED",
}

var batteryLevels = map[byte]string{
	1: "dead",
	2: "low",
	3: "middle",
	4: "full",
}

func parseStatus(payload []byte) (PanelStatus, error) {
	if len(payload) < 143 {
		return PanelStatus{}, fmt.Errorf("status payload too short: %d", len(payload))
	}

	statusByte := payload[20]
	partitions := make([]Partition, 0, 16)
	for i := range 16 {
		b := payload[21+i]
		if b&0x80 == 0 {
			continue
		}
		partitions = append(partitions, Partition{
			Index:   i,
			Enabled: true,
			Stay:    b&0x40 != 0,
			Fired:   b&0x08 != 0,
			Firing:  b&0x04 != 0,
			Armed:   b&0x01 != 0,
		})
	}

	zones := make([]Zone, 0, 64)
	for i := range 64 {
		bi := i / 8
		bit := byte(1 << (i % 8))
		if payload[12+bi]&bit == 0 {
			continue
		}
		zones = append(zones, Zone{
			Index:      i + 1,
			Open:       payload[38+bi]&bit != 0,
			Violated:   payload[46+bi]&bit != 0,
			Bypassed:   payload[54+bi]&bit != 0,
			Tamper:     payload[89+bi]&bit != 0,
			LowBattery: payload[105+bi]&bit != 0,
		})
	}

	state := states[(statusByte>>5)&0x03]
	if state == "" {
		state = "UNKNOWN"
	}
	battery := batteryLevels[payload[134]]
	if battery == "" {
		battery = "unknown"
	}

	return PanelStatus{
		Model:       payload[0],
		Version:     fmt.Sprintf("%d.%d.%d", payload[1], payload[2], payload[3]),
		State:       state,
		SirenLive:   statusByte&0x02 != 0,
		ZonesFiring: statusByte&0x08 != 0,
		ZonesClosed: statusByte&0x04 != 0,
		Battery:     battery,
		Tamper:      payload[71]&0x02 != 0,
		Partitions:  partitions,
		Zones:       zones,
	}, nil
}
