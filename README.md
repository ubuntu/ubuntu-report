# Ubuntu Report
Report hardware and other collected metrics like installer or upgrade information.

This information can't be used to identify a single machine and is presented before being sent to the server.

[![Build & Tests Status](https://travis-ci.org/ubuntu/ubuntu-report.svg?branch=master)](https://travis-ci.org/ubuntu/ubuntu-report)
[![codecov](https://codecov.io/gh/ubuntu/ubuntu-report/branch/master/graph/badge.svg)](https://codecov.io/gh/ubuntu/ubuntu-report)
[![Go Report Card](https://goreportcard.com/badge/github.com/ubuntu/ubuntu-report)](https://goreportcard.com/report/github.com/ubuntu/ubuntu-report)

APIS:
 * Go: [![Go API](https://godoc.org/github.com/ubuntu/ubuntu-report?status.svg)](https://godoc.org/github.com/ubuntu/ubuntu-report/pkg/sysmetrics)
 * C: [![C API](https://godoc.org/github.com/ubuntu/ubuntu-report?status.svg)](https://godoc.org/github.com/ubuntu/ubuntu-report/pkg/sysmetrics/C)


## About

The tool will show you what is going to be reported and ask for your acknowledgement before uploading it. It will be mainly
invoked by a GUI, but it provides also a command line tool.

The Ubuntu welcome UI has a dedicated panel for this report collection and upload.

The command line tool as well as the Go and C API have different modes:
* Interactive mode (prompt displaying the data being sent and ask if sending or opting out)
* Only show the report
* Report automatically the collected data without prompting
* Report that you have opted out of data collection

By default, you can only report your data collection once per distribution version.

To execute the interactive command line interface manually, just use `ubuntu-report`.

## Command line usage

### ubuntu-report

Report metrics from your system, install and upgrades

#### Synopsis

This tool will collect and report metrics from current hardware, partition and session information.
This information can't be used to identify a single machine and is presented before being sent to the server.

```
ubuntu-report [flags]
```

#### Options

```
  -f, --force           collect and send new report even if already reported
  -h, --help            help for ubuntu-report
  -u, --url string      server url to send report to. Leave empty for default. (default "https://metrics.ubuntu.com")
  -v, --verbose count   issue INFO (-v) and DEBUG (-vv) output
```

### ubuntu-report interactive

Interactive mode, alias to running this tool without any subcommands.

### ubuntu-report send

Send or opt-out directly from metric reports without interactions

#### Synopsis

Send or opt-out directly from metric reports without interactions

```
ubuntu-report send yes|no [flags]
```

#### Options

```
  -h, --help         help for send
  -u, --url string   server url to send report to. Leave empty for default. (default "https://metrics.ubuntu.com")
```

#### Options inherited from parent commands

```
  -f, --force           collect and send new report even if already reported
  -v, --verbose count   issue INFO (-v) and DEBUG (-vv) output
```

### ubuntu-report service

Try to send periodically previously unsent but collected data once network is avaiable

#### Synopsis

Try to send periodically previously unsent but collected data once network is avaiable

```
ubuntu-report service [flags]
```

#### Options

```
  -h, --help         help for service
  -u, --url string   server url to send report to. Leave empty for default. (default "https://metrics.ubuntu.com")
```

#### Options inherited from parent commands

```
  -f, --force           collect and send new report even if already reported
  -v, --verbose count   issue INFO (-v) and DEBUG (-vv) output
```

### ubuntu-report show

Only collect and display metrics without sending

#### Synopsis

Only collect and display metrics without sending

```
ubuntu-report show [flags]
```

#### Options

```
  -h, --help   help for show
```

#### Options inherited from parent commands

```
  -f, --force           collect and send new report even if already reported
  -v, --verbose count   issue INFO (-v) and DEBUG (-vv) output
```

## Service

In case we can't report (due to limited network or other networking conditions) your report when you act on it,
a little service will kick at login, and try to send the pending report data again. Note that it will exponentially
back off.

The service won't be active once the pending report is sent.

## APIS

### Go API

The Go API is used by the command line, but can be embedded as well by 3rd parties. Doc reference is available at
[![this link](https://godoc.org/github.com/ubuntu/ubuntu-report?status.svg)](https://godoc.org/github.com/ubuntu/ubuntu-report/pkg/sysmetrics).

### C API

The C API is provided for embedding the library in C code. Doc reference is available at
[![this link](https://godoc.org/github.com/ubuntu/ubuntu-report?status.svg)](https://godoc.org/github.com/ubuntu/ubuntu-report/pkg/sysmetrics/C).

You can generate the shared library and headers by running `go generate`.

## Command line options

You can regenerate previous README section, shell completion support and man pages by simply running `go generate`.

## Send data to server

### Example of data being sent if agreement is accepted

The data are pretty printed here to be more readable.

```json
{
  "Version": "18.04",
  "OEM": {
    "Vendor": "Vendor Name",
    "Product": "4287CTO"
  },
  "BIOS": {
    "Vendor": "Vendor Name",
    "Version": "8DET52WW (1.27)"
  },
  "CPU": {
    "OpMode": "32-bit, 64-bit",
    "CPUs": "8",
    "Threads": "2",
    "Cores": "4",
    "Sockets": "1",
    "Vendor": "Genuine",
    "Family": "6",
    "Model": "158",
    "Stepping": "10",
    "Name": "Intius Corus i5-8300H CPU @ 2.30GHz",
    "Virtualization": "VT-x"
  },
  "Arch": "amd64",
  "GPU": [
    {
      "Vendor": "8086",
      "Model": "0126"
    }
  ],
  "RAM": 8,
  "Disks": [
    240.1,
    500.1
  ],
  "Partitions": [
    229.2,
    479.7
  ],
  "Screens": [
    {
      "Size": "277mmx156mm",
      "Resolution": "1366x768",
      "Frequency": "60.02"
    },
    {
      "Resolution": "1920x1080",
      "Frequency": "60.00"
    }
  ],
  "Autologin": false,
  "LivePatch": true,
  "Session": {
    "DE": "ubuntu:GNOME",
    "Name": "ubuntu",
    "Type": "x11"
  },
  "Language": "fr_FR",
  "Timezone": "Europe/Paris",
  "Install": {
    "Media": "Ubuntu 18.04 LTS \"Bionic Beaver\" - Alpha amd64 (20180305)",
    "Type": "GTK",
    "PartitionMethod": "use_device",
    "DownloadUpdates": true,
    "Language": "fr",
    "Minimal": false,
    "RestrictedAddons": false,
    "Stages": {
      "0": "language",
      "3": "language",
      "10": "console_setup",
      "15": "prepare",
      "25": "partman",
      "27": "start_install",
      "37": "timezone",
      "49": "usersetup",
      "57": "user_done",
      "829": "done"
    }
  }
}
```

### Data being sent if agreement is denied

The data are pretty printed here to be more readable.

```json
{
  "OptOut": true
}
```
