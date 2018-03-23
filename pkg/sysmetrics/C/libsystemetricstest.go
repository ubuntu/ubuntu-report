package main

// #include <stdbool.h>
// #include <stdio.h>
// #include <stdlib.h>
// extern char* Collect(char** p0);
// typedef enum {
//     ReportInteractive = 0,
//     ReportAuto = 1,
//     ReportOptOut = 2,
// } ReportType;
// typedef unsigned char GoUint8;
// extern char* CollectAndSend(ReportType p0, GoUint8 p1, char* p2);
import "C"

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/ubuntu/ubuntu-report/internal/helper"
	"github.com/ubuntu/ubuntu-report/pkg/sysmetrics"
)

/*
The C API is calling the Go API, which is heavily tested. Consequently, we only test
main cases.
*/

const (
	expectedReportItem = `"Version":`
	optOutJSON         = `{"OptOut": true}`
)

func testCollect(t *testing.T) {
	t.Parallel()

	var res *C.char
	defer C.free(unsafe.Pointer(res))

	err := C.Collect(&res)
	defer C.free(unsafe.Pointer(err))

	if err != nil {
		t.Fatal("we didn't expect an error and got one", C.GoString(err))
	}
	data := C.GoString(res)
	if !strings.Contains(data, expectedReportItem) {
		t.Errorf("we expected at least %s in output, got: '%s", expectedReportItem, data)
	}
}

func testNonInteractiveCollectAndSend(t *testing.T) {
	// we change current path and env variable: not parallelizable tests
	helper.SkipIfShort(t)

	testCases := []struct {
		name string
		r    sysmetrics.ReportType

		shouldHitServer bool
		wantErr         bool
	}{
		{"regular report auto", sysmetrics.ReportAuto, true, false},
		{"regular report opt-out", sysmetrics.ReportOptOut, true, false},
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

			url := C.CString(ts.URL)
			defer C.free(unsafe.Pointer(url))

			err := C.CollectAndSend(C.ReportType(tc.r), C.uchar(0), url)
			defer C.free(unsafe.Pointer(err))

			if err != nil {
				t.Fatal("we didn't expect getting an error, got:", err)
			}

			a.Equal(serverHit, tc.shouldHitServer)
			p := filepath.Join(out, helper.FindInDirectory(t, "", out))
			data, errread := ioutil.ReadFile(p)
			if errread != nil {
				t.Fatalf("couldn't open report file %s", out)
			}
			d := string(data)
			switch tc.r {
			case sysmetrics.ReportAuto:
				if !strings.Contains(d, expectedReportItem) {
					t.Errorf("we expected to find %s in report file, got: %s", expectedReportItem, d)
				}
			case sysmetrics.ReportOptOut:
				if !strings.Contains(d, optOutJSON) {
					t.Errorf("we expected to find %s in report file, got: %s", optOutJSON, d)
				}
			}
		})
	}
}

func testInteractiveCollectAndSend(t *testing.T) {
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
				fmt.Println("HIT")
				serverHit = true
			}))
			defer ts.Close()

			stdout, tearDown := helper.CaptureStdout(t)
			defer tearDown()
			stdin, tearDown := helper.CaptureStdin(t)
			defer tearDown()

			cmdErrs := helper.RunFunctionWithTimeout(t, func() error {
				url := C.CString(ts.URL)
				defer C.free(unsafe.Pointer(url))

				errstr := C.CollectAndSend(C.ReportType(sysmetrics.ReportInteractive), C.uchar(0), url)
				defer C.free(unsafe.Pointer(errstr))
				var err error
				if errstr != nil {
					err = errors.New(C.GoString(errstr))
				}
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
