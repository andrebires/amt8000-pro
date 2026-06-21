package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	defaultPanelPort = 9009
	defaultProxyAddr = "127.0.0.1:19009"
	maxFrameLength   = 4096
	cmdAuth          = 0xf0f0
)

type captureRecord struct {
	CapturedAtUTC string `json:"capturedAtUtc"`
	SessionID     string `json:"sessionId"`
	Direction     string `json:"direction"`
	Kind          string `json:"kind,omitempty"`
	Command       string `json:"command,omitempty"`
	Dst           string `json:"dst,omitempty"`
	Src           string `json:"src,omitempty"`
	PayloadHex    string `json:"payloadHex,omitempty"`
	PayloadLength int    `json:"payloadLength,omitempty"`
	FrameHex      string `json:"frameHex,omitempty"`
	FrameLength   int    `json:"frameLength,omitempty"`
	ChunkHex      string `json:"chunkHex,omitempty"`
	ChunkLength   int    `json:"chunkLength,omitempty"`
	Redacted      bool   `json:"redacted,omitempty"`
	SkippedBytes  int    `json:"skippedBytes,omitempty"`
	Error         string `json:"error,omitempty"`
}

type recorder struct {
	mu      sync.Mutex
	encoder *json.Encoder
}

type frameParser struct {
	sessionID string
	direction string
	recorder  *recorder
	buffer    []byte
}

func main() {
	target, err := panelTarget()
	if err != nil {
		fatal(err)
	}
	proxyAddr := getenvDefault("AMT_PROXY_ADDR", defaultProxyAddr)
	outPath := getenvDefault("AMT_CAPTURE_OUT", defaultCapturePath())

	if err := os.MkdirAll(filepath.Dir(outPath), 0o700); err != nil {
		fatal(err)
	}
	out, err := os.Create(outPath)
	if err != nil {
		fatal(err)
	}
	defer out.Close()

	rec := &recorder{encoder: json.NewEncoder(out)}
	listener, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		fatal(err)
	}
	defer listener.Close()

	log.Printf("capture proxy listening on %s", listener.Addr().String())
	log.Printf("forwarding to panel %s", target)
	log.Printf("writing redacted JSONL capture to %s", outPath)
	log.Printf("point AMT Remoto/Programador at this proxy address, then download the event buffer")

	for {
		client, err := listener.Accept()
		if err != nil {
			fatal(err)
		}
		go handleConnection(client, target, rec)
	}
}

func panelTarget() (string, error) {
	host := os.Getenv("AMT_HOST")
	if host == "" {
		return "", errors.New("AMT_HOST is required")
	}
	port := defaultPanelPort
	if raw := os.Getenv("AMT_PORT"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 65535 {
			return "", errors.New("AMT_PORT must be between 1 and 65535")
		}
		port = parsed
	}
	return net.JoinHostPort(host, strconv.Itoa(port)), nil
}

func handleConnection(client net.Conn, target string, rec *recorder) {
	defer client.Close()
	panel, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		log.Printf("panel dial failed: %v", err)
		return
	}
	defer panel.Close()

	sessionID := time.Now().UTC().Format("20060102T150405.000000000Z")
	log.Printf("accepted session %s from %s", sessionID, client.RemoteAddr())

	done := make(chan struct{}, 2)
	go proxyStream(panel, client, newFrameParser(sessionID, "client_to_panel", rec), done)
	go proxyStream(client, panel, newFrameParser(sessionID, "panel_to_client", rec), done)
	<-done
}

func proxyStream(dst net.Conn, src net.Conn, parser *frameParser, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()
	buf := make([]byte, 8192)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			parser.Write(chunk)
			if _, writeErr := dst.Write(chunk); writeErr != nil {
				parser.RecordError(writeErr)
				return
			}
		}
		if readErr != nil {
			if !errors.Is(readErr, io.EOF) {
				parser.RecordError(readErr)
			}
			return
		}
	}
}

func newFrameParser(sessionID string, direction string, rec *recorder) *frameParser {
	return &frameParser{
		sessionID: sessionID,
		direction: direction,
		recorder:  rec,
	}
}

