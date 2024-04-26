// Package daemon represents the oidc broker binary
package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubuntu/decorate"
	"github.com/ubuntu/ubuntu-report/internal/consts"
	"github.com/ubuntu/ubuntu-report/internal/daemon"
)

// cmdName is the binary name for the agent.
var cmdName = filepath.Base(os.Args[0])

// App encapsulate commands and options of the daemon, which can be controlled by env variables and config files.
type App struct {
	rootCmd cobra.Command
	viper   *viper.Viper
	config  daemonConfig

	daemon *daemon.Daemon

	ready chan struct{}
}

// only overriable for tests.
type systemPaths struct {
	DaemonConf  string
	LogDir      string
	IncomingDir string
}

// daemonConfig defines configuration parameters of the daemon.
type daemonConfig struct {
	Verbosity  int
	Paths      systemPaths
	ServerPort int
	Distros    []string
	Variants   []string
}

// New registers commands and return a new App.
func New() *App {
	a := App{ready: make(chan struct{})}
	a.rootCmd = cobra.Command{
		Use:   fmt.Sprintf("%s COMMAND", cmdName),
		Short: fmt.Sprintf("%s ubuntu report collector service", cmdName),
		Long:  fmt.Sprintf("Collector service %s to receive report data from ubuntu-report.", cmdName),
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Command parsing has been successful. Returns to not print usage anymore.
			a.rootCmd.SilenceUsage = true

			// Set config defaults
			systemLogDir := filepath.Join("/var", "log", cmdName)
			a.config = daemonConfig{
				Paths: systemPaths{
					DaemonConf: filepath.Join(consts.DefaultCollectorConfPath, cmdName),
					LogDir:     systemLogDir,
				},
			}

			// Install and unmarshall configuration
			if err := initViperConfig(cmdName, &a.rootCmd, a.viper); err != nil {
				return err
			}
			if err := a.viper.Unmarshal(&a.config); err != nil {
				return fmt.Errorf("unable to decode configuration into struct: %w", err)
			}

			setVerboseMode(a.config.Verbosity)
			slog.Debug("Debug mode is enabled")

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.serve(a.config)
		},
		// We display usage error ourselves
		SilenceErrors: true,
	}
	viper := viper.New()

	a.viper = viper

	installVerbosityFlag(&a.rootCmd, a.viper)
	installConfigFlag(&a.rootCmd)

	// subcommands
	a.installVersion()

	return &a
}

// serve creates an instance of the daemon. This call is blocking until we quit it.
func (a *App) serve(config daemonConfig) error {
	ctx := context.Background()

	if err := ensureDirWithPerms(config.Paths.LogDir, 0755); err != nil {
		close(a.ready)
		return fmt.Errorf("error initializing log directory at %q: %v", config.Paths.LogDir, err)
	}

	config.Paths.IncomingDir = filepath.Join(config.Paths.LogDir, "incoming")
	if err := ensureDirWithPerms(config.Paths.IncomingDir, 0755); err != nil {
		close(a.ready)
		return fmt.Errorf("error initializing log incoming directory at %q: %v", config.Paths.IncomingDir, err)
	}

	var daemonopts []daemon.Option
	daemon, err := daemon.New(ctx, daemonopts...)
	if err != nil {
		close(a.ready)
		return err
	}

	a.daemon = daemon
	close(a.ready)

	slog.Debug(fmt.Sprintf("Accepted distros: %v", config.Distros))
	slog.Debug(fmt.Sprintf("Accepted variants %v", config.Variants))
	return daemon.Serve(ctx, config.ServerPort, config.Paths.IncomingDir, config.Distros, config.Variants)
}

// installVerbosityFlag adds the -v and -vv options and returns the reference to it.
func installVerbosityFlag(cmd *cobra.Command, viper *viper.Viper) *int {
	r := cmd.PersistentFlags().CountP("verbosity", "v", "issue INFO (-v), DEBUG (-vv) or DEBUG with caller (-vvv) output")
	decorate.LogOnError(viper.BindPFlag("verbosity", cmd.PersistentFlags().Lookup("verbosity")))
	return r
}

// Run executes the command and associated process. It returns an error on syntax/usage error.
func (a *App) Run() error {
	return a.rootCmd.Execute()
}

// UsageError returns if the error is a command parsing or runtime one.
func (a App) UsageError() bool {
	return !a.rootCmd.SilenceUsage
}

// Hup prints all goroutine stack traces and return false to signal you shouldn't quit.
func (a App) Hup() error {
	return a.daemon.RotateLog()
}

// Quit gracefully shutdown the service.
func (a *App) Quit() {
	a.WaitReady()
	if a.daemon == nil {
		return
	}
	a.daemon.Quit()
}

// WaitReady signals when the daemon is ready
// Note: we need to use a pointer to not copy the App object before the daemon is ready, and thus, creates a data race.
func (a *App) WaitReady() {
	<-a.ready
}

// RootCmd returns a copy of the root command for the app. Shouldn't be in general necessary apart when running generators.
func (a App) RootCmd() cobra.Command {
	return a.rootCmd
}
