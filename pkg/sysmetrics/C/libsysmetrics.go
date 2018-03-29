package main

/*
// sysmetrics_report_type define the desired kind of interaction in sysmetrics_collect_and_send()
typedef enum {
    // sysmetrics_report_interactive will show report content on stdout and read anwser on stdin
    sysmetrics_report_interactive = 0,
    // sysmetrics_report_auto will send a report without printing report
    sysmetrics_report_auto = 1,
    // sysmetrics_report_optout will send opt-out message without printing report
    sysmetrics_report_optout = 2,
} sysmetrics_report_type;
*/
import "C"

import (
	"github.com/ubuntu/ubuntu-report/pkg/sysmetrics"
)

// generate shared library and header
//go:generate sh -c "go build -o ../../../build/libsysmetrics.so.1 -buildmode=c-shared -ldflags \"-extldflags -Wl,-soname,libsysmetrics.so.1\" libsysmetrics.go && mv ../../../build/libsysmetrics.so.h ../../../build/libsysmetrics.h"

// sysmetrics_collect system info and return a pretty printed version of collected data
//export sysmetrics_collect
func sysmetrics_collect(res **C.char) *C.char {
	b, err := sysmetrics.Collect()
	*res = C.CString(string(b))
	if err != nil {
		*res = C.CString("") // scratch data
		return C.CString(err.Error())
	}
	return nil
}

// sysmetrics_collect_and_send gather system info and send them
// alwaysReports bypass previous report already be sent for current version check
// It can be send to an alternative url via baseURL to send the report to, if not empty
//export sysmetrics_collect_and_send
func sysmetrics_collect_and_send(r C.sysmetrics_report_type, alwaysReport bool, baseURL *C.char) *C.char {
	err := sysmetrics.CollectAndSend(sysmetrics.ReportType(r), alwaysReport, C.GoString(baseURL))
	if err != nil {
		return C.CString(err.Error())
	}
	return nil
}

func main() {

}