func (p *frameParser) Write(chunk []byte) {
	if p.direction == "client_to_panel" {
		p.recordClientChunk(chunk)
	}
	p.buffer = append(p.buffer, chunk...)
	for {
		if len(p.buffer) < 6 {
			return
		}
		length := int(binary.BigEndian.Uint16(p.buffer[4:6]))
		if length < 2 || length > maxFrameLength {
			offset := p.nextPlausibleFrameOffset()
			if offset <= 0 {
				if len(p.buffer) > 6 {
					p.record(captureRecord{
						Error:        fmt.Sprintf("invalid frame length %d", length),
						SkippedBytes: len(p.buffer) - 5,
					})
					p.buffer = append([]byte(nil), p.buffer[len(p.buffer)-5:]...)
				}
				return
			}
			p.record(captureRecord{
				Error:        fmt.Sprintf("invalid frame length %d", length),
				SkippedBytes: offset,
			})
			p.buffer = p.buffer[offset:]
			continue
		}
		expected := 6 + length + 1
		if len(p.buffer) < expected {
			return
		}
		raw := append([]byte(nil), p.buffer[:expected]...)
		p.buffer = p.buffer[expected:]
		p.recordFrame(raw, length)
	}
}

func (p *frameParser) nextPlausibleFrameOffset() int {
	for i := 1; i+8 <= len(p.buffer); i++ {
		length := int(binary.BigEndian.Uint16(p.buffer[i+4 : i+6]))
		if length < 2 || length > maxFrameLength {
			continue
		}
		command := binary.BigEndian.Uint16(p.buffer[i+6 : i+8])
		if command == 0 {
			continue
		}
		expected := i + 6 + length + 1
		if expected <= len(p.buffer) || len(p.buffer)-i < 6+length+1 {
			return i
		}
	}
	return -1
}

func (p *frameParser) recordFrame(raw []byte, length int) {
	command := binary.BigEndian.Uint16(raw[6:8])
	record := captureRecord{
		Command:       fmt.Sprintf("0x%04x", command),
		Dst:           fmt.Sprintf("0x%04x", binary.BigEndian.Uint16(raw[0:2])),
		Src:           fmt.Sprintf("0x%04x", binary.BigEndian.Uint16(raw[2:4])),
		PayloadLength: length - 2,
		FrameLength:   len(raw),
	}
	if command == cmdAuth {
		record.Redacted = true
	} else {
		payload := raw[8 : 8+length-2]
		record.PayloadHex = hex.EncodeToString(payload)
		record.FrameHex = hex.EncodeToString(raw)
	}
	p.record(record)
}

func (p *frameParser) RecordError(err error) {
	if p.direction == "client_to_panel" && len(p.buffer) > 0 {
		p.recordClientChunk(p.buffer)
		p.buffer = nil
	}
	p.record(captureRecord{Error: err.Error()})
}

func (p *frameParser) recordClientChunk(chunk []byte) {
	record := captureRecord{
		Kind:        "client_chunk",
		ChunkLength: len(chunk),
	}
	if containsAuthCommand(chunk) {
		record.Redacted = true
	} else {
		record.ChunkHex = hex.EncodeToString(chunk)
	}
	p.record(record)
}

func containsAuthCommand(data []byte) bool {
	for i := 0; i+1 < len(data); i++ {
		if data[i] == 0xf0 && data[i+1] == 0xf0 {
			return true
		}
	}
	return false
}

func (p *frameParser) record(record captureRecord) {
	record.CapturedAtUTC = time.Now().UTC().Format(time.RFC3339Nano)
	record.SessionID = p.sessionID
	record.Direction = p.direction
	p.recorder.Write(record)
}

func (r *recorder) Write(record captureRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.encoder.Encode(record); err != nil {
		log.Printf("capture write failed: %v", err)
	}
}

func defaultCapturePath() string {
	return filepath.Join(os.TempDir(), "amt8000-captures", time.Now().UTC().Format("20060102T150405Z")+"-isecnet.jsonl")
}

func getenvDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
