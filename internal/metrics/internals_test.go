package metrics

import (
	"flag"
	"path"
	"testing"

	"github.com/ubuntu/ubuntu-report/internal/helper"
)

/*
 * Tests here some private functions to gather metrics
 * Collect() public API is calling out a lot of functions,
 * that's why we add some unit tests on direct Collect() callees here
 * for finer-graind results in case of failure.
 */

var update = flag.Bool("update", false, "update golden files")

func TestInstallerInfo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string
	}{
		{"regular", "testdata/good"},
		{"empty file", "testdata/empty"},
		{"doesn't exist", "testdata/none"},
		{"garbage content", "testdata/garbage"},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			got := []byte(*m.installerInfo())
			want := helper.LoadOrUpdateGolden(path.Join(m.root, "gold", "intallerInfo"), got, *update, t)

			a.Equal(string(got), string(want))
		})
	}
}

func TestUpgradeInfo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string
	}{
		{"regular", "testdata/good"},
		{"empty file", "testdata/empty"},
		{"doesn't exist", "testdata/none"},
		{"garbage content", "testdata/garbage"},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			got := []byte(*m.upgradeInfo())
			want := helper.LoadOrUpdateGolden(path.Join(m.root, "gold", "upgradeInfo"), got, *update, t)

			a.Equal(string(got), string(want))
		})
	}
}

func TestGetVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string

		want string
	}{
		{"regular", "testdata/good", "18.04"},
		{"empty file", "testdata/empty", ""},
		{"missing", "testdata/missing-fields/ids/version", ""},
		{"empty", "testdata/empty-fields/ids/version", ""},
		{"doesn't exist", "testdata/none", ""},
		{"garbage content", "testdata/garbage", ""},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			got := m.getVersion()

			a.Equal(got, tc.want)
		})
	}
}

func TestGetRAM(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string

		want string
	}{
		{"regular", "testdata/good", "8048100"},
		{"empty file", "testdata/empty", ""},
		{"missing", "testdata/missing-fields/ram", ""},
		{"empty", "testdata/empty-fields/ram", ""},
		{"doesn't exist", "testdata/none", ""},
		{"garbage content", "testdata/garbage", ""},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			got := m.getRAM()

			a.Equal(got, tc.want)
		})
	}
}

func TestGetTimeZone(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string

		want string
	}{
		{"regular", "testdata/good", "Europe/Paris"},
		{"empty file", "testdata/empty", ""},
		{"doesn't exist", "testdata/none", ""},
		{"garbage content", "testdata/garbage", ""},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			got := m.getTimeZone()

			a.Equal(got, tc.want)
		})
	}
}

func TestGetAutologin(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string

		want string
	}{
		{"regular", "testdata/good", "false"},
		{"empty file", "testdata/empty", "false"},
		{"missing", "testdata/missing-fields/autologin", "false"},
		{"empty", "testdata/empty-fields/autologin", "false"},
		{"enabled", "testdata/specials/autologin/true", "true"},
		{"disabled", "testdata/specials/autologin/false", "false"},
		{"enabled no space", "testdata/specials/autologin/true-no-space", "true"},
		{"uppercase", "testdata/specials/autologin/true-uppercase", "true"},
		{"doesn't exist", "testdata/none", "false"},
		{"garbage content", "testdata/garbage", "false"},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			got := m.getAutologin()

			a.Equal(got, tc.want)
		})
	}
}

func TestGetOEM(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string

		wantVendor  string
		wantProduct string
	}{
		{"regular", "testdata/good", "DID", "4287CTO"},
		{"empty vendor", "testdata/empty-fields/oem/vendor", "", "4287CTO"},
		{"empty product", "testdata/empty-fields/oem/product", "DID", ""},
		{"empty both", "testdata/empty", "", ""},
		{"doesn't exist", "testdata/none", "", ""},
		{"garbage content", "testdata/garbage", "", ""},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			vendor, product := m.getOEM()

			a.Equal(vendor, tc.wantVendor)
			a.Equal(product, tc.wantProduct)
		})
	}
}

