package main

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/ubuntu-report/internal/metrics"
)

type reportType int

const (
	reportInteractive reportType = iota
	reportAuto
	reportOptOut
)

// Collect system info and return a pretty printed version of collected data
func Collect() ([]byte, error) {
	log.Debug("collect system information")

	m, err := metrics.New()
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't create a metric collector")
	}
	return metricsCollect(m)
}

// Report system info after collect
// reportType can be:
// - reportInteractive (will show report content on stdout and read anwser on stdin)
// - reportAuto (will report without printing report)
// - reportOptOut (will send opt-out message without printing report)
// - baseURL is alternative url to send the report to, if not empty
// alwaysReport forces a report even if a previous report was already done
func Report(r reportType, alwaysReport bool, baseURL string) error {
	log.Debug("collect and report system information")

	m, err := metrics.New()
	if err != nil {
		return errors.Wrapf(err, "couldn't create a metric collector")
	}
	return metricsReport(m, r, alwaysReport, baseURL)
}
