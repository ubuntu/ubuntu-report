package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ubuntu/ubuntu-report/internal/sender"
	"github.com/ubuntu/ubuntu-report/internal/utils"
	"github.com/ubuntu/ubuntu-report/pkg/sysmetrics"
)

// generate README, shell completion and manpages
//go:generate go test . --generate --path ../../build/

func main() {
	rootCmd := generateRootCmd()

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func generateRootCmd() *cobra.Command {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	log.SetLevel(log.ErrorLevel)

	var flagForce bool
	var flagVerbosity int
	var flagServerURL string

	var rootCmd = &cobra.Command{
		Use:   "ubuntu-report",
		Short: "Report metrics from your system, install and upgrades",
		Long: `This tool will collect and report metrics from current hardware, ` +
			`partition and session information.` + "\n" +
			`This information can't be used to identify a single machine and ` +
			`is presented before being sent to the server.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if flagVerbosity == 1 {
				log.SetLevel(log.InfoLevel)
			} else if flagVerbosity > 1 {
				log.SetFormatter(&log.TextFormatter{})
				log.SetLevel(log.DebugLevel)
				log.Debug("verbosity set to debug and will print stacktraces")
				utils.ErrFormat = "%+v"
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := sysmetrics.CollectAndSend(sysmetrics.ReportInteractive, flagForce, flagServerURL); err != nil {
				log.Errorf(utils.ErrFormat, err)
				os.Exit(1)
			}
		},
	}
	rootCmd.PersistentFlags().CountVarP(&flagVerbosity, "verbose", "v", "issue INFO (-v) and DEBUG (-vv) output")
	rootCmd.PersistentFlags().BoolVarP(&flagForce, "force", "f", false, "collect and send new report even if already reported")

	rootCmd.Flags().StringVarP(&flagServerURL, "url", "u", sender.BaseURL, "server url to send report to. Leave empty for default.")

	show := &cobra.Command{
		Use:   "show",
		Short: "Only collect and display metrics without sending",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := sysmetrics.Collect()
			if err != nil {
				log.Errorf(utils.ErrFormat, err)
				os.Exit(1)
			}
			fmt.Println(string(data))
		},
	}
	rootCmd.AddCommand(show)

	send := &cobra.Command{
		Use:   "send yes|no",
		Short: "Send or opt-out directly from metric reports without interactions",

		// we want exactly one arg by in ValidArgs list or upgrade internal command
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || !stringInSlice(args[0], append(cmd.ValidArgs, "upgrade")) {
				return fmt.Errorf("Only accept one argument: yes or no, received '%s'", strings.Join(args, " "))
			}
			return nil
		},
		ValidArgs: []string{"yes", "no"},

		Run: func(cmd *cobra.Command, args []string) {
			var r sysmetrics.ReportType
			if args[0] == "yes" {
				r = sysmetrics.ReportAuto
			} else if args[0] == "no" {
				r = sysmetrics.ReportOptOut
			} else if args[0] == "upgrade" {
				if err := sysmetrics.CollectAndSendOnUpgrade(flagForce, flagServerURL); err != nil {
					// log a warning, but don't error out as this is an automated upgrade call
					log.Warningf(utils.ErrFormat, err)
				}
				return
			} else {
				log.Error("Invalid arg")
				os.Exit(1)
			}

			if err := sysmetrics.CollectAndSend(r, flagForce, flagServerURL); err != nil {
				log.Errorf(utils.ErrFormat, err)
				os.Exit(1)
			}
		},
	}
	send.Flags().StringVarP(&flagServerURL, "url", "u", sender.BaseURL, "server url to send report to. Leave empty for default.")
	rootCmd.AddCommand(send)

	service := &cobra.Command{
		Use:    "service",
		Short:  "Try to send periodically previously collected data once network if previous send was unsuccessful",
		Args:   cobra.NoArgs,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := sysmetrics.SendPendingReport(flagServerURL)
			if err != nil {
				log.Errorf(utils.ErrFormat, err)
				os.Exit(1)
			}
		},
	}
	service.Flags().StringVarP(&flagServerURL, "url", "u", sender.BaseURL, "server url to send report to. Leave empty for default.")
	rootCmd.AddCommand(service)

	interactiveCmd := &cobra.Command{
		Use:   "interactive",
		Short: "Interactive mode, alias to running this tool without any subcommands.",
		Run:   rootCmd.Run,
	}
	interactiveCmd.Flags().StringVarP(&flagServerURL, "url", "u", sender.BaseURL, "server url to send report to. Leave empty for default.")
	rootCmd.AddCommand(interactiveCmd)

	return rootCmd
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
