package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ubuntu/ubuntu-report/internal/helper"
)

const (
	expectedReportItem = `"Version":`
	optOutJSON         = `{"OptOut": true}`
)

func TestShow(t *testing.T) {
	helper.SkipIfShort(t)
	a := helper.Asserter{T: t}
	stdout, restoreStdout := helper.CaptureStdout(t)
	defer restoreStdout()

	cmd := generateRootCmd()
	cmd.SetArgs([]string{"show"})

	var c *cobra.Command
	cmdErrs := helper.RunFunctionWithTimeout(t, func() error {
		var err error
		c, err = cmd.ExecuteC()
		restoreStdout() // close stdout to release ReadAll()
		return err
	})

	if err := <-cmdErrs; err != nil {
		t.Fatal("got an error when expecting none:", err)
	}
	a.Equal(c.Name(), "show")
	got, err := ioutil.ReadAll(stdout)
	if err != nil {
		t.Error("couldn't read from stdout", err)
	}
	if !strings.Contains(string(got), expectedReportItem) {
		t.Errorf("Expected %s to be in output, but got: %s", expectedReportItem, string(got))
	}
}

// Test Verbosity level with Show
func TestVerbosity(t *testing.T) {
	helper.SkipIfShort(t)

	testCases := []struct {
		verbosity string
	}{
		{""},
		{"-v"},
		{"-vv"},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run("verbosity level "+tc.verbosity, func(t *testing.T) {
			a := helper.Asserter{T: t}
			out, restoreLogs := helper.CaptureLogs(t)
			defer restoreLogs()

			cmd := generateRootCmd()
			args := []string{"show"}
			if tc.verbosity != "" {
				args = append(args, tc.verbosity)
			}
			cmd.SetArgs(args)

			cmdErrs := helper.RunFunctionWithTimeout(t, func() error {
				var err error
				_, err = cmd.ExecuteC()
				restoreLogs() // send EOF to log to release io.Copy()
				return err
			})

			var got bytes.Buffer
			io.Copy(&got, out)

			if err := <-cmdErrs; err != nil {
				t.Fatal("got an error when expecting none:", err)
			}

			switch tc.verbosity {
			case "":
				a.Equal(got.String(), "")
			case "-v":
				// empty logs, apart info on dcd, installer or upgrade telemetry (file can be missing)
				// and other GPU, screen and autologin that you won't have in Travis CI.
				scanner := bufio.NewScanner(bytes.NewReader(got.Bytes()))
				for scanner.Scan() {
					l := scanner.Text()
					if strings.Contains(l, "level=info") {
						allowedLog := false
						for _, msg := range []string{"/telemetry", "DCD", "GPU info", "Disk info", "Screen info", "CPU info", "autologin information", "/sys/class/dmi/id/"} {
							if strings.Contains(l, msg) {
								allowedLog = true
							}
						}
						if allowedLog {
							continue
						}
						t.Errorf("Expected no log output with -v apart from missing telemetry, GPU, Disk, Screen, sys and autologin information, but got: %s", l)
					}
				}
			case "-vv":
				if !strings.Contains(got.String(), "level=debug") {
					t.Errorf("Expected some debug log to be printed, but got: %s", got.String())
				}
			}
		})
	}
}

func TestSend(t *testing.T) {
	helper.SkipIfShort(t)

	testCases := []struct {
		name   string
		answer string

		shouldHitServer bool
		wantErr         bool
	}{
		{"regular report auto", "yes", true, false},
		{"regular report opt-out", "no", true, false},
		{"dist-upgrade report", "upgrade", true, false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			a := helper.Asserter{T: t}

			out, tearDown := helper.TempDir(t)
			defer tearDown()
			defer helper.ChangeEnv("XDG_CACHE_HOME", out)()
			out = filepath.Join(out, "ubuntu-report")
			// create a previous report with fake json data (which isn't optout)
			if err := os.MkdirAll(out, 0700); err != nil {
				t.Fatalf("couldn't create ubuntu-report directory: %v", err)
			}
			if err := ioutil.WriteFile(filepath.Join(out, "ubuntu.14.04"), []byte(`{ "some-opt-in-data': true}`), 0644); err != nil {
				t.Fatalf("couldn't setup previous report file: %v", err)
			}

			// we don't really care where we hit for this API integration test, internal ones test it
			// and we don't really control /etc/os-release version and id.
			// Same for report file
			serverHit := false
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverHit = true
			}))
			defer ts.Close()

			cmd := generateRootCmd()
			args := []string{"send", tc.answer, "--url", ts.URL}
			cmd.SetArgs(args)

			cmdErrs := helper.RunFunctionWithTimeout(t, func() error {
				var err error
				_, err = cmd.ExecuteC()
				return err
			})

			if err := <-cmdErrs; err != nil {
				t.Fatal("got an error when expecting none:", err)
			}

			a.Equal(serverHit, tc.shouldHitServer)
			// get highest report path
			reportP := ""
			files, err := ioutil.ReadDir(out)
			if err != nil {
				t.Fatalf("couldn't scan %s: %v", out, err)
			}
			for _, f := range files {
				if f.Name() > reportP {
					reportP = f.Name()
				}
			}
			data, err := ioutil.ReadFile(filepath.Join(out, reportP))
			if err != nil {
				t.Fatalf("couldn't open report file %s", reportP)
			}
			d := string(data)

			switch tc.answer {
			case "yes":
				fallthrough
			case "upgrade":
				if !strings.Contains(d, expectedReportItem) {
					t.Errorf("we expected to find %s in report file, got: %s", expectedReportItem, d)
				}
			case "no":
				if !strings.Contains(d, optOutJSON) {
					t.Errorf("we expected to find %s in report file, got: %s", optOutJSON, d)
				}
			}
		})
	}
}

