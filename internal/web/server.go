package web

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/andrebires/amt8000-pro/internal/isecnet"
)

type StatusClient interface {
	GetStatus() (isecnet.PanelStatus, error)
}

type Server struct {
	client StatusClient
}

func NewServer(client StatusClient) *Server {
	return &Server{client: client}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /api/status", s.handleStatus)
	return mux
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.client.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	status, err := s.client.GetStatus()
	data := struct {
		Status isecnet.PanelStatus
		Error  string
	}{Status: status}
	if err != nil {
		data.Error = err.Error()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pageTemplate.Execute(w, data)
}

var pageTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>AMT 8000 Pro</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: #f7f7f4;
      --fg: #1f2428;
      --muted: #687076;
      --line: #d9ddd8;
      --panel: #ffffff;
      --accent: #087f5b;
      --warn: #b42318;
    }
    @media (prefers-color-scheme: dark) {
      :root { --bg:#111411; --fg:#f1f4ef; --muted:#a8b0a6; --line:#30372f; --panel:#181d18; --accent:#58d68d; --warn:#ff8a7a; }
    }
    * { box-sizing: border-box; }
    body { margin: 0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: var(--bg); color: var(--fg); }
    header { border-bottom: 1px solid var(--line); padding: 20px; }
    main { max-width: 1120px; margin: 0 auto; padding: 20px; }
    h1 { margin: 0; font-size: 24px; }
    h2 { margin: 0 0 12px; font-size: 18px; }
    .muted { color: var(--muted); }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 12px; margin: 16px 0; }
    .card { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; padding: 16px; }
    .metric { font-size: 28px; font-weight: 700; }
    .ok { color: var(--accent); }
    .bad { color: var(--warn); }
    table { width: 100%; border-collapse: collapse; font-size: 14px; }
    th, td { border-bottom: 1px solid var(--line); padding: 8px; text-align: left; }
    th { color: var(--muted); font-weight: 600; }
    .error { border-color: var(--warn); color: var(--warn); }
  </style>
</head>
<body>
  <header>
    <h1>AMT 8000 Pro</h1>
    <div class="muted">Local LAN status console</div>
  </header>
  <main>
    {{if .Error}}
      <section class="card error">
        <h2>Panel connection failed</h2>
        <p>{{.Error}}</p>
      </section>
    {{else}}
      <section class="grid">
        <div class="card"><h2>State</h2><div class="metric">{{.Status.State}}</div></div>
        <div class="card"><h2>Firmware</h2><div class="metric">{{.Status.Version}}</div><div class="muted">Model 0x{{printf "%x" .Status.Model}}</div></div>
        <div class="card"><h2>Battery</h2><div class="metric">{{.Status.Battery}}</div></div>
        <div class="card"><h2>Siren</h2><div class="metric {{if .Status.SirenLive}}bad{{else}}ok{{end}}">{{if .Status.SirenLive}}Live{{else}}Quiet{{end}}</div></div>
      </section>
      <section class="card">
        <h2>Partitions</h2>
        <table>
          <thead><tr><th>#</th><th>Armed</th><th>Stay</th><th>Fired</th><th>Firing</th></tr></thead>
          <tbody>
            {{range .Status.Partitions}}<tr><td>{{.Index}}</td><td>{{.Armed}}</td><td>{{.Stay}}</td><td>{{.Fired}}</td><td>{{.Firing}}</td></tr>{{end}}
          </tbody>
        </table>
      </section>
      <section class="card" style="margin-top:12px">
        <h2>Zones</h2>
        <table>
          <thead><tr><th>#</th><th>Open</th><th>Violated</th><th>Bypassed</th><th>Tamper</th><th>Low Battery</th></tr></thead>
          <tbody>
            {{range .Status.Zones}}<tr><td>{{.Index}}</td><td>{{.Open}}</td><td>{{.Violated}}</td><td>{{.Bypassed}}</td><td>{{.Tamper}}</td><td>{{.LowBattery}}</td></tr>{{end}}
          </tbody>
        </table>
      </section>
    {{end}}
  </main>
</body>
</html>`))
