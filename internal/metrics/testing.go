package metrics

import "os/exec"

// GetenvFn is used to mock os.Getenv() for testing only
type GetenvFn func(key string) string

// NewTestMetrics create a full mocking testing metrics element.
// This is only a mock for testing, not for real use
func NewTestMetrics(root string,
	cmdGPU *exec.Cmd, cmdScreen *exec.Cmd, cmdPartition *exec.Cmd,
	getenv GetenvFn) Metrics {
	// do not use helper as in _test.go package
	return Metrics{
		root:          root,
		gpuInfoCmd:    cmdGPU,
		screenInfoCmd: cmdScreen,
		spaceInfoCmd:  cmdPartition,
		getenv:        getenv,
	}
}
