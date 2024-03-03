// Package daemon handles the http service.
package daemon

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/ubuntu/decorate"
)

// Daemon is an http server.
type Daemon struct {
	httpServer http.Server
	logDir     string     // LogDir  stores the log files. Files are actually stored in /incoming/
	logFile    *os.File   // logFile is a pointer to the currently open log file
	logMutex   sync.Mutex // logMutex ensures that there is no concurrent write to the log file

	distros  []string
	variants []string
}

type options struct {
}

// Option is the function signature used to tweak the daemon creation.
type Option func(*options)

// New returns an new, initialized daemon server, which handles systemd activation.
// If systemd activation is used, it will override any socket passed here.
func New(ctx context.Context, args ...Option) (d *Daemon, err error) {
	defer decorate.OnError(&err, "can't create daemon")

	slog.Debug("Building new daemon")

	// Set default options.
	opts := options{}
	// Apply given args.
	for _, f := range args {
		f(&opts)
	}

	return &Daemon{}, nil
}

// Serve starts the http server.
func (d *Daemon) Serve(ctx context.Context, httpPort int, logDir string, distros, variants []string) (err error) {
	defer decorate.OnError(&err, "error while serving")

	d.logDir = logDir
	d.distros = distros
	d.variants = variants

	// Setup HTTP server
	r := mux.NewRouter()
	r.Use(httpLogger)
	r.HandleFunc("/health", d.healthCheckHandler)
	r.HandleFunc("/{distro}/{variant}/{version}", d.submitHandler).Methods("POST")

	slog.Debug(fmt.Sprintf("Starting server on %d", httpPort))
	d.httpServer = http.Server{
		Addr:         ":" + strconv.Itoa(httpPort),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		if err := d.httpServer.ListenAndServe(); err != nil {
			log.Fatal("Server failed to start: ", err)
		}
	}()
	slog.Debug("Starting to serve requests")
	d.initializeLogFile()

	// Set up a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a signal
	<-stop
	return nil
}

// Quit gracefully quits listening loop and stops the grpc server.
// It can drops any existing connexion is force is true.
func (d *Daemon) Quit() {
	slog.Info("Stopping daemon requested.")
	d.httpServer.Shutdown(context.Background())
}
