package render

import (
	"fmt"
	"github.com/xzw/llmfetch/internal/model"
	"github.com/xzw/llmfetch/internal/system"
	"strconv"
	"strings"
)

type Options struct {
	Color   bool
	Emoji   bool
	Unicode bool
}

type Renderer struct {
	opts Options
}

func New(opts Options) Renderer {
	return Renderer{opts: opts}
}

func (r Renderer) Dashboard(info system.Info, models []model.Model) string {
	var b strings.Builder
	b.WriteString(r.style("LLMFetch", cyan, true))
	b.WriteString(r.style("  AI workstation dashboard", gray, false))
	b.WriteString("\n\n")
	b.WriteString(r.fetchPanel(info, len(models)))
	b.WriteString("\n\n")
	b.WriteString(r.style("Best Fits For This Machine", cyan, true))
	b.WriteString("\n")
	b.WriteString(r.modelTable(firstN(models, 10), info.MemoryGB))
	return b.String()
}

func (r Renderer) fetchPanel(info system.Info, modelCount int) string {
	logo := appleLogo
	leftW := 32
	systemW := 58
	aiW := 44
	systemRows := [][2]string{
		{"__title__", "System"},
		{"Host", info.Host},
		{"OS", info.OS},
		{"Kernel", info.Kernel},
		{"Uptime", info.Uptime},
		{"Packages", info.Packages},
		{"Terminal", terminalSummary(info.Terminal, info.Shell)},
		{"Display", info.Display},
		{"CPU", info.CPU},
		{"GPU", info.GPU},
		{"Memory", info.Memory},
		{"Disk (/)", info.Storage},
		{localIPLabel(info.LocalIPIface), info.LocalIP},
		{batteryLabel(info.BatteryName), info.Battery},
		{"Power", info.Power},
		{"Locale", info.Locale},
	}
	aiRows := [][2]string{
		{"__title__", "AI Stack"},
		{"Runtimes", r.runtimeLine(info.Runtimes, 0, 3)},
		{"Engines", r.runtimeLine(info.Runtimes, 3, len(info.Runtimes))},
		{"Best RT", bestRuntime(info)},
		{"Accel", accelerator(info)},
		{"Max Fit", maxFit(info.MemoryGB)},
		{"Quant", recommendedQuant(info.MemoryGB)},
		{"Registry", fmt.Sprintf("%d models", modelCount)},
		{"AI Tier", info.AITier},
		{"Profile", info.Profile},
		{"AI Score", fmt.Sprintf("%d", info.AIScore)},
		{"Capability", info.Capability},
	}

	height := len(logo)
	if len(systemRows) > height {
		height = len(systemRows)
	}
	if len(aiRows) > height {
		height = len(aiRows)
	}

	var b strings.Builder
	b.WriteString(r.fetchBorder("top", leftW, systemW, aiW))
	for i := 0; i < height; i++ {
		left := ""
		if i < len(logo) {
			left = r.style(logo[i], white, false)
		}
		systemCol := ""
		if i < len(systemRows) {
			systemCol = r.fetchRow(systemRows[i][0], systemRows[i][1], systemW)
		}
		aiCol := ""
		if i < len(aiRows) {
			aiCol = r.fetchRow(aiRows[i][0], aiRows[i][1], aiW)
		}
		b.WriteString(r.v())
		b.WriteString(pad(left, leftW))
		b.WriteString(r.v())
		b.WriteString(pad(systemCol, systemW))
		b.WriteString(r.v())
		b.WriteString(pad(aiCol, aiW))
		b.WriteString(r.v())
		b.WriteString("\n")
	}
	b.WriteString(r.fetchBorder("bottom", leftW, systemW, aiW))
	return b.String()
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

func (r Renderer) fetchRow(label, value string, width int) string {
	if label == "__title__" {
		return r.style(trimText(value, width), green, true)
	}
	if label == "" && strings.HasPrefix(value, "-") {
		return r.style(value, gray, false)
	}
	labelW := 11
	if visibleWidth(label) > labelW {
		labelW = visibleWidth(label) + 2
	}
	valueW := max(0, width-labelW)
	if label != "Runtimes" && label != "Engines" {
		value = trimText(value, valueW)
	}
	return r.style(pad(label, labelW), cyan, true) + r.valueStyle(label, value)
}

func (r Renderer) runtimeLine(runtimes []system.RuntimeStatus, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(runtimes) {
		end = len(runtimes)
	}
	if start >= end {
		return "-"
	}
	parts := make([]string, 0, len(runtimes))
	for _, rt := range runtimes[start:end] {
		mark := r.style("NO", red, true)
		if r.opts.Unicode {
			if rt.Found {
				mark = r.style("✓", green, true)
			} else {
				mark = r.style("×", red, true)
			}
		} else if rt.Found {
			mark = r.style("OK", green, true)
		}
		parts = append(parts, r.style(rt.Name, cyan, false)+" "+mark)
	}
	return strings.Join(parts, "  ")
}

func (r Renderer) valueStyle(label, value string) string {
	switch label {
	case "Runtimes", "Engines":
		return value
	case "AI Tier":
		return r.style(value, yellow, true)
	case "AI Score":
		return r.style(value, green, true)
	case "Profile":
		return r.style(value, magenta, true)
	case "GPU", "CPU":
		return r.style(value, yellow, true)
	default:
		if strings.HasPrefix(label, "Battery") {
			return r.batteryPercentValue(value)
		}
		return r.percentValue(value)
	}
}

func (r Renderer) percentValue(value string) string {
	percent, start, end, ok := percentRange(value)
	if !ok {
		return r.style(value, white, false)
	}
	code := green
	if percent >= 85 {
		code = red
	} else if percent >= 70 {
		code = yellow
	}
	return r.style(value[:start], white, false) + r.style(value[start:end], code, true) + r.style(value[end:], white, false)
}

func (r Renderer) batteryPercentValue(value string) string {
	percent, start, end, ok := percentRange(value)
	if !ok {
		return r.style(value, white, false)
	}
	code := green
	if percent <= 20 {
		code = red
	} else if percent <= 50 {
		code = yellow
	}
	return r.style(value[:start], white, false) + r.style(value[start:end], code, true) + r.style(value[end:], white, false)
}

func (r Renderer) modelTable(models []model.Model, systemMemoryGB int) string {
	headers := []string{"Rank", "Model", "Provider", "Best", "Type", "Bit", "Score", "Runtime", "In/s", "Out tok/s", "TPM", "Mem", "Mem%", "Fit", "Context", "License", "Trend"}
	widths := []int{4, 22, 8, 8, 8, 5, 5, 10, 5, 9, 6, 5, 5, 7, 7, 8, 6}
	var rows [][]string
	for _, m := range models {
		rows = append(rows, []string{
			r.style(fmt.Sprintf("%d", m.Rank), white, m.Rank <= 3),
			r.style(trimText(m.Name, widths[1]), white, m.Rank <= 3),
			r.style(trimText(m.Provider, widths[2]), gray, false),
			r.style(trimText(m.BestFor, widths[3]), bestForColor(m.BestFor), false),
			r.style(trimText(m.Type, widths[4]), cyan, false),
			r.style(bitLabel(m), bitColor(bitLabel(m)), true),
			r.score(m.Score),
			r.runtime(trimText(m.Runtime, widths[7])),
			r.style(fmt.Sprintf("%d", m.InTPS), white, false),
			r.speed(m.OutTPS),
			r.style(fmt.Sprintf("%d", tokensPerMinute(m.OutTPS)), white, false),
			r.style(fmt.Sprintf("%dGB", m.MemoryGB), white, false),
			r.style(memoryPercent(m.MemoryGB, systemMemoryGB), yellow, false),
			r.fit(m.Fit),
			r.style(trimText(m.Context, widths[14]), blue, false),
			r.style(trimText(m.License, widths[15]), licenseColor(m.License), false),
			r.trend(m.Trend),
		})
	}

	var b strings.Builder
	b.WriteString(r.tableBorder("top", widths))
	b.WriteString(r.tableRow(headers, widths, true))
	b.WriteString(r.tableBorder("mid", widths))
	for _, row := range rows {
		b.WriteString(r.tableRow(row, widths, false))
	}
	b.WriteString(r.tableBorder("bottom", widths))
	return b.String()
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

func bitLabel(m model.Model) string {
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

func (r Renderer) tableRow(values []string, widths []int, header bool) string {
	cells := make([]string, 0, len(values))
	for i, value := range values {
		if header {
			value = r.style(value, gray, true)
		}
		cells = append(cells, " "+pad(value, widths[i])+" ")
	}
	return r.v() + strings.Join(cells, r.v()) + r.v() + "\n"
}

func (r Renderer) tableBorder(kind string, widths []int) string {
	var left, mid, right, line string
	if r.opts.Unicode {
		switch kind {
		case "top":
			left, mid, right, line = "┌", "┬", "┐", "─"
		case "mid":
			left, mid, right, line = "├", "┼", "┤", "─"
		default:
			left, mid, right, line = "└", "┴", "┘", "─"
		}
	} else {
		left, mid, right, line = "+", "+", "+", "-"
	}
	parts := make([]string, 0, len(widths))
	for _, w := range widths {
		parts = append(parts, strings.Repeat(line, w+2))
	}
	return r.style(left+strings.Join(parts, mid)+right, gray, false) + "\n"
}

func (r Renderer) fetchBorder(kind string, widths ...int) string {
	var left, mid, right, line string
	if r.opts.Unicode {
		line = "─"
		if kind == "top" {
			left, mid, right = "┌", "┬", "┐"
		} else {
			left, mid, right = "└", "┴", "┘"
		}
	} else {
		left, mid, right, line = "+", "+", "+", "-"
	}
	parts := make([]string, 0, len(widths))
	for _, w := range widths {
		parts = append(parts, strings.Repeat(line, w))
	}
	return r.style(left+strings.Join(parts, mid)+right, gray, false) + "\n"
}

func (r Renderer) borderTop(leftW, rightW int) string {
	if r.opts.Unicode {
		return r.style("┌"+strings.Repeat("─", leftW)+"┬"+strings.Repeat("─", rightW)+"┐", gray, false) + "\n"
	}
	return r.style("+"+strings.Repeat("-", leftW)+"+"+strings.Repeat("-", rightW)+"+", gray, false) + "\n"
}

func (r Renderer) borderMid(leftW, rightW int) string {
	if r.opts.Unicode {
		return r.style("├"+strings.Repeat("─", leftW)+"┼"+strings.Repeat("─", rightW)+"┤", gray, false) + "\n"
	}
	return r.style("+"+strings.Repeat("-", leftW)+"+"+strings.Repeat("-", rightW)+"+", gray, false) + "\n"
}

func (r Renderer) borderBottom(leftW, rightW int) string {
	if r.opts.Unicode {
		return r.style("└"+strings.Repeat("─", leftW)+"┴"+strings.Repeat("─", rightW)+"┘", gray, false) + "\n"
	}
	return r.style("+"+strings.Repeat("-", leftW)+"+"+strings.Repeat("-", rightW)+"+", gray, false) + "\n"
}

func (r Renderer) v() string {
	if r.opts.Unicode {
		return r.style("│", gray, false)
	}
	return r.style("|", gray, false)
}

func (r Renderer) score(score int) string {
	switch {
	case score >= 95:
		return r.style(fmt.Sprintf("%d", score), green, true)
	case score >= 90:
		return r.style(fmt.Sprintf("%d", score), cyan, true)
	default:
		return r.style(fmt.Sprintf("%d", score), yellow, false)
	}
}

func (r Renderer) runtime(value string) string {
	if strings.Contains(value, "MLX") {
		return r.style(value, magenta, true)
	}
	if value == "Ollama" {
		return r.style(value, green, false)
	}
	return r.style(value, yellow, false)
}

func (r Renderer) speed(value int) string {
	text := fmt.Sprintf("%d", value)
	if r.opts.Emoji {
		text = "⚡" + text
	}
	return r.style(text, yellow, false)
}

func (r Renderer) fit(value string) string {
	text := value
	if r.opts.Emoji {
		switch value {
		case "Best":
			text = "✅ Best"
		case "Good":
			text = "👍 Good"
		case "Near":
			text = "🟡 Near"
		}
	}
	switch value {
	case "Best":
		return r.style(text, green, true)
	case "Good":
		return r.style(text, cyan, true)
	case "Near":
		return r.style(text, yellow, true)
	default:
		return r.style(text, red, true)
	}
}

func (r Renderer) trend(value int) string {
	switch {
	case value > 0:
		if !r.opts.Unicode {
			return r.style(fmt.Sprintf("UP%d", value), green, false)
		}
		return r.style(fmt.Sprintf("↑%d", value), green, false)
	case value < 0:
		if !r.opts.Unicode {
			return r.style(fmt.Sprintf("DN%d", -value), red, false)
		}
		return r.style(fmt.Sprintf("↓%d", -value), red, false)
	default:
		if r.opts.Unicode {
			return r.style("→0", gray, false)
		}
		return r.style("0", gray, false)
	}
}

func (r Renderer) style(value string, color string, bold bool) string {
	if !r.opts.Color {
		return value
	}
	code := color
	if bold {
		code = "1;" + color
	}
	return "\033[" + code + "m" + value + "\033[0m"
}

func pad(value string, width int) string {
	current := visibleWidth(value)
	if current >= width {
		return value
	}
	return value + strings.Repeat(" ", width-current)
}

func trimText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if visibleWidth(value) <= width {
		return value
	}
	runes := []rune(value)
	for len(runes) > 0 && visibleWidth(string(runes)) > width-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
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
		if r > 0x1100 {
			width += 2
		} else {
			width++
		}
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

func bestForColor(value string) string {
	switch value {
	case "Coding", "Reasoning", "Vision", "Vision OCR", "RAG":
		return cyan
	default:
		return white
	}
}

func licenseColor(value string) string {
	if strings.Contains(value, "NC") {
		return yellow
	}
	return gray
}

func firstN(models []model.Model, n int) []model.Model {
	if len(models) <= n {
		return models
	}
	return models[:n]
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
