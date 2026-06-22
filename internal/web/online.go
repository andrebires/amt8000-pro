package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andrebires/amt8000-pro/internal/isecnet"
)

type onlineCommandRequest struct {
	Confirm  bool   `json:"confirm"`
	Mode     string `json:"mode,omitempty"`
	Active   *bool  `json:"active,omitempty"`
	Bypassed *bool  `json:"bypassed,omitempty"`
	Time     string `json:"time,omitempty"`
}

type onlineCommandResponse struct {
	OK      bool                 `json:"ok"`
	Status  *isecnet.PanelStatus `json:"status,omitempty"`
	AuditID string               `json:"auditId"`
	Error   string               `json:"error,omitempty"`
}

func (s *Server) handleArmPartition(w http.ResponseWriter, r *http.Request) {
	partition, ok := parseOnlineIndex(w, r.PathValue("partition"), "partition", 0, 16)
	if !ok {
		return
	}
	req, ok := parseOnlineCommandRequest(w, r)
	if !ok {
		return
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "away"
	}
	if mode != "away" && mode != "stay" {
		writeOnlineResponse(w, http.StatusBadRequest, onlineCommandResponse{OK: false, Error: "mode must be away or stay"})
		return
	}
	s.runOnlineCommand(w, r, onlineCommandSpec{
		Action:         "arm_partition",
		Target:         fmt.Sprintf("partition:%d", partition),
		RequestedState: mode,
		Confirmed:      req.Confirm,
		Run: func(client PanelClient) (isecnet.PanelStatus, error) {
			return client.ArmPartition(partition, mode)
		},
	})
}

func (s *Server) handleDisarmPartition(w http.ResponseWriter, r *http.Request) {
	partition, ok := parseOnlineIndex(w, r.PathValue("partition"), "partition", 0, 16)
	if !ok {
		return
	}
	req, ok := parseOnlineCommandRequest(w, r)
	if !ok {
		return
	}
	s.runOnlineCommand(w, r, onlineCommandSpec{
		Action:         "disarm_partition",
		Target:         fmt.Sprintf("partition:%d", partition),
		RequestedState: "disarmed",
		Confirmed:      req.Confirm,
		Run: func(client PanelClient) (isecnet.PanelStatus, error) {
			return client.DisarmPartition(partition)
		},
	})
}

func (s *Server) handleClockSync(w http.ResponseWriter, r *http.Request) {
	req, ok := parseOnlineCommandRequest(w, r)
	if !ok {
		return
	}
	value := time.Now()
	requestedState := "server_time"
	if strings.TrimSpace(req.Time) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.Time))
		if err != nil {
			writeOnlineResponse(w, http.StatusBadRequest, onlineCommandResponse{OK: false, Error: "time must be RFC3339"})
			return
		}
		value = parsed
		requestedState = parsed.Format(time.RFC3339)
	}
	s.runOnlineCommand(w, r, onlineCommandSpec{
		Action:         "sync_panel_clock",
		Target:         "panel_clock",
		RequestedState: requestedState,
		Confirmed:      req.Confirm,
		Run: func(client PanelClient) (isecnet.PanelStatus, error) {
			return client.SetPanelDateTime(value)
		},
	})
}

func (s *Server) handleZoneBypass(w http.ResponseWriter, r *http.Request) {
	zone, ok := parseOnlineIndex(w, r.PathValue("zone"), "zone", 1, 64)
	if !ok {
		return
	}
	req, ok := parseOnlineCommandRequest(w, r)
	if !ok {
		return
	}
	if req.Bypassed == nil {
		writeOnlineResponse(w, http.StatusBadRequest, onlineCommandResponse{OK: false, Error: "bypassed is required"})
		return
	}
	requestedState := "unbypassed"
	if *req.Bypassed {
		requestedState = "bypassed"
	}
	s.runOnlineCommand(w, r, onlineCommandSpec{
		Action:         "set_zone_bypass",
		Target:         fmt.Sprintf("zone:%d", zone),
		RequestedState: requestedState,
		Confirmed:      req.Confirm,
		Run: func(client PanelClient) (isecnet.PanelStatus, error) {
			return client.SetZoneBypass(zone, *req.Bypassed)
		},
	})
}

func (s *Server) handleClearAlarmMemory(w http.ResponseWriter, r *http.Request) {
	req, ok := parseOnlineCommandRequest(w, r)
	if !ok {
		return
	}
	s.runOnlineCommand(w, r, onlineCommandSpec{
		Action:         "clear_alarm_memory",
		Target:         "alarm_memory",
		RequestedState: "cleared",
		Confirmed:      req.Confirm,
		Run: func(client PanelClient) (isecnet.PanelStatus, error) {
			return client.ClearAlarmMemory()
		},
	})
}

