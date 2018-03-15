package utils_test

import (
	"os"
	"os/user"
	"testing"

	"github.com/ubuntu/ubuntu-report/internal/utils"
)

func TestReportPath(t *testing.T) {

	// get current user for some tests
	u, err := user.Current()
	if err != nil {
		t.Fatalf("couldn't get current user for testing: %v", err)
	}

	testCases := []struct {
		name          string
		home          string
		xdg_cache_dir string
		distro        string
		version       string

		want    string
		wantErr bool
	}{
		{"regular", "/some/dir", "", "distroname", "versionnumber", "/some/dir/.cache/ubuntu-report/distroname.versionnumber", false},
		{"relative xdg path", "/some/dir", "xdg_cache_path", "distroname", "versionnumber", "/some/dir/xdg_cache_path/ubuntu-report/distroname.versionnumber", false},
		{"absolute xdg path", "/some/dir", "/xdg_cache_path", "distroname", "versionnumber", "/xdg_cache_path/ubuntu-report/distroname.versionnumber", false},
		{"no home dir", "", "", "distroname", "versionnumber", u.HomeDir + "/.cache/ubuntu-report/distroname.versionnumber", false},
		{"no distro name", "/some/dir", "", "", "versionnumber", "/some/dir/.cache/ubuntu-report/.versionnumber", false},
		{"no version name", "/some/dir", "", "distroname", "", "/some/dir/.cache/ubuntu-report/distroname.", false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer changeEnv(t, "HOME", tc.home)()
			defer changeEnv(t, "XDG_CACHE_HOME", tc.xdg_cache_dir)()

			got, err := utils.ReportPath(tc.distro, tc.version)

			if err != nil && !tc.wantErr {
				t.Fatal("got an unexpected err:", err)
			}
			if err == nil && tc.wantErr {
				t.Error("expected an error and got none")
			}
			if got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}

func changeEnv(t *testing.T, key, value string) func() {
	orig := os.Getenv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("Couldn't change %s env to %s: %v", key, value, err)
	}

	return func() {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Couldn't restore %s env to %s: %v", key, orig, err)
		}
	}
}
