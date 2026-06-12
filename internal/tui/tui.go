package tui

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	llmmodel "github.com/xzw/llmfetch/internal/model"
	"github.com/xzw/llmfetch/internal/render"
	"github.com/xzw/llmfetch/internal/system"
)

type state struct {
	info       system.Info
	models     []llmmodel.Model
	query      string
	searchMode bool
	fitIndex   int
	sortIndex  int
	selected   int
	offset     int
	showDetail bool
	width      int
	height     int
	tick       int
	opts       render.Options
}

var fitFilters = []string{"All", "Best", "Good", "Near"}
var sortModes = []string{"Score", "Out tok/s", "Memory", "Context", "Fit", "Trend"}
var pendingInput []byte

func Run(info system.Info, models []llmmodel.Model, opts render.Options) error {
	if err := cbreakMode(); err != nil {
		return err
	}
	defer restoreMode()

	fmt.Print("\033[?1049h\033[?25l")
	defer fmt.Print("\033[?25h\033[?1049l\033[0m")

	s := state{info: info, models: models, opts: opts}
	keys := make(chan keyResult)
	resize := make(chan os.Signal, 1)
	signal.Notify(resize, syscall.SIGWINCH)
	defer signal.Stop(resize)
	ticker := time.NewTicker(450 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		for {
			key, err := readKey()
			keys <- keyResult{key: key, err: err}
			if err != nil {
				return
			}
		}
	}()

	for {
		s.width, s.height = terminalSize()
		s = s.clamp()
		fmt.Print(s.view())
		select {
		case pressed := <-keys:
			if pressed.err != nil {
				return pressed.err
			}
			var quit bool
			s, quit = s.update(pressed.key)
			if quit {
				return nil
			}
		case <-resize:
			s.width, s.height = terminalSize()
		case <-ticker.C:
			s.tick++
		}
	}
}

type keyResult struct {
	key string
	err error
}

func (s state) update(key string) (state, bool) {
	if s.searchMode {
		switch key {
		case "esc", "enter":
			s.searchMode = false
		case "backspace":
			if len(s.query) > 0 {
				s.query = s.query[:len(s.query)-1]
			}
			s.selected = 0
			s.offset = 0
		case "ctrl+u":
			s.query = ""
			s.selected = 0
			s.offset = 0
		default:
			if len([]rune(key)) == 1 && key >= " " && key <= "~" {
				s.query += key
				s.selected = 0
				s.offset = 0
			}
		}
		return s.clamp(), false
	}

	switch key {
	case "q", "ctrl+c", "esc":
		return s, true
	case "down", "j":
		s.selected++
	case "up", "k":
		s.selected--
	case "pgdown":
		s.selected += s.visibleRows()
	case "pgup":
		s.selected -= s.visibleRows()
	case "g":
		s.selected = 0
	case "G":
		s.selected = len(s.filtered()) - 1
	case "/":
		s.searchMode = true
	case "s":
		s.sortIndex = (s.sortIndex + 1) % len(sortModes)
		s.selected = 0
		s.offset = 0
	case "f":
		s.fitIndex = (s.fitIndex + 1) % len(fitFilters)
		s.selected = 0
		s.offset = 0
	case "enter", "d":
		s.showDetail = !s.showDetail
	case "c":
		s.query = ""
		s.selected = 0
		s.offset = 0
	}
	return s.clamp(), false
}

func (s state) view() string {
	models := s.filtered()
	var b strings.Builder
	b.WriteString("\033[H\033[2J")
	b.WriteString(s.header())
	b.WriteString(s.filters(len(models)))
	b.WriteString(s.table(models))
	if s.showDetail && len(models) > 0 {
		b.WriteString(s.detail(models[s.selected]))
	}
	b.WriteString(s.footer(models))
	return b.String()
}

func (s state) header() string {
	if s.showFetchHeader() {
		return s.fetchHeader()
	}
	return s.compactHeader()
}

func (s state) showFetchHeader() bool {
	return s.width >= 128 && s.height >= 34
}

func (s state) compactHeader() string {
	line1 := "LLMFetch  AI workstation model fit browser"
	line2 := fmt.Sprintf(
		"CPU: %s  |  RAM: %s  |  GPU: %s  |  Profile: %s",
		s.info.Chip,
		s.info.Memory,
		s.info.GPU,
		s.info.Profile,
	)
	parts := make([]string, 0, len(s.info.Runtimes))
	for _, rt := range s.info.Runtimes {
		mark := color("x", red, true)
		if rt.Found {
			mark = color("ok", green, true)
		}
		parts = append(parts, color(rt.Name, cyan, false)+": "+mark)
	}
	return box(line1+"\n"+line2+"\n"+strings.Join(parts, "  |  "), s.width, gray)
}

