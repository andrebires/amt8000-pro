package isecnet

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestEventRequestPayload(t *testing.T) {
	payload, err := eventRequestPayload([]int{21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := hex.DecodeString("001500140013001200110010000f000e000d000c000b000a0009000800070006")
	if !bytes.Equal(payload, want) {
		t.Fatalf("payload = %x, want %x", payload, want)
	}
}

func TestParseEventRecord(t *testing.T) {
	record, _ := hex.DecodeString("01d52605170719231133fffaa2aa00")

	event, ok := parseEventRecord(record)
	if !ok {
		t.Fatal("parseEventRecord returned ok=false")
	}
	if event.Index != 469 {
		t.Fatalf("index = %d, want 469", event.Index)
	}
	if event.Timestamp != "2026-05-17T07:19:23" {
		t.Fatalf("timestamp = %q, want 2026-05-17T07:19:23", event.Timestamp)
	}
	if event.Code != "0x1133" {
		t.Fatalf("code = %q, want 0x1133", event.Code)
	}
	if event.Description != "24h zone alarm - Zone 02" {
		t.Fatalf("description = %q", event.Description)
	}
	if event.Zone == nil || *event.Zone != 2 {
		t.Fatalf("zone = %v, want 2", event.Zone)
	}
	if event.Raw != "01d52605170719231133fffaa2aa00" {
		t.Fatalf("raw = %q", event.Raw)
	}
}

func TestDescribeCapturedEventRecords(t *testing.T) {
	tests := []struct {
		name           string
		raw            string
		description    string
		zone           *int
		partition      *int
		user           *int
		deliveryStatus string
	}{
		{
			name:        "periodic test",
			raw:         "001126062006300016a2fffaaaaa00",
			description: "Periodic test",
		},
		{
			name:        "power grid failure",
			raw:         "000f26061916411213a1fffaaaaa00",
			description: "Power grid failure",
		},
		{
			name:        "power grid restored",
			raw:         "001026061918114233a1fffaaaaa00",
			description: "Restoration power grid failure",
		},
		{
			name:        "zone alarm",
			raw:         "01fa2606070724471133fffaa1aa00",
			description: "24h zone alarm - Zone 01",
			zone:        intPtr(1),
		},
		{
			name:        "zone alarm restored",
			raw:         "01fb2606070724513133fffaa1aa00",
			description: "Restoration zone alarm 24 hours - Zone 01",
			zone:        intPtr(1),
		},
		{
			name:        "user activation",
			raw:         "01db26052009075434a1fff2a1aa00",
			description: "User activation - Partition 0 - User 01",
			partition:   intPtr(0),
			user:        intPtr(1),
		},
		{
			name:        "user deactivation",
			raw:         "01da26052009070514a1fff2a1aa00",
			description: "User deactivation - Partition 0 - User 01",
			partition:   intPtr(0),
			user:        intPtr(1),
		},
		{
			name:        "master activation",
			raw:         "000226060716525034a7fff1aaaa00",
			description: "Activation via keyboard or phone - Partition 0 - User Master",
			partition:   intPtr(0),
		},
		{
			name:        "master deactivation",
			raw:         "01fc26060707245214a7fff1aaaa00",
			description: "Deactivation via keyboard or phone - Partition 0 - User Master",
			partition:   intPtr(0),
		},
		{
			name:        "programming changed",
			raw:         "000126060716523113a6fff199aa00",
			description: "Programming changed - User 99",
			user:        intPtr(99),
		},
		{
			name:        "tamper violation",
			raw:         "00632506120307501145fffaaaaa00",
			description: "Tamper violation - Panel",
		},
		{
			name:        "tamper restoration",
			raw:         "00662506120915373145fffaaaaa00",
			description: "Tamper restoration - Panel",
		},
		{
			name:        "incorrect password",
			raw:         "00602506120306561461fffaaaaa00",
			description: "Incorrect password event",
		},
		{
			name:           "failure to report",
			raw:            "00282504190641061354fffaaaaa00",
			description:    "Failure to report events",
			deliveryStatus: "failed",
		},
		{
			name:        "reset clock",
			raw:         "00151001010000011625fff2a2aa00",
			description: "Reset date and time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, _ := hex.DecodeString(tt.raw)
			event, ok := parseEventRecord(record)
			if !ok {
				t.Fatal("parseEventRecord returned ok=false")
			}
			if event.Description != tt.description {
				t.Fatalf("description = %q, want %q", event.Description, tt.description)
			}
			assertOptionalInt(t, "zone", event.Zone, tt.zone)
			assertOptionalInt(t, "partition", event.Partition, tt.partition)
			assertOptionalInt(t, "user", event.User, tt.user)
			if event.DeliveryStatus != tt.deliveryStatus {
				t.Fatalf("delivery status = %q, want %q", event.DeliveryStatus, tt.deliveryStatus)
			}
		})
	}
}

func TestParseEventPayload(t *testing.T) {
	payload, _ := hex.DecodeString(
		"01d52605170719231133fffaa2aa00" +
			"01d426051706300016a2fffaaaaa00",
	)

	events, err := parseEventPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}
	if events[0].Index != 469 || events[1].Index != 468 {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestLatestEvents(t *testing.T) {
	events := []PanelEvent{
		{Index: 1, Timestamp: "2026-05-17T07:19:23"},
		{Index: 2, Timestamp: "2026-06-20T06:30:00"},
		{Index: 3, Timestamp: "2026-06-19T18:11:42"},
	}

	got := latestEvents(events, 2)
	if len(got) != 2 {
		t.Fatalf("events len = %d, want 2", len(got))
	}
	if got[0].Index != 2 || got[1].Index != 3 {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func assertOptionalInt(t *testing.T, name string, got *int, want *int) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("%s = %v, want %v", name, got, want)
		}
		return
	}
	if *got != *want {
		t.Fatalf("%s = %d, want %d", name, *got, *want)
	}
}
