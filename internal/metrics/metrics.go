package metrics

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/ubuntu-report/internal/utils"
)

const (
	installerLogsPath = "var/log/installer/telemetry"
	upgradeLogsPath   = "var/log/upgrade/telemetry"
)

// Metrics collect system, upgrade and installer data
type Metrics struct {
	root          string
	screenInfoCmd *exec.Cmd
	spaceInfoCmd  *exec.Cmd
	gpuInfoCmd    *exec.Cmd
}

// New return a new metrics element with optional testing functions
func New(options ...func(*Metrics) error) (Metrics, error) {
	m := Metrics{
		root:          "/",
		screenInfoCmd: setCommand("xrandr"),
		spaceInfoCmd:  setCommand("df", "-h"),
		gpuInfoCmd:    setCommand("lspci", "-n"),
	}

	for _, options := range options {
		if err := options(&m); err != nil {
			return m, err
		}
	}

	return m, nil
}

func setCommand(cmds ...string) *exec.Cmd {
	if len(cmds) == 1 {
		return exec.Command(cmds[0])
	}
	return exec.Command(cmds[0], cmds[1:]...)
}

// Collect system, installer and update info, returning a json formatted byte
func (m Metrics) Collect() ([]byte, error) {
	log.Debugf("Collecting metrics on system with root set to %s", m.root)
	r := metrics{}

	r.Version = getVersion(m.root)

	vendor, product := getOEM(m.root)
	r.OEM = struct {
		Vendor  string
		Product string
	}{vendor, product}
	vendor, version := getBIOS(m.root)
	r.BIOS = struct {
		Vendor  string
		Version string
	}{vendor, version}

	r.CPU = getCPUInfo(m.root)
	r.GPU = m.getGPU()
	r.RAM = getRAM(m.root)
	r.Partitions = m.getPartitions()
	r.Screens = m.getScreensInfo()

	r.Autologin = getAutologin(m.root)
	// TODO: LivePatch
	r.Session = struct {
		DE   string
		Name string
		Type string
	}{
		os.Getenv("XDG_CURRENT_DESKTOP"),
		os.Getenv("XDG_SESSION_DESKTOP"),
		os.Getenv("XDG_SESSION_TYPE")}
	r.Timezone = getTimeZone(m.root)

	r.Install = installerInfo(m.root)
	r.Upgrade = upgradeInfo(m.root)

	d, err := json.Marshal(r)
	return d, errors.Wrapf(err, "can't be converted to a valid json")
}

func installerInfo(root string) *json.RawMessage {
	b, err := getFromFile(filepath.Join(root, installerLogsPath))
	if err != nil {
		log.Infof("no installer data found: "+utils.ErrFormat, err)
		b = []byte("{}")
	}
	if !json.Valid(b) {
		log.Infof("installer data found, but not valid json.")
		b = []byte("{}")
	}
	c := json.RawMessage(b)
	json.Valid(b)
	return &c
}

func upgradeInfo(root string) *json.RawMessage {
	b, err := getFromFile(filepath.Join(root, upgradeLogsPath))
	if err != nil {
		log.Infof("no upgrade data found: "+utils.ErrFormat, err)
		b = []byte("{}")
	}
	if !json.Valid(b) {
		log.Infof("upgrade data found, but not valid json.")
		b = []byte("{}")
	}
	c := json.RawMessage(b)
	return &c
}
