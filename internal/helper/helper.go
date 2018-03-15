package helper

import (
	"io/ioutil"
	"testing"
)

/*
 * Contains test helpers across package
 */

// Asserter for testing purposes
type Asserter struct {
	*testing.T
}

// Equal checks that the 2 values are equals
func (m Asserter) Equal(got, want interface{}) {
	m.Helper()
	if got != want {
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
func LoadOrUpdateGolden(p string, data []byte, update bool, t *testing.T) []byte {
	t.Helper()

	if update {
		t.Log("update golden file at", p)
		if err := ioutil.WriteFile(p, data, 0666); err != nil {
			t.Fatalf("can't update golden file %s: %v", p, err)
		}
	}

	var content []byte
	var err error
	if content, err = ioutil.ReadFile(p); err != nil {
		t.Fatalf("got an error loading golden file %s: %v", p, err)
	}
	return content
}
