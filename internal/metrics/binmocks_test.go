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
			os.Exit(1)
		}
		regularOutput := `00:00.0 0600: 8086:0104 (rev 09)
00:02.0 0300: 8086:0126 (rev 09)
00:16.0 0780: 8086:1c3a (rev 04)
00:16.3 0700: 8086:1c3d (rev 04)
00:19.0 0200: 8086:1502 (rev 04)`
		switch args[1] {
		case "one gpu":
			fmt.Println(regularOutput)
		case "multiple gpus":
			fmt.Println(regularOutput)
			fmt.Println("00:02.0 0300: 8086:0127 (rev 09)")
		case "no revision number":
			fmt.Println("00:02.0 0300: 8086:0126")
		case "no gpu":
			fmt.Println(`00:00.0 0600: 8086:0104 (rev 09)
00:16.0 0780: 8086:1c3a (rev 04)
00:16.3 0700: 8086:1c3d (rev 04)
00:19.0 0200: 8086:1502 (rev 04)`)
		case "hexa numbers":
			fmt.Println("00:02.0 0300: 8b86:a126 (rev 09)")
		case "empty":
		case "malformed gpu line":
			fmt.Println("00:02.0 0300: 80860127 (rev 09)")
		case "garbage":
			fmt.Println(garbageOutput)
		case "fail":
			fmt.Println(regularOutput) // still print content
			os.Exit(1)
		}

	case "lscpu":
		if args[0] != "-J" {
			fmt.Fprintf(os.Stderr, "Unexpected lscpu arguments: %v\n", args)
			os.Exit(1)
		}
		regularOutput := `{
   "lscpu": [
      {"field": "Architecture:", "data": "x86_64"},
      {"field": "CPU op-mode(s):", "data": "32-bit, 64-bit"},
      {"field": "Byte Order:", "data": "Little Endian"},
      {"field": "CPU(s):", "data": "8"},
      {"field": "On-line CPU(s) list:", "data": "0-7"},
      {"field": "Thread(s) per core:", "data": "2"},
      {"field": "Core(s) per socket:", "data": "4"},
      {"field": "Socket(s):", "data": "1"},
      {"field": "NUMA node(s):", "data": "1"},
      {"field": "Vendor ID:", "data": "Genuine"},
      {"field": "CPU family:", "data": "6"},
      {"field": "Model:", "data": "158"},
      {"field": "Model name:", "data": "Intuis Corus i5-8300H CPU @ 2.30GHz"},
      {"field": "Stepping:", "data": "10"},
      {"field": "CPU MHz:", "data": "3419.835"},
      {"field": "CPU max MHz:", "data": "4000.0000"},
      {"field": "CPU min MHz:", "data": "800.0000"},
      {"field": "BogoMIPS:", "data": "4608.00"},
      {"field": "Virtualization:", "data": "VT-x"},
      {"field": "L1d cache:", "data": "32K"},
      {"field": "L1i cache:", "data": "32K"},
      {"field": "L2 cache:", "data": "256K"},
      {"field": "L3 cache:", "data": "8192K"},
      {"field": "NUMA node0 CPU(s):", "data": "0-7"},
      {"field": "Flags:", "data": "fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc art arch_perfmon pebs bts rep_good nopl xtopology nonstop_tsc cpuid aperfmperf tsc_known_freq pni pclmulqdq dtes64 monitor ds_cpl vmx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand lahf_lm abm 3dnowprefetch cpuid_fault epb invpcid_single pti ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid mpx rdseed adx smap clflushopt intel_pt xsaveopt xsavec xgetbv1 xsaves dtherm ida arat pln pts hwp hwp_notify hwp_act_window hwp_epp"}
   ]
}`
		switch args[1] {
		case "regular":
			fmt.Println(regularOutput)
		case "missing one expected field":
			fmt.Println(strings.Replace(regularOutput, "Vendor ID:", "", -1))
		case "missing one optional field":
			fmt.Println(strings.Replace(regularOutput, "Virtualization vendor:", "", -1))
		case "virtualized":
			fmt.Println(regularOutput)
			fmt.Println(`{"field": "Hypervisor vendor:", "data": "KVM"},
				{"field": "Virtualization type:", "data": "full"},`)
		case "without space":
			fmt.Println(`{
   "lscpu": [
      {"field":"Architecture:","data":"x86_64"},
      {"field":"CPU op-mode(s):","data":"32-bit, 64-bit"},
      {"field":"Byte Order:","data":"Little Endian"},
      {"field":"CPU(s):","data":"8"},
      {"field":"On-line CPU(s) list:","data":"0-7"},
      {"field":"Thread(s) per core:","data":"2"},
      {"field":"Core(s) per socket:","data":"4"},
      {"field":"Socket(s):","data":"1"},
      {"field":"NUMA node(s):","data":"1"},
      {"field":"Vendor ID:","data":"Genuine"},
      {"field":"CPU family:","data":"6"},
      {"field":"Model:","data":"158"},
      {"field":"Model name:","data":"Intuis Corus i5-8300H CPU @ 2.30GHz"},
      {"field":"Stepping:","data":"10"},
      {"field":"CPU MHz:","data":"3419.835"},
      {"field":"CPU max MHz:","data":"4000.0000"},
      {"field":"CPU min MHz:","data":"800.0000"},
      {"field":"BogoMIPS:","data":"4608.00"},
      {"field":"Virtualization:","data":"VT-x"},
      {"field":"L1d cache:","data":"32K"},
      {"field":"L1i cache:","data":"32K"},
      {"field":"L2 cache:","data":"256K"},
      {"field":"L3 cache:","data":"8192K"},
      {"field":"NUMA node0 CPU(s):","data":"0-7"},
      {"field":"Flags:","data":"fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc art arch_perfmon pebs bts rep_good nopl xtopology nonstop_tsc cpuid aperfmperf tsc_known_freq pni pclmulqdq dtes64 monitor ds_cpl vmx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand lahf_lm abm 3dnowprefetch cpuid_fault epb invpcid_single pti ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid mpx rdseed adx smap clflushopt intel_pt xsaveopt xsavec xgetbv1 xsaves dtherm ida arat pln pts hwp hwp_notify hwp_act_window hwp_epp"}
   ]
}`)
		case "empty":
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
		case "chosen resolution not preferred":
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis) 510mm x 287mm
   1920x1080     60.00* 
   1600x1200     60.00 +
   1680x1050     59.95  `)
		case "multiple frequencies for resolution":
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis) 510mm x 287mm
   1920x1080     60.00*+  59.94    50.00    60.05    60.00    50.04`)
		case "multiple frequencies select other resolution":
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis) 510mm x 287mm
   1920x1080     60.00 +  59.94    50.00*   60.05    60.00    50.04`)
		case "multiple frequencies select other resolution on non preferred":
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis) 510mm x 287mm
   1920x1080     60.00    59.94    50.00*   60.05    60.00    50.04
   1600x1200     60.00 +`)
		case "no specified screen size":
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis)
   1920x1080     60.00*+
   1600x1200     60.00  
   1680x1050     59.95  `)
		case "no chosen resolution":
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis) 510mm x 287mm
   1920x1080     60.00  
   1600x1200     60.00  
   1680x1050     59.95  `)
		case "empty":
		case "malformed screen line":
			fmt.Println(`VGA-1 connected 1920x1080+0+0 (normal left inverted right x axis y axis) 510m x 287mm
   1920x108160.00*+`)
		case "garbage":
			fmt.Println(garbageOutput)
		case "fail":
			fmt.Println(regularOutput) // still print content
			os.Exit(1)
		}

	case "df":
		regularOutput := `Sys. de fichiers blocs de 1K   Utilisé Disponible Uti% Monté sur
udev                 3992524         0    3992524   0% /dev
tmpfs                 804812      2104     802708   1% /run
/dev/sda5          159431364 142492784    8816880  95% /
tmpfs                4024048    152728    3871320   4% /dev/shm
tmpfs                   5120         4       5116   1% /run/lock`
		switch args[0] {
		case "one partition":
			fmt.Println(regularOutput)
		case "multiple partitions":
			fmt.Println(regularOutput)
			fmt.Println(`/dev/sdc2          309681364 102492784    2816880   5% /something`)
		case "no partitions":
			fmt.Println("")
		case "filters loop devices":
			fmt.Println(regularOutput)
			fmt.Println(`/dev/loop0            132480    132480          0 100% /snap/gnome-3-26-1604/27
/dev/loop2             83584     83584          0 100% /snap/core/4110`)
		case "empty":
		case "malformed partition line string":
			fmt.Println(`/dev/sda5          a159431364 142492784    8816880  95% /`)
		case "malformed partition line one field":
			fmt.Println(`/dev/sda5`)
		case "garbage":
			fmt.Println(garbageOutput)
		case "fail":
			fmt.Println(regularOutput) // still print content
			os.Exit(1)
		}

	case "dpkg":
		if args[0] != "--print-architecture" && args[0] != "--status" {
			fmt.Fprintf(os.Stderr, "Unexpected dpkg arguments: %v\n", args)
			os.Exit(1)
		}
		switch args[0] {
		case "--print-architecture":
			switch args[1] {
			case "regular":
				fmt.Println("amd64")
			case "empty":
			case "fail":
				fmt.Println("amd64") // still print content
				os.Exit(1)
			}
		case "--status":
			if args[1] != "libc6" {
				fmt.Fprintf(os.Stderr, "Unexpected dpkg --status arguments: %v\n", args)
				os.Exit(1)
			}
			regularOutput := `Package: libc6
Status: install ok installed
Priority: optional
Section: libs
Installed-Size: 13111
Maintainer: Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>
Architecture: amd64
Multi-Arch: same
Source: glibc
Version: 2.33-0ubuntu5
Replaces: libc6-amd64
Depends: libgcc-s1, libcrypt1 (>= 1:4.4.10-10ubuntu4)
Recommends: libidn2-0 (>= 2.0.5~), libnss-nis, libnss-nisplus
Suggests: glibc-doc, debconf | debconf-2.0, locales
Breaks: busybox (<< 1.30.1-6), fakeroot (<< 1.25.3-1.1ubuntu2~), hurd (<< 1:0.9.git20170910-1), ioquake3 (<< 1.36+u20200211.f2c61c1~dfsg-2~), iraf-fitsutil (<< 2018.07.06-4), libgegl-0.4-0 (<< 0.4.18), libtirpc1 (<< 0.2.3), locales (<< 2.33), locales-all (<< 2.33), macs (<< 2.2.7.1-3~), nocache (<< 1.1-1~), nscd (<< 2.33), openarena (<< 0.8.8+dfsg-4~), openssh-server (<< 1:8.2p1-4), r-cran-later (<< 0.7.5+dfsg-2), wcc (<< 0.0.2+dfsg-3)
Conffiles:
 /etc/ld.so.conf.d/x86_64-linux-gnu.conf d4e7a7b88a71b5ffd9e2644e71a0cfab
Description: GNU C Library: Shared libraries
 Contains the standard libraries that are used by nearly all programs on
 the system. This package includes shared versions of the standard C library
 and the standard math library, as well as many others.
Homepage: https://www.gnu.org/software/libc/libc.html
Original-Maintainer: GNU Libc Maintainers <debian-glibc@lists.debian.org>
Original-Vcs-Browser: https://salsa.debian.org/glibc-team/glibc
Original-Vcs-Git: https://salsa.debian.org/glibc-team/glibc.git`

			switch args[2] {
			case "regular", "no hwcap":
				fmt.Println(regularOutput)

			case "old version":
				fmt.Println(`Package: libc6
Status: install ok installed
Priority: optional
Section: libs
Installed-Size: 13111
Maintainer: Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>
Architecture: amd64
Multi-Arch: same
Source: glibc
Version: 2.32
Replaces: libc6-amd64
Depends: libgcc-s1, libcrypt1 (>= 1:4.4.10-10ubuntu4)
Recommends: libidn2-0 (>= 2.0.5~), libnss-nis, libnss-nisplus
Suggests: glibc-doc, debconf | debconf-2.0, locales
Breaks: busybox (<< 1.30.1-6), fakeroot (<< 1.25.3-1.1ubuntu2~), hurd (<< 1:0.9.git20170910-1), ioquake3 (<< 1.36+u20200211.f2c61c1~dfsg-2~), iraf-fitsutil (<< 2018.07.06-4), libgegl-0.4-0 (<< 0.4.18), libtirpc1 (<< 0.2.3), locales (<< 2.33), locales-all (<< 2.33), macs (<< 2.2.7.1-3~), nocache (<< 1.1-1~), nscd (<< 2.33), openarena (<< 0.8.8+dfsg-4~), openssh-server (<< 1:8.2p1-4), r-cran-later (<< 0.7.5+dfsg-2), wcc (<< 0.0.2+dfsg-3)
Conffiles:
 /etc/ld.so.conf.d/x86_64-linux-gnu.conf d4e7a7b88a71b5ffd9e2644e71a0cfab
Description: GNU C Library: Shared libraries
 Contains the standard libraries that are used by nearly all programs on
 the system. This package includes shared versions of the standard C library
 and the standard math library, as well as many others.
Homepage: https://www.gnu.org/software/libc/libc.html
Original-Maintainer: GNU Libc Maintainers <debian-glibc@lists.debian.org>
Original-Vcs-Browser: https://salsa.debian.org/glibc-team/glibc
Original-Vcs-Git: https://salsa.debian.org/glibc-team/glibc.git`)

			case "not installed":
				fmt.Println(`dpkg-query: package 'libc6' is not installed and no information is available
Use dpkg --info (= dpkg-deb --info) to examine archive files.`)
			case "empty:":
			case "fail":
				// still print content
				fmt.Println(regularOutput)
				os.Exit(1)
			}

		}

	case "/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2":
		if args[0] != "--help" {
			fmt.Fprintf(os.Stderr, "Unexpected ld arguments: %v\n", args)
			os.Exit(1)
		}
		regularOutput := `Usage: /lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 [OPTION]... EXECUTABLE-FILE [ARGS-FOR-PROGRAM...]
You have invoked 'ld.so', the program interpreter for dynamically-linked
ELF programs.  Usually, the program interpreter is invoked automatically
when a dynamically-linked executable is started.

You may invoke the program interpreter program directly from the command
line to load and run an ELF executable file; this is like executing that
file itself, but always uses the program interpreter you invoked,
instead of the program interpreter specified in the executable file you
run.  Invoking the program interpreter directly provides access to
additional diagnostics, and changing the dynamic linker behavior without
setting environment variables (which would be inherited by subprocesses).

  --list                list all dependencies and how they are resolved
  --verify              verify that given object really is a dynamically linked
                        object we can handle
  --inhibit-cache       Do not use /etc/ld.so.cache
  --library-path PATH   use given PATH instead of content of the environment
                        variable LD_LIBRARY_PATH
  --glibc-hwcaps-prepend LIST
                        search glibc-hwcaps subdirectories in LIST
  --glibc-hwcaps-mask LIST
                        only search built-in subdirectories if in LIST
  --inhibit-rpath LIST  ignore RUNPATH and RPATH information in object names
                        in LIST
  --audit LIST          use objects named in LIST as auditors
  --preload LIST        preload objects named in LIST
  --argv0 STRING        set argv[0] to STRING before running
  --list-tunables       list all tunables with minimum and maximum values
  --list-diagnostics    list diagnostics information
  --help                display this help and exit
  --version             output version information and exit

This program interpreter self-identifies as: /lib64/ld-linux-x86-64.so.2

Shared library search path:
  (libraries located via /etc/ld.so.cache)
  /lib/x86_64-linux-gnu (system search path)
  /usr/lib/x86_64-linux-gnu (system search path)
  /lib (system search path)
  /usr/lib (system search path)

Subdirectories of glibc-hwcaps directories, in priority order:
  x86-64-v4
  x86-64-v3 (supported, searched)
  x86-64-v2 (supported, searched)

Legacy HWCAP subdirectories under library search path directories:
  x86_64 (AT_PLATFORM; supported, searched)
  tls (supported, searched)
  avx512_1
  x86_64 (supported, searched)`

		switch args[1] {
		case "regular":
			fmt.Println(regularOutput)
		case "no hwcap":
			fmt.Println(`Usage: /lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 [OPTION]... EXECUTABLE-FILE [ARGS-FOR-PROGRAM...]
You have invoked 'ld.so', the program interpreter for dynamically-linked
ELF programs.  Usually, the program interpreter is invoked automatically
when a dynamically-linked executable is started.

You may invoke the program interpreter program directly from the command
line to load and run an ELF executable file; this is like executing that
file itself, but always uses the program interpreter you invoked,
instead of the program interpreter specified in the executable file you
run.  Invoking the program interpreter directly provides access to
additional diagnostics, and changing the dynamic linker behavior without
setting environment variables (which would be inherited by subprocesses).

  --list                list all dependencies and how they are resolved
  --verify              verify that given object really is a dynamically linked
                        object we can handle
  --inhibit-cache       Do not use /etc/ld.so.cache
  --library-path PATH   use given PATH instead of content of the environment
                        variable LD_LIBRARY_PATH
  --glibc-hwcaps-prepend LIST
                        search glibc-hwcaps subdirectories in LIST
  --glibc-hwcaps-mask LIST
                        only search built-in subdirectories if in LIST
  --inhibit-rpath LIST  ignore RUNPATH and RPATH information in object names
                        in LIST
  --audit LIST          use objects named in LIST as auditors
  --preload LIST        preload objects named in LIST
  --argv0 STRING        set argv[0] to STRING before running
  --list-tunables       list all tunables with minimum and maximum values
  --list-diagnostics    list diagnostics information
  --help                display this help and exit
  --version             output version information and exit

This program interpreter self-identifies as: /lib64/ld-linux-x86-64.so.2

Shared library search path:
  (libraries located via /etc/ld.so.cache)
  /lib/x86_64-linux-gnu (system search path)
  /usr/lib/x86_64-linux-gnu (system search path)
  /lib (system search path)
  /usr/lib (system search path)

Subdirectories of glibc-hwcaps directories, in priority order:
  x86-64-v4
  x86-64-v3
  x86-64-v2

Legacy HWCAP subdirectories under library search path directories:
  x86_64 (AT_PLATFORM; supported, searched)
  tls (supported, searched)
  avx512_1
  x86_64 (supported, searched)`)

		case "empty":
		case "fail":
			fmt.Println(regularOutput)
			os.Exit(1)
		}
	}
}