func (s state) fetchHeader() string {
	logo := appleLogo
	leftW := 32
	inner := s.width - 2
	if inner < 126 {
		return s.compactHeader()
	}
	systemW := 58
	aiW := inner - leftW - systemW - 2
	if aiW < 34 {
		return s.compactHeader()
	}

	systemRows, aiRows := s.fetchColumns(systemW, aiW)
	height := len(logo)
	if len(systemRows) > height {
		height = len(systemRows)
	}
	if len(aiRows) > height {
		height = len(aiRows)
	}

	var b strings.Builder
	b.WriteString(color("┌"+strings.Repeat("─", leftW)+"┬"+strings.Repeat("─", systemW)+"┬"+strings.Repeat("─", aiW)+"┐", gray, false) + "\n")
	for i := 0; i < height; i++ {
		left := ""
		if i < len(logo) {
			left = colorLogoLine(i, logo[i])
		}
		systemCol := ""
		if i < len(systemRows) {
			systemCol = systemRows[i]
		}
		aiCol := ""
		if i < len(aiRows) {
			aiCol = aiRows[i]
		}
		b.WriteString(color("│", gray, false))
		b.WriteString(pad(clipLine(left, leftW), leftW))
		b.WriteString(color("│", gray, false))
		b.WriteString(pad(clipLine(systemCol, systemW), systemW))
		b.WriteString(color("│", gray, false))
		b.WriteString(pad(clipLine(aiCol, aiW), aiW))
		b.WriteString(color("│", gray, false))
		b.WriteString("\n")
	}
	b.WriteString(color("└"+strings.Repeat("─", leftW)+"┴"+strings.Repeat("─", systemW)+"┴"+strings.Repeat("─", aiW)+"┘", gray, false) + "\n")
	return b.String()
}

func (s state) fetchColumns(systemW, aiW int) ([]string, []string) {
	runtimes := runtimeSummary(s.info.Runtimes, 0, 3)
	engines := runtimeSummary(s.info.Runtimes, 3, len(s.info.Runtimes))
	systemRows := []string{
		fetchTitle("System", systemW),
		fetchRow("Host", empty(s.info.Host), systemW),
		fetchRow("OS", empty(s.info.OS), systemW),
		fetchRow("Kernel", empty(s.info.Kernel), systemW),
		fetchRow("Uptime", empty(s.info.Uptime), systemW),
		fetchRow("Packages", empty(s.info.Packages), systemW),
		fetchRow("Terminal", empty(terminalSummary(s.info.Terminal, s.info.Shell)), systemW),
		fetchRow("Display", empty(s.info.Display), systemW),
		fetchRow("CPU", empty(s.info.CPU), systemW),
		fetchRow("GPU", empty(s.info.GPU), systemW),
		fetchRow("Memory", empty(s.info.Memory), systemW),
		fetchRow("Disk (/)", empty(s.info.Storage), systemW),
		fetchRow(localIPLabel(s.info.LocalIPIface), empty(s.info.LocalIP), systemW),
		fetchRow(batteryLabel(s.info.BatteryName), empty(s.info.Battery), systemW),
		fetchRow("Power", empty(s.info.Power), systemW),
		fetchRow("Locale", empty(s.info.Locale), systemW),
	}
	aiRows := []string{
		fetchTitle("AI Stack", aiW),
		fetchRow("Runtimes", runtimes, aiW),
		fetchRow("Engines", engines, aiW),
		fetchRow("Best RT", bestRuntime(s.info), aiW),
		fetchRow("Accel", accelerator(s.info), aiW),
		fetchRow("Max Fit", maxFit(s.info.MemoryGB), aiW),
		fetchRow("Quant", recommendedQuant(s.info.MemoryGB), aiW),
		fetchRow("Registry", fmt.Sprintf("%d models", len(s.models)), aiW),
		fetchRow("AI Tier", empty(s.info.AITier), aiW),
		fetchRow("Profile", empty(s.info.Profile), aiW),
		fetchRow("AI Score", strconv.Itoa(s.info.AIScore), aiW),
		fetchRow("Capability", empty(s.info.Capability), aiW),
	}
	return systemRows, aiRows
}

