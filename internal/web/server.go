package web

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/andrebires/amt8000-pro/internal/isecnet"
)

const sessionCookieName = "amt8000_panel"

type PanelConnection struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

type StatusClient interface {
	GetStatus() (isecnet.PanelStatus, error)
}

type ClientFactory func(PanelConnection) StatusClient

type Server struct {
	newClient ClientFactory
}

func NewServer(newClient ClientFactory) *Server {
	return &Server{newClient: newClient}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /login", s.handleLogin)
	mux.HandleFunc("POST /login", s.handleLoginPost)
	mux.HandleFunc("POST /logout", s.handleLogout)
	mux.HandleFunc("GET /api/status", s.handleStatus)
	return mux
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	conn, ok := connectionFromRequest(r)
	if !ok {
		http.Error(w, "login required", http.StatusUnauthorized)
		return
	}
	status, err := s.newClient(conn).GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	conn, ok := connectionFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	status, err := s.newClient(conn).GetStatus()
	data := struct {
		Connection PanelConnection
		Status     isecnet.PanelStatus
		Error      string
	}{Connection: conn, Status: status}
	if err != nil {
		data.Error = err.Error()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pageTemplate.Execute(w, data)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	data := loginPageData{Port: 9009}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = loginTemplate.Execute(w, data)
}

func (s *Server) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderLogin(w, loginPageData{Port: 9009, Error: "Invalid form submission."})
		return
	}

	port, err := strconv.Atoi(strings.TrimSpace(r.FormValue("port")))
	if err != nil || port <= 0 || port > 65535 {
		renderLogin(w, loginPageData{Host: r.FormValue("host"), Port: 9009, Error: "Port must be between 1 and 65535."})
		return
	}
	conn := PanelConnection{
		Host:     strings.TrimSpace(r.FormValue("host")),
		Port:     port,
		Password: strings.TrimSpace(r.FormValue("password")),
	}
	if err := validateConnection(conn); err != nil {
		renderLogin(w, loginPageData{Host: conn.Host, Port: conn.Port, Error: err.Error()})
		return
	}

	status, err := s.newClient(conn).GetStatus()
	if err != nil {
		renderLogin(w, loginPageData{Host: conn.Host, Port: conn.Port, Error: "Panel connection failed: " + err.Error()})
		return
	}
	setConnectionCookie(w, conn)
	_ = status
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

type loginPageData struct {
	Host  string
	Port  int
	Error string
}

func renderLogin(w http.ResponseWriter, data loginPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_ = loginTemplate.Execute(w, data)
}

func validateConnection(conn PanelConnection) error {
	if conn.Host == "" {
		return errors.New("IP address is required")
	}
	if conn.Port <= 0 || conn.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if conn.Password == "" {
		return errors.New("remote password is required")
	}
	return nil
}

func connectionFromRequest(r *http.Request) (PanelConnection, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return PanelConnection{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return PanelConnection{}, false
	}
	var conn PanelConnection
	if err := json.Unmarshal(raw, &conn); err != nil {
		return PanelConnection{}, false
	}
	if validateConnection(conn) != nil {
		return PanelConnection{}, false
	}
	return conn, true
}

func setConnectionCookie(w http.ResponseWriter, conn PanelConnection) {
	raw, _ := json.Marshal(conn)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    base64.RawURLEncoding.EncodeToString(raw),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
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
    h3 { margin: 0 0 8px; font-size: 15px; }
    .muted { color: var(--muted); }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 12px; margin: 16px 0; }
    .card { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; padding: 16px; }
    .metric { font-size: 28px; font-weight: 700; }
    .ok { color: var(--accent); }
    .bad { color: var(--warn); }
    .toolbar { display:flex; align-items:center; justify-content:space-between; gap:12px; flex-wrap:wrap; margin:16px 0; }
    .statusline { display:flex; gap:16px; flex-wrap:wrap; font-size:14px; }
    .section-grid { display:grid; grid-template-columns: minmax(0, 1fr); gap:12px; }
    .trouble-list { margin:0; padding-left:18px; }
    .pill { display:inline-flex; align-items:center; min-height:24px; padding:2px 8px; border:1px solid var(--line); border-radius:999px; font-size:12px; font-weight:700; }
    table { width: 100%; border-collapse: collapse; font-size: 14px; }
    th, td { border-bottom: 1px solid var(--line); padding: 8px; text-align: left; }
    th { color: var(--muted); font-weight: 600; }
    .error { border-color: var(--warn); color: var(--warn); }
    button { min-height:36px; border:1px solid var(--line); border-radius:6px; padding:6px 12px; font:inherit; font-weight:700; background:transparent; color:inherit; cursor:pointer; }
  </style>
