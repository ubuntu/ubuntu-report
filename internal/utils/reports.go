package utils

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	defaultCacheDir = ".cache"
)

var (
	reportPath = filepath.Join("ubuntu-report", "report")

	// ErrorFormat used to print debug messages
	ErrorFormat = "%v"
)

// ReportPath of last saved report
func ReportPath() (string, error) {
	d, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, reportPath), nil
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
