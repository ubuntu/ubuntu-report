package sysmetrics_test

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ubuntu/ubuntu-report/internal/helper"
	"github.com/ubuntu/ubuntu-report/pkg/sysmetrics"
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

func TestSendReport(t *testing.T) {
	// we change current path and env variable: not parallelizable tests
	helper.SkipIfShort(t)

	testCases := []struct {
		name         string
		alwaysReport bool

		shouldHitServer bool
		wantErr         bool
	}{
		{"regular send", false, true, false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			a := helper.Asserter{T: t}

			out, tearDown := helper.TempDir(t)
			defer tearDown()
			defer helper.ChangeEnv("XDG_CACHE_HOME", out)()
			out = filepath.Join(out, "ubuntu-report")
			// we don't really care where we hit for this API integration test, internal ones test it
			// and we don't really control /etc/os-release version and id.
			// Same for report file
			serverHit := false
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverHit = true
			}))
			defer ts.Close()

			err := sysmetrics.SendReport([]byte(fmt.Sprintf(`{ %s: "18.04" }`, sysmetrics.ExpectedReportItem)),
				tc.alwaysReport, ts.URL)

			a.CheckWantedErr(err, tc.wantErr)
			a.Equal(serverHit, tc.shouldHitServer)
			p := filepath.Join(out, helper.FindInDirectory(t, "", out))
			data, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatalf("couldn't open report file %s", out)
			}
			d := string(data)
			if !strings.Contains(d, sysmetrics.ExpectedReportItem) {
				t.Errorf("we expected to find %s in report file, got: %s", sysmetrics.ExpectedReportItem, d)
			}
		})
	}
}
func TestSendDecline(t *testing.T) {
	// we change current path and env variable: not parallelizable tests
	helper.SkipIfShort(t)

	testCases := []struct {
		name         string
		alwaysReport bool

		shouldHitServer bool
		wantErr         bool
	}{
		{"regular send opt-out", false, true, false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			a := helper.Asserter{T: t}

			out, tearDown := helper.TempDir(t)
			defer tearDown()
			defer helper.ChangeEnv("XDG_CACHE_HOME", out)()
			out = filepath.Join(out, "ubuntu-report")
			// we don't really care where we hit for this API integration test, internal ones test it
			// and we don't really control /etc/os-release version and id.
			// Same for report file
			serverHit := false
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverHit = true
			}))
			defer ts.Close()

			err := sysmetrics.SendDecline(tc.alwaysReport, ts.URL)

			if err != nil {
				t.Fatal("we didn't expect getting an error, got:", err)
			}

			a.Equal(serverHit, tc.shouldHitServer)
			p := filepath.Join(out, helper.FindInDirectory(t, "", out))
			data, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatalf("couldn't open report file %s", out)
			}
			d := string(data)
			if !strings.Contains(d, sysmetrics.OptOutJSON) {
				t.Errorf("we expected to find %s in report file, got: %s", sysmetrics.OptOutJSON, d)
			}
		})
	}
}

func TestSendReportTwice(t *testing.T) {
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
			defer helper.ChangeEnv("XDG_CACHE_HOME", out)()
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
			err := sysmetrics.SendReport([]byte(fmt.Sprintf(`{ %s: "18.04" }`, sysmetrics.ExpectedReportItem)),
				tc.alwaysReport, ts.URL)
			if err != nil {
				t.Fatal("we didn't expect getting an error, got:", err)
			}
			a.Equal(serverHit, true)
			p := filepath.Join(out, helper.FindInDirectory(t, "", out))
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
			err = sysmetrics.SendReport([]byte(fmt.Sprintf(`{ %s: "18.04" }`, sysmetrics.ExpectedReportItem)),
				tc.alwaysReport, ts.URL)
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

func TestSendDeclineTwice(t *testing.T) {
	// we change current path and env variable: not parallelizable tests
	helper.SkipIfShort(t)

	testCases := []struct {
		name         string
		alwaysReport bool

		wantErr bool
	}{
		{"fail decline twice", false, true},
		{"forcing decline twice", true, false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			a := helper.Asserter{T: t}

			out, tearDown := helper.TempDir(t)
			defer tearDown()
			defer helper.ChangeEnv("XDG_CACHE_HOME", out)()
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
			err := sysmetrics.SendDecline(tc.alwaysReport, ts.URL)
			if err != nil {
				t.Fatal("we didn't expect getting an error, got:", err)
			}
			a.Equal(serverHit, true)
			p := filepath.Join(out, helper.FindInDirectory(t, "", out))
			data, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatalf("couldn't open report file %s", out)
			}
			d := string(data)
			if !strings.Contains(d, sysmetrics.OptOutJSON) {
				t.Errorf("we expected to find %s in report file, got: %s", sysmetrics.OptOutJSON, d)
			}

			// scratch data file
			if err != ioutil.WriteFile(p, []byte(""), 0644) {
				t.Fatalf("couldn't reset %s: %v", p, err)
			}

			// second call, reset server
			serverHit = false
			err = sysmetrics.SendDecline(tc.alwaysReport, ts.URL)
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
				if !strings.Contains(d, sysmetrics.OptOutJSON) {
					t.Errorf("we expected to find %s in second report file, got: %s", sysmetrics.OptOutJSON, d)
				}
			case false:
				if d != "" {
					t.Errorf("we expected to have an untouched report file on second report, got: %s", d)
				}
			}

		})
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
			defer helper.ChangeEnv("XDG_CACHE_HOME", out)()
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
				t.Fatal("we didn't expect getting an error, got:", err)
			}

			a.Equal(serverHit, tc.shouldHitServer)
			p := filepath.Join(out, helper.FindInDirectory(t, "", out))
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
			defer helper.ChangeEnv("XDG_CACHE_HOME", out)()
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
				t.Fatal("we didn't expect getting an error, got:", err)
			}
			a.Equal(serverHit, true)
			p := filepath.Join(out, helper.FindInDirectory(t, "", out))
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