</head>
<body>
  <header>
    <h1>AMT 8000 Pro</h1>
    <div class="muted">Connected to {{.Connection.Host}}:{{.Connection.Port}}</div>
    <form method="post" action="/logout" style="margin-top:12px">
      <button type="submit">Log out</button>
    </form>
  </header>
  <main>
    {{if .Error}}
      <section class="card error">
        <h2>Panel connection failed</h2>
        <p>{{.Error}}</p>
      </section>
    {{else}}
      <section class="toolbar">
        <div class="statusline">
          <span>Connection elapsed: <strong id="elapsed">0s</strong></span>
          <span>Last refresh: <strong id="last-refresh">initial load</strong></span>
          <span id="refresh-error" class="bad"></span>
        </div>
        <button type="button" id="refresh-button">Refresh</button>
      </section>
      <section class="grid">
        <div class="card"><h2>State</h2><div class="metric" id="state">{{.Status.State}}</div></div>
        <div class="card"><h2>Firmware</h2><div class="metric" id="version">{{.Status.Version}}</div><div class="muted">Model <span id="model">0x{{printf "%x" .Status.Model}}</span></div></div>
        <div class="card"><h2>Panel Clock</h2><div class="metric" id="panel-date-time">{{if .Status.PanelDateTime}}{{.Status.PanelDateTime}}{{else}}unsupported{{end}}</div></div>
        <div class="card"><h2>Battery</h2><div class="metric" id="battery">{{.Status.Battery}}</div><div class="muted">Voltage <span id="battery-voltage">unsupported</span></div></div>
        <div class="card"><h2>Source</h2><div class="metric" id="source-voltage">unsupported</div><div class="muted">AC/source voltage</div></div>
        <div class="card"><h2>Siren</h2><div class="metric {{if .Status.SirenLive}}bad{{else}}ok{{end}}" id="siren">{{if .Status.SirenLive}}Live{{else}}Quiet{{end}}</div></div>
        <div class="card"><h2>Troubles</h2><div class="metric" id="trouble-count">{{len .Status.Troubles}}</div><div class="muted">Known derived problems</div></div>
      </section>
      <div class="section-grid">
        <section class="card">
          <h2>Partitions</h2>
          <table>
            <thead><tr><th>#</th><th>State</th><th>Armed</th><th>Stay</th><th>Fired</th><th>Firing</th></tr></thead>
            <tbody id="partitions">
              {{range .Status.Partitions}}<tr><td>{{.Index}}</td><td><span class="pill">{{.State}}</span></td><td>{{.Armed}}</td><td>{{.Stay}}</td><td>{{.Fired}}</td><td>{{.Firing}}</td></tr>{{end}}
            </tbody>
          </table>
        </section>
        <section class="card">
          <h2>Zones</h2>
          <table>
            <thead><tr><th>#</th><th>State</th><th>Open</th><th>Fired</th><th>Bypassed</th><th>Tamper</th><th>Low Battery</th></tr></thead>
            <tbody id="zones">
              {{range .Status.Zones}}<tr><td>{{.Index}}</td><td><span class="pill">{{.State}}</span></td><td>{{.Open}}</td><td>{{.Violated}}</td><td>{{.Bypassed}}</td><td>{{.Tamper}}</td><td>{{.LowBattery}}</td></tr>{{end}}
            </tbody>
          </table>
        </section>
        <section class="card">
          <h2>Pending Problems</h2>
          <ul class="trouble-list" id="troubles">
            {{range .Status.Troubles}}<li>{{.Message}}{{if .Zone}} - Zone {{.Zone}}{{end}}</li>{{else}}<li class="muted">No known derived problems.</li>{{end}}
          </ul>
        </section>
      </div>
    {{end}}
  </main>
  <script>
    const startedAt = Date.now();
    const elapsedEl = document.getElementById("elapsed");
    const refreshEl = document.getElementById("last-refresh");
    const errorEl = document.getElementById("refresh-error");
    const refreshButton = document.getElementById("refresh-button");

    function boolText(value) {
      return value ? "true" : "false";
    }

    function voltageText(value) {
      return value === null || value === undefined ? "unsupported" : Number(value).toFixed(2) + " V";
    }

    function setText(id, value) {
      const el = document.getElementById(id);
      if (el) el.textContent = value;
    }

    function renderStatus(status) {
      setText("state", status.state);
      setText("version", status.version);
      setText("model", "0x" + Number(status.model).toString(16));
      setText("panel-date-time", status.panelDateTime || "unsupported");
      setText("battery", status.battery);
      setText("battery-voltage", voltageText(status.batteryVoltage));
      setText("source-voltage", voltageText(status.sourceVoltage));
      setText("siren", status.sirenLive ? "Live" : "Quiet");
      document.getElementById("siren")?.classList.toggle("bad", status.sirenLive);
      document.getElementById("siren")?.classList.toggle("ok", !status.sirenLive);
      setText("trouble-count", status.troubles ? status.troubles.length : 0);

      const partitions = document.getElementById("partitions");
      if (partitions) {
        partitions.innerHTML = (status.partitions || []).map((partition) => "<tr><td>" + partition.index + "</td><td><span class=\"pill\">" + partition.state + "</span></td><td>" + boolText(partition.armed) + "</td><td>" + boolText(partition.stay) + "</td><td>" + boolText(partition.fired) + "</td><td>" + boolText(partition.firing) + "</td></tr>").join("");
      }

      const zones = document.getElementById("zones");
      if (zones) {
        zones.innerHTML = (status.zones || []).map((zone) => "<tr><td>" + zone.index + "</td><td><span class=\"pill\">" + zone.state + "</span></td><td>" + boolText(zone.open) + "</td><td>" + boolText(zone.violated) + "</td><td>" + boolText(zone.bypassed) + "</td><td>" + boolText(zone.tamper) + "</td><td>" + boolText(zone.lowBattery) + "</td></tr>").join("");
      }

      const troubles = document.getElementById("troubles");
      if (troubles) {
        if (!status.troubles || status.troubles.length === 0) {
          troubles.innerHTML = "<li class=\"muted\">No known derived problems.</li>";
        } else {
          troubles.innerHTML = status.troubles.map((trouble) => "<li>" + trouble.message + (trouble.zone ? " - Zone " + trouble.zone : "") + "</li>").join("");
        }
      }
    }

    async function refreshStatus() {
      if (!refreshButton) return;
      refreshButton.disabled = true;
      try {
        const response = await fetch("/api/status", {headers: {"Accept": "application/json"}});
        if (!response.ok) throw new Error(await response.text());
        renderStatus(await response.json());
        errorEl.textContent = "";
        refreshEl.textContent = new Date().toLocaleTimeString();
      } catch (error) {
        errorEl.textContent = "Refresh failed: " + String(error.message || error).trim();
      } finally {
        refreshButton.disabled = false;
      }
    }

    function updateElapsed() {
      if (!elapsedEl) return;
      const total = Math.floor((Date.now() - startedAt) / 1000);
      const minutes = Math.floor(total / 60);
      const seconds = total % 60;
      elapsedEl.textContent = minutes > 0 ? minutes + "m " + seconds + "s" : seconds + "s";
    }

    refreshButton?.addEventListener("click", refreshStatus);
    setInterval(updateElapsed, 1000);
    setInterval(refreshStatus, 10000);
    updateElapsed();
  </script>
