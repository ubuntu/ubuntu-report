package metrics

import "encoding/json"

type metrics struct {
	Version string `json:",omitempty"`

	OEM *struct {
		Vendor  string
		Product string
		DCD     string `json:",omitempty"`
	} `json:",omitempty"`
	BIOS *struct {
		Vendor  string
		Version string
	} `json:",omitempty"`
	CPU        *cpuInfo     `json:",omitempty"`
	Arch       string       `json:",omitempty"`
	GPU        []gpuInfo    `json:",omitempty"`
	RAM        *float64     `json:",omitempty"`
	Partitions []float64    `json:",omitempty"`
	Screens    []screenInfo `json:",omitempty"`

	Autologin *bool `json:",omitempty"`
	LivePatch *bool `json:",omitempty"`
	Session   *struct {
		DE   string
		Name string
		Type string
	} `json:",omitempty"`
	Language string `json:",omitempty"`
	Timezone string `json:",omitempty"`

	Install json.RawMessage `json:",omitempty"`
	Upgrade json.RawMessage `json:",omitempty"`
}

type gpuInfo struct {
	Vendor string
	Model  string
}

type screenInfo struct {
	Size       string
	Resolution string
	Frequency  string
}

type cpuInfo struct {
	OpMode             string
	CPUs               string
	Threads            string
	Cores              string
	Sockets            string
	Vendor             string
	Family             string
	Model              string
	Stepping           string
	Name               string
	Virtualization     string `json:",omitempty"`
	Hypervisor         string `json:",omitempty"`
	VirtualizationType string `json:",omitempty"`
}
