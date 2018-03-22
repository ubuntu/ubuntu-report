package sysmetrics_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ubuntu/ubuntu-report/internal/helper"
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

func TestNonInteractiveCollectAndSend(t *testing.T) {
	// we change current path and env variable: not parallelizable tests
	helper.SkipIfShort(t)

	testCases := []struct {
		name         string
		r            sysmetrics.ReportType
		alwaysReport bool

		shouldHitServer bool
		wantErr         bool
	}{
		{"regular report auto", sysmetrics.ReportAuto, false, true, false},
		{"regular report opt-out", sysmetrics.ReportOptOut, false, true, false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			a := helper.Asserter{T: t}

			out, tearDown := helper.TempDir(t)
			defer tearDown()
			defer changeEnv("XDG_CACHE_HOME", out)()
			out = filepath.Join(out, "ubuntu-report")
			// we don't really care where we hit for this API integration test, internal ones test it
			// and we don't really control /etc/os-release version and id.
			// Same for report file
			serverHit := false
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverHit = true
			}))
			defer ts.Close()

			err := sysmetrics.CollectAndSend(tc.r, tc.alwaysReport, ts.URL)

			if err != nil {
				t.Fatal("we didn't get an error in collect/sending where we didn't expect one", err)
			}

			a.Equal(serverHit, tc.shouldHitServer)
			p := filepath.Join(out, findInDirectory(t, "", out))
			data, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatalf("couldn't open report file %s", out)
			}
			d := string(data)
			switch tc.r {
			case sysmetrics.ReportAuto:
				if !strings.Contains(d, sysmetrics.ExpectedReportItem) {
					t.Errorf("we expected to find %s in report file, got: %s", sysmetrics.ExpectedReportItem, d)
				}
			case sysmetrics.ReportOptOut:
				if !strings.Contains(d, sysmetrics.OptOutJSON) {
					t.Errorf("we expected to find %s in report file, got: %s", sysmetrics.OptOutJSON, d)
				}
			}
		})
	}
}

func TestCollectAndSendTwice(t *testing.T) {
	// we change current path and env variable: not parallelizable tests
	helper.SkipIfShort(t)

	testCases := []struct {
		name         string
		alwaysReport bool

		wantErr bool
	}{
		{"fail report twice", false, true},
		{"forcing report twice", true, false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			a := helper.Asserter{T: t}

			out, tearDown := helper.TempDir(t)
			defer tearDown()
			defer changeEnv("XDG_CACHE_HOME", out)()
			out = filepath.Join(out, "ubuntu-report")
			// we don't really care where we hit for this API integration test, internal ones test it
			// and we don't really control /etc/os-release version and id.
			// Same for report file
			serverHit := false
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverHit = true
			}))
			defer ts.Close()

			// first call
			err := sysmetrics.CollectAndSend(sysmetrics.ReportAuto, tc.alwaysReport, ts.URL)
			if err != nil {
				t.Fatal("we didn't get an error in collect/sending where we didn't expect one", err)
			}
			a.Equal(serverHit, true)
			p := filepath.Join(out, findInDirectory(t, "", out))
			data, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatalf("couldn't open report file %s", out)
			}
			d := string(data)
			if !strings.Contains(d, sysmetrics.ExpectedReportItem) {
				t.Errorf("we expected to find %s in report file, got: %s", sysmetrics.ExpectedReportItem, d)
			}

			// scratch data file
			if err != ioutil.WriteFile(p, []byte(""), 0644) {
				t.Fatalf("couldn't reset %s: %v", p, err)
			}

			// second call, reset server
			serverHit = false
			err = sysmetrics.CollectAndSend(sysmetrics.ReportAuto, tc.alwaysReport, ts.URL)
			a.CheckWantedErr(err, tc.wantErr)

			a.Equal(serverHit, tc.alwaysReport)
			// reread the same file
			data, err = ioutil.ReadFile(p)
			if err != nil {
				t.Fatalf("couldn't open report file %s", out)
			}
			d = string(data)
			switch tc.alwaysReport {
			case true:
				if !strings.Contains(d, sysmetrics.ExpectedReportItem) {
					t.Errorf("we expected to find %s in second report file, got: %s", sysmetrics.ExpectedReportItem, d)
				}
			case false:
				if d != "" {
					t.Errorf("we expected to have an untouched report file on second report, got: %s", d)
				}
			}

		})
	}
}

func changeEnv(key, value string) func() {
	old := os.Getenv(key)
	os.Setenv(key, value)

	return func() {
		os.Setenv(key, old)
	}
}

// findInDirectory return first match of prefix in d
func findInDirectory(t *testing.T, prefix, d string) string {
	t.Helper()

	files, err := ioutil.ReadDir(d)
	if err != nil {
		t.Fatalf("couldn't scan %s: %v", d, err)
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), prefix) {
			return f.Name()
		}
	}
	t.Fatalf("didn't find %s in %s. Only got: %v", prefix, d, files)
	return ""
}
