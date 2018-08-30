package metrics

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
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

func (m Metrics) getOEM() (string, string, string) {
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
	dcd, err := matchFromFile(filepath.Join(m.root, "var/lib/ubuntu_dist_channel"), `^([^\s#]+)$`, true)
	if err != nil {
		log.Infof("no DCD information: "+utils.ErrFormat, err)
	}
	return v, p, dcd
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

func (m Metrics) getDisks() []float64 {
	var sizes []float64

	blockFolder := filepath.Join(m.root, "sys/block")
	dirs, err := ioutil.ReadDir(blockFolder)
	if err != nil {
		log.Infof("couldn't get disk block information: "+utils.ErrFormat, err)
		return nil
	}

	for _, d := range dirs {
		if !(strings.HasPrefix(d.Name(), "hd") || strings.HasPrefix(d.Name(), "sd") || strings.HasPrefix(d.Name(), "vd")) {
			continue
		}

		v, err := getFromFileTrimmed(filepath.Join(blockFolder, d.Name(), "size"))
		if err != nil {
			log.Infof("couldn't get disk block information for %s: "+utils.ErrFormat, d.Name(), err)
			continue
		}
		s, err := strconv.Atoi(v)
		if err != nil {
			log.Infof("number of block for disk %s isn't an integer: "+utils.ErrFormat, d.Name(), err)
			continue
		}

		v, err = getFromFileTrimmed(filepath.Join(blockFolder, d.Name(), "queue/logical_block_size"))
		if err != nil {
			log.Infof("couldn't get disk block information for %s: "+utils.ErrFormat, d.Name(), err)
			continue
		}
		bs, err := strconv.Atoi(v)
		if err != nil {
			log.Infof("block size for disk %s isn't an integer: "+utils.ErrFormat, d.Name(), err)
			continue
		}

		// convert in Gib in .1 precision
		size := float64(s) * float64(bs) / (1000 * 1000 * 1000)
		size = math.Round(size*10) / 10

		sizes = append(sizes, size)
	}

	return sizes
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