func TestInteractiveCollectAndSend(t *testing.T) {
	// we change current path and env variable: not parallelizable tests
	helper.SkipIfShort(t)

	testCases := []struct {
		name    string
		answers []string

		sendOnlyOptOutData bool
		wantWriteAndUpload bool
	}{
		{"yes", []string{"yes"}, false, true},
		{"y", []string{"y"}, false, true},
		{"YES", []string{"YES"}, false, true},
		{"Y", []string{"Y"}, false, true},
		{"no", []string{"no"}, true, true},
		{"n", []string{"n"}, true, true},
		{"NO", []string{"NO"}, true, true},
		{"n", []string{"N"}, true, true},
		{"quit", []string{"quit"}, false, false},
		{"q", []string{"q"}, false, false},
		{"QUIT", []string{"QUIT"}, false, false},
		{"Q", []string{"Q"}, false, false},
		{"default-quit", []string{""}, false, false},
		{"garbage-then-quit", []string{"garbage", "yesgarbage", "nogarbage", "quitgarbage", "Q"}, false, false},
		{"ctrl-c-input", []string{"CTRL-C"}, false, false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			a := helper.Asserter{T: t}

			out, tearDown := helper.TempDir(t)
			defer tearDown()
			defer helper.ChangeEnv("XDG_CACHE_HOME", out)()
			out = filepath.Join(out, "ubuntu-report")
			// we don't really care where we hit for this API integration test, internal ones test it
			// and we don't really control /etc/os-release version and id.
			// Same for report file
			serverHit := false
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverHit = true
			}))
			defer ts.Close()

			stdout, tearDown := helper.CaptureStdout(t)
			defer tearDown()
			stdin, tearDown := helper.CaptureStdin(t)
			defer tearDown()

			cmdErrs := helper.RunFunctionWithTimeout(t, func() error { return sysmetrics.CollectAndSend(sysmetrics.ReportInteractive, false, ts.URL) })

			gotJSONReport := false
			answerIndex := 0
			scanner := bufio.NewScanner(stdout)
			scanner.Split(sysmetrics.ScanLinesOrQuestion)
			for scanner.Scan() {
				txt := scanner.Text()
				// first, we should have a known element
				if strings.Contains(txt, sysmetrics.ExpectedReportItem) {
					gotJSONReport = true
				}
				if !strings.Contains(txt, "Do you agree to report this?") {
					continue
				}
				a := tc.answers[answerIndex]
				if a == "CTRL-C" {
					stdin.Close()
					break
				} else {
					stdin.Write([]byte(tc.answers[answerIndex] + "\n"))
				}
				answerIndex = answerIndex + 1
				// all answers have be provided
				if answerIndex >= len(tc.answers) {
					stdin.Close()
					break
				}
			}

			if err := <-cmdErrs; err != nil {
				t.Fatal("didn't expect to get an error, got:", err)
			}
			a.Equal(gotJSONReport, true)
			a.Equal(serverHit, tc.wantWriteAndUpload)

			if !tc.wantWriteAndUpload {
				if _, err := os.Stat(filepath.Join(out, "ubuntu-report")); err == nil || (err != nil && !os.IsNotExist(err)) {
					t.Fatal("we didn't want to get a report but we got one")
				}
				return
			}
			p := filepath.Join(out, helper.FindInDirectory(t, "", out))
			data, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatalf("couldn't open report file %s", out)
			}
			d := string(data)
			expected := sysmetrics.ExpectedReportItem
			if tc.sendOnlyOptOutData {
				expected = sysmetrics.OptOutJSON
			}
			if !strings.Contains(d, expected) {
				t.Errorf("we expected to find %s in report file, got: %s", expected, d)
			}
		})
	}
}

func TestSendPendingReport(t *testing.T) {
	// we change current path and env variable: not parallelizable tests
	helper.SkipIfShort(t)

	testCases := []struct {
		name string

		shouldHitServer bool
		wantErr         bool
	}{
		{"regular send", true, false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			a := helper.Asserter{T: t}

			out, tearDown := helper.TempDir(t)
			defer tearDown()
			defer helper.ChangeEnv("XDG_CACHE_HOME", out)()
			out = filepath.Join(out, "ubuntu-report")

			pendingReportData, err := ioutil.ReadFile(filepath.Join("testdata", "good", "ubuntu-report", "pending"))
			if err != nil {
				t.Fatalf("couldn't open pending report file: %v", err)
			}
			pendingReportPath := filepath.Join(out, "pending")
			if err := os.MkdirAll(out, 0700); err != nil {
				t.Fatal("couldn't create parent directory of pending report", err)
			}
			if err := ioutil.WriteFile(pendingReportPath, pendingReportData, 0644); err != nil {
				t.Fatalf("couldn't copy pending report file to cache directory: %v", err)
			}

			// we don't really care where we hit for this API integration test, internal ones test it
			// and we don't really control /etc/os-release version and id.
			// Same for report file
			serverHit := false
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverHit = true
			}))
			defer ts.Close()

			err = sysmetrics.SendPendingReport(ts.URL)

			a.CheckWantedErr(err, tc.wantErr)
			a.Equal(serverHit, tc.shouldHitServer)

			if _, pendingReportErr := os.Stat(pendingReportPath); os.IsExist(pendingReportErr) {
				t.Errorf("we expected the pending report to be removed and it wasn't")
			}

			p := filepath.Join(out, helper.FindInDirectory(t, "", out))
			got, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatalf("couldn't open report file %s", out)
			}
			a.Equal(got, pendingReportData)
		})
	}
}
