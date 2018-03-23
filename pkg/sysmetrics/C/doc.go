// C bindings to sysmetrics package
//
// Building as a shared library
//
// The following command (in the pkg/sysmetrics/C directory) will provide a .h and .so file:
//   go build -o libsysmetrics.so.1 -buildmode=c-shared -ldflags '-extldflags -Wl,-soname,libsysmetrics.so.1' libsysmetrics.go
//
// Then, you can simply build your example program with:
//   gcc main.c ./libsysmetrics.so.1
//
// Collect system info
//
// res will be the pretty printed version of collected data.
// err will be != NULL in case any error occurred.
//
//   char* Collect(char** res);
//
// Example:
//   #include <stdio.h>
//   #include <stdlib.h>
//   #include "libsysmetrics1.h"
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
// alwaysReports bypass previous report already be sent for current version check.
// It can be send to an alternative url via baseURL to send the report to, if not empty.
//
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
//
// Example (in autoreport mode, without reporting twice the same data and using default server URL):
//   #include <stdbool.h>
//   #include <stdio.h>
//   #include <stdlib.h>
//   #include "libsysmetrics1.h"
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
package main
