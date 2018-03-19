package metrics_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const (
	garbageOutput = `fdsofhoidshf fods gfpds
gpofgipogifd
fdspfds

gfoidgo
gfdojoi`
)

// TestMetricsHelperProcess is available to both internal and package test
// to mock binaries (it's sub-executed)
func TestMetricsHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "lspci":
		if args[0] != "-n" {
			fmt.Fprintf(os.Stderr, "Unexpected lspci arguments: %v\n", args)
		}
		regularOutput := `00:00.0 0600: 8086:0104 (rev 09)
00:02.0 0300: 8086:0126 (rev 09)
00:16.0 0780: 8086:1c3a (rev 04)
00:16.3 0700: 8086:1c3d (rev 04)
00:19.0 0200: 8086:1502 (rev 04)`
		switch args[1] {
		case "onegpu":
			fmt.Println(regularOutput)
		case "multiplegpus":
			fmt.Println(regularOutput)
			fmt.Println("00:02.0 0300: 8086:0127 (rev 09)")
		case "nogpu":
			fmt.Println(`00:00.0 0600: 8086:0104 (rev 09)
00:16.0 0780: 8086:1c3a (rev 04)
00:16.3 0700: 8086:1c3d (rev 04)
00:19.0 0200: 8086:1502 (rev 04)`)
		case "empty":
		case "malformed gpu line":
			fmt.Println("00:02.0 0300: 80860127 (rev 09)")
		case "garbage":
			fmt.Println(garbageOutput)
		case "fail":
			fmt.Println(regularOutput) // still print content
			os.Exit(1)
		}

	case "xrandr":
		regularOutput := `Screen 0: minimum 320 x 200, current 1920 x 1848, maximum 8192 x 8192
LVDS-1 connected primary 1366x768+0+1080 (normal left inverted right x axis y axis) 277mm x 156mm
   1366x768      60.02*+
   1360x768      59.80    59.96  
   1024x768      60.04    60.00  
HDMI-1 disconnected (normal left inverted right x axis y axis)
DP-1 disconnected (normal left inverted right x axis y axis)`
		switch args[0] {
		case "one screen":
			fmt.Println(regularOutput)
		case "multiple screens":
			fmt.Println(regularOutput)
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis) 510mm x 287mm
   1920x1080     60.00*+
   1600x1200     60.00  
   1680x1050     59.95  `)
		case "no screen":
			fmt.Println("")
		case "chosen resolution not first":
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis) 510mm x 287mm
   1920x1080     60.00  
   1600x1200     60.00*+
   1680x1050     59.95  `)
		case "no chosen resolution":
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis) 510mm x 287mm
   1920x1080     60.00  
   1600x1200     60.00  
   1680x1050     59.95  `)
		case "empty":
		case "malformed screen line":
			fmt.Println(strings.Replace(regularOutput, " ", "", -1))
		case "garbage":
			fmt.Println(garbageOutput)
		case "fail":
			fmt.Println(regularOutput) // still print content
			os.Exit(1)
		}
	}
}
