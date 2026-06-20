package isecnet

type PanelStatus struct {
	Model       byte        `json:"model"`
	Version     string      `json:"version"`
	State       string      `json:"state"`
	SirenLive   bool        `json:"sirenLive"`
	ZonesFiring bool        `json:"zonesFiring"`
	ZonesClosed bool        `json:"zonesClosed"`
	Battery     string      `json:"battery"`
	Tamper      bool        `json:"tamper"`
	Partitions  []Partition `json:"partitions"`
	Zones       []Zone      `json:"zones"`
}

type Partition struct {
	Index   int  `json:"index"`
	Enabled bool `json:"enabled"`
	Stay    bool `json:"stay"`
	Fired   bool `json:"fired"`
	Firing  bool `json:"firing"`
	Armed   bool `json:"armed"`
}

type Zone struct {
	Index      int  `json:"index"`
	Open       bool `json:"open"`
	Violated   bool `json:"violated"`
	Bypassed   bool `json:"bypassed"`
	Tamper     bool `json:"tamper"`
	LowBattery bool `json:"lowBattery"`
}
