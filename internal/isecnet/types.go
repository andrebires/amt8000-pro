package isecnet

type PanelStatus struct {
	Model          byte        `json:"model"`
	Version        string      `json:"version"`
	State          string      `json:"state"`
	PanelDateTime  string      `json:"panelDateTime,omitempty"`
	SirenLive      bool        `json:"sirenLive"`
	ZonesFiring    bool        `json:"zonesFiring"`
	ZonesClosed    bool        `json:"zonesClosed"`
	Battery        string      `json:"battery"`
	BatteryVoltage *float64    `json:"batteryVoltage"`
	SourceVoltage  *float64    `json:"sourceVoltage"`
	Tamper         bool        `json:"tamper"`
	Troubles       []Trouble   `json:"troubles"`
	Partitions     []Partition `json:"partitions"`
	Zones          []Zone      `json:"zones"`
}

type Partition struct {
	Index   int    `json:"index"`
	Enabled bool   `json:"enabled"`
	State   string `json:"state"`
	Stay    bool   `json:"stay"`
	Fired   bool   `json:"fired"`
	Firing  bool   `json:"firing"`
	Armed   bool   `json:"armed"`
}

type Zone struct {
	Index      int    `json:"index"`
	State      string `json:"state"`
	Open       bool   `json:"open"`
	Violated   bool   `json:"violated"`
	Bypassed   bool   `json:"bypassed"`
	Tamper     bool   `json:"tamper"`
	LowBattery bool   `json:"lowBattery"`
}

type Trouble struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Zone    int    `json:"zone,omitempty"`
}

type StatusCapture struct {
	Status PanelStatus `json:"status"`
	Frame  Frame       `json:"-"`
}
