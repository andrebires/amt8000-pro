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
	return isecnet.PanelStatus{
		Model:         0x8b,
		Version:       "3.2.5",
		State:         "DISARMED",
		PanelDateTime: "2026-06-20T19:15:21",
		Battery:       "full",
		Troubles: []isecnet.Trouble{
			{Code: "ZONE_LOW_BATTERY", Message: "Zone battery is low", Zone: 2},
		},
		Partitions: []isecnet.Partition{
			{Index: 0, Enabled: true, State: "DISARMED"},
		},
		Zones: []isecnet.Zone{
			{Index: 2, State: "OPEN", Open: true},
		},
	}, nil
}

func TestStatusEndpoint(t *testing.T) {
	server := NewServer(func(PanelConnection) StatusClient { return fakeClient{} })
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	setConnectionCookie(httptest.NewRecorder(), PanelConnection{})
	req.AddCookie(encodedTestCookie(t, PanelConnection{Host: "192.168.1.50", Port: 9009, Password: "878787"}))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"version":"3.2.5"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"state":"OPEN"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestIndexRedirectsToLoginWithoutConnection(t *testing.T) {
	server := NewServer(func(PanelConnection) StatusClient { return fakeClient{} })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "/login" {
		t.Fatalf("Location = %q, want /login", location)
	}
}

func TestLoginPage(t *testing.T) {
	server := NewServer(func(PanelConnection) StatusClient { return fakeClient{} })
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Panel IP") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestIndexRendersReadOnlyOnlineStatus(t *testing.T) {
	server := NewServer(func(PanelConnection) StatusClient { return fakeClient{} })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(encodedTestCookie(t, PanelConnection{Host: "192.168.1.50", Port: 9009, Password: "878787"}))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	for _, want := range []string{"Connection elapsed", "Last refresh", "Panel Clock", "2026-06-20T19:15:21", "Pending Problems", "Zone battery is low", "OPEN"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("body does not contain %q: %s", want, rec.Body.String())
		}
	}
}

func encodedTestCookie(t *testing.T, conn PanelConnection) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	setConnectionCookie(rec, conn)
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies len = %d", len(cookies))
	}
	return cookies[0]
}
