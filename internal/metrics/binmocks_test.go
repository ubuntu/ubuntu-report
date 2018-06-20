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
		case "chosen resolution not prefered":
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
		if args[0] != "--print-architecture" {
			fmt.Fprintf(os.Stderr, "Unexpected dpkg arguments: %v\n", args)
			os.Exit(1)
		}
		switch args[1] {
		case "regular":
			fmt.Println("amd64")
		case "empty":
		case "fail":
			fmt.Println("amd64") // still print content
			os.Exit(1)
		}
	}

}