func TestInteractive(t *testing.T) {
	helper.SkipIfShort(t)

	testCases := []struct {
		name    string
		cmd     string
		answers []string

		sendOnlyOptOutData bool
		wantWriteAndUpload bool
	}{
		{"root yes command", "", []string{"yes"}, false, true},
		{"root YES", "", []string{"YES"}, false, true},
		{"root Y", "", []string{"Y"}, false, true},
		{"root no", "", []string{"no"}, true, true},
		{"root n", "", []string{"n"}, true, true},
		{"root NO", "", []string{"NO"}, true, true},
		{"root n", "", []string{"N"}, true, true},
		{"root quit", "", []string{"quit"}, false, false},
		{"root q", "", []string{"q"}, false, false},
		{"root QUIT", "", []string{"QUIT"}, false, false},
		{"root Q", "", []string{"Q"}, false, false},
		{"root default-quit", "", []string{""}, false, false},
		{"root garbage-then-quit", "", []string{"garbage", "yesgarbage", "nogarbage", "quitgarbage", "Q"}, false, false},
		{"root ctrl-c-input", "", []string{"CTRL-C"}, false, false},
		{"interactive yes command", "interactive", []string{"yes"}, false, true},
		{"interactive no command", "interactive", []string{"no"}, true, true},
		{"interactive ctrl-c-input", "interactive", []string{"CTRL-C"}, false, false},
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

			stdout, restoreStdout := helper.CaptureStdout(t)
			defer restoreStdout()
			stdin, tearDown := helper.CaptureStdin(t)
			defer tearDown()

			cmd := generateRootCmd()
			args := []string{}
			if tc.cmd != "" {
				args = append(args, tc.cmd)
			}
			args = append(args, "--url", ts.URL)
			cmd.SetArgs(args)

			cmdErrs := helper.RunFunctionWithTimeout(t, func() error {
				var err error
				_, err = cmd.ExecuteC()
				restoreStdout()
				return err
			})

			gotJSONReport := false
			answerIndex := 0
			scanner := bufio.NewScanner(stdout)
			scanner.Split(scanLinesOrQuestion)
			for scanner.Scan() {
				txt := scanner.Text()
				// first, we should have a known element
				if strings.Contains(txt, expectedReportItem) {
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
			expected := expectedReportItem
			if tc.sendOnlyOptOutData {
				expected = optOutJSON
			}
			if !strings.Contains(d, expected) {
				t.Errorf("we expected to find %s in report file, got: %s", expected, d)
			}
		})
	}
}

func TestService(t *testing.T) {
	helper.SkipIfShort(t)

	testCases := []struct {
		name string

		shouldHitServer bool
	}{
		{"regular send", true},
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

			cmd := generateRootCmd()
			args := []string{"service", "--url", ts.URL}
			cmd.SetArgs(args)

			cmdErrs := helper.RunFunctionWithTimeout(t, func() error {
				var err error
				_, err = cmd.ExecuteC()
				return err
			})

			if err := <-cmdErrs; err != nil {
				t.Fatal("got an error when expecting none:", err)
			}

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

// scanLinesOrQuestion is copy of ScanLines, adding the expected question string as we don't return here
func scanLinesOrQuestion(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	if i := bytes.IndexByte(data, ']'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
