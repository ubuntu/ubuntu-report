// Package sysmetrics Golang bindings: collect and report system and hardware metrics
// from your system.
package sysmetrics

import (
	"os"

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

// SendReport POST to the baseURL server data coming from a previous collect.
// The report will not be sent if a report has already been sent for this version unless "alwaysReport" is true.
// If "baseURL" is not an empty string, this overrides the server the report is sent to.
func SendReport(data []byte, alwaysReport bool, baseURL string) error {
	log.Debug("report system information")

	m, err := metrics.New()
	if err != nil {
		return errors.Wrapf(err, "couldn't create a metric collector")
	}
	return metricsSend(m, data, true, alwaysReport, baseURL, "", os.Stdin, os.Stdout)
}

// SendDecline POST to the baseURL server data denial report message.
// The denial message will not be sent if a report has already been sent for this version unless "alwaysReport" is true.
// If "baseURL" is not an empty string, this overrides the server the report is sent to.
func SendDecline(alwaysReport bool, baseURL string) error {
	log.Debug("report system information")

	m, err := metrics.New()
	if err != nil {
		return errors.Wrapf(err, "couldn't create a metric collector")
	}
	return metricsSend(m, nil, false, alwaysReport, baseURL, "", os.Stdin, os.Stdout)
}

// CollectAndSend gather system info and send them
// The report will not be sent if a report has already been sent for this version unless "alwaysReport" is true.
// If "baseURL" is not an empty string, this overrides the server the report is sent to.
func CollectAndSend(r ReportType, alwaysReport bool, baseURL string) error {
	log.Debug("collect and report system information")

	m, err := metrics.New()
	if err != nil {
		return errors.Wrapf(err, "couldn't create a metric collector")
	}
	return metricsCollectAndSend(m, r, alwaysReport, baseURL, "", os.Stdin, os.Stdout)
}

// SendPendingReport will try to send any pending report which didn't suceed previously due to network issues.
// It will try sending and exponentially back off until a send is successful.
func SendPendingReport(baseURL string) error {
	log.Debug("try sending previous report")

	m, err := metrics.New()
	if err != nil {
		return errors.Wrapf(err, "couldn't create a metric collector")
	}
	return metricsSendPendingReport(m, baseURL, "", os.Stdin, os.Stdout)
}