func bestRuntime(info system.Info) string {
	for _, preferred := range []string{"MLX", "Ollama", "llama.cpp", "LM Studio", "vLLM"} {
		for _, rt := range info.Runtimes {
			if rt.Name == preferred && rt.Found {
				return preferred
			}
		}
	}
	if strings.Contains(info.Chip, "Apple") {
		return "MLX"
	}
	return "llama.cpp"
}

func accelerator(info system.Info) string {
	if strings.Contains(info.GPU, "Apple") || strings.Contains(info.Chip, "Apple") {
		return "Metal"
	}
	return "CPU"
}

func maxFit(memoryGB int) string {
	switch {
	case memoryGB >= 96:
		return "70B Q4 / 32B 8-bit"
	case memoryGB >= 64:
		return "70B Q4 / 32B FP16"
	case memoryGB >= 32:
		return "32B Q4 / 14B FP16"
	case memoryGB >= 16:
		return "14B Q4 / 7B FP16"
	default:
		return "7B Q4"
	}
}

func recommendedQuant(memoryGB int) string {
	switch {
	case memoryGB >= 64:
		return "Q5_K_M / 8-bit"
	case memoryGB >= 32:
		return "Q4_K_M / 8-bit"
	case memoryGB >= 16:
		return "Q4_K_M"
	default:
		return "Q3_K_M"
	}
}

func terminalSummary(terminal, shell string) string {
	terminal = strings.TrimSpace(terminal)
	shell = strings.TrimSpace(shell)
	if terminal == "" {
		return shell
	}
	if shell == "" {
		return terminal
	}
	return terminal + " / " + shell
}

func fetchTitle(title string, width int) string {
	return color(trim(title, width), green, true)
}

func fetchRow(label, value string, width int) string {
	labelW := 11
	if visibleWidth(label) > labelW {
		labelW = visibleWidth(label) + 2
	}
	valueW := max(0, width-labelW)
	if label != "Runtimes" && label != "Engines" {
		value = trim(value, valueW)
	}
	return color(pad(label, labelW), cyan, true) + fetchValue(label, value)
}

func fetchValue(label, value string) string {
	switch label {
	case "Runtimes", "Engines":
		return value
	case "AI Tier":
		return color(value, yellow, true)
	case "AI Score":
		return color(value, green, true)
	case "Profile":
		return color(value, magenta, true)
	case "GPU", "CPU":
		return color(value, yellow, true)
	default:
		if strings.HasPrefix(label, "Battery") {
			return colorBatteryPercentValue(value)
		}
		return colorPercentValue(value)
	}
}

func localIPLabel(iface string) string {
	if strings.TrimSpace(iface) == "" {
		return "Local IP"
	}
	return "Local IP (" + iface + ")"
}

func batteryLabel(name string) string {
	if strings.TrimSpace(name) == "" {
		return "Battery"
	}
	return "Battery (" + name + ")"
}

func colorPercentValue(value string) string {
	percent, start, end, ok := percentRange(value)
	if !ok {
		return color(value, white, false)
	}
	code := green
	if percent >= 85 {
		code = red
	} else if percent >= 70 {
		code = yellow
	}
	return color(value[:start], white, false) + color(value[start:end], code, true) + color(value[end:], white, false)
}

func colorBatteryPercentValue(value string) string {
	percent, start, end, ok := percentRange(value)
	if !ok {
		return color(value, white, false)
	}
	code := green
	if percent <= 20 {
		code = red
	} else if percent <= 50 {
		code = yellow
	}
	return color(value[:start], white, false) + color(value[start:end], code, true) + color(value[end:], white, false)
}

func percentRange(value string) (int, int, int, bool) {
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			continue
		}
		j := i
		for j < len(value) && value[j] >= '0' && value[j] <= '9' {
			j++
		}
		if j < len(value) && value[j] == '%' {
			n, _ := strconv.Atoi(value[i:j])
			start := i
			if i > 0 && value[i-1] == '(' {
				start = i - 1
			}
			end := j + 1
			if end < len(value) && value[end] == ')' {
				end++
			}
			return n, start, end, true
		}
		i = j
	}
	return 0, 0, 0, false
}

func runtimeSummary(runtimes []system.RuntimeStatus, start, end int) string {
	if end > len(runtimes) {
		end = len(runtimes)
	}
	if start >= end {
		return "-"
	}
	parts := make([]string, 0, end-start)
	for _, rt := range runtimes[start:end] {
		mark := color("x", red, true)
		if rt.Found {
			mark = color("ok", green, true)
		}
		parts = append(parts, color(rt.Name, cyan, false)+" "+mark)
	}
	return strings.Join(parts, "  ")
}

