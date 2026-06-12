package system

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type RuntimeStatus struct {
	Name      string `json:"name"`
	Found     bool   `json:"found"`
	Detail    string `json:"detail,omitempty"`
	InstallBy string `json:"install_by,omitempty"`
}

type Info struct {
	Host         string          `json:"host"`
	OS           string          `json:"os"`
	Arch         string          `json:"arch"`
	Kernel       string          `json:"kernel"`
	Uptime       string          `json:"uptime"`
	Packages     string          `json:"packages"`
	Display      string          `json:"display"`
	CPU          string          `json:"cpu"`
	Chip         string          `json:"chip"`
	GPU          string          `json:"gpu"`
	Memory       string          `json:"memory"`
	MemoryGB     int             `json:"memory_gb"`
	Storage      string          `json:"storage"`
	Swap         string          `json:"swap"`
	LocalIP      string          `json:"local_ip"`
	LocalIPIface string          `json:"local_ip_iface"`
	Battery      string          `json:"battery"`
	BatteryName  string          `json:"battery_name"`
	Power        string          `json:"power"`
	Locale       string          `json:"locale"`
	Terminal     string          `json:"terminal"`
	Shell        string          `json:"shell"`
	Runtimes     []RuntimeStatus `json:"runtimes"`
	AITier       string          `json:"ai_tier"`
	Profile      string          `json:"profile"`
	Capability   string          `json:"capability"`
	AIScore      int             `json:"ai_score"`
}

func Detect() Info {
	host, _ := os.Hostname()
	info := Info{
		Host:     host,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Kernel:   strings.TrimSpace(run("uname", "-sr")),
		Uptime:   detectUptime(),
		Packages: detectPackages(),
		Display:  "Unknown",
		CPU:      "Unknown",
		Chip:     "Unknown",
		GPU:      "Unknown",
		Memory:   "Unknown",
		Storage:  detectStorage(),
		Swap:     detectSwap(),
		LocalIP:  "Unknown",
		Battery:  "Unknown",
		Power:    detectPowerAdapter(),
		Locale:   detectLocale(),
		Terminal: getenvAny("TERM_PROGRAM", "TERM"),
		Shell:    baseName(os.Getenv("SHELL")),
	}

	if runtime.GOOS == "darwin" {
		applyMacHardware(&info)
	} else {
		info.Chip = cpuName()
		info.CPU = info.Chip
		info.MemoryGB = detectLinuxMemoryGB()
		if info.MemoryGB > 0 {
			info.Memory = strconv.Itoa(info.MemoryGB) + "GB"
		}
	}
	info.LocalIP, info.LocalIPIface = detectLocalIP()
	info.Battery, info.BatteryName = detectBattery()

	info.Runtimes = []RuntimeStatus{
		detectRuntime("MLX", "mlx_lm", ""),
		detectRuntime("Ollama", "ollama", ""),
		detectRuntime("LM Studio", "lms", "/Applications/LM Studio.app"),
		detectRuntime("llama.cpp", "llama-cli", ""),
		detectRuntime("vLLM", "vllm", ""),
	}
	info.AITier, info.AIScore = tier(info.MemoryGB)
	info.Profile = profile(info)
	info.Capability = capability(info.MemoryGB)
	return info
}

func detectRuntime(name, command, appPath string) RuntimeStatus {
	_, err := exec.LookPath(command)
	found := err == nil
	if !found && appPath != "" {
		_, statErr := os.Stat(appPath)
		found = statErr == nil
	}
	return RuntimeStatus{Name: name, Found: found}
}

func applyMacHardware(info *Info) {
	out := run("system_profiler", "SPHardwareDataType")
	lines := parseColonMap(out)
	if v := lines["Model Name"]; v != "" {
		info.Host = v
	}
	if v := lines["Chip"]; v != "" {
		info.Chip = v
	}
	if cores := lines["Total Number of Cores"]; cores != "" {
		coreCount := strings.Fields(cores)[0]
		info.CPU = fmt.Sprintf("%s (%s cores)", info.Chip, coreCount)
	} else {
		info.CPU = info.Chip
	}
	if v := lines["Memory"]; v != "" {
		info.MemoryGB = parseLeadingInt(v)
	}
	if memory := detectMemory(); memory != "" {
		info.Memory = memory
	} else if v := lines["Memory"]; v != "" {
		info.Memory = v
	}
	if strings.Contains(info.Chip, "Apple") {
		info.GPU = "Apple GPU"
	}
	if gpu := detectMacGPU(); gpu != "" {
		info.GPU = gpu
	}
	info.Display = detectMacDisplay()
	product := strings.TrimSpace(run("sw_vers", "-productName"))
	version := strings.TrimSpace(run("sw_vers", "-productVersion"))
	build := strings.TrimSpace(run("sw_vers", "-buildVersion"))
	info.OS = strings.TrimSpace(product + " " + version + " (" + build + ") " + runtime.GOARCH)
}

