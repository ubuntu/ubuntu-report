package metrics_test

import (
	"testing"

	"github.com/ubuntu/ubuntu-report/internal/helper"
	"github.com/ubuntu/ubuntu-report/internal/metrics"
)

func TestGetIDS(t *testing.T) {

	testCases := []struct {
		name string
		root string

		wantDistro  string
		wantVersion string
		wantErr     bool
	}{
		{"regular", "testdata/good", "ubuntu", "18.04", false},
		{"doesn't exist", "testdata/none", "", "", true},
		{"empty file", "testdata/empty", "", "", true},
		{"missing distro", "testdata/missing/ids/distro", "", "", true},
		{"missing version", "testdata/missing/ids/version", "", "", true},
		{"missing both", "testdata/missing/ids/both", "", "", true},
		{"empty distro", "testdata/empty-fields/ids/distro", "", "", true},
		{"empty version", "testdata/empty-fields/ids/version", "", "", true},
		{"empty both", "testdata/empty-fields/ids/both", "", "", true},
		{"garbage content", "testdata/garbage", "", "", true},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, metrics.WithRootAt(tc.root))
			d, v, err := m.GetIDS()

			a.CheckWantedErr(err, tc.wantErr)
			a.Equal(d, tc.wantDistro)
			a.Equal(v, tc.wantVersion)
		})
	}
}

func newTestMetrics(t *testing.T, fixtures ...func(m *metrics.Metrics) error) metrics.Metrics {
	t.Helper()
	m, err := metrics.New(fixtures...)
	if err != nil {
		t.Fatal("can't create metrics object", err)
	}
	return m
}
