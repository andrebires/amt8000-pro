package isecnet

import "testing"

func TestZoneNameRequestPayload(t *testing.T) {
	payload := zoneNameRequestPayload(16)
	want := []byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f}
	if string(payload) != string(want) {
		t.Fatalf("payload = % x, want % x", payload, want)
	}
}

func TestParseZoneNamesPayload(t *testing.T) {
	payload := []byte{
		0x00, 'I', 'V', 'A', ' ', 'D', 'I', 'V', 'I', 'S', 'A', ' ', ' ', ' ', ' ',
		0x01, 'I', 'V', 'A', ' ', 'f', 'u', 'n', 'd', 'o', 's', ' ', ' ', ' ', ' ',
		0x02, 'I', 'V', 'A', ' ', 'F', 'U', 'N', 'D', 'O', ' ', 'b', 'x', ' ', ' ',
		0x03, 'I', 'V', 'A', ' ', 'L', 'A', 'T', 'E', 'R', 'A', 'L', ' ', ' ', ' ',
		0x04, 'I', 'V', 'A', ' ', 'F', 'R', 'E', 'N', 'T', 'E', ' ', ' ', ' ', ' ',
	}

	names, err := parseZoneNamesPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	wants := map[int]string{
		1: "IVA DIVISA",
		2: "IVA fundos",
		3: "IVA FUNDO bx",
		4: "IVA LATERAL",
		5: "IVA FRENTE",
	}
	for index, want := range wants {
		if names[index] != want {
			t.Fatalf("names[%d] = %q, want %q", index, names[index], want)
		}
	}
}

func TestApplyZoneNames(t *testing.T) {
	status := PanelStatus{Zones: []Zone{{Index: 1}, {Index: 2}, {Index: 3}}}

	status = applyZoneNames(status, map[int]string{1: "IVA DIVISA", 3: "IVA FUNDO bx"})

	if status.Zones[0].Name != "IVA DIVISA" || status.Zones[1].Name != "" || status.Zones[2].Name != "IVA FUNDO bx" {
		t.Fatalf("unexpected zones: %+v", status.Zones)
	}
}

func TestParseZoneNamesPayloadRejectsInvalidShape(t *testing.T) {
	if _, err := parseZoneNamesPayload([]byte{0x00}); err == nil {
		t.Fatal("parseZoneNamesPayload returned nil error for short payload")
	}
}
