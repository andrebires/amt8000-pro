package isecnet

import (
	"fmt"
	"net"
	"strings"
)

const (
	cmdZoneNames       uint16 = 0x33e0
	zoneNameBatchSize         = 16
	zoneNameRecordSize        = 15
	zoneNameTextSize          = 14
	zoneNameMaxZones          = 64
)

func (c *Client) readZoneNamesOnConn(conn net.Conn) (map[int]string, error) {
	names := make(map[int]string, zoneNameMaxZones)
	for start := 0; start < zoneNameMaxZones; start += zoneNameBatchSize {
		payload := zoneNameRequestPayload(start)
		if err := c.writeFrame(conn, cmdZoneNames, payload); err != nil {
			return nil, err
		}
		frame, err := c.readFrame(conn)
		if err != nil {
			return nil, err
		}
		if frame.Command != cmdZoneNames {
			return nil, fmt.Errorf("unexpected zone names response command 0x%04x", frame.Command)
		}
		batchNames, err := parseZoneNamesPayload(frame.Payload)
		if err != nil {
			return nil, err
		}
		for index, name := range batchNames {
			names[index] = name
		}
	}
	return names, nil
}

func zoneNameRequestPayload(start int) []byte {
	payload := make([]byte, zoneNameBatchSize)
	for i := range zoneNameBatchSize {
		payload[i] = byte(start + i)
	}
	return payload
}

func parseZoneNamesPayload(payload []byte) (map[int]string, error) {
	if len(payload)%zoneNameRecordSize != 0 {
		return nil, fmt.Errorf("zone names payload length %d is not a multiple of %d", len(payload), zoneNameRecordSize)
	}
	names := make(map[int]string, len(payload)/zoneNameRecordSize)
	for offset := 0; offset < len(payload); offset += zoneNameRecordSize {
		record := payload[offset : offset+zoneNameRecordSize]
		index := int(record[0]) + 1
		if index < 1 || index > zoneNameMaxZones {
			return nil, fmt.Errorf("zone name index %d out of range", index)
		}
		name := strings.TrimRight(string(record[1:1+zoneNameTextSize]), " \x00\xff")
		if name != "" {
			names[index] = name
		}
	}
	return names, nil
}

func applyZoneNames(status PanelStatus, names map[int]string) PanelStatus {
	if len(names) == 0 {
		return status
	}
	for i := range status.Zones {
		status.Zones[i].Name = names[status.Zones[i].Index]
	}
	return status
}
