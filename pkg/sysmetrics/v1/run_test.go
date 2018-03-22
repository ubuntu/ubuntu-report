package sysmetrics

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ubuntu/ubuntu-report/internal/helper"
	"github.com/ubuntu/ubuntu-report/internal/metrics"
)

var Update = flag.Bool("update", false, "update golden files")

func TestMetricsCollect(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		root          string
		caseGPU       string
		caseScreen    string
		casePartition string
		env           map[string]string

		// note that only an internal json package error can make it returning an error
		wantErr bool
	}{
		{"regular",
			"testdata/good", "one gpu", "one screen", "one partition",
			map[string]string{"XDG_CURRENT_DESKTOP": "some:thing", "XDG_SESSION_DESKTOP": "ubuntusession", "XDG_SESSION_TYPE": "x12"},
			false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m, cancelGPU, cancelScreen, cancelPartition := newTestMetricsWithCommands(t, tc.root,
				tc.caseGPU, tc.caseScreen, tc.casePartition, tc.env)
			defer cancelGPU()
			defer cancelScreen()
			defer cancelPartition()
			b1, err1 := metricsCollect(m)

			want := helper.LoadOrUpdateGolden(t, filepath.Join(tc.root, "gold", "metricscollect"), b1, *Update)
			a.CheckWantedErr(err1, tc.wantErr)
			a.Equal(b1, want)

			// second run should return the same thing (idemnpotence)
			m, cancelGPU, cancelScreen, cancelPartition = newTestMetricsWithCommands(t, tc.root,
				tc.caseGPU, tc.caseScreen, tc.casePartition, tc.env)
			defer cancelGPU()
			defer cancelScreen()
			defer cancelPartition()
			b2, err2 := metricsCollect(m)

			a.CheckWantedErr(err2, tc.wantErr)
			var got1, got2 json.RawMessage
			json.Unmarshal(b1, got1)
			json.Unmarshal(b2, got2)
			a.Equal(got1, got2)
		})
	}
}

func TestMetricsReport(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		root            string
		caseGPU         string
		caseScreen      string
		casePartition   string
		env             map[string]string
		r               ReportType
		manualServerURL string

		cacheReportP    string
		shouldHitServer bool
		sHitHat         string
		wantErr         bool
	}{
		{"regular report auto",
			"testdata/good", "one gpu", "one screen", "one partition",
			map[string]string{"XDG_CURRENT_DESKTOP": "some:thing", "XDG_SESSION_DESKTOP": "ubuntusession", "XDG_SESSION_TYPE": "x12"},
			ReportAuto, "",
			"ubuntu-report/ubuntu.18.04", true, "/ubuntu/desktop/18.04", false},
		{"regular report OptOut",
			"testdata/good", "one gpu", "one screen", "one partition",
			map[string]string{"XDG_CURRENT_DESKTOP": "some:thing", "XDG_SESSION_DESKTOP": "ubuntusession", "XDG_SESSION_TYPE": "x12"},
			ReportOptOut, "",
			"ubuntu-report/ubuntu.18.04", true, "/ubuntu/desktop/18.04", false},
		{"No IDs (mandatory)",
			"testdata/no-ids", "", "", "", nil, ReportAuto, "", "ubuntu-report", false, "", true},
		{"Other URL",
			"testdata/good", "", "", "", nil, ReportAuto, "localhost:4299", "ubuntu-report", false, "", true},
		{"Invalid URL",
			"testdata/good", "", "", "", nil, ReportAuto, "http://a b.com/", "ubuntu-report", false, "", true},
		{"Unwritable path",
			"testdata/good", "", "", "", nil, ReportAuto, "", "/unwritable/cache/path", true, "/ubuntu/desktop/18.04", true},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m, cancelGPU, cancelScreen, cancelPartition := newTestMetricsWithCommands(t, tc.root,
				tc.caseGPU, tc.caseScreen, tc.casePartition, tc.env)
			defer cancelGPU()
			defer cancelScreen()
			defer cancelPartition()
			var out string
			out, tearDown := helper.TempDir(t)
			defer tearDown()
			if strings.HasPrefix(tc.cacheReportP, "/") {
				// absolute path, override temporary one
				out = tc.cacheReportP
			}
			serverHitAt := ""
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverHitAt = r.URL.String()
			}))
			defer ts.Close()
			url := tc.manualServerURL
			if url == "" {
				url = ts.URL
			}

			err := metricsReport(m, tc.r, false, url, out)

			a.CheckWantedErr(err, tc.wantErr)
			// check we didn't do too much work on error
			if err != nil {
				if !tc.shouldHitServer {
					a.Equal(serverHitAt, "")
				}
				if tc.shouldHitServer && serverHitAt == "" {
					t.Error("we should have hit the local server and it wasn't")
				}
				if _, err := os.Stat(filepath.Join(out, tc.cacheReportP)); !os.IsNotExist(err) {
					t.Errorf("we didn't expect finding a cache report path as we erroring out")
				}
				return
			}
			a.Equal(serverHitAt, tc.sHitHat)
			gotF, err := os.Open(filepath.Join(out, tc.cacheReportP))
			if err != nil {
				t.Fatal("didn't generate a report file on disk", err)
			}
			got, err := ioutil.ReadAll(gotF)
			if err != nil {
				t.Fatal("couldn't read generated report file", err)
			}
			want := helper.LoadOrUpdateGolden(t, filepath.Join(tc.root, "gold", fmt.Sprintf("cachereport.ReportType%d", int(tc.r))), got, *Update)
			a.Equal(got, want)
		})
	}
}

func newMockShortCmd(t *testing.T, s ...string) (*exec.Cmd, context.CancelFunc) {
	t.Helper()
	return helper.ShortProcess(t, "TestMetricsHelperProcess", s...)
}

func newTestMetricsWithCommands(t *testing.T, root, caseGPU, caseScreen, casePartition string, env map[string]string) (m metrics.Metrics,
	cancelGPU, cancelSreen, cancelPartition context.CancelFunc) {
	t.Helper()
	cmdGPU, cancelGPU := newMockShortCmd(t, "lspci", "-n", caseGPU)
	cmdScreen, cancelScreen := newMockShortCmd(t, "xrandr", caseScreen)
	cmdPartition, cancelPartition := newMockShortCmd(t, "df", casePartition)
	return metrics.NewTestMetrics(root, cmdGPU, cmdScreen, cmdPartition, helper.GetenvFromMap(env)),
		cancelGPU, cancelScreen, cancelPartition
}