func detectMacGPU() string {
	out := run("system_profiler", "SPDisplaysDataType")
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chipset Model:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Chipset Model:"))
		}
	}
	return ""
}

func detectMacDisplay() string {
	out := run("system_profiler", "SPDisplaysDataType")
	displays := []string{}
	inDisplays := false
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Displays:" {
			inDisplays = true
			continue
		}
		if !inDisplays {
			continue
		}
		if strings.HasPrefix(trimmed, "UI Looks like:") {
			resolution := normalizeResolution(strings.TrimSpace(strings.TrimPrefix(trimmed, "UI Looks like:")))
			if len(displays) > 0 {
				displays[len(displays)-1] = resolution
			} else {
				displays = append(displays, resolution)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "Resolution:") {
			resolution := strings.TrimSpace(strings.TrimPrefix(trimmed, "Resolution:"))
			displays = append(displays, normalizeResolution(resolution))
		}
	}
	if len(displays) == 0 {
		return "Unknown"
	}
	return strings.Join(displays, ", ")
}

func normalizeResolution(value string) string {
	value = strings.ReplaceAll(value, " x ", "x")
	value = strings.ReplaceAll(value, " @ ", "@")
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return value
	}
	resolution, _, _ := strings.Cut(fields[0], "@")
	return resolution
}

func detectUptime() string {
	if runtime.GOOS == "darwin" {
		raw := run("sysctl", "-n", "kern.boottime")
		if start := strings.Index(raw, "sec = "); start >= 0 {
			start += len("sec = ")
			end := strings.Index(raw[start:], ",")
			if end > 0 {
				sec, _ := strconv.ParseInt(raw[start:start+end], 10, 64)
				if sec > 0 {
					return formatDuration(time.Since(time.Unix(sec, 0)))
				}
			}
		}
	}
	return "Unknown"
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "Unknown"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	parts := []string{}
	if days > 0 {
		parts = append(parts, plural(days, "day"))
	}
	if hours > 0 {
		parts = append(parts, plural(hours, "hour"))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, plural(minutes, "min"))
	}
	return strings.Join(parts, ", ")
}

func plural(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

func detectPackages() string {
	brew, err := exec.LookPath("brew")
	if err != nil {
		return "Unknown"
	}
	formula := countCommandLines(brew, "list", "--formula")
	casks := countCommandLines(brew, "list", "--cask")
	if formula == 0 && casks == 0 {
		return "Unknown"
	}
	return fmt.Sprintf("%d (brew), %d (cask)", formula, casks)
}

func countCommandLines(name string, args ...string) int {
	out := run(name, args...)
	if out == "" {
		return 0
	}
	return len(strings.Split(strings.TrimSpace(out), "\n"))
}

func detectMemory() string {
	totalRaw := run("sysctl", "-n", "hw.memsize")
	totalBytes, err := strconv.ParseFloat(strings.TrimSpace(totalRaw), 64)
	if err != nil || totalBytes <= 0 {
		return ""
	}
	vm := parseVMStat(run("vm_stat"))
	pageSize := float64(vm["page_size"])
	if pageSize <= 0 {
		pageSize = 16384
	}
	availablePages := float64(vm["Pages free"] + vm["Pages speculative"] + vm["Pages inactive"])
	availableBytes := availablePages * pageSize
	usedBytes := totalBytes - availableBytes
	if usedBytes < 0 || usedBytes > totalBytes {
		return formatGiB(totalBytes)
	}
	percent := int((usedBytes/totalBytes)*100 + 0.5)
	return fmt.Sprintf("%s / %s (%d%%)", formatGiB(usedBytes), formatGiB(totalBytes), percent)
}

func parseVMStat(raw string) map[string]int64 {
	result := map[string]int64{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Mach Virtual Memory Statistics:") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "page" && i+3 < len(fields) {
					size, _ := strconv.ParseInt(fields[i+3], 10, 64)
					result["page_size"] = size
					break
				}
			}
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimSuffix(strings.TrimSpace(value), ".")
		n, _ := strconv.ParseInt(value, 10, 64)
		result[strings.TrimSpace(key)] = n
	}
	return result
}

func detectSwap() string {
	raw := run("sysctl", "-n", "vm.swapusage")
	if raw == "" {
		return "Unknown"
	}
	fields := strings.Fields(strings.ReplaceAll(raw, "=", " "))
	total := ""
	used := ""
	for i, field := range fields {
		switch strings.TrimSuffix(field, ":") {
		case "total":
			if i+1 < len(fields) {
				total = fields[i+1]
			}
		case "used":
			if i+1 < len(fields) {
				used = fields[i+1]
			}
		}
	}
	if used == "" {
		return "Unknown"
	}
	if strings.HasPrefix(used, "0.00") || strings.HasPrefix(used, "0B") {
		return "Unused"
	}
	if total != "" {
		return used + " / " + total
	}
	return used
}

