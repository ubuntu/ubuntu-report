package metrics

import (
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
	v, err := getFromFileTrimmed(filepath.Join(m.root, "sys/class/dmi/id/sys_vendor"))
	if err != nil {
		log.Infof("couldn't get sys vendor information: "+utils.ErrFormat, err)
	}
	if strings.Contains(v, "\n") {
		log.Infof(utils.ErrFormat, errors.Errorf("malformed sys vendor information, file contains: %s", v))
		v = ""
	}
	p, err := getFromFileTrimmed(filepath.Join(m.root, "sys/class/dmi/id/product_name"))
	if err != nil {
		log.Infof("couldn't get sys product name information: "+utils.ErrFormat, err)
	}
	if strings.Contains(p, "\n") {
		log.Infof(utils.ErrFormat, errors.Errorf("malformed sys product name information, file contains: %s", p))
		p = ""
	}
	return v, p
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

func (m Metrics) installerInfo() json.RawMessage {
	return getAndValidateJSONFromFile(filepath.Join(m.root, installerLogsPath), "install")
}

func (m Metrics) upgradeInfo() json.RawMessage {
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

func getAndValidateJSONFromFile(p string, errmsg string) json.RawMessage {
	b, err := getFromFile(p)
	if err != nil {
		log.Infof("no %s data found: "+utils.ErrFormat, errmsg, err)
		return nil
	}
	if !json.Valid(b) {
		log.Infof("%s data found, but not valid json.", errmsg)
		return nil
	}
	return json.RawMessage(b)
}
