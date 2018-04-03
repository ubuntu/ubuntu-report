// Package sysmetrics C bindings: collect and report system and hardware metrics
// from your system.
//
// Collect system info
//
// Command signature:
//   char* sysmetrics_collect(char** res);
// "res" will be the pretty printed version of collected data. The return "err" will be != NULL in case
// any error occurred during data collection.
//
// Example:
//   #include <stdio.h>
//   #include <stdlib.h>
//   #include <libsysmetrics.h>
//
//   int main() {
//       char *res, *err;
//
//       err = sysmetrics_collect(&res);
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
// Send provided metrics data to server
//
// Command signature:
//   char* sysmetrics_send_report(char* data, bool alwaysReport, char* baseUrl);
//
// The report will not be sent if a report has already been sent for this version unless "alwaysReport" is true.
// If "baseURL" is not an empty string, this overrides the server the report is sent to.
// The return "err" will be != NULL in case any error occurred during POST.
//
// Example (sending provided metrics data):
//   #include <stdbool.h>
//   #include <stdio.h>
//   #include <stdlib.h>
//   #include <libsysmetrics.h>
//
//   int main() {
//       char *err;
//
//       err = sysmetrics_send_report("{ \"Version\": \"18.04\" }", false, "");
//
//       if (err != NULL) {
//           printf("ERR: %s\n", err);
//       } else {
//           printf("Report sent to default server");
//       }
//       free(err);
//   }
//
// Collect and send system info to server
//
// Command signature:
//   char* sysmetrics_collect_and_send(sysmetrics_report_type r, bool alwaysReport, char* url);
//
// sysmetrics_report_type is the following enum:
//    typedef enum {
//      // sysmetrics_report_interactive will show report content on stdout and read anwser on stdin
//      sysmetrics_report_interactive = 0,
//      // sysmetrics_report_auto will send a report without printing report
//      sysmetrics_report_auto = 1,
//      // sysmetrics_report_optout will send opt-out message without printing report
//      sysmetrics_report_optout = 2,
//    } sysmetrics_report_type;
// You should generally prefer in bindings the auto or optout report. Interactive is based on stdout and stdin.
// The report will not be sent if a report has already been sent for this version unless "alwaysReport" is true.
// It can be sent to an alternative url via "baseURL" to send the report to. Empty string will send to default server.
//
// Example (in autoreport mode, without reporting twice the same data and using default server URL):
//   #include <stdbool.h>
//   #include <stdio.h>
//   #include <stdlib.h>
//   #include <libsysmetrics.h>
//
//   int main() {
//       sysmetrics_report_type r = sysmetrics_report_auto;
//       char *err;
//
//       err = sysmetrics_collect_and_send(r, false, "");
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
