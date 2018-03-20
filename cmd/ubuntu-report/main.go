package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ubuntu/ubuntu-report/internal/utils"
)

func main() {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	log.SetLevel(log.ErrorLevel)

	rootCmd := generateRootCmd()

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func generateRootCmd() *cobra.Command {
	var flagForce bool
	var flagVerbosity int

	var rootCmd = &cobra.Command{
		Use:   "ubuntu-report",
		Short: "Report metrics from your system, install and upgrades",
		Long: `This tool will collect and report metrics from current hardware,` +
			`partition and session information.` + "\n" +
			`Those information can't be used to identify a single machine and` +
			`are presented before being sent to the server.`,
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
			if err := Report(reportInteractive, flagForce, ""); err != nil {
				log.Errorf(utils.ErrFormat, err)
				os.Exit(1)
			}
		},
	}
	rootCmd.PersistentFlags().CountVarP(&flagVerbosity, "verbose", "v", "issue INFO (-v) and DEBUG (-vv) output")
	rootCmd.PersistentFlags().BoolVarP(&flagForce, "force", "f", false, "collect and send new report even if already reported")

	show := &cobra.Command{
		Use:   "show",
		Short: "Only collect and display metrics without sending",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := Collect()
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
		Short: "Send or opt-out directly from metrics report without interactions",

		// we want exactly one arg by in ValidArgs list
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || !stringInSlice(args[0], cmd.ValidArgs) {
				return fmt.Errorf("Only accept one argument: yes or no, received '%s'", strings.Join(args, " "))
			}
			return nil
		},
		ValidArgs: []string{"yes", "no"},

		Run: func(cmd *cobra.Command, args []string) {
			var r reportType
			if args[0] == "yes" {
				r = reportAuto
			} else if args[0] == "no" {
				r = reportOptOut
			} else {
				log.Error("Invalid arg")
				os.Exit(1)
			}

			if err := Report(r, flagForce, ""); err != nil {
				log.Errorf(utils.ErrFormat, err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(send)

	interactiveCmd := &cobra.Command{
		Use:   "interactive",
		Short: "Interactive mode, similar to running this tool without any subcommands",
		Run:   rootCmd.Run,
	}
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
