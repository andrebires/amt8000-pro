package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andrebires/amt8000-pro/internal/isecnet"
)

type fakeClient struct{}

func (fakeClient) GetStatus() (isecnet.PanelStatus, error) {
	return isecnet.PanelStatus{Model: 0x8b, Version: "3.2.5", State: "DISARMED", Battery: "full"}, nil
}

func TestStatusEndpoint(t *testing.T) {
	server := NewServer(fakeClient{})
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"version":"3.2.5"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
