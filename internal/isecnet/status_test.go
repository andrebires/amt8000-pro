package isecnet

import "testing"

func TestParseStatus(t *testing.T) {
	payload := make([]byte, 143)
	payload[0] = 0x8b
	payload[1], payload[2], payload[3] = 3, 2, 5
	payload[12] = 0b00000011
	payload[20] = 0b01100110
	payload[21] = 0b10000001
	payload[22] = 0b10001100
	payload[38] = 0b00000010
	payload[46] = 0b00000011
	payload[54] = 0b00000010
	copy(payload[64:70], []byte{0x20, 0x06, 0x26, 0x19, 0x15, 0x21})
	payload[71] = 0x02
	payload[89] = 0b00000001
	payload[105] = 0b00000010
	payload[134] = 4

	status, err := parseStatus(payload)
	if err != nil {
		t.Fatal(err)
	}
	if status.Model != 0x8b || status.Version != "3.2.5" {
		t.Fatalf("unexpected model/version: %#x %s", status.Model, status.Version)
	}
	if status.State != "ARMED" || !status.SirenLive || !status.ZonesClosed {
		t.Fatalf("unexpected global status: %+v", status)
	}
	if status.PanelDateTime != "2026-06-20T19:15:21" {
		t.Fatalf("panel date/time = %q, want 2026-06-20T19:15:21", status.PanelDateTime)
	}
	if len(status.Partitions) != 2 || !status.Partitions[0].Armed {
		t.Fatalf("unexpected partitions: %+v", status.Partitions)
	}
	if status.Partitions[0].State != "ARMED" || status.Partitions[1].State != "FIRING" {
		t.Fatalf("unexpected partition states: %+v", status.Partitions)
	}
	if len(status.Zones) != 2 {
		t.Fatalf("zones len = %d, want 2", len(status.Zones))
	}
	if status.Zones[0].State != "FIRED_CLOSED" || !status.Zones[0].Violated || !status.Zones[0].Tamper {
		t.Fatalf("zone 1 not parsed: %+v", status.Zones[0])
	}
	if status.Zones[1].State != "FIRED_OPEN" || !status.Zones[1].Open || !status.Zones[1].Bypassed || !status.Zones[1].LowBattery {
		t.Fatalf("zone 2 not parsed: %+v", status.Zones[1])
	}
	if len(status.Troubles) != 3 {
		t.Fatalf("troubles len = %d, want 3: %+v", len(status.Troubles), status.Troubles)
	}
	if status.Troubles[0].Code != "PANEL_TAMPER" {
		t.Fatalf("unexpected troubles: %+v", status.Troubles)
	}
}

func TestParsePanelDateTime(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		want string
	}{
		{name: "valid", raw: []byte{0x20, 0x06, 0x26, 0x19, 0x15, 0x21}, want: "2026-06-20T19:15:21"},
		{name: "leap day", raw: []byte{0x29, 0x02, 0x24, 0x00, 0x00, 0x00}, want: "2024-02-29T00:00:00"},
		{name: "invalid bcd", raw: []byte{0x2a, 0x06, 0x26, 0x19, 0x15, 0x21}},
		{name: "invalid day", raw: []byte{0x31, 0x04, 0x26, 0x19, 0x15, 0x21}},
		{name: "invalid time", raw: []byte{0x20, 0x06, 0x26, 0x24, 0x15, 0x21}},
		{name: "too short", raw: []byte{0x20, 0x06}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePanelDateTime(tt.raw); got != tt.want {
				t.Fatalf("parsePanelDateTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestZoneState(t *testing.T) {
	tests := []struct {
		name     string
		open     bool
		violated bool
		want     string
	}{
		{name: "closed", want: "CLOSED"},
		{name: "open", open: true, want: "OPEN"},
		{name: "fired closed", violated: true, want: "FIRED_CLOSED"},
		{name: "fired open", open: true, violated: true, want: "FIRED_OPEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := zoneState(tt.open, tt.violated); got != tt.want {
				t.Fatalf("zoneState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPartitionState(t *testing.T) {
	tests := []struct {
		name string
		raw  byte
		want string
	}{
		{name: "disarmed", raw: 0x80, want: "DISARMED"},
		{name: "armed", raw: 0x81, want: "ARMED"},
		{name: "stay", raw: 0xc1, want: "STAY"},
		{name: "fired", raw: 0x89, want: "FIRED"},
		{name: "firing", raw: 0x8d, want: "FIRING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := partitionState(tt.raw); got != tt.want {
				t.Fatalf("partitionState() = %q, want %q", got, tt.want)
			}
		})
	}
}