func TestGetBIOS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string

		wantVendor  string
		wantVersion string
	}{
		{"regular", "testdata/good", "DID", "42 (maybe 43)"},
		{"empty vendor", "testdata/empty-fields/bios/vendor", "", "42 (maybe 43)"},
		{"empty product", "testdata/empty-fields/bios/version", "DID", ""},
		{"empty both", "testdata/empty", "", ""},
		{"doesn't exist", "testdata/none", "", ""},
		{"garbage content", "testdata/garbage", "", ""},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			vendor, version := m.getBIOS()

			a.Equal(vendor, tc.wantVendor)
			a.Equal(version, tc.wantVersion)
		})
	}
}

func TestGetLivePatch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string

		want string
	}{
		{"regular", "testdata/good", "true"},
		{"disabled", "testdata/none", "false"},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			enabled := m.getLivePatch()

			a.Equal(enabled, tc.want)
		})
	}
}
func TestGetCPU(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string

		wantInfo []cpuInfo
	}{
		{"regular multi-core", "testdata/good", []cpuInfo{{"Genuine", "6", "42", "7"}}},
		{"one cpu one core", "testdata/specials/cpu/onecpu-onecore", []cpuInfo{{"Genuine", "6", "42", "7"}}},
		{"multi cpus", "testdata/specials/cpu/multicpus",
			[]cpuInfo{
				{"Genuine", "6", "42", "7"},
				{"Genuine2", "7", "42", "7"},
				{"Genuine3", "6", "1337", "7"},
				{"Genuine4", "6", "42", "8"},
			}},
		{"multi cpus multi core", "testdata/specials/cpu/multicpus-multicores",
			[]cpuInfo{
				{"Genuine", "6", "42", "7"},
				{"Genuine2", "7", "42", "7"},
				{"Genuine4", "6", "42", "8"},
			}},
		{"missing physical id", "testdata/missing-fields/cpu/physical-id", []cpuInfo{{"Genuine", "6", "42", "7"}}},
		{"missing vendor", "testdata/missing-fields/cpu/vendor", []cpuInfo{{"", "6", "42", "7"}}},
		{"missing family", "testdata/missing-fields/cpu/family", []cpuInfo{{"Genuine", "", "42", "7"}}},
		{"missing model", "testdata/missing-fields/cpu/model", []cpuInfo{{"Genuine", "6", "", "7"}}},
		{"missing stepping", "testdata/missing-fields/cpu/stepping", []cpuInfo{{"Genuine", "6", "42", ""}}},
		{"missing all", "testdata/missing-fields/cpu/all", nil},
		{"empty physical id", "testdata/empty-fields/cpu/physical-id", []cpuInfo{{"Genuine", "6", "42", "7"}}},
		{"empty vendor", "testdata/empty-fields/cpu/vendor", []cpuInfo{{"", "6", "42", "7"}}},
		{"empty family", "testdata/empty-fields/cpu/family", []cpuInfo{{"Genuine", "", "42", "7"}}},
		{"empty model", "testdata/empty-fields/cpu/model", []cpuInfo{{"Genuine", "6", "", "7"}}},
		{"empty stepping", "testdata/empty-fields/cpu/stepping", []cpuInfo{{"Genuine", "6", "42", ""}}},
		{"empty all", "testdata/empty-fields/cpu/all", nil},
		{"doesn't exist", "testdata/none", nil},
		{"garbage content", "testdata/garbage", nil},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, WithRootAt(tc.root))
			info := m.getCPUInfo()

			a.Equal(info, tc.wantInfo)
		})
	}
}

func newTestMetrics(t *testing.T, fixtures ...func(m *Metrics) error) Metrics {
	t.Helper()
	m, err := New(fixtures...)
	if err != nil {
		t.Fatal("can't create metrics object", err)
	}
	return m
}