func (s state) filters(count int) string {
	search := s.query
	if search == "" {
		search = "Press / to search..."
	}
	specs := []struct {
		label  string
		value  string
		width  int
		active bool
	}{
		{"Search", search, 30, s.searchMode},
		{"Sort [s]", sortModes[s.sortIndex], 18, false},
		{"Fit [f]", fitFilters[s.fitIndex], 16, false},
		{"Use Case", "All", 18, false},
		{"Runtime", "All", 18, false},
		{"Rows", fmt.Sprintf("%d/%d", count, len(s.models)), 16, false},
	}
	var lines [3]string
	used := 0
	for _, spec := range specs {
		if used+spec.width > s.width-1 {
			break
		}
		w := spec.width
		parts := filterBoxLines(spec.label, spec.value, w, spec.active)
		for i := 0; i < 3 && i < len(parts); i++ {
			lines[i] += parts[i]
		}
		used += w
	}
	return color(clipLine(lines[0], s.width), gray, false) + "\n" +
		color(clipLine(lines[1], s.width), cyan, s.searchMode) + "\n" +
		color(clipLine(lines[2], s.width), gray, false) + "\n"
}

func (s state) table(models []llmmodel.Model) string {
	headers, widths := tableLayout(s.width)
	var b strings.Builder
	b.WriteString(color(clipLine(modelsTitle(len(models), len(s.models)), s.width), gray, false) + "\n")
	b.WriteString(color(clipLine(rowLine(headers, widths), s.tableLineWidth()), cyan, true) + color(scrollbar(0, s.offset, s.visibleRows(), len(models)), gray, false) + "\n")
	if len(models) == 0 {
		b.WriteString(color("No models match the current filters.\n", red, true))
		return b.String()
	}

	start := s.offset
	end := start + s.visibleRows()
	if end > len(models) {
		end = len(models)
	}
	for i := start; i < end; i++ {
		m := models[i]
		modelName := trim(m.Name, widths[1])
		if i == s.selected {
			modelName = marquee(m.Name, widths[1], s.tick)
		}
		values := []string{
			strconv.Itoa(m.Rank),
			modelName,
		}
		if s.width >= 150 {
			values = append(values,
				m.Provider,
				m.BestFor,
				m.Type,
				bitLabel(m),
				strconv.Itoa(m.Score),
				m.Runtime,
				strconv.Itoa(m.InTPS),
				strconv.Itoa(m.OutTPS),
				strconv.Itoa(tokensPerMinute(m.OutTPS)),
				fmt.Sprintf("%dGB", m.MemoryGB),
				memoryPercent(m.MemoryGB, s.info.MemoryGB),
				fitLabel(m.Fit),
				m.Context,
				m.License,
				trendText(m.Trend),
			)
		} else if s.width >= 118 {
			values = append(values,
				m.Provider,
				m.BestFor,
				strconv.Itoa(m.Score),
				m.Runtime,
				bitLabel(m),
				strconv.Itoa(m.OutTPS),
				fmt.Sprintf("%dGB", m.MemoryGB),
				fitLabel(m.Fit),
				m.Context,
				m.License,
				trendText(m.Trend),
			)
		} else {
			values = append(values,
				strconv.Itoa(m.Score),
				m.Runtime,
				strconv.Itoa(m.OutTPS),
				fmt.Sprintf("%dGB", m.MemoryGB),
				fitLabel(m.Fit),
				m.Context,
			)
		}
		scroll := color(scrollbar(i-start+1, s.offset, s.visibleRows(), len(models)), gray, false)
		if i == s.selected {
			line := clipLine(rowLine(values, widths), s.tableLineWidth())
			line = reverse(pad(line, s.tableLineWidth()))
			b.WriteString(line + scroll + "\n")
		} else {
			values = colorModelValues(values, m, s.width >= 150, s.width >= 118, s.info.MemoryGB)
			line := clipLine(rowLine(values, widths), s.tableLineWidth())
			b.WriteString(line + scroll + "\n")
		}
	}
	return b.String()
}

