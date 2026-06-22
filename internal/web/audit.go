package web

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AuditSink interface {
	WriteOnlineCommand(record OnlineCommandAuditRecord) (string, error)
}

type OnlineCommandAuditRecord struct {
	ID             string `json:"id"`
	TimestampUTC   string `json:"timestampUtc"`
	PanelHost      string `json:"panelHost"`
	PanelPort      int    `json:"panelPort"`
	Action         string `json:"action"`
	Target         string `json:"target,omitempty"`
	RequestedState string `json:"requestedState,omitempty"`
	Result         string `json:"result"`
	Error          string `json:"error,omitempty"`
}

type JSONLAuditSink struct {
	path string
	mu   sync.Mutex
}

func NewJSONLAuditSink(path string) *JSONLAuditSink {
	return &JSONLAuditSink{path: path}
}

func (s *JSONLAuditSink) WriteOnlineCommand(record OnlineCommandAuditRecord) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record.ID == "" {
		record.ID = newAuditID()
	}
	if record.TimestampUTC == "" {
		record.TimestampUTC = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return record.ID, err
	}
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return record.ID, err
	}
	defer file.Close()
	return record.ID, json.NewEncoder(file).Encode(record)
}

type noopAuditSink struct{}

func (noopAuditSink) WriteOnlineCommand(record OnlineCommandAuditRecord) (string, error) {
	if record.ID == "" {
		record.ID = newAuditID()
	}
	return record.ID, nil
}

func defaultAuditPath() string {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "amt8000-pro", "online-audit.jsonl")
}

func newAuditID() string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err == nil {
		return time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + hex.EncodeToString(raw[:])
	}
	return fmt.Sprintf("%s-%d", time.Now().UTC().Format("20060102T150405.000000000Z"), time.Now().UnixNano())
}

func logAuditFailure(err error) {
	if err != nil {
		log.Printf("online command audit write failed: %v", err)
	}
}
