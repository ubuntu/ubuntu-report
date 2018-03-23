package helper

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
 * Contains test helpers across package
 */

// Asserter for testing purposes
type Asserter struct {
	*testing.T
}

// Equal checks that the 2 values are equals
// slices and arrays can be of different orders
func (m Asserter) Equal(got, want interface{}) {
	m.Helper()

	same := false
	switch t := reflect.TypeOf(got); t.Kind() {
	case reflect.Slice:
		// We treat slice of bytes differently, order is important
		a, gotIsBytes := got.([]byte)
		b, wantIsBytes := want.([]byte)
		if gotIsBytes && wantIsBytes {
			// convert them to string for easier comparaison once
			// they don'tmatch
			if same = reflect.DeepEqual(a, b); !same {
				m.Errorf("got: %s (converted from []byte), wants %s (converted from []byte)", string(a), string(b))
			}
			return
		}
		same = unsortedEqualsSliceArray(got, want)
	case reflect.Array:
		same = unsortedEqualsSliceArray(got, want)
	case reflect.Map, reflect.Ptr:
		same = reflect.DeepEqual(got, want)
	default:
		same = got == want
	}

	if !same {
		m.Errorf("got: %#v (%T), wants %#v (%T)", got, got, want, want)
	}
}

// CheckWantedErr checks that we received an error when desired or none other
func (m Asserter) CheckWantedErr(err error, wantErr bool) {
	m.Helper()
	if err != nil && !wantErr {
		m.Fatal("got an unexpected err:", err)
	}
	if err == nil && wantErr {
		m.Error("expected an error and got none")
	}
}

// LoadOrUpdateGolden returns golden file content.
// It will update it beforehand if requested.
func LoadOrUpdateGolden(t *testing.T, p string, data []byte, update bool) []byte {
	t.Helper()

	if update {
		t.Log("update golden file at", p)
		if data == nil {
			t.Logf("No file to create as data is nil")
			os.Remove(p)
			return nil
		}
		if err := ioutil.WriteFile(p, data, 0666); err != nil {
			t.Fatalf("can't update golden file %s: %v", p, err)
		}
	}

	var content []byte
	var err error
	if content, err = ioutil.ReadFile(p); os.IsExist(err) && err != nil {
		t.Fatalf("got an error loading golden file %s: %v", p, err)
	}
	return content
}

func unsortedEqualsSliceArray(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}

	a1 := reflect.ValueOf(a)
	a2 := reflect.ValueOf(b)

	if a1.Len() != a2.Len() {
		return false
	}

	// mark indexes in b that we already matched against
	seen := make([]bool, a2.Len())
	for i := 0; i < a1.Len(); i++ {
		cur := a1.Index(i).Interface()

		found := false
		for j := 0; j < a2.Len(); j++ {
			if seen[j] {
				continue
			}

			if reflect.DeepEqual(a2.Index(j).Interface(), cur) {
				seen[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// ShortProcess helper is mocking a command supposed to return quickly
// (within 100 milliseconds)
// (inspired by stdlib)
func ShortProcess(t *testing.T, helper string, s ...string) (*exec.Cmd, context.CancelFunc) {
	t.Helper()

	cs := []string{"-test.run=" + helper, "--"}
	cs = append(cs, s...)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	return cmd, cancel
}

// CopyFile for testing from src to dst
func CopyFile(t *testing.T, src, dst string) {
	t.Helper()

	s, err := os.Open(src)
	if err != nil {
		t.Fatalf("couldn't open %s: %v", src, err)
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		t.Fatalf("couldn't create %s: %v", dst, err)
	}
	defer func() {
		if err := d.Close(); err != nil {
			t.Fatalf("couldn't close properly %s: %v", dst, err)
		}
	}()

	if _, err := io.Copy(d, s); err != nil {
		t.Fatalf("couldn't copy %s content to %s: %v", src, dst, err)
	}
}

// SkipIfShort will skip current test if -short isn't passed
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("short tests only, skipping")
	}
}

// GetenvFromMap generates a getenv function mock from a map[string]string
// no value returns empty string
func GetenvFromMap(env map[string]string) func(key string) string {
	return func(key string) string {
		value, ok := env[key]
		if !ok {
			value = ""
		}
		return value
	}
}

// TempDir creates and give defer to remove temporary dir safely for testing
func TempDir(t *testing.T) (string, func()) {
	t.Helper()
	d, err := ioutil.TempDir("", "ubuntu-report-tests")
	if err != nil {
		t.Fatal("couldn't create temporary directory", err)
	}
	return d, func() {
		if err = os.RemoveAll(d); err != nil {
			t.Fatalf("couldn't clean temporary directory %s, %v", d, err)
		}
	}
}

// CaptureStdout returns an io.Reader to read what was printed, and teardown
func CaptureStdout(t *testing.T) (io.Reader, func()) {
	t.Helper()
	stdout, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatal("couldn't create stdout pipe", err)
	}
	oldStdout := os.Stdout
	os.Stdout = stdoutW
	closed := false
	return stdout, func() {
		// only teardown once
		if closed {
			return
		}
		os.Stdout = oldStdout
		stdoutW.Close()
		closed = true
	}
}

// CaptureStdin returns an i.Writer to write, as stdin, and teardown
func CaptureStdin(t *testing.T) (io.WriteCloser, func()) {
	t.Helper()
	stdin, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal("couldn't create stdin pipe", err)
	}
	oldStdin := os.Stdin
	os.Stdin = stdin
	return stdinW, func() { os.Stdin = oldStdin }
}

// CaptureLogs returns an io.Reader to read what was logged, and teardown
func CaptureLogs(t *testing.T) (io.Reader, func()) {
	t.Helper()

	pr, pw := io.Pipe()
	log.SetOutput(pw)
	old := log.StandardLogger().Out
	closed := false
	return pr, func() {
		// only teardown once
		if closed {
			return
		}
		log.SetOutput(old)
		pw.Close()
		closed = true
	}
}

// RunFunctionWithTimeout run in a go routing the fn functionL
// There is a timeout as a maximum limit for the function, to run
// The returned channel of errors is closed once the command ends
// or if timeout is reached
func RunFunctionWithTimeout(t *testing.T, fn func() error) chan error {
	errs := make(chan error)
	go func() {
		defer close(errs)

		// add a timeout
		cmd := func() chan error {
			cmderr := make(chan error)
			go func() {
				defer close(cmderr)
				err := fn()
				cmderr <- err
			}()
			return cmderr
		}

		select {
		case err := <-cmd():
			errs <- err
		case <-time.After(5 * time.Second):
			errs <- errors.New("function under test timed out")
		}
	}()
	return errs
}