func colorModelValues(values []string, m llmmodel.Model, wide, medium bool, systemMemoryGB int) []string {
	out := append([]string(nil), values...)
	if len(out) < 2 {
		return out
	}
	out[0] = color(out[0], gray, false)
	out[1] = color(out[1], white, m.Rank <= 3)
	if wide {
		out[2] = color(out[2], gray, false)
		out[3] = color(out[3], bestForColor(m.BestFor), true)
		out[4] = color(out[4], blue, false)
		out[5] = color(out[5], bitColor(out[5]), true)
		out[6] = scoreColor(out[6])
		out[7] = runtimeColor(out[7])
		out[8] = color(out[8], white, false)
		out[9] = color(out[9], yellow, true)
		out[10] = color(out[10], blue, false)
		out[11] = color(out[11], white, false)
		out[12] = memoryPercentColor(out[12], m.MemoryGB, systemMemoryGB)
		out[13] = color(out[13], fitColor(m.Fit), true)
		out[14] = color(out[14], blue, false)
		out[15] = color(out[15], licenseColor(m.License), false)
		out[16] = trendColor(out[16], m.Trend)
		return out
	}
	if medium {
		out[2] = color(out[2], gray, false)
		out[3] = color(out[3], bestForColor(m.BestFor), true)
		out[4] = scoreColor(out[4])
		out[5] = runtimeColor(out[5])
		out[6] = color(out[6], bitColor(out[6]), true)
		out[7] = color(out[7], yellow, true)
		out[8] = color(out[8], white, false)
		out[9] = color(out[9], fitColor(m.Fit), true)
		out[10] = color(out[10], blue, false)
		out[11] = color(out[11], licenseColor(m.License), false)
		out[12] = trendColor(out[12], m.Trend)
		return out
	}
	out[2] = scoreColor(out[2])
	out[3] = runtimeColor(out[3])
	out[4] = color(out[4], yellow, true)
	out[5] = color(out[5], white, false)
	out[6] = color(out[6], fitColor(m.Fit), true)
	out[7] = color(out[7], blue, false)
	return out
}

func (s state) tableLineWidth() int {
	if s.width <= 1 {
		return s.width
	}
	return s.width - 1
}

func modelsTitle(count, total int) string {
	return "Models (" + strconv.Itoa(count) + "/" + strconv.Itoa(total) + ")"
}

func scrollbar(row, offset, visible, total int) string {
	if total <= visible || visible <= 0 {
		return " "
	}
	trackRows := visible + 1
	thumbSize := max(1, (visible*trackRows)/total)
	maxStart := trackRows - thumbSize
	start := 0
	if total > visible {
		start = (offset * maxStart) / (total - visible)
	}
	if row >= start && row < start+thumbSize {
		return "█"
	}
	return "│"
}

func tokensPerMinute(outTPS int) int {
	return outTPS * 60
}

func memoryPercent(modelGB, systemGB int) string {
	if systemGB <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d%%", (modelGB*100+systemGB/2)/systemGB)
}

func marquee(value string, width, tick int) string {
	if width <= 0 {
		return ""
	}
	if visibleWidth(value) <= width {
		return value
	}
	gap := "   "
	runes := []rune(value + gap)
	offset := tick % len(runes)
	stream := []rune(value + gap + value + gap)
	if offset+width > len(stream) {
		offset = 0
	}
	return string(stream[offset : offset+width])
}

func bitLabel(m llmmodel.Model) string {
	text := strings.ToLower(m.Name + " " + m.Runtime)
	switch {
	case strings.Contains(text, "fp16") || strings.Contains(text, "f16"):
		return "FP16"
	case strings.Contains(text, "fp8"):
		return "FP8"
	case strings.Contains(text, "q8") || strings.Contains(text, "8bit") || strings.Contains(text, "8-bit"):
		return "8-bit"
	case strings.Contains(text, "q6") || strings.Contains(text, "6bit") || strings.Contains(text, "6-bit"):
		return "6-bit"
	case strings.Contains(text, "q5") || strings.Contains(text, "5bit") || strings.Contains(text, "5-bit"):
		return "5-bit"
	case strings.Contains(text, "q4") || strings.Contains(text, "4bit") || strings.Contains(text, "4-bit"):
		return "4-bit"
	case strings.Contains(text, "q3") || strings.Contains(text, "3bit") || strings.Contains(text, "3-bit"):
		return "3-bit"
	default:
		return "Auto"
	}
}

func bitColor(value string) string {
	switch value {
	case "FP16", "8-bit", "FP8":
		return magenta
	case "4-bit", "5-bit", "6-bit":
		return yellow
	case "3-bit":
		return red
	default:
		return blue
	}
}

func scoreColor(value string) string {
	n, _ := strconv.Atoi(value)
	switch {
	case n >= 96:
		return color(value, green, true)
	case n >= 90:
		return color(value, yellow, true)
	default:
		return color(value, white, false)
	}
}

