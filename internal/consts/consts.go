// Package consts defines the constants used by the project.
package consts

import "log/slog"

var (
	// Version is the version of the executable.
	Version = "Dev"
)

const (

	// DefaultLevelLog is the default logging level selected without any option.
	DefaultLevelLog = slog.LevelWarn

	// DefaultCollectorConfPath is the default path to the configuration file of the collector service
	DefaultCollectorConfPath    = "/etc/ubuntu-report/"
	DefaultCollectorLogFilename = "metrics.log"
)
