// Package sysmetrics C bindings: collect and report system and hardware metrics
// from your system.
//
// Collect system info
//
// Command signature:
//   char* Collect(char** res);
// "res" will be the pretty printed version of collected data. The return "err" err will be != NULL in case
// any error occurred during data collection.
//
// Example:
//   #include <stdio.h>
//   #include <stdlib.h>
//   #include "libsysmetrics.h"
//
//   int main() {
//       char *res, *err;
//
//       err = Collect(&res);
//
//       if (err != NULL) {
//           printf("ERR: %s\n", err);
//       } else {
//           printf("GOT: %s\n", res);
//       }
//       free(res);
//       free(err);
//   }
//
// CollectAndSend gather system info and send them
//
// Command signature:
//   char* CollectAndSend(ReportType r, bool alwaysreport, char* url);
//
// ReportType is the following enum:
//    typedef enum {
//      // ReportInteractive will show report content on stdout and read anwser on stdin
//      ReportInteractive = 0,
//      // ReportAuto will send a report without printing report
//      ReportAuto = 1,
//      // ReportOptOut will send opt-out message without printing report
//      ReportOptOut = 2,
//    } ReportType;
// You should generally prefer in bindings the Auto or OptOut report. Interactive is based on stdout and stdin.
// "alwaysReports" bypass detection if a report has already been be sent for current version check.
// It can be sent to an alternative url via "baseURL" to send the report to. Empty string will send to default server.
//
// Example (in autoreport mode, without reporting twice the same data and using default server URL):
//   #include <stdbool.h>
//   #include <stdio.h>
//   #include <stdlib.h>
//   #include "libsysmetrics.h"
//
//   int main() {
//       ReportType r = ReportAuto;
//       char *err;
//
//       err = CollectAndSend(r, false, "");
//
//       if (err != NULL) {
//           printf("ERR: %s\n", err);
//       } else {
//           printf("Report sent to default server");
//       }
//       free(err);
//   }
//
// Building as a shared library
//
// The following command (in the pkg/sysmetrics/C directory) will provide a .h and .so file:
//   go build -o libsysmetrics.so.1 -buildmode=c-shared -ldflags '-extldflags -Wl,-soname,libsysmetrics.so.1' libsysmetrics.go
// You will probably want to rename libsysmetrics.so.h to libsysmetrics.h. Note that go generate will proceed
// this for you.
//
// Then, you can simply build your example program with:
//   gcc main.c ./libsysmetrics.so.1
//
package main
