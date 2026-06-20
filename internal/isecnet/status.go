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
			State:   partitionState(b),
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
		open := payload[38+bi]&bit != 0
		violated := payload[46+bi]&bit != 0
		zones = append(zones, Zone{
			Index:      i + 1,
			State:      zoneState(open, violated),
			Open:       open,
			Violated:   violated,
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
		Model:         payload[0],
		Version:       fmt.Sprintf("%d.%d.%d", payload[1], payload[2], payload[3]),
		State:         state,
		PanelDateTime: parsePanelDateTime(payload[64:70]),
		SirenLive:     statusByte&0x02 != 0,
		ZonesFiring:   statusByte&0x08 != 0,
		ZonesClosed:   statusByte&0x04 != 0,
		Battery:       battery,
		Tamper:        payload[71]&0x02 != 0,
		Troubles:      deriveTroubles(payload[71]&0x02 != 0, battery, zones),
		Partitions:    partitions,
		Zones:         zones,
	}, nil
}

func parsePanelDateTime(b []byte) string {
	if len(b) < 6 {
		return ""
	}
	day := bcd(b[0])
	month := bcd(b[1])
	year := 2000 + bcd(b[2])
	hour := bcd(b[3])
	minute := bcd(b[4])
	second := bcd(b[5])
	if !validPanelDateTime(year, month, day, hour, minute, second) {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d", year, month, day, hour, minute, second)
}

func bcd(b byte) int {
	high := int(b >> 4)
	low := int(b & 0x0f)
	if high > 9 || low > 9 {
		return -1
	}
	return high*10 + low
}

func validPanelDateTime(year, month, day, hour, minute, second int) bool {
	if year < 2000 || year > 2099 || month < 1 || month > 12 || hour > 23 || minute > 59 || second > 59 {
		return false
	}
	daysInMonth := [...]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if month == 2 && isLeapYear(year) {
		return day >= 1 && day <= 29
	}
	return day >= 1 && day <= daysInMonth[month]
}

func isLeapYear(year int) bool {
	return year%400 == 0 || (year%4 == 0 && year%100 != 0)
}

func zoneState(open, violated bool) string {
	switch {
	case violated && open:
		return "FIRED_OPEN"
	case violated:
		return "FIRED_CLOSED"
	case open:
		return "OPEN"
	default:
		return "CLOSED"
	}
}

func partitionState(b byte) string {
	switch {
	case b&0x04 != 0:
		return "FIRING"
	case b&0x08 != 0:
		return "FIRED"
	case b&0x40 != 0:
		return "STAY"
	case b&0x01 != 0:
		return "ARMED"
	default:
		return "DISARMED"
	}
}

func deriveTroubles(panelTamper bool, battery string, zones []Zone) []Trouble {
	troubles := make([]Trouble, 0)
	if panelTamper {
		troubles = append(troubles, Trouble{
			Code:    "PANEL_TAMPER",
			Message: "Panel tamper is active",
		})
	}
	if battery == "low" || battery == "dead" {
		troubles = append(troubles, Trouble{
			Code:    "PANEL_BATTERY_" + batteryCodeSuffix(battery),
			Message: "Panel battery is " + battery,
		})
	}
	for _, zone := range zones {
		zoneIndex := zone.Index
		if zone.Tamper {
			troubles = append(troubles, Trouble{
				Code:    "ZONE_TAMPER",
				Message: "Zone tamper is active",
				Zone:    zoneIndex,
			})
		}
		if zone.LowBattery {
			troubles = append(troubles, Trouble{
				Code:    "ZONE_LOW_BATTERY",
				Message: "Zone battery is low",
				Zone:    zoneIndex,
			})
		}
	}
	return troubles
}

func batteryCodeSuffix(battery string) string {
	switch battery {
	case "dead":
		return "DEAD"
	case "low":
		return "LOW"
	default:
		return "UNKNOWN"
	}
}
