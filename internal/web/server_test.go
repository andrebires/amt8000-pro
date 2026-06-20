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