func runtimeColor(value string) string {
	if strings.Contains(value, "MLX") {
		return color(value, magenta, true)
	}
	if strings.Contains(value, "Ollama") {
		return color(value, blue, true)
	}
	return color(value, yellow, false)
}

func memoryPercentColor(value string, modelGB, systemGB int) string {
	if systemGB <= 0 {
		return color(value, gray, false)
	}
	pct := (modelGB*100 + systemGB/2) / systemGB
	code := green
	if pct >= 75 {
		code = yellow
	}
	if pct >= 90 {
		code = red
	}
	return color(value, code, true)
}

func trendColor(value string, trend int) string {
	if trend > 0 {
		return color(value, green, true)
	}
	if trend < 0 {
		return color(value, red, true)
	}
	return color(value, gray, false)
}

func bestForColor(value string) string {
	switch value {
	case "Coding":
		return blue
	case "Reasoning":
		return magenta
	case "Vision", "Vision OCR":
		return yellow
	case "RAG":
		return green
	default:
		return white
	}
}

func licenseColor(value string) string {
	if strings.Contains(value, "NC") {
		return yellow
	}
	if value == "unknown" || value == "-" {
		return gray
	}
	return blue
}

func (s state) detail(m llmmodel.Model) string {
	install := installHint(m)
	body := strings.Join([]string{
		"Detail: " + m.Name,
		fmt.Sprintf("Provider: %s   Best For: %s   Type: %s   Score: %d", m.Provider, m.BestFor, m.Type, m.Score),
		fmt.Sprintf("Runtime: %s   Input tok/s: %d   Output tok/s: %d   TPM: %d   Memory: %dGB", m.Runtime, m.InTPS, m.OutTPS, tokensPerMinute(m.OutTPS), m.MemoryGB),
		fmt.Sprintf("Memory fit: %s of system RAM   Context: %s", memoryPercent(m.MemoryGB, s.info.MemoryGB), m.Context),
		fmt.Sprintf("Fit: %s   License: %s   Trend: %s", fitLabel(m.Fit), m.License, trendText(m.Trend)),
		"Install hint: " + install,
	}, "\n")
	return box(body, s.width, gray)
}

func (s state) footer(models []llmmodel.Model) string {
	selected := ""
	if len(models) > 0 {
		m := models[s.selected]
		selected = fmt.Sprintf("▶ %s  %s  %s  Output tok/s %d  TPM %d  %dGB  Fit %s", m.Name, m.Provider, m.Runtime, m.OutTPS, tokensPerMinute(m.OutTPS), m.MemoryGB, m.Fit)
	}
	mode := "NORMAL"
	if s.searchMode {
		mode = "SEARCH /" + s.query
	}
	help := mode + "  /:search  c:clear  f:fit  s:sort  d/Enter:detail  ↑↓/j/k:move  q:quit"
	return "\n" + color(clipLine(selected, s.width), cyan, true) + "\n" + reverse(pad(clipLine(help, s.width), s.width)) + "\n"
}

func (s state) filtered() []llmmodel.Model {
	result := make([]llmmodel.Model, 0, len(s.models))
	query := strings.ToLower(strings.TrimSpace(s.query))
	fit := fitFilters[s.fitIndex]
	for _, m := range s.models {
		if fit != "All" && m.Fit != fit {
			continue
		}
		haystack := strings.ToLower(strings.Join([]string{m.Name, m.Provider, m.BestFor, m.Type, m.Runtime, m.License}, " "))
		if query != "" && !strings.Contains(haystack, query) {
			continue
		}
		result = append(result, m)
	}
	switch sortModes[s.sortIndex] {
	case "Score":
		sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })
	case "Out tok/s":
		sort.Slice(result, func(i, j int) bool { return result[i].OutTPS > result[j].OutTPS })
	case "Memory":
		sort.Slice(result, func(i, j int) bool { return result[i].MemoryGB < result[j].MemoryGB })
	case "Context":
		sort.Slice(result, func(i, j int) bool { return contextRank(result[i].Context) > contextRank(result[j].Context) })
	case "Fit":
		sort.Slice(result, func(i, j int) bool { return llmmodel.FitRank(result[i].Fit) > llmmodel.FitRank(result[j].Fit) })
	case "Trend":
		sort.Slice(result, func(i, j int) bool { return result[i].Trend > result[j].Trend })
	}
	return result
}

func (s state) clamp() state {
	models := s.filtered()
	if s.selected < 0 {
		s.selected = 0
	}
	if len(models) > 0 && s.selected >= len(models) {
		s.selected = len(models) - 1
	}
	rows := s.visibleRows()
	if s.selected < s.offset {
		s.offset = s.selected
	}
	if s.selected >= s.offset+rows {
		s.offset = s.selected - rows + 1
	}
	if s.offset < 0 {
		s.offset = 0
	}
	return s
}

