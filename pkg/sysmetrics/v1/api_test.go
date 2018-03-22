package sysmetrics_test

import (
	"strings"
	"testing"

	"github.com/ubuntu/ubuntu-report/pkg/sysmetrics/v1"
)

func TestCollect(t *testing.T) {
	t.Parallel()

	data, err := sysmetrics.Collect()

	if err != nil {
		t.Fatal("we didn't expect an error and got one", err)
	}

	if !strings.Contains(string(data), sysmetrics.ExpectedReportItem) {
		t.Errorf("we expected at least %s in output, got: '%s", sysmetrics.ExpectedReportItem, string(data))
	}
}
