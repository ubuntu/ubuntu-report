package main

import "testing"

// The actual test functions are in non-_test.go files
// so that they can use cgo (import "C").
// These wrappers are here for gotest to find.
// Similar technic than in https://golang.org/misc/cgo/test/cgo_test.go
func TestCollect(t *testing.T)                      { testCollect(t) }
func TestNonInteractiveCollectAndSend(t *testing.T) { testNonInteractiveCollectAndSend(t) }
func TestInteractiveCollectAndSend(t *testing.T)    { testInteractiveCollectAndSend(t) }
