package metrics

import "encoding/json"

type metrics struct {
	Version string `json:",omitempty"`

	OEM *struct {
		Vendor  string
		Product string
	} `json:",omitempty"`
	BIOS *struct {
		Vendor  string
		Version string
	} `json:",omitempty"`
	CPU        []cpuInfo    `json:",omitempty"`
	GPU        []gpuInfo    `json:",omitempty"`
	RAM        string       `json:",omitempty"`
	Partitions []string     `json:",omitempty"`
	Screens    []screenInfo `json:",omitempty"`

	Autologin string `json:",omitempty"`
	LivePatch string `json:",omitempty"`
	Session   *struct {
		DE   string
		Name string
		Type string
	} `json:",omitempty"`
	Timezone string `json:",omitempty"`

	Install *json.RawMessage `json:",omitempty"`
	Upgrade *json.RawMessage `json:",omitempty"`
}

type gpuInfo struct {
	Vendor string
	Model  string
}

type screenInfo struct {
	Resolution string
	Frequence  string
}

type cpuInfo struct {
	Vendor   string
	Family   string
	Model    string
	Stepping string
}
