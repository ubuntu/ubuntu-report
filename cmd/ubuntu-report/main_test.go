package main

import (
	"io/ioutil"
	"os"
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
	stdout, tearDown := helper.CaptureStdout(t)
	defer tearDown()

	cmd := generateRootCmd()
	cmd.SetArgs([]string{"show"})

	var c *cobra.Command
	cmdErrs := helper.RunFunctionWithTimeout(t, func() error {
		var err error
		c, err = cmd.ExecuteC()
		os.Stdout.Close() // close stdout to release ReadAll()
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
