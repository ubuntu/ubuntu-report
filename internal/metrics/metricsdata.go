package metrics

import "encoding/json"

type metrics struct {
	Version string

	OEM struct {
		Vendor  string
		Product string
	}
	BIOS struct {
		Vendor  string
		Version string
	}
	CPU        []cpuInfo
	GPU        []gpuInfo
	RAM        string
	Partitions []string
	Screens    []screenInfo

	Autologin string
	LivePatch string
	Session   struct {
		DE   string
		Name string
		Type string
	}
	Timezone string

	Install *json.RawMessage
	Upgrade *json.RawMessage
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
