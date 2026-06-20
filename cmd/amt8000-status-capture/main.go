package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/andrebires/amt8000-pro/internal/isecnet"
)

type fixture struct {
	CapturedAtUTC   string              `json:"capturedAtUtc"`
	Command         string              `json:"command"`
	Response        responseFixture     `json:"response"`
	ParsedStatus    isecnet.PanelStatus `json:"parsedStatus"`
	SanitizedFields map[string]string   `json:"sanitizedFields"`
}

type responseFixture struct {
	FrameHex      string `json:"frameHex"`
	PayloadHex    string `json:"payloadHex"`
	PayloadLength int    `json:"payloadLength"`
}

func main() {
	host := os.Getenv("AMT_HOST")
	password := os.Getenv("AMT_PASSWORD")
	if host == "" {
		fatal("AMT_HOST is required")
	}
	if password == "" {
		fatal("AMT_PASSWORD is required")
	}

	port := 9009
	if raw := os.Getenv("AMT_PORT"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 65535 {
			fatal("AMT_PORT must be between 1 and 65535")
		}
		port = parsed
	}

	client := isecnet.NewClient(host, port, password, 5*time.Second)
	capture, err := client.GetStatusCapture()
	if err != nil {
		fatal(err.Error())
	}

	out := fixture{
		CapturedAtUTC: time.Now().UTC().Format(time.RFC3339),
		Command:       "0x0b4a",
		Response: responseFixture{
			FrameHex:      hex.EncodeToString(capture.Frame.Raw),
			PayloadHex:    hex.EncodeToString(capture.Frame.Payload),
			PayloadLength: len(capture.Frame.Payload),
		},
		ParsedStatus: capture.Status,
		SanitizedFields: map[string]string{
			"host":     "omitted",
			"password": "omitted",
			"cookie":   "omitted",
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(out); err != nil {
		fatal(err.Error())
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
