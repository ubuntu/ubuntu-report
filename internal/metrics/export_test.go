package metrics

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/ubuntu-report/internal/helper"
)

// WithRootAt tweaks the root directory of the file system
func WithRootAt(p string) func(*Metrics) error {
	log.Debugf("Setting root directory to %s", p)
	return func(m *Metrics) error {
		m.root = p
		return nil
	}
}

// WithCPUInfoCommand tweaks the default cpu info command
func WithCPUInfoCommand(cmd *exec.Cmd) func(*Metrics) error {
	log.Debugf("Setting cpu info command to '%s'", cmd.Args)
	return func(m *Metrics) error {
		m.cpuInfoCmd = cmd
		return nil
	}
}

// WithGPUInfoCommand tweaks the default gpu info command
func WithGPUInfoCommand(cmd *exec.Cmd) func(*Metrics) error {
	log.Debugf("Setting gpu info command to '%s'", cmd.Args)
	return func(m *Metrics) error {
		m.gpuInfoCmd = cmd
		return nil
	}
}

// WithScreenInfoCommand tweaks the default screen info command
func WithScreenInfoCommand(cmd *exec.Cmd) func(*Metrics) error {
	log.Debugf("Setting screen info command to '%s'", cmd.Args)
	return func(m *Metrics) error {
		m.screenInfoCmd = cmd
		return nil
	}
}

// WithSpaceInfoCommand tweaks the default space info command
func WithSpaceInfoCommand(cmd *exec.Cmd) func(*Metrics) error {
	log.Debugf("Setting space info command to '%s'", cmd.Args)
	return func(m *Metrics) error {
		m.spaceInfoCmd = cmd
		return nil
	}
}

// WithArchitectureCommand tweaks the current given architecture
func WithArchitectureCommand(cmd *exec.Cmd) func(*Metrics) error {
	log.Debugf("Setting architecture command to '%s'", cmd.Args)
	return func(m *Metrics) error {
		m.archCmd = cmd
		return nil
	}
}

// WithMapForEnv replace system getenv with given environ hashmap
func WithMapForEnv(env map[string]string) func(*Metrics) error {
	log.Debugf("Setting new environment to '%v'", env)
	return func(m *Metrics) error {
		m.getenv = helper.GetenvFromMap(env)
		return nil
	}
}
