package metrics

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/ubuntu-report/internal/utils"
)

func (m Metrics) getVersion() string {
	v, err := matchFromFile(filepath.Join(m.root, "etc/os-release"), `^VERSION_ID="(.*)"$`, false)
	if err != nil {
		log.Infof("couldn't get version information from os-release: "+utils.ErrFormat, err)
		return ""
	}
	return v
}

func (m Metrics) getRAM() *float64 {
	s, err := matchFromFile(filepath.Join(m.root, "proc/meminfo"), `^MemTotal: +(\d+) kB$`, false)
	if err != nil {
		log.Infof("couldn't get RAM information from meminfo: "+utils.ErrFormat, err)
		return nil
	}
	v, err := convKBToGB(s)
	if err != nil {
		log.Infof("partition size should be an integer: "+utils.ErrFormat, err)
		return nil
	}
	return &v
}

func (m Metrics) getTimeZone() string {
	v, err := getFromFileTrimmed(filepath.Join(m.root, "etc/timezone"))
	if err != nil {
		log.Infof("couldn't get timezone information: "+utils.ErrFormat, err)
		return ""
	}
	if strings.Contains(v, "\n") {
		log.Infof(utils.ErrFormat, errors.Errorf("malformed timezone information, file contains: %s", v))
		return ""
	}
	return v
}

func (m Metrics) getAutologin() bool {
	v, err := matchFromFile(filepath.Join(m.root, "etc/gdm3/custom.conf"), `^AutomaticLoginEnable ?= ?(.*)$`, true)
	if err != nil {
		log.Infof("couldn't get autologin information from gdm: "+utils.ErrFormat, err)
		return false
	}
	if strings.ToLower(v) != "true" {
		return false
	}
	return true
}

func (m Metrics) getOEM() (string, string) {
	v, err := getFromFileTrimmed(filepath.Join(m.root, "sys/class/dmi/id/chassis_vendor"))
	if err != nil {
		log.Infof("couldn't get chassis vendor information: "+utils.ErrFormat, err)
	}
	if strings.Contains(v, "\n") {
		log.Infof(utils.ErrFormat, errors.Errorf("malformed chassis vendor information, file contains: %s", v))
		v = ""
	}
	p, err := getFromFileTrimmed(filepath.Join(m.root, "sys/class/dmi/id/product_name"))
	if err != nil {
		log.Infof("couldn't get chassis product name information: "+utils.ErrFormat, err)
	}
	if strings.Contains(p, "\n") {
		log.Infof(utils.ErrFormat, errors.Errorf("malformed chassis product name information, file contains: %s", p))
		p = ""
	}
	return v, p
}

func (m Metrics) getCPU() []cpuInfo {
	indexedCPUInfo := make(map[string]cpuInfo)

	p := filepath.Join(m.root, "proc/cpuinfo")
	f, err := os.Open(p)
	if err != nil {
		err = errors.Wrapf(err, "couldn't open %s", p)
		log.Infof(utils.ErrFormat, err)
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	i := cpuInfo{}
	var physicalID string
	for scanner.Scan() {
		t := scanner.Text()

		fields := strings.Split(t, ":")
		if len(fields) > 2 {
			log.Debug("fields are expected to have one element only")
			continue
		} else if len(fields) < 2 {
			if (i != cpuInfo{}) {
				// we only store the CPU info once per physical unit (multiple core)
				indexedCPUInfo[physicalID] = i
			}
			i = cpuInfo{}
			continue
		}
		v := strings.TrimSpace(fields[1])

		switch strings.TrimSpace(fields[0]) {
		case "vendor_id":
			i.Vendor = v
		case "cpu family":
			i.Family = v
		case "model":
			i.Model = v
		case "stepping":
			i.Stepping = v
		case "physical id":
			physicalID = v
		}
	}
	// Store last cpu
	if (i != cpuInfo{}) {
		indexedCPUInfo[physicalID] = i
	}

	var r []cpuInfo
	for _, v := range indexedCPUInfo {
		r = append(r, v)
	}
	if len(r) < 1 {
		log.Infof("Didn't find any CPU info in %s", p)
	}
	return r
}

func (m Metrics) getBIOS() (string, string) {
	vd, err := getFromFileTrimmed(filepath.Join(m.root, "sys/class/dmi/id/bios_vendor"))
	if err != nil {
		log.Infof("couldn't get bios vendor information: "+utils.ErrFormat, err)
		vd = ""
	}
	if strings.Contains(vd, "\n") {
		log.Infof(utils.ErrFormat, errors.Errorf("malformed bios vendor information, file contains: %s", vd))
		vd = ""
	}
	ve, err := getFromFileTrimmed(filepath.Join(m.root, "sys/class/dmi/id/bios_version"))
	if err != nil {
		log.Infof("couldn't get bios version: "+utils.ErrFormat, err)
		ve = ""
	}
	if strings.Contains(ve, "\n") {
		log.Infof(utils.ErrFormat, errors.Errorf("malformed bios version information, file contains: %s", ve))
		ve = ""
	}
	return vd, ve
}

func (m Metrics) getLivePatch() bool {
	if _, err := os.Stat(filepath.Join(m.root, "var/snap/canonical-livepatch/common/machine-token")); err != nil {
		return false
	}
	return true
}

func (m Metrics) installerInfo() *json.RawMessage {
	return getAndValidateJSONFromFile(filepath.Join(m.root, installerLogsPath), "install")
}

func (m Metrics) upgradeInfo() *json.RawMessage {
	return getAndValidateJSONFromFile(filepath.Join(m.root, upgradeLogsPath), "upgrade")
}

func matchFromFile(p, regex string, notFoundOk bool) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", errors.Wrapf(err, "couldn't open %s", p)
	}
	defer f.Close()

	return filterFirst(f, regex, notFoundOk)
}

func getFromFile(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't open %s", p)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read %s", p)
	}

	return b, nil
}

func getFromFileTrimmed(p string) (string, error) {
	b, err := getFromFile(p)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func getAndValidateJSONFromFile(p string, errmsg string) *json.RawMessage {
	b, err := getFromFile(p)
	if err != nil {
		log.Infof("no %s data found: "+utils.ErrFormat, errmsg, err)
		return nil
	}
	if !json.Valid(b) {
		log.Infof("%s data found, but not valid json.", errmsg)
		return nil
	}
	c := json.RawMessage(b)
	return &c
}
