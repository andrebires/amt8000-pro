package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

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

func (fakeClient) GetEvents() (isecnet.PanelEvents, error) {
	partitionOne := 1
	zoneTwo := 2
	userThree := 3
	return isecnet.PanelEvents{
		Limit: isecnet.EventLimit,
		Total: 2,
		Events: []isecnet.PanelEvent{
			{
				Index:          1,
				Timestamp:      "2026-06-20T19:20:00",
				Code:           "E130",
				Description:    "Zone alarm",
				Partition:      &partitionOne,
				Zone:           &zoneTwo,
				DeliveryStatus: "sent",
			},
			{
				Index:              2,
				Timestamp:          "2026-06-20T19:21:00",
				Code:               "R401",
				Description:        "User disarm",
				Partition:          &partitionOne,
				User:               &userThree,
				DeliveryStatus:     "blocked",
				ReceptorIPBlocked:  true,
				ReceptorIPDisabled: true,
			},
		},
	}, nil
}

type failingEventsClient struct {
	fakeClient
}

func (failingEventsClient) GetEvents() (isecnet.PanelEvents, error) {
	return isecnet.PanelEvents{}, errors.New("panel read failed")
}

type serializedClient struct {
	mu        sync.Mutex
	active    int
	maxActive int
}

func (c *serializedClient) GetStatus() (isecnet.PanelStatus, error) {
	c.enter()
	defer c.exit()
	return fakeClient{}.GetStatus()
}

func (c *serializedClient) GetEvents() (isecnet.PanelEvents, error) {
	c.enter()
	defer c.exit()
	return fakeClient{}.GetEvents()
}

func (c *serializedClient) enter() {
	c.mu.Lock()
	c.active++
	if c.active > c.maxActive {
		c.maxActive = c.active
	}
	c.mu.Unlock()
	time.Sleep(20 * time.Millisecond)
}

func (c *serializedClient) exit() {
	c.mu.Lock()
	c.active--
	c.mu.Unlock()
}

func (c *serializedClient) max() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.maxActive
}

func TestStatusEndpoint(t *testing.T) {
	server := NewServer(func(PanelConnection) PanelClient { return fakeClient{} })
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
	server := NewServer(func(PanelConnection) PanelClient { return fakeClient{} })
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
	server := NewServer(func(PanelConnection) PanelClient { return fakeClient{} })
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
	server := NewServer(func(PanelConnection) PanelClient { return fakeClient{} })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(encodedTestCookie(t, PanelConnection{Host: "192.168.1.50", Port: 9009, Password: "878787"}))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	for _, want := range []string{"Connection elapsed", "Last refresh", "Panel Clock", "2026-06-20T19:15:21", "Pending Problems", "Events", "CSV", "JSON", "Zone battery is low", "OPEN"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("body does not contain %q: %s", want, rec.Body.String())
		}
	}
	if strings.Contains(rec.Body.String(), "Receptor IP") {
		t.Fatalf("events table should not show redundant Receptor IP column: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "updateEventExportLinks();\n    refreshEvents();") {
		t.Fatalf("events should not auto-download on page load: %s", rec.Body.String())
	}
}

func TestEventsEndpoint(t *testing.T) {
	server := NewServer(func(PanelConnection) PanelClient { return fakeClient{} })
	req := httptest.NewRequest(http.MethodGet, "/api/events?blocked=true", nil)
	req.AddCookie(encodedTestCookie(t, PanelConnection{Host: "192.168.1.50", Port: 9009, Password: "878787"}))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{`"total":1`, `"code":"R401"`, `"receptorIpBlocked":true`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body does not contain %q: %s", want, body)
		}
	}
	if strings.Contains(body, `"code":"E130"`) {
		t.Fatalf("filtered event leaked into body: %s", body)
	}
}

func TestEventsEndpointSortsNewestFirst(t *testing.T) {
	server := NewServer(func(PanelConnection) PanelClient { return fakeClient{} })
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.AddCookie(encodedTestCookie(t, PanelConnection{Host: "192.168.1.50", Port: 9009, Password: "878787"}))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	newest := strings.Index(body, `"code":"R401"`)
	older := strings.Index(body, `"code":"E130"`)
	if newest < 0 || older < 0 || newest > older {
		t.Fatalf("events are not newest first: %s", body)
	}
}

func TestPanelCommandsAreSerializedPerConnection(t *testing.T) {
	client := &serializedClient{}
	server := NewServer(func(PanelConnection) PanelClient { return client })
	handler := server.Routes()
	conn := PanelConnection{Host: "192.168.1.50", Port: 9009, Password: "878787"}
	cookie := encodedTestCookie(t, conn)
	paths := []string{"/api/status", "/api/events", "/api/status", "/api/events"}
	var wg sync.WaitGroup

	for _, path := range paths {
		path := path
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.AddCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("%s status = %d body=%s", path, rec.Code, rec.Body.String())
			}
		}()
	}
	wg.Wait()

	if client.max() != 1 {
		t.Fatalf("panel commands overlapped; max active = %d", client.max())
	}
}

func TestEventsEndpointReportsDownloadFailure(t *testing.T) {
	server := NewServer(func(PanelConnection) PanelClient { return failingEventsClient{} })
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.AddCookie(encodedTestCookie(t, PanelConnection{Host: "192.168.1.50", Port: 9009, Password: "878787"}))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "EVENT_DOWNLOAD_FAILED") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestEventsCSVExport(t *testing.T) {
	server := NewServer(func(PanelConnection) PanelClient { return fakeClient{} })
	req := httptest.NewRequest(http.MethodGet, "/api/events/export?format=csv&q=zone", nil)
	req.AddCookie(encodedTestCookie(t, PanelConnection{Host: "192.168.1.50", Port: 9009, Password: "878787"}))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/csv") {
		t.Fatalf("Content-Type = %q, want text/csv", contentType)
	}
	body := rec.Body.String()
	for _, want := range []string{"index,timestamp,code,description,partition,zone,user,delivery_status,receptor_ip_blocked,receptor_ip_disabled,raw", "1,2026-06-20T19:20:00,E130,Zone alarm,1,2,,sent,false,false,"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body does not contain %q: %s", want, body)
		}
	}
}

func TestEventsJSONExport(t *testing.T) {
	server := NewServer(func(PanelConnection) PanelClient { return fakeClient{} })
	req := httptest.NewRequest(http.MethodGet, "/api/events/export?format=json&delivery=sent", nil)
	req.AddCookie(encodedTestCookie(t, PanelConnection{Host: "192.168.1.50", Port: 9009, Password: "878787"}))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"code":"E130"`) || strings.Contains(body, `"code":"R401"`) {
		t.Fatalf("unexpected body: %s", body)
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
