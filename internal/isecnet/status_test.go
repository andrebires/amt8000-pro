package isecnet

import "testing"

func TestParseStatus(t *testing.T) {
	payload := make([]byte, 143)
	payload[0] = 0x8b
	payload[1], payload[2], payload[3] = 3, 2, 5
	payload[12] = 0b00000011
	payload[20] = 0b01100110
	payload[21] = 0b10000001
	payload[22] = 0b10000000
	payload[38] = 0b00000010
	payload[46] = 0b00000001
	payload[54] = 0b00000010
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
	if len(status.Partitions) != 2 || !status.Partitions[0].Armed {
		t.Fatalf("unexpected partitions: %+v", status.Partitions)
	}
	if len(status.Zones) != 2 {
		t.Fatalf("zones len = %d, want 2", len(status.Zones))
	}
	if !status.Zones[0].Violated || !status.Zones[0].Tamper {
		t.Fatalf("zone 1 not parsed: %+v", status.Zones[0])
	}
	if !status.Zones[1].Open || !status.Zones[1].Bypassed || !status.Zones[1].LowBattery {
		t.Fatalf("zone 2 not parsed: %+v", status.Zones[1])
	}
}
