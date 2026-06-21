package web

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/andrebires/amt8000-pro/internal/isecnet"
)

type eventFilters struct {
	Query     string
	Partition *int
	Delivery  string
	Blocked   *bool
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	conn, ok := connectionFromRequest(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "login required")
		return
	}
	events, err := s.newClient(conn).GetEvents()
	if err != nil {
		writeEventsError(w, err)
		return
	}
	events = filterEvents(events, eventFiltersFromRequest(r))
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleEventsExport(w http.ResponseWriter, r *http.Request) {
	conn, ok := connectionFromRequest(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "login required")
		return
	}
	events, err := s.newClient(conn).GetEvents()
	if err != nil {
		writeEventsError(w, err)
		return
	}
	events = filterEvents(events, eventFiltersFromRequest(r))

	switch strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format"))) {
	case "", "csv":
		writeEventsCSV(w, events)
	case "json":
		w.Header().Set("Content-Disposition", `attachment; filename="amt8000-events.json"`)
		writeJSON(w, http.StatusOK, events)
	default:
		writeJSONError(w, http.StatusBadRequest, "INVALID_EXPORT_FORMAT", "format must be csv or json")
	}
}

func writeEventsError(w http.ResponseWriter, err error) {
	writeJSONError(w, http.StatusBadGateway, "EVENT_DOWNLOAD_FAILED", err.Error())
}

func eventFiltersFromRequest(r *http.Request) eventFilters {
	query := r.URL.Query()
	filters := eventFilters{
		Query:    strings.TrimSpace(query.Get("q")),
		Delivery: strings.TrimSpace(query.Get("delivery")),
	}
	if raw := strings.TrimSpace(query.Get("partition")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			filters.Partition = &value
		}
	}
	if raw := strings.TrimSpace(query.Get("blocked")); raw != "" {
		if value, err := strconv.ParseBool(raw); err == nil {
			filters.Blocked = &value
		}
	}
	return filters
}

func filterEvents(events isecnet.PanelEvents, filters eventFilters) isecnet.PanelEvents {
	out := events
	out.Events = make([]isecnet.PanelEvent, 0, len(events.Events))
	query := strings.ToLower(filters.Query)
	delivery := strings.ToLower(filters.Delivery)
	for _, event := range events.Events {
		if query != "" && !eventMatchesQuery(event, query) {
			continue
		}
		if filters.Partition != nil && (event.Partition == nil || *event.Partition != *filters.Partition) {
			continue
		}
		if delivery != "" && strings.ToLower(event.DeliveryStatus) != delivery {
			continue
		}
		if filters.Blocked != nil && event.ReceptorIPBlocked != *filters.Blocked {
			continue
		}
		out.Events = append(out.Events, event)
	}
	sort.SliceStable(out.Events, func(i, j int) bool {
		if out.Events[i].Timestamp == out.Events[j].Timestamp {
			return out.Events[i].Index > out.Events[j].Index
		}
		return out.Events[i].Timestamp > out.Events[j].Timestamp
	})
	out.Total = len(out.Events)
	if out.Limit == 0 {
		out.Limit = isecnet.EventLimit
	}
	return out
}

func eventMatchesQuery(event isecnet.PanelEvent, query string) bool {
	parts := []string{
		event.Timestamp,
		event.Code,
		event.Description,
		event.DeliveryStatus,
		event.Raw,
		strconv.Itoa(event.Index),
	}
	if event.Partition != nil {
		parts = append(parts, strconv.Itoa(*event.Partition))
	}
	if event.Zone != nil {
		parts = append(parts, strconv.Itoa(*event.Zone))
	}
	if event.User != nil {
		parts = append(parts, strconv.Itoa(*event.User))
	}
	return strings.Contains(strings.ToLower(strings.Join(parts, " ")), query)
}

func writeEventsCSV(w http.ResponseWriter, events isecnet.PanelEvents) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="amt8000-events.csv"`)
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{
		"index",
		"timestamp",
		"code",
		"description",
		"partition",
		"zone",
		"user",
		"delivery_status",
		"receptor_ip_blocked",
		"receptor_ip_disabled",
		"raw",
	})
	for _, event := range events.Events {
		_ = writer.Write([]string{
			strconv.Itoa(event.Index),
			event.Timestamp,
			event.Code,
			event.Description,
			optionalInt(event.Partition),
			optionalInt(event.Zone),
			optionalInt(event.User),
			event.DeliveryStatus,
			strconv.FormatBool(event.ReceptorIPBlocked),
			strconv.FormatBool(event.ReceptorIPDisabled),
			event.Raw,
		})
	}
	writer.Flush()
}

func optionalInt(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeJSONError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]string{
		"code":  code,
		"error": message,
	})
}
