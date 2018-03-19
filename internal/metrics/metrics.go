package metrics

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
	getenv        func(string) string
}

// New return a new metrics element with optional testing functions
func New(options ...func(*Metrics) error) (Metrics, error) {
	m := Metrics{
		root:          "/",
		screenInfoCmd: setCommand("xrandr"),
		spaceInfoCmd:  setCommand("df", "-h"),
		gpuInfoCmd:    setCommand("lspci", "-n"),
		getenv:        os.Getenv,
	}

	for _, options := range options {
		if err := options(&m); err != nil {
			return m, err
		}
	}

	return m, nil
}

// GetIDS returns distro and version information
func (m Metrics) GetIDS() (string, string, error) {
	p := filepath.Join(m.root, "etc", "os-release")
	f, err := os.Open(p)
	if err != nil {
		return "", "", errors.Wrapf(err, "couldn't open %s", p)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	dRe := regexp.MustCompile(`^ID=(.*)$`)
	vRe := regexp.MustCompile(`^VERSION_ID="(.*)"$`)
	var distro, version string
	for scanner.Scan() {
		v := dRe.FindStringSubmatch(scanner.Text())
		if v != nil {
			distro = strings.TrimSpace(v[1])
		}
		v = vRe.FindStringSubmatch(scanner.Text())
		if v != nil {
			version = strings.TrimSpace(v[1])
		}
	}

	if err := scanner.Err(); (distro == "" || version == "") && err != nil {
		return "", "", errors.Wrap(err, "error while scanning")
	}

	if distro == "" || version == "" {
		return "", "", errors.Errorf("distribution '%s' or version '%s' information missing", distro, version)
	}

	return distro, version, nil
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

	r.Version = m.getVersion()

	if vendor, product := m.getOEM(); vendor != "" || product != "" {
		r.OEM = &struct {
			Vendor  string
			Product string
		}{vendor, product}
	}
	if vendor, version := m.getBIOS(); vendor != "" || version != "" {
		r.BIOS = &struct {
			Vendor  string
			Version string
		}{vendor, version}
	}

	r.CPU = m.getCPU()
	r.GPU = m.getGPU()
	r.RAM = m.getRAM()
	r.Partitions = m.getPartitions()
	r.Screens = m.getScreens()

	a := m.getAutologin()
	r.Autologin = &a
	l := m.getLivePatch()
	r.LivePatch = &l

	de := m.getenv("XDG_CURRENT_DESKTOP")
	sessionName := m.getenv("XDG_SESSION_DESKTOP")
	sessionType := m.getenv("XDG_SESSION_TYPE")
	if de != "" || sessionName != "" || sessionType != "" {
		r.Session = &struct {
			DE   string
			Name string
			Type string
		}{de, sessionName, sessionType}
	}
	r.Timezone = m.getTimeZone()

	r.Install = m.installerInfo()
	r.Upgrade = m.upgradeInfo()

	d, err := json.Marshal(r)
	return d, errors.Wrapf(err, "can't be converted to a valid json")
}

func convKBToGB(s string) (float64, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.Wrapf(err, "couldn't convert %s to an integer", s)
	}
	// convert in GB (SI) and round it to 0.1
	f := float64(v) / (1000 * 1000)
	return math.Round(f*10) / 10, nil
}
