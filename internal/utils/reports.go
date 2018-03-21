package utils

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	defaultCacheDir = ".cache"
	reportDir       = "ubuntu-report"
)

var (
	// ErrFormat used to print debug messages.
	// Only for log.() msg, not errors.() error wrapping!
	ErrFormat = "%v"
)

// ReportPath of last saved report
func ReportPath(distro, version string, cacheP string) (string, error) {
	if cacheP == "" {
		var err error
		if cacheP, err = cacheDir(); err != nil {
			return "", err
		}
	}
	return filepath.Join(cacheP, reportDir, distro+"."+version), nil
}

func cacheDir() (string, error) {
	d := os.Getenv("XDG_CACHE_HOME")
	if filepath.IsAbs(d) {
		return d, nil
	}

	if d == "" {
		d = defaultCacheDir
	}
	h, err := getHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, d), nil
}

func getHomeDir() (string, error) {
	d := os.Getenv("HOME")

	if d == "" {
		usr, err := user.Current()
		if err != nil {
			return "", errors.Wrapf(err, "couldn't get user home directory")
		}
		d = usr.HomeDir
	}

	return d, nil
}
