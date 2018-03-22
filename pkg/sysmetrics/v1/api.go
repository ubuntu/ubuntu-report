package sysmetrics

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/ubuntu-report/internal/metrics"
)

// ReportType define the desired kind of interaction in CollectAndSend()
type ReportType int

const (
	// ReportInteractive will show report content on stdout and read anwser on stdin
	ReportInteractive ReportType = iota
	// ReportAuto will send a report without printing report
	ReportAuto
	// ReportOptOut will send opt-out message without printing report
	ReportOptOut
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

// CollectAndSend gather system info and send them
// alwaysReports bypass previous report already be sent for current version check
// It can be send to an alternative url via baseURL to send the report to, if not empty
func CollectAndSend(r ReportType, alwaysReport bool, baseURL string) error {
	log.Debug("collect and report system information")

	m, err := metrics.New()
	if err != nil {
		return errors.Wrapf(err, "couldn't create a metric collector")
	}
	return metricsReport(m, r, alwaysReport, baseURL, "")
}
