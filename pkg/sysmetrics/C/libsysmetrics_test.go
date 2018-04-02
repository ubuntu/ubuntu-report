package main

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ubuntu/ubuntu-report/internal/helper"
)

// The actual test functions are in non-_test.go files
// so that they can use cgo (import "C").
// These wrappers are here for gotest to find.
// Similar technic than in https://golang.org/misc/cgo/test/cgo_test.go
func TestCollect(t *testing.T)                      { testCollect(t) }
func TestSend(t *testing.T)                         { testSend(t) }
func TestNonInteractiveCollectAndSend(t *testing.T) { testNonInteractiveCollectAndSend(t) }
func TestInteractiveCollectAndSend(t *testing.T)    { testInteractiveCollectAndSend(t) }

func TestCollectExample(t *testing.T) {
	helper.SkipIfShort(t)
	t.Parallel()
	ensureGCC(t)

	out, tearDown := helper.TempDir(t)
	defer tearDown()
	lib := buildLib(t, out)
	p := extractExampleFromDoc(t, out, "Collect system info", "", "")
	binary := buildExample(t, out, p, lib)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+out)
	data, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal("we didn't expect an error and got one", err)
	}

	if !strings.Contains(string(data), expectedReportItem) {
		t.Errorf("we expected at least %s in output, got: '%s", expectedReportItem, string(data))
	}
}

func TestSendExample(t *testing.T) {
	helper.SkipIfShort(t)
	t.Parallel()
	ensureGCC(t)

	a := helper.Asserter{T: t}

	serverHit := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHit = true
	}))
	defer ts.Close()

	out, tearDown := helper.TempDir(t)
	defer tearDown()

	lib := buildLib(t, out)
	p := extractExampleFromDoc(t, out, "Send", `""`, `"`+ts.URL+`"`)
	binary := buildExample(t, out, p, lib)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+out, "XDG_CACHE_HOME="+out)
	err := cmd.Run()

	if err != nil {
		t.Fatal("we didn't expect an error and got one", err)
	}

	// There isn't a data race as only the external binary can hit test server,
	// but Go can't know it. To prevent that, shutdown the test server explicitely
	ts.Close()

	a.Equal(serverHit, true)
	xdgP := filepath.Join(out, "ubuntu-report")
	p = filepath.Join(xdgP, helper.FindInDirectory(t, "", xdgP))
	data, err := ioutil.ReadFile(p)
	if err != nil {
		t.Fatalf("couldn't open report file %s", p)
	}
	d := string(data)

	if !strings.Contains(d, expectedReportItem) {
		t.Errorf("we expected to find %s in report file, got: %s", expectedReportItem, d)
	}
}

func TestCollectAndSendExample(t *testing.T) {
	helper.SkipIfShort(t)
	t.Parallel()
	ensureGCC(t)

	a := helper.Asserter{T: t}

	serverHit := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHit = true
	}))
	defer ts.Close()

	out, tearDown := helper.TempDir(t)
	defer tearDown()

	lib := buildLib(t, out)
	p := extractExampleFromDoc(t, out, "Collect and send system info to server", `""`, `"`+ts.URL+`"`)
	binary := buildExample(t, out, p, lib)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+out, "XDG_CACHE_HOME="+out)
	err := cmd.Run()

	if err != nil {
		t.Fatal("we didn't expect an error and got one", err)
	}

	// There isn't a data race as only the external binary can hit test server,
	// but Go can't know it. To prevent that, shutdown the test server explicitely
	ts.Close()

	a.Equal(serverHit, true)
	xdgP := filepath.Join(out, "ubuntu-report")
	p = filepath.Join(xdgP, helper.FindInDirectory(t, "", xdgP))
	data, err := ioutil.ReadFile(p)
	if err != nil {
		t.Fatalf("couldn't open report file %s", p)
	}
	d := string(data)

	if !strings.Contains(d, expectedReportItem) {
		t.Errorf("we expected to find %s in report file, got: %s", expectedReportItem, d)
	}
}

func ensureGCC(t *testing.T) {
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("skipping test: no gcc found:", err)
	}
}

func buildExample(t *testing.T, dest, example, lib string) string {
	t.Helper()

	d := filepath.Join(dest, "example")
	cmd := exec.Command("gcc", "-o", d, example, lib)
	var out bytes.Buffer
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatal("couldn't build example binary:", err, "\n", out.String())
	}
	return d
}

func buildLib(t *testing.T, p string) string {
	t.Helper()
	libName := "libsysmetrics.so.1"
	d := filepath.Join(p, libName)
	cmd := exec.Command("go", "build", "-o", d, "-buildmode=c-shared", "-ldflags", "-extldflags -Wl,-soname,"+libName, "libsysmetrics.go")
	if err := cmd.Run(); err != nil {
		t.Fatal("couldn't build library:", err)
	}
	if err := os.Rename(filepath.Join(p, "libsysmetrics.so.h"), filepath.Join(p, "libsysmetrics.h")); err != nil {
		t.Fatal("couldn't rename header file", err)
	}
	return d
}

func extractExampleFromDoc(t *testing.T, dir, title, pattern, replace string) string {
	t.Helper()

	f, err := os.Open("doc.go")
	if err != nil {
		t.Fatal("couldn't open documentation file:", err)
	}
	defer f.Close()

	p := filepath.Join(dir, strings.Replace(strings.ToLower(title), " ", "_", -1)+".c")
	w, err := os.Create(p)
	if err != nil {
		t.Fatal("couldn't create example file:", err)
	}
	defer w.Close()

	scanner := bufio.NewScanner(f)
	correctSection := false
	inExample := false
	for scanner.Scan() {
		txt := strings.TrimPrefix(scanner.Text(), "//")
		if strings.HasPrefix(txt, " "+title) {
			correctSection = true
			continue
		}
		if !correctSection {
			continue
		}
		if strings.HasPrefix(txt, " Example") {
			inExample = true
			continue
		}
		if !inExample {
			continue
		}
		// end of example: no space separated content, nor empty line
		if !(strings.HasPrefix(txt, "   ") || txt == "") {
			break
		}
		txt = strings.Replace(strings.TrimPrefix(txt, "   "), pattern, replace, -1)
		if _, err := w.WriteString(txt + "\n"); err != nil {
			t.Fatalf("couldn't write '%s' to destination example file: %v", txt, err)
		}
	}

	return p
}