func (s *Server) handlePGM(w http.ResponseWriter, r *http.Request) {
	pgm, ok := parseOnlineIndex(w, r.PathValue("pgm"), "pgm", 1, 16)
	if !ok {
		return
	}
	req, ok := parseOnlineCommandRequest(w, r)
	if !ok {
		return
	}
	if req.Active == nil {
		writeOnlineResponse(w, http.StatusBadRequest, onlineCommandResponse{OK: false, Error: "active is required"})
		return
	}
	requestedState := "inactive"
	if *req.Active {
		requestedState = "active"
	}
	s.runOnlineCommand(w, r, onlineCommandSpec{
		Action:         "set_pgm",
		Target:         fmt.Sprintf("pgm:%d", pgm),
		RequestedState: requestedState,
		Confirmed:      req.Confirm,
		Run: func(client PanelClient) (isecnet.PanelStatus, error) {
			return client.SetPGM(pgm, *req.Active)
		},
	})
}

type onlineCommandSpec struct {
	Action         string
	Target         string
	RequestedState string
	Confirmed      bool
	Run            func(PanelClient) (isecnet.PanelStatus, error)
}

func (s *Server) runOnlineCommand(w http.ResponseWriter, r *http.Request, spec onlineCommandSpec) {
	conn, ok := connectionFromRequest(r)
	if !ok {
		writeOnlineResponse(w, http.StatusUnauthorized, onlineCommandResponse{OK: false, Error: "login required"})
		return
	}
	if !spec.Confirmed {
		auditID := s.auditOnlineCommand(conn, spec, "rejected", "confirmation required")
		writeOnlineResponse(w, http.StatusBadRequest, onlineCommandResponse{OK: false, AuditID: auditID, Error: "confirmation required"})
		return
	}

	status, err := s.executeOnlineCommand(conn, spec.Run)
	if err != nil {
		auditID := s.auditOnlineCommand(conn, spec, "failed", err.Error())
		statusCode := http.StatusBadGateway
		if errors.Is(err, isecnet.ErrOnlineCommandUnsupported) {
			statusCode = http.StatusNotImplemented
		}
		writeOnlineResponse(w, statusCode, onlineCommandResponse{OK: false, AuditID: auditID, Error: err.Error()})
		return
	}
	auditID := s.auditOnlineCommand(conn, spec, "succeeded", "")
	writeOnlineResponse(w, http.StatusOK, onlineCommandResponse{OK: true, Status: &status, AuditID: auditID})
}

func (s *Server) executeOnlineCommand(conn PanelConnection, run func(PanelClient) (isecnet.PanelStatus, error)) (isecnet.PanelStatus, error) {
	unlock := s.lockPanel(conn)
	defer unlock()
	return run(s.newClient(conn))
}

func (s *Server) auditOnlineCommand(conn PanelConnection, spec onlineCommandSpec, result string, errorText string) string {
	if s.auditSink == nil {
		s.auditSink = noopAuditSink{}
	}
	auditID, err := s.auditSink.WriteOnlineCommand(OnlineCommandAuditRecord{
		PanelHost:      conn.Host,
		PanelPort:      conn.Port,
		Action:         spec.Action,
		Target:         spec.Target,
		RequestedState: spec.RequestedState,
		Result:         result,
		Error:          errorText,
	})
	logAuditFailure(err)
	return auditID
}

func parseOnlineCommandRequest(w http.ResponseWriter, r *http.Request) (onlineCommandRequest, bool) {
	var req onlineCommandRequest
	if r.Body == nil {
		return req, true
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeOnlineResponse(w, http.StatusBadRequest, onlineCommandResponse{OK: false, Error: "invalid JSON body"})
		return onlineCommandRequest{}, false
	}
	return req, true
}

func parseOnlineIndex(w http.ResponseWriter, raw string, name string, min int, max int) (int, bool) {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < min || value > max {
		writeOnlineResponse(w, http.StatusBadRequest, onlineCommandResponse{OK: false, Error: fmt.Sprintf("%s must be between %d and %d", name, min, max)})
		return 0, false
	}
	return value, true
}

func writeOnlineResponse(w http.ResponseWriter, status int, response onlineCommandResponse) {
	writeJSON(w, status, response)
}
