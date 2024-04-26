package daemon

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ubuntu/ubuntu-report/internal/consts"
	"github.com/ubuntu/ubuntu-report/internal/metrics"
)

// RotateLog is a publicly facing function to rotate the log file. It's called on SIGHUP.
func (d *Daemon) RotateLog() error {
	return d.initializeLogFile()
}

// initializeLogFile creates a new log file and closes previously opened one if necessary.
func (d *Daemon) initializeLogFile() error {
	var err error
	d.logMutex.Lock()
	defer d.logMutex.Unlock()

	if d.logFile != nil {
		slog.Debug("Closing log file")
		d.logFile.Close()
	}

	logpath := filepath.Join(d.logDir, consts.DefaultCollectorLogFilename)
	d.logFile, err = os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	slog.Debug(fmt.Sprintf("Opening log file %q", logpath))
	if err != nil {
		return err
	}

	return nil
}

// writeToLogFile dumps the record to the currently open log file
func (d *Daemon) writeToLogFile(reject bool, distro, variant, version string, data metrics.MetricsData) error {
	d.logMutex.Lock()
	defer d.logMutex.Unlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	currentTime := time.Now().Format(time.RFC3339)

	rejectStr := "OK"
	if reject {
		rejectStr = "REJ"

	}

	line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
		rejectStr,
		currentTime,
		distro,
		variant,
		version,
		strings.ReplaceAll(string(jsonData), "\n", ""))

	slog.Debug(fmt.Sprintf("Writing data to %s", d.logFile.Name()))
	_, err = d.logFile.WriteString(line + "\n")
	if err != nil {
		return err
	}

	return nil
}