func (s state) visibleRows() int {
	reserved := s.headerHeight() + 3 + 2 + 3
	if s.showDetail {
		reserved += 8
	}
	rows := s.height - reserved
	if rows < 4 {
		return 4
	}
	return rows
}

func (s state) headerHeight() int {
	if s.showFetchHeader() {
		return len(appleLogo) + 3
	}
	return 5
}

func readKey() (string, error) {
	if len(pendingInput) == 0 {
		buf := make([]byte, 32)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return "", err
		}
		if n == 0 {
			return "", nil
		}
		pendingInput = append(pendingInput, buf[:n]...)
	}
	b := pendingInput
	if b[0] == 3 {
		pendingInput = pendingInput[1:]
		return "ctrl+c", nil
	}
	if b[0] == 21 {
		pendingInput = pendingInput[1:]
		return "ctrl+u", nil
	}
	if b[0] == 13 || b[0] == 10 {
		pendingInput = pendingInput[1:]
		return "enter", nil
	}
	if b[0] == 127 || b[0] == 8 {
		pendingInput = pendingInput[1:]
		return "backspace", nil
	}
	if b[0] == 27 {
		if len(b) < 3 {
			pendingInput = pendingInput[1:]
			return "esc", nil
		}
		seq := string(b[:3])
		switch {
		case strings.Contains(seq, "[A"):
			pendingInput = pendingInput[3:]
			return "up", nil
		case strings.Contains(seq, "[B"):
			pendingInput = pendingInput[3:]
			return "down", nil
		case strings.Contains(seq, "[5"):
			pendingInput = trimCSISequence(pendingInput)
			return "pgup", nil
		case strings.Contains(seq, "[6"):
			pendingInput = trimCSISequence(pendingInput)
			return "pgdown", nil
		default:
			pendingInput = pendingInput[1:]
			return "esc", nil
		}
	}
	key := string(b[0])
	pendingInput = pendingInput[1:]
	return key, nil
}

func trimCSISequence(input []byte) []byte {
	for i, b := range input {
		if b >= '@' && b <= '~' {
			return input[i+1:]
		}
	}
	return nil
}

func cbreakMode() error {
	// Keep output post-processing enabled. Full raw mode disables ONLCR/OPOST,
	// which makes "\n" move down without returning to column zero in many
	// terminals and causes the classic staircase layout failure.
	cmd := exec.Command("stty", "-echo", "-icanon", "min", "1", "time", "0")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func restoreMode() {
	cmd := exec.Command("stty", "sane")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func terminalSize() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 140, 40
	}
	fields := strings.Fields(string(out))
	if len(fields) != 2 {
		return 140, 40
	}
	h, _ := strconv.Atoi(fields[0])
	w, _ := strconv.Atoi(fields[1])
	if w == 0 {
		w = 140
	}
	if h == 0 {
		h = 40
	}
	return w, h
}

func filterBoxLines(label, value string, width int, active bool) []string {
	top := "┌ " + label + " " + strings.Repeat("─", max(0, width-len([]rune(label))-4)) + "┐"
	mid := "│" + pad(" "+trim(value, width-3), width-2) + "│"
	bot := "└" + strings.Repeat("─", width-2) + "┘"
	return []string{
		top,
		mid,
		bot,
	}
}

func tableLayout(width int) ([]string, []int) {
	if width >= 150 {
		return []string{"Rank", "Model", "Provider", "Best", "Type", "Bit", "Score", "Runtime", "In/s", "Out tok/s", "TPM", "Mem", "Mem%", "Fit", "Ctx", "License", "Trend"},
			[]int{4, 27, 7, 7, 7, 5, 5, 9, 5, 9, 5, 5, 5, 7, 5, 7, 5}
	}
	if width >= 118 {
		return []string{"Rank", "Model", "Provider", "Best", "Score", "Runtime", "Bit", "Tok/s", "Memory", "Fit", "Context", "License", "Trend"},
			[]int{4, 22, 9, 8, 5, 10, 5, 5, 6, 7, 7, 8, 5}
	}
	return []string{"Rank", "Model", "Score", "Runtime", "Tok/s", "Mem", "Fit", "Ctx"},
		[]int{4, 19, 5, 10, 5, 5, 6, 5}
}

