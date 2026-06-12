package tui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/xzw/llmfetch/internal/registry"
	"github.com/xzw/llmfetch/internal/render"
	"github.com/xzw/llmfetch/internal/system"
)

func TestViewDoesNotOverflowCommonWidths(t *testing.T) {
	models, err := registry.LoadBundledModels()
	if err != nil {
		t.Fatal(err)
	}
	info := system.Info{
		Host:       "MacBook Pro",
		OS:         "macOS Sequoia 15.7.3 (24G419) arm64",
		Kernel:     "Darwin 24.6.0",
		Uptime:     "4 days, 4 hours, 2 mins",
		Packages:   "163 (brew), 6 (cask)",
		Display:    "Color LCD 3024 x 1964 Retina",
		CPU:        "Apple M3 Max (14 cores)",
		Chip:       "Apple M3 Max",
		GPU:        "Apple M3 Max (30 cores)",
		Memory:     "36 GB",
		Storage:    "217GB free / 926GB",
		Swap:       "Unused",
		LocalIP:    "192.168.110.106",
		Battery:    "100% [AC Connected]",
		Power:      "60W",
		Locale:     "zh_CN.UTF-8",
		Terminal:   "Ghostty",
		Shell:      "zsh",
		AITier:     "A",
		Profile:    "MLX Powerhouse",
		Capability: "Coding A  Reasoning A-  Vision B+",
		AIScore:    82,
		Runtimes:   []system.RuntimeStatus{{Name: "MLX"}, {Name: "Ollama"}, {Name: "LM Studio"}, {Name: "llama.cpp"}, {Name: "vLLM"}},
	}
	for _, size := range []struct {
		width  int
		height int
	}{
		{80, 24},
		{100, 30},
		{118, 34},
		{153, 37},
		{160, 44},
	} {
		s := state{
			info:   info,
			models: models,
			width:  size.width,
			height: size.height,
			opts:   render.Options{Color: true, Emoji: false, Unicode: true},
		}
		assertNoOverflow(t, "default", size.width, s.view())

		s.searchMode = true
		s.query = "qwen"
		assertNoOverflow(t, "search", size.width, s.view())

		s.searchMode = false
		s.sortIndex = 1
		s.fitIndex = 1
		s.showDetail = true
		assertNoOverflow(t, "detail", size.width, s.view())
	}
}

func TestReadKeyDoesNotDropPastedSearchInput(t *testing.T) {
	pendingInput = []byte("/qwen\r")
	want := []string{"/", "q", "w", "e", "n", "enter"}
	for _, expected := range want {
		got, err := readKey()
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}
	if len(pendingInput) != 0 {
		t.Fatalf("pending input not drained: %q", string(pendingInput))
	}
}

func assertNoOverflow(t *testing.T, label string, width int, view string) {
	t.Helper()
	for i, line := range strings.Split(view, "\n") {
		clean := stripANSI(line)
		if visibleWidth(clean) > width {
			t.Fatalf("%s width %d line %d overflows: got %d: %q", label, width, i+1, visibleWidth(clean), clean)
		}
	}
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

func stripANSI(value string) string {
	return ansiPattern.ReplaceAllString(value, "")
}
