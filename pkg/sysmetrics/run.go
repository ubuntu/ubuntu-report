package sysmetrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/ubuntu-report/internal/metrics"
	"github.com/ubuntu/ubuntu-report/internal/sender"
	"github.com/ubuntu/ubuntu-report/internal/utils"
)

// optOutJSON is the data sent in case of Opt-Out choice
const optOutJSON = `{"OptOut": true}`

var (
	initialReportTimeoutDuration = 30 * time.Second
)

func metricsCollect(m metrics.Metrics) ([]byte, error) {
	data, err := m.Collect()
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't collect system minimal info")
	}

	log.Debug("pretty print format the collected data to the user")
	h := json.RawMessage(data)
	return json.MarshalIndent(&h, "", "  ")
}

func metricsSend(m metrics.Metrics, data []byte, acknowledgement, alwaysReport bool, baseURL string, reportBasePath string, in io.Reader, out io.Writer) error {
	distro, version, err := m.GetIDS()
	if err != nil {
		return errors.Wrapf(err, "couldn't get mandatory information")
	}

	reportP, err := checkPreviousReport(distro, version, reportBasePath, alwaysReport)
	if err != nil {
		return err
	}

	// erase potential collected data
	if !acknowledgement {
		data = []byte(optOutJSON)
	}

	if baseURL == "" {
		baseURL = sender.BaseURL
	}
	u, err := sender.GetURL(baseURL, distro, version)
	if err != nil {
		return errors.Wrapf(err, "report destination url is invalid")
	}
	if err := sender.Send(u, data); err != nil {
		returnErr := errors.Wrapf(err, "data were not delivered successfully to metrics server, saving for a later automated report")
		p, err := utils.PendingReportPath(reportBasePath)
		if err != nil {
			return errors.Wrapf(err, "couldn't get where pending reported metrics should be stored on disk: %v", returnErr)
		}
		if err := saveMetrics(p, data); err != nil {
			return errors.Wrapf(err, "couldn't save pending reported are on disk: %v", returnErr)
		}
		return returnErr
	}

	return saveMetrics(reportP, data)
}

func metricsCollectAndSend(m metrics.Metrics, r ReportType, alwaysReport bool, baseURL string, reportBasePath string, in io.Reader, out io.Writer) error {
	distro, version, err := m.GetIDS()
	if err != nil {
		return errors.Wrapf(err, "couldn't get mandatory information")
	}

	if _, err := checkPreviousReport(distro, version, reportBasePath, alwaysReport); err != nil {
		return err
	}

	var data []byte
	if r != ReportOptOut {
		if data, err = metricsCollect(m); err != nil {
			return errors.Wrapf(err, "couldn't collect system minimal info and format it")
		}
	}

	sendMetrics := true
	if r == ReportInteractive {
		fmt.Fprintln(out, "This is the result of hardware and optional installer/upgrader that we collected:")
		fmt.Fprintln(out, string(data))

		validAnswer := false
		scanner := bufio.NewScanner(in)
		for validAnswer != true {
			fmt.Fprintf(out, "Do you agree to report this? [y (send metrics)/n (send opt out message)/Q (quit)] ")
			if !scanner.Scan() {
				log.Info("programm interrupted")
				return nil
			}
			text := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if text == "n" || text == "no" {
				log.Debug("sending report was denied")
				sendMetrics = false
				validAnswer = true
			} else if text == "y" || text == "yes" {
				log.Debug("sending report was accepted")
				sendMetrics = true
				validAnswer = true
			} else if text == "q" || text == "quit" || text == "" {
				return nil
			}
			if validAnswer != true {
				log.Error("we didn't understand your answer")
			}
		}
	} else if r == ReportAuto {
		log.Debug("auto report requested")
		sendMetrics = true
	} else {
		log.Debug("opt-out report requested")
		sendMetrics = false
	}

	return metricsSend(m, data, sendMetrics, alwaysReport, baseURL, reportBasePath, in, out)
}

func saveMetrics(p string, data []byte) error {
	log.Debugf("save sent metrics to %s", p)

	d := filepath.Dir(p)
	if err := os.MkdirAll(d, 0700); err != nil {
		return errors.Wrap(err, "couldn't create parent directory to save reported metrics")
	}

	if err := ioutil.WriteFile(p, data, 0666); err != nil {
		return errors.Wrap(err, "couldn't save reported or pending metrics on disk")
	}

	return nil
}

func checkPreviousReport(distro, version, reportBasePath string, alwaysReport bool) (string, error) {
	p, err := utils.ReportPath(distro, version, reportBasePath)
	if err != nil {
		return "", errors.Wrapf(err, "couldn't get where to save reported metrics on disk")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		log.Infof("previous report found in %s", p)
		if !alwaysReport {
			return "", errors.Errorf("metrics from this machine have already been reported and can be found in: %s", p)
		}
		log.Debug("ignore previous report requested")
	}
	return p, nil
}

func metricsSendPendingReport(m metrics.Metrics, baseURL, reportBasePath string, in io.Reader, out io.Writer) error {
	distro, version, err := m.GetIDS()
	if err != nil {
		return errors.Wrapf(err, "couldn't get mandatory information")
	}

	reportP, err := utils.ReportPath(distro, version, reportBasePath)
	if err != nil {
		return errors.Wrapf(err, "couldn't get where to save reported metrics on disk")
	}

	pending, err := utils.PendingReportPath(reportBasePath)
	if err != nil {
		return errors.Wrapf(err, "couldn't get where to previous reported metrics are on disk")
	}
	data, err := ioutil.ReadFile(pending)
	if err != nil {
		return errors.Wrapf(err, "no pending report found")
	}

	if baseURL == "" {
		baseURL = sender.BaseURL
	}
	u, err := sender.GetURL(baseURL, distro, version)
	if err != nil {
		return errors.Wrapf(err, "report destination url is invalid")
	}

	wait := time.Duration(initialReportTimeoutDuration)
	for {
		if err := sender.Send(u, data); err != nil {
			log.Errorf("data were not delivered successfully to metrics server, retrying in %ds", wait/(1000*1000*1000))
			time.Sleep(wait)
			wait = wait * 2
			if wait > time.Duration(30*time.Minute) {
				wait = time.Duration(30 * time.Minute)
			}
			continue
		}
		break
	}

	if err := os.Remove(pending); err != nil {
		return errors.Wrapf(err, "couldn't remove pending report after a successful report")
	}
	return saveMetrics(reportP, data)
}
