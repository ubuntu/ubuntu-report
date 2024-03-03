package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"

	"github.com/gorilla/mux"
	"github.com/ubuntu/ubuntu-report/internal/metrics"
	"golang.org/x/exp/slog"
)

// httpLogger logs the incoming request
func httpLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info(fmt.Sprintf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// healthCheckHandler handles the health check requests
func (d *Daemon) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// Perform any self-test or health check logic here
	// For simplicity, just sending a confirmation response
	// But ultimately we should verify that we can log records
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Health Check: Service is running")
}

// submitHandler handles the POST requests
func (d *Daemon) submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	reject := false

	// Retrieve distro/variant/version from URL
	urlvars := mux.Vars(r)
	distro := urlvars["distro"]
	variant := urlvars["variant"]
	version := urlvars["version"]

	// Reject unsupported distros and variants

	if !slices.Contains(d.distros, distro) {
		reject = true
	}
	if !slices.Contains(d.variants, variant) {
		reject = true
	}

	// TODO Reject invalid versions
	if !validateVersion(version) {
		reject = true
	}

	// Read the body of the request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Reject invalid json
	var metrics metrics.MetricsData
	if err := json.Unmarshal(body, &metrics); err != nil {
		http.Error(w, "Error parsing JSON body", http.StatusBadRequest)
		reject = true
	}

	// Serialize and write to a log file
	d.writeToLogFile(reject, distro, variant, version, metrics)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("POST request processed\n"))
}

// validateVersion verifies the the version of Ubuntu is valid
// it's broader than the standard Ubuntu version for flavours
// that amy use slightly different numbering
func validateVersion(s string) bool {
	pattern := `^\d\d\.(0[1-9]|1[0-2])$`
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		slog.Error(fmt.Sprintf("Error compiling regex: %v", err))
		return false
	}

	return matched
}
