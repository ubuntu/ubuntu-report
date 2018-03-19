package metrics

import (
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/ubuntu-report/internal/utils"
)

func (m Metrics) getGPU() []gpuInfo {
	var gpus []gpuInfo

	r := runCmd(m.gpuInfoCmd)

	results, err := filterAll(r, `^.* 0300: (.*) \(rev .*\)$`)
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

func (m Metrics) getScreens() []screenInfo {
	var screens []screenInfo

	r := runCmd(m.screenInfoCmd)

	results, err := filterAll(r, `^ +(.*)\*\+$`)
	if err != nil {
		log.Infof("couldn't get Screen info: "+utils.ErrFormat, err)
		return nil
	}

	for _, screeninfo := range results {
		i := strings.Fields(screeninfo)
		if len(i) != 2 {
			log.Infof("screen info should be of form 'resolution     freq*+', got: %s", screeninfo)
			continue
		}
		screens = append(screens, screenInfo{Resolution: i[0], Frequence: i[1]})
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
