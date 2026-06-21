package isecnet

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"time"
)

const (
	EventLimit      = 256
	eventBufferSize = 512
	eventBatchSize  = 16
	eventRecordSize = 15
	eventBatchDelay = 200 * time.Millisecond
	cmdEvents       = 0x3900
)

type eventCapture struct {
	Events []PanelEvent
	Frames []Frame
}

func (c *Client) GetEvents() (PanelEvents, error) {
	capture, err := c.GetEventsCapture()
	if err != nil {
		return PanelEvents{}, err
	}
	events := latestEvents(capture.Events, EventLimit)
	return PanelEvents{
		Events: events,
		Limit:  EventLimit,
		Total:  len(events),
	}, nil
}

func (c *Client) GetEventsCapture() (eventCapture, error) {
	conn, err := c.connectAndAuth()
	if err != nil {
		return eventCapture{}, err
	}
	defer conn.Close()
	defer func() {
		_ = c.writeFrame(conn, cmdDisconnect, nil)
	}()

	events := make([]PanelEvent, 0, eventBufferSize)
	frames := make([]Frame, 0, eventBufferSize/eventBatchSize)
	for start := eventBufferSize - eventBatchSize; start >= 0; start -= eventBatchSize {
		batch := make([]int, 0, eventBatchSize)
		for index := start + eventBatchSize - 1; index >= start; index-- {
			batch = append(batch, index)
		}
		frame, batchEvents, err := c.readEventBatch(conn, batch)
		if err != nil {
			return eventCapture{}, err
		}
		frames = append(frames, frame)
		events = append(events, batchEvents...)
		if start > 0 {
			time.Sleep(eventBatchDelay)
		}
	}

	return eventCapture{
		Events: events,
		Frames: frames,
	}, nil
}

func (c *Client) readEventBatch(conn net.Conn, indexes []int) (Frame, []PanelEvent, error) {
	payload, err := eventRequestPayload(indexes)
	if err != nil {
		return Frame{}, nil, err
	}
	if err := c.writeFrame(conn, cmdEvents, payload); err != nil {
		return Frame{}, nil, err
	}
	frame, err := c.readFrame(conn)
	if err != nil {
		return Frame{}, nil, err
	}
	if frame.Command != cmdEvents {
		return Frame{}, nil, fmt.Errorf("unexpected events response command 0x%04x", frame.Command)
	}
	events, err := parseEventPayload(frame.Payload)
	if err != nil {
		return Frame{}, nil, err
	}
	return frame, events, nil
}

func eventRequestPayload(indexes []int) ([]byte, error) {
	if len(indexes) != eventBatchSize {
		return nil, fmt.Errorf("event request needs %d indexes, got %d", eventBatchSize, len(indexes))
	}
	payload := make([]byte, eventBatchSize*2)
	for i, index := range indexes {
		if index < 0 || index >= eventBufferSize {
			return nil, fmt.Errorf("event index %d out of range", index)
		}
		binary.BigEndian.PutUint16(payload[i*2:i*2+2], uint16(index))
	}
	return payload, nil
}

func parseEventPayload(payload []byte) ([]PanelEvent, error) {
	if len(payload)%eventRecordSize != 0 {
		return nil, fmt.Errorf("events payload length %d is not a multiple of %d", len(payload), eventRecordSize)
	}
	events := make([]PanelEvent, 0, len(payload)/eventRecordSize)
	for offset := 0; offset < len(payload); offset += eventRecordSize {
		event, ok := parseEventRecord(payload[offset : offset+eventRecordSize])
		if ok {
			events = append(events, event)
		}
	}
	return events, nil
}

func parseEventRecord(record []byte) (PanelEvent, bool) {
	if len(record) != eventRecordSize {
		return PanelEvent{}, false
	}
	timestamp := parseEventDateTime(record[2:8])
	if timestamp == "" {
		return PanelEvent{}, false
	}
	codeBytes := record[8:10]
	event := PanelEvent{
		Index:     int(binary.BigEndian.Uint16(record[0:2])),
		Timestamp: timestamp,
		Code:      "0x" + hex.EncodeToString(codeBytes),
		Raw:       hex.EncodeToString(record),
	}
	describeEventRecord(record, &event)
	return event, true
}

func describeEventRecord(record []byte, event *PanelEvent) {
	code := binary.BigEndian.Uint16(record[8:10])
	partition0 := 0

	switch code {
	case 0x1133:
		setZoneEvent(event, record, "24h zone alarm")
	case 0x3133:
		setZoneEvent(event, record, "Restoration zone alarm 24 hours")
	case 0x1145:
		event.Description = "Tamper violation - Panel"
	case 0x3145:
		event.Description = "Tamper restoration - Panel"
	case 0x1354:
		event.Description = "Failure to report events"
		event.DeliveryStatus = "failed"
	case 0x13a1:
		event.Description = "Power grid failure"
	case 0x33a1:
		event.Description = "Restoration power grid failure"
	case 0x13a6:
		user, ok := eventParamNumber(record[12])
		if ok {
			event.User = intPtr(user)
			event.Description = fmt.Sprintf("Programming changed - User %02d", user)
		} else {
			event.Description = "Programming changed"
		}
	case 0x1461:
		event.Description = "Incorrect password event"
	case 0x14a1:
		user, ok := eventParamNumber(record[12])
		event.Partition = &partition0
		if ok {
			event.User = intPtr(user)
			event.Description = fmt.Sprintf("User deactivation - Partition 0 - User %02d", user)
		} else {
			event.Description = "User deactivation - Partition 0"
		}
	case 0x34a1:
		user, ok := eventParamNumber(record[12])
		event.Partition = &partition0
		if ok {
			event.User = intPtr(user)
			event.Description = fmt.Sprintf("User activation - Partition 0 - User %02d", user)
		} else {
			event.Description = "User activation - Partition 0"
		}
	case 0x14a7:
		event.Partition = &partition0
		event.Description = "Deactivation via keyboard or phone - Partition 0 - User Master"
	case 0x34a7:
		event.Partition = &partition0
		event.Description = "Activation via keyboard or phone - Partition 0 - User Master"
	case 0x1625:
		event.Description = "Reset date and time"
	case 0x16a2:
		event.Description = "Periodic test"
	}
}

func setZoneEvent(event *PanelEvent, record []byte, label string) {
	zone, ok := eventParamNumber(record[12])
	if !ok {
		event.Description = label
		return
	}
	event.Zone = intPtr(zone)
	event.Description = fmt.Sprintf("%s - Zone %02d", label, zone)
}

func eventParamNumber(value byte) (int, bool) {
	if value == 0xaa || value == 0xff {
		return 0, false
	}
	hi := value >> 4
	lo := value & 0x0f
	if hi <= 9 && lo <= 9 {
		return int(hi*10 + lo), true
	}
	if hi == 0x0a && lo <= 9 {
		return int(lo), true
	}
	return 0, false
}

func intPtr(value int) *int {
	return &value
}

func parseEventDateTime(b []byte) string {
	if len(b) < 6 {
		return ""
	}
	year := 2000 + bcd(b[0])
	month := bcd(b[1])
	day := bcd(b[2])
	hour := bcd(b[3])
	minute := bcd(b[4])
	second := bcd(b[5])
	if !validPanelDateTime(year, month, day, hour, minute, second) {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d", year, month, day, hour, minute, second)
}

func latestEvents(events []PanelEvent, limit int) []PanelEvent {
	out := append([]PanelEvent(nil), events...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Timestamp == out[j].Timestamp {
			return out[i].Index > out[j].Index
		}
		return out[i].Timestamp > out[j].Timestamp
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