func detectLocalIP() (string, string) {
	for _, iface := range []string{"en0", "en1"} {
		ip := run("ipconfig", "getifaddr", iface)
		if ip != "" {
			cidr := ip + detectCIDRSuffix(iface)
			return cidr, iface
		}
	}
	return "Unknown", ""
}

func detectCIDRSuffix(iface string) string {
	out := run("ifconfig", iface)
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		for i, field := range fields {
			if field == "netmask" && i+1 < len(fields) {
				mask := strings.TrimPrefix(fields[i+1], "0x")
				if bits := hexNetmaskBits(mask); bits > 0 {
					return fmt.Sprintf("/%d", bits)
				}
			}
		}
	}
	return ""
}

func hexNetmaskBits(mask string) int {
	value, err := strconv.ParseUint(mask, 16, 32)
	if err != nil {
		return 0
	}
	bits := 0
	for value > 0 {
		bits += int(value & 1)
		value >>= 1
	}
	return bits
}

func detectBattery() (string, string) {
	raw := run("pmset", "-g", "batt")
	if raw == "" {
		return "Unknown", ""
	}
	source := "Unknown"
	if strings.Contains(raw, "'AC Power'") {
		source = "AC Connected"
	} else if strings.Contains(raw, "'Battery Power'") {
		source = "Battery"
	}
	name := detectBatteryName()
	for _, line := range strings.Split(raw, "\n") {
		if !strings.Contains(line, "%") {
			continue
		}
		fields := strings.Fields(line)
		for _, field := range fields {
			if strings.Contains(field, "%") {
				percent := strings.TrimRight(strings.TrimSpace(field), ";")
				return percent + " [" + source + "]", name
			}
		}
	}
	return source, name
}

func detectBatteryName() string {
	raw := run("ioreg", "-r", "-c", "AppleSmartBattery")
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "\"DeviceName\"") {
			continue
		}
		_, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), "\"")
	}
	return ""
}

func detectPowerAdapter() string {
	raw := run("system_profiler", "SPPowerDataType")
	lines := parseColonMap(raw)
	if wattage := lines["Wattage (W)"]; wattage != "" {
		return wattage + "W"
	}
	if connected := lines["Connected"]; connected != "" {
		return "Connected: " + connected
	}
	return "Unknown"
}

func detectLocale() string {
	if runtime.GOOS == "darwin" {
		if value := run("defaults", "read", "-g", "AppleLocale"); value != "" {
			if strings.Contains(value, ".") {
				return value
			}
			return value + ".UTF-8"
		}
	}
	if value := os.Getenv("LANG"); value != "" {
		return value
	}
	if value := os.Getenv("LC_ALL"); value != "" {
		return value
	}
	return "Unknown"
}

func parseColonMap(raw string) map[string]string {
	result := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		result[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return result
}

func tier(memoryGB int) (string, int) {
	switch {
	case memoryGB >= 128:
		return "S+", 98
	case memoryGB >= 96:
		return "S", 94
	case memoryGB >= 64:
		return "A+", 90
	case memoryGB >= 32:
		return "A", 82
	default:
		return "B", 72
	}
}

func profile(info Info) string {
	if runtime.GOOS == "darwin" && strings.Contains(info.Chip, "Apple") {
		return "MLX Powerhouse"
	}
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return "CUDA Beast"
	}
	return "Local AI Builder"
}

func capability(memoryGB int) string {
	switch {
	case memoryGB >= 64:
		return "Coding A+  Reasoning A+  Vision A"
	case memoryGB >= 32:
		return "Coding A  Reasoning A-  Vision B+"
	default:
		return "Coding B  Reasoning B-  Vision B"
	}
}

func cpuName() string {
	raw := run("uname", "-p")
	if raw == "" {
		return runtime.GOARCH
	}
	return raw
}

func detectLinuxMemoryGB() int {
	raw, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "MemTotal:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		kb, _ := strconv.Atoi(fields[1])
		return kb / 1024 / 1024
	}
	return 0
}

func detectStorage() string {
	path := "/"
	if runtime.GOOS == "darwin" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			path = home
		}
	}
	out := run("df", "-k", path)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return "Unknown"
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return "Unknown"
	}
	totalKB, _ := strconv.ParseFloat(fields[1], 64)
	availableKB, _ := strconv.ParseFloat(fields[3], 64)
	usedKB := totalKB - availableKB
	percent := int((usedKB/totalKB)*100 + 0.5)
	return fmt.Sprintf("%s / %s (%d%%)", formatGiB(usedKB*1024), formatGiB(totalKB*1024), percent)
}

func formatGiB(bytes float64) string {
	return fmt.Sprintf("%.2f GiB", bytes/1024/1024/1024)
}

func run(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(stdout.String())
}

func parseLeadingInt(value string) int {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(fields[0])
	return n
}

func getenvAny(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return "unknown"
}

func baseName(path string) string {
	if path == "" {
		return "unknown"
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
