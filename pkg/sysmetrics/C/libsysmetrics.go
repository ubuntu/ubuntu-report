package main

/*
// ReportType define the desired kind of interaction in CollectAndSend()
typedef enum {
    // ReportInteractive will show report content on stdout and read anwser on stdin
    ReportInteractive = 0,
    // ReportAuto will send a report without printing report
    ReportAuto = 1,
    // ReportOptOut will send opt-out message without printing report
    ReportOptOut = 2,
} ReportType;
*/
import "C"

import (
	"github.com/ubuntu/ubuntu-report/pkg/sysmetrics"
)

// generate shared library and header
//go:generate sh -c "go build -o ../../../build/libsysmetrics.so.1 -buildmode=c-shared -ldflags \"-extldflags -Wl,-soname,libsysmetrics.so.1\" libsysmetrics.go && mv ../../../build/libsysmetrics.so.h ../../../build/libsysmetrics.h"

// Collect system info and return a pretty printed version of collected data
//export Collect
func Collect(res **C.char) *C.char {
	b, err := sysmetrics.Collect()
	*res = C.CString(string(b))
	if err != nil {
		*res = C.CString("") // scratch data
		return C.CString(err.Error())
	}
	return nil
}

// CollectAndSend gather system info and send them
// alwaysReports bypass previous report already be sent for current version check
// It can be send to an alternative url via baseURL to send the report to, if not empty
//export CollectAndSend
func CollectAndSend(r C.ReportType, alwaysReport bool, baseURL *C.char) *C.char {
	err := sysmetrics.CollectAndSend(sysmetrics.ReportType(r), alwaysReport, C.GoString(baseURL))
	if err != nil {
		return C.CString(err.Error())
	}
	return nil
}

func main() {

}