</body>
</html>`))

var loginTemplate = template.Must(template.New("login").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Connect AMT 8000 Pro</title>
  <style>
    :root { color-scheme: light dark; --bg:#f7f7f4; --fg:#1f2428; --muted:#687076; --line:#d9ddd8; --panel:#fff; --warn:#b42318; }
    @media (prefers-color-scheme: dark) { :root { --bg:#111411; --fg:#f1f4ef; --muted:#a8b0a6; --line:#30372f; --panel:#181d18; --warn:#ff8a7a; } }
    * { box-sizing: border-box; }
    body { margin:0; min-height:100vh; display:grid; place-items:center; font-family:ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background:var(--bg); color:var(--fg); padding:20px; }
    main { width:min(100%, 420px); background:var(--panel); border:1px solid var(--line); border-radius:8px; padding:20px; }
    h1 { margin:0 0 6px; font-size:24px; }
    p { margin:0 0 20px; color:var(--muted); }
    label { display:block; margin:14px 0 6px; font-weight:600; }
    input { width:100%; min-height:42px; border:1px solid var(--line); border-radius:6px; padding:8px 10px; font:inherit; background:transparent; color:inherit; }
    button { width:100%; min-height:42px; margin-top:18px; border:0; border-radius:6px; font:inherit; font-weight:700; background:#087f5b; color:white; cursor:pointer; }
    .error { color:var(--warn); border:1px solid var(--warn); border-radius:6px; padding:10px; margin-bottom:14px; }
  </style>
</head>
<body>
  <main>
    <h1>Connect Panel</h1>
    <p>Enter the AMT 8000 Pro local IP and remote access password.</p>
    {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
    <form method="post" action="/login">
      <label for="host">Panel IP</label>
      <input id="host" name="host" inputmode="decimal" autocomplete="off" required value="{{.Host}}">
      <label for="port">Port</label>
      <input id="port" name="port" inputmode="numeric" autocomplete="off" required value="{{.Port}}">
      <label for="password">Remote password</label>
      <input id="password" name="password" type="password" inputmode="numeric" autocomplete="current-password" required>
      <button type="submit">Connect</button>
    </form>
  </main>
</body>
</html>`))
