package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ubuntu/ubuntu-report/internal/metrics"
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
	var flagCollectOnly, flagReportYes, flagForce bool
	var flagVerbosity int

	var rootCmd = &cobra.Command{
		Use:   "ubuntu-report",
		Short: "Report metrics from your system, install and upgrades",
		Long: `This tool will collect and report metrics from current hardware,` +
			`partition and session information.` + "\n" +
			`Those information can't be used to identify a single machine and` +
			`are presented before being sent to the server.`,
		Run: func(cmd *cobra.Command, args []string) {
			if flagVerbosity == 1 {
				log.SetLevel(log.InfoLevel)
			} else if flagVerbosity > 1 {
				log.SetFormatter(&log.TextFormatter{})
				log.SetLevel(log.DebugLevel)
				log.Debug("verbosity set to debug and will print stacktraces")
				utils.ErrFormat = "%+v"
			}

			if flagCollectOnly && flagReportYes {
				log.Error("can't use show-only and yes flags together as the first option disable reporting")
				cmd.Usage()
				os.Exit(1)
			}

			if err := runTelemetry(flagCollectOnly, flagReportYes, flagForce); err != nil {
				log.Errorf(utils.ErrFormat, err)
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().BoolVarP(&flagCollectOnly, "show-only", "s", false, "only show what would be reported")
	rootCmd.Flags().BoolVarP(&flagReportYes, "yes", "y", false, "report automatically metrics without prompting")
	rootCmd.Flags().CountVarP(&flagVerbosity, "verbose", "v", "issue INFO (-v) and DEBUG (-vv) output")
	rootCmd.Flags().BoolVarP(&flagForce, "force", "f", false, "install if even already reported")

	return rootCmd
}

func runTelemetry(collectOnly, autoReport, ignorePreviousReport bool) error {

	m, err := metrics.New()
	if err != nil {
		return errors.Wrapf(err, "couldn't create a metric collector")
	}
	distro, version, err := m.GetIDS()
	if err != nil {
		return errors.Wrapf(err, "couldn't get mandatory information")
	}
	// this error isn't a stopping us from reporting
	reportP, err := utils.ReportPath(distro, version)
	if err != nil {
		return errors.Wrapf(err, "couldn't get where to save reported metrics on disk")
	}
	if _, err := os.Stat(reportP); !os.IsNotExist(err) {
		log.Infof("previous report found in %s", reportP)
		if !ignorePreviousReport {
			return errors.Errorf("metrics from this machine have already been reported and can be found in: %s, "+
				"please use the --force flag if you really want to report them again.", reportP)
		}
		log.Debug("ignore previous report flag was set")
	}

	data, err := m.Collect()
	if err != nil {
		return errors.Wrapf(err, "couldn't collect system minimal info")
	}

	if !collectOnly {
		fmt.Println("This is the result of hardware and optional installer/upgrader that we collected:")
	}

	if err := displayToUser(data); err != nil {
		return errors.Wrapf(err, "couldn't prettify json data")
	}

	if collectOnly {
		log.Debug("show only flag, no more to do")
		return nil
	}

	if !autoReport {
		validAnswer := false
		scanner := bufio.NewScanner(os.Stdin)
		for validAnswer != true {
			fmt.Printf("Do you agree to report this? [y/N] ")
			if !scanner.Scan() {
				log.Info("programm interrupted")
				return nil
			}
			text := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if text == "n" || text == "no" || text == "" {
				log.Debug("sending report was denied")
				return nil
			} else if text == "y" || text == "yes" {
				log.Debug("sending report was accepted")
				validAnswer = true
			}
			if validAnswer != true {
				log.Error("we didn't understand your answer")
			}
		}
	} else {
		log.Debug("report yes flag was set")
	}

	/*if err := sender.Send(sender.URL, data); err != nil {
		return errors.Errorf("data were not delivered successfully to metrics server: "+errorFormat, err)
	}*/

	return saveMetrics(reportP, data)
}

func displayToUser(d []byte) error {
	log.Debug("pretty print the collected data to the user")
	h := json.RawMessage(d)
	b, err := json.MarshalIndent(&h, "", "  ")
	if err != nil {
		return err
	}
	os.Stdout.Write(b)
	fmt.Println()
	return nil
}

func saveMetrics(p string, data []byte) error {
	log.Debugf("save sent metrics to %s", p)

	d := filepath.Dir(p)
	if err := os.MkdirAll(d, 0700); err != nil {
		return errors.Wrap(err, "couldn't create parent directory to save reported metrics")
	}

	if err := ioutil.WriteFile(p, data, 0666); err != nil {
		return errors.Wrap(err, "couldn't save reported metrics on disk")
	}

	return nil
}
