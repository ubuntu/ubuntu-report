package helper

import "testing"

// Asserter for testing purposes
type Asserter struct {
	*testing.T
}

// Equal checks that the 2 values are equals
func (m Asserter) Equal(a, b interface{}) {
	m.Helper()
	if a != b {
		m.Errorf("got: %#v (%T), was %#v (%T)", a, a, b, b)
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
