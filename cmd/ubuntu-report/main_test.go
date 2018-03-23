package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
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
				// empty logs, apart info on installer or upgrade telemetry (file can be missing)
				scanner := bufio.NewScanner(bytes.NewReader(got.Bytes()))
				for scanner.Scan() {
					l := scanner.Text()
					if strings.Contains(l, "level=info") && strings.Contains(l, "/telemetry") {
						continue
					}
					t.Errorf("Expected no log output with -v apart from missing telemetry installer or updater logs, but got: %s", l)
				}
			case "-vv":
				if !strings.Contains(got.String(), "level=debug") {
					t.Errorf("Expected some debug log to be printed, but got: %s", got.String())
				}
			}
		})
	}
}
