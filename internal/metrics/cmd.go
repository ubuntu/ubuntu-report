package metrics

import (
	"io"
	"os/exec"
	"strings"
	"bytes"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/ubuntu-report/internal/utils"
)

func (m Metrics) getGPU() []gpuInfo {
	var gpus []gpuInfo

	r := runCmd(m.gpuInfoCmd)

	results, err := filterAll(r, `^.* 0300: ([a-zA-Z0-9]+:[a-zA-Z0-9]+)( \(rev .*\))?$`)
	if err != nil {
		log.Infof("couldn't get GPU info: "+utils.ErrFormat, err)
		return nil
	}

	for _, gpuinfo := range results {
		i := strings.SplitN(gpuinfo, ":", 2)
		if len(i) != 2 {
			log.Infof("GPU info should of form vendor:model, got: %s", gpuinfo)
			continue
		}
		gpus = append(gpus, gpuInfo{Vendor: i[0], Model: i[1]})
	}

	return gpus
}

func (m Metrics) getCPU() cpuInfo {
	c := cpuInfo{}

	r := runCmd(m.cpuInfoCmd)

	for result := range filter(r, `{"field": *"(.*)", *"data": *"(.*)"},`, true) {
		if result.err != nil {
			log.Infof("Couldn't get CPU info: "+utils.ErrFormat, result.err)
			return cpuInfo{}
		}

		key, v := result.r[0], result.r[1]

		switch strings.TrimSpace(key) {
		case "CPU op-mode(s):":
			c.OpMode = v
		case "CPU(s):":
			c.CPUs = v
		case "Thread(s) per core:":
			c.Threads = v
		case "Core(s) per socket:":
			c.Cores = v
		case "Socket(s):":
			c.Sockets = v
		case "Vendor ID:":
			c.Vendor = v
		case "CPU family:":
			c.Family = v
		case "Model:":
			c.Model = v
		case "Stepping:":
			c.Stepping = v
		case "Model name:":
			c.Name = v
		case "Virtualization:":
			c.Virtualization = v
		case "Hypervisor vendor:":
			c.Hypervisor = v
		case "Virtualization type:":
			c.VirtualizationType = v
		}
	}

	return c
}

func (m Metrics) getScreens() []screenInfo {
	var screens []screenInfo

	r := runCmd(m.screenInfoCmd)

	var results []string
	results, err := filterAll(r, `^(?: +(.*)\*|.* connected .* (\d+mm x \d+mm))`)
	if err != nil {
		log.Infof("couldn't get Screen info: "+utils.ErrFormat, err)
		return nil
	}

	var lastSize string
	for _, screeninfo := range results {
		if strings.Index(screeninfo, "mm") > -1 {
			lastSize = strings.Replace(screeninfo, " ", "", -1)
			continue
		}
		i := strings.Fields(screeninfo)
		if len(i) < 2 {
			log.Infof("screen info should be either a screen physical size (connected) or a a resolution + freq, got: %s", screeninfo)
			continue
		}
		if lastSize == "" {
			log.Infof("We couldn't get physical info size prior to Resolution and Frequency information.")
			continue
		}
		screens = append(screens, screenInfo{Size: lastSize, Resolution: i[0], Frequency: i[len(i)-1]})
	}

	return screens
}

func (m Metrics) getPartitions() []float64 {
	var sizes []float64

	r := runCmd(m.spaceInfoCmd)

	results, err := filterAll(r, `^/dev/([^\s]+ +[^\s]*).*$`)
	if err != nil {
		log.Infof("couldn't get Disk info: "+utils.ErrFormat, err)
		return nil
	}

	for _, size := range results {
		// negative lookahead isn't supported in go, so exclude loop devices manually
		if strings.HasPrefix(size, "loop") {
			continue
		}
		s := strings.Fields(size)
		if len(s) != 2 {
			log.Infof("partition size should be of form 'block device      size', got: %s", size)
			continue
		}
		v, err := convKBToGB(s[1])
		if err != nil {
			log.Infof("partition size should be an integer: "+utils.ErrFormat, err)
			continue
		}
		sizes = append(sizes, v)
	}

	return sizes
}

func (m Metrics) getArch() string {
	b, err := m.archCmd.CombinedOutput()
	if err != nil {
		log.Infof("couldn't get Architecture: "+utils.ErrFormat, err)
		return ""
	}

	return strings.TrimSpace(string(b))
}

func (m Metrics) getHwCap() string {
	// Make sure we have glibc version > 2.33
	r := runCmd(m.libc6Cmd)
	result, err := filterFirst(r, `^(?: +Installed: (.*))`, false)
	if err != nil {
		log.Infof("Couldn't get glibc version: "+utils.ErrFormat, err)
		return ""
	}
	// This successfully handles a version of (none)
	if strings.Compare(result, "2.33") < 0 {
		return "N/A"
	}

	// Find the hwcap
	rSupported := runCmd(m.hwCapCmd)

	// remove the legacy hwcap section, as we don't want to report those
	legacyBytes := []byte("Legacy HWCAP subdirectories")
	bytesSupported, err := io.ReadAll(rSupported)
	if err != nil {
		log.Infof("Couldn't get hwcap: "+utils.ErrFormat, err)
		return ""
	}
	legacyIndex := bytes.Index(bytesSupported, legacyBytes)
	if legacyIndex >= 0 {
		bytesSupported = bytesSupported[0:legacyIndex]
	}
	// convert back to io.Reader for the filter functions
	newSupported := bytes.NewReader(bytesSupported)

	// now find which version is supported
	resultSupported, err := filterFirst(newSupported, `^(?:(.*) +.*supported, searched.*)`, false)
	if err != nil {
		log.Infof("Couldn't get hwcap: "+utils.ErrFormat, err)
		return ""
	}

	return resultSupported
}

func runCmd(cmd *exec.Cmd) io.Reader {
	pr, pw := io.Pipe()
	cmd.Stdout = pw

	go func() {
		err := cmd.Run()
		if err != nil {
			pw.CloseWithError(errors.Wrapf(err, "'%s' return an error", cmd.Args))
			return
		}
		pw.Close()
	}()
	return pr
}