func box(value string, width int, borderColor string) string {
	if width < 40 {
		width = 40
	}
	inner := width - 2
	lines := strings.Split(value, "\n")
	var b strings.Builder
	b.WriteString(color("┌"+strings.Repeat("─", inner)+"┐", borderColor, false) + "\n")
	for _, line := range lines {
		line = clipLine(line, inner)
		b.WriteString(color("│", borderColor, false) + pad(line, inner) + color("│", borderColor, false) + "\n")
	}
	b.WriteString(color("└"+strings.Repeat("─", inner)+"┘", borderColor, false) + "\n")
	return b.String()
}

func rowLine(values []string, widths []int) string {
	cells := make([]string, 0, len(values))
	for i, value := range values {
		cells = append(cells, pad(trim(value, widths[i]), widths[i]))
	}
	return strings.Join(cells, "  ")
}

func fitLabel(fit string) string {
	switch fit {
	case "Best":
		return "Best"
	case "Good":
		return "Good"
	case "Near":
		return "Near"
	default:
		return fit
	}
}

func trendText(value int) string {
	switch {
	case value > 0:
		return fmt.Sprintf("↑%d", value)
	case value < 0:
		return fmt.Sprintf("↓%d", -value)
	default:
		return "→0"
	}
}

func installHint(m llmmodel.Model) string {
	if strings.Contains(m.Runtime, "MLX") {
		return "mlx-lm install " + m.Name
	}
	if m.Runtime == "Ollama" {
		return "ollama pull " + strings.ToLower(m.Name)
	}
	return "llama.cpp --model " + strings.ToLower(m.Name) + ".gguf"
}

func contextRank(value string) int {
	value = strings.TrimSuffix(value, "K")
	n, _ := strconv.Atoi(value)
	return n
}

func color(value string, code string, bold bool) string {
	if bold {
		code = "1;" + code
	}
	return "\033[" + code + "m" + value + "\033[0m"
}

func colorLogoLine(index int, value string) string {
	palette := []string{green, green, yellow, yellow, cyan, cyan, blue, blue, magenta, magenta, red, red}
	if value == "" {
		return value
	}
	return color(value, palette[index%len(palette)], true)
}

func reverse(value string) string {
	return "\033[7m" + value + "\033[0m"
}

func fitColor(fit string) string {
	switch fit {
	case "Best":
		return green
	case "Good":
		return cyan
	case "Near":
		return yellow
	default:
		return red
	}
}

func trim(value string, width int) string {
	if visibleWidth(value) <= width {
		return value
	}
	runes := []rune(value)
	for len(runes) > 0 && visibleWidth(string(runes)) > width-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func clipLine(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if visibleWidth(value) <= width {
		return value
	}
	runes := []rune(value)
	for len(runes) > 0 && visibleWidth(string(runes)) > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}

func pad(value string, width int) string {
	w := visibleWidth(value)
	if w >= width {
		return value
	}
	return value + strings.Repeat(" ", width-w)
}

func visibleWidth(value string) int {
	width := 0
	inANSI := false
	for _, r := range value {
		if r == '\033' {
			inANSI = true
			continue
		}
		if inANSI {
			if r == 'm' {
				inANSI = false
			}
			continue
		}
		width++
	}
	return width
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func empty(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

var appleLogo = []string{
	"                    'c.",
	"                 ,xNMM.",
	"               .OMMMMo",
	"               lMM\"",
	"     .;loddo:.  .olloddol;.",
	"   cKMMMMMMMMMMNWMMMMMMMMMM0:",
	" .KMMMMMMMMMMMMMMMMMMMMMMMWd.",
	" XMMMMMMMMMMMMMMMMMMMMMMMX.",
	";MMMMMMMMMMMMMMMMMMMMMMMM:",
	":MMMMMMMMMMMMMMMMMMMMMMMM:",
	".MMMMMMMMMMMMMMMMMMMMMMMX.",
	" kMMMMMMMMMMMMMMMMMMMMMMMMWd.",
	" 'XMMMMMMMMMMMMMMMMMMMMMMMMMMk",
	"  'XMMMMMMMMMMMMMMMMMMMMMMMMK.",
	"    kMMMMMMMMMMMMMMMMMMMMMMd",
	"     ;KMMMMMMMWXXWMMMMMMMk.",
	"       \"cooc*\"    \"*coo'\"",
}

const (
	gray    = "38;5;245"
	red     = "38;5;203"
	green   = "38;5;48"
	yellow  = "38;5;220"
	blue    = "38;5;75"
	magenta = "38;5;207"
	cyan    = "38;5;81"
	white   = "97"
)
