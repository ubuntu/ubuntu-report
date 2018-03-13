package metrics

import (
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/ubuntu-report/internal/utils"
)

func getGPU() []gpuInfo {
	var gpus []gpuInfo

	r := runCmd("lspci", "-n")

	results, err := filterAll(r, `^.* 0300: (.*) \(rev .*\)$`)
	if err != nil {
		log.Infof("couldn't get GPU info: "+utils.ErrorFormat, err)
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

func getScreensInfo() []screenInfo {
	var screens []screenInfo

	r := runCmd("xrandr")

	results, err := filterAll(r, `^ +(.*)\*\+$`)
	if err != nil {
		log.Infof("couldn't get Screen info: "+utils.ErrorFormat, err)
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

func getPartitions() []string {
	var sizes []string

	r := runCmd("df", "-h")

	results, err := filterAll(r, `^/dev/([^\s]+ +[^\s]*).*$`)
	if err != nil {
		log.Infof("couldn't get Disk info: "+utils.ErrorFormat, err)
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
		sizes = append(sizes, s[1])
	}

	return sizes
}

func runCmd(cmds ...string) io.Reader {
	var cmd *exec.Cmd
	if len(cmds) == 1 {
		cmd = exec.Command(cmds[0])
	} else {
		cmd = exec.Command(cmds[0], cmds[1:]...)
	}

	pr, pw := io.Pipe()
	cmd.Stdout = pw

	go func() {
		err := cmd.Run()
		if err != nil {
			pw.CloseWithError(errors.Wrapf(err, "Running %s return an error", cmds))
			return
		}
		pw.Close()
	}()
	return pr
}
