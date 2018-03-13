package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"

	"github.com/ubuntu/ubuntu-report/internal/metrics"
	"github.com/ubuntu/ubuntu-report/internal/utils"
)

func main() {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	log.SetLevel(log.ErrorLevel)

	flagCollectOnly := flag.BoolP("show-only", "s", false, "only show what would be reported")
	flagReportYes := flag.BoolP("yes", "y", false, "report automatically metrics without prompting")
	flagVerbosity := flag.CountP("verbose", "v", "report issue INFO (-v) or DEBUG (-vv) output")
	flagHelp := flag.BoolP("help", "h", false, "get this help")
	flagForce := flag.BoolP("force", "f", false, "install if even already reported")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [flags]:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if *flagHelp {
		flag.Usage()
		os.Exit(2)
	}

	if *flagVerbosity == 1 {
		log.SetLevel(log.InfoLevel)
	} else if *flagVerbosity > 1 {
		log.SetFormatter(&log.TextFormatter{})
		log.SetLevel(log.DebugLevel)
		log.Debug("verbosity set to debug and will print stacktraces")
		utils.ErrorFormat = "%+v"
	}

	if *flagCollectOnly && *flagReportYes {
		log.Error("couldn't use show-only and yes flags together as the first option disable reporting")
		flag.Usage()
		os.Exit(1)
	}

	if err := runTelemetry(*flagCollectOnly, *flagReportYes, *flagForce, utils.ErrorFormat); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func runTelemetry(collectOnly, autoReport, ignorePreviousReport bool, errorFormat string) error {

	// this error isn't a stopping us from reporting
	p, _ := utils.ReportPath()
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		log.Infof("previous report found in %s", p)
		if !ignorePreviousReport {
			return errors.Errorf("metrics from this machine have already been reported and can be found in: %s, "+
				"please use the --force flag if you really want to report them again.", p)
		}
		log.Debug("ignore previous report flag was set")
	}

	m, err := metrics.New()
	if err != nil {
		return errors.Errorf("couldn't create a metric collector: "+errorFormat, err)
	}

	data, err := m.Collect()
	if err != nil {
		return errors.Errorf("couldn't collect system minimal info: "+errorFormat, err)
	}

	if !collectOnly {
		fmt.Println("This is the result of hardware and optional installer/upgrader that we collected:")
	}

	if err := displayToUser(data); err != nil {
		return errors.Errorf("couldn't prettify json data: "+errorFormat, err)
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

	if p, err = utils.ReportPath(); err != nil {
		return errors.Errorf("couldn't get where to save reported metrics on disk: "+errorFormat, err)
	}
	if err := ioutil.WriteFile(p, data, 0666); err != nil {
		return errors.Errorf("couldn't save reported metrics on disk: "+errorFormat, err)
	}

	return nil
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
