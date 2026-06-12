package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/xzw/llmfetch/internal/logo"
	"github.com/xzw/llmfetch/internal/registry"
	"github.com/xzw/llmfetch/internal/render"
	"github.com/xzw/llmfetch/internal/system"
	"github.com/xzw/llmfetch/internal/tui"
)

type Config struct {
	Interactive bool
	Snapshot    bool
	JSON        bool
	Logos       bool
	NoColor     bool
	ASCII       bool
	NoEmoji     bool
}

type Snapshot struct {
	System system.Info `json:"system"`
	Models any         `json:"models"`
}

func Run(args []string) error {
	cfg := Config{}
	fs := flag.NewFlagSet("llmfetch", flag.ContinueOnError)
	fs.BoolVar(&cfg.Interactive, "interactive", false, "open interactive TUI (default)")
	fs.BoolVar(&cfg.Interactive, "i", false, "open interactive TUI (default)")
	fs.BoolVar(&cfg.Snapshot, "snapshot", false, "print dashboard snapshot")
	fs.BoolVar(&cfg.Snapshot, "s", false, "print dashboard snapshot")
	fs.BoolVar(&cfg.JSON, "json", false, "print JSON snapshot")
	fs.BoolVar(&cfg.Logos, "logos", false, "print logo catalog")
	fs.BoolVar(&cfg.NoColor, "no-color", false, "disable ANSI colors")
	fs.BoolVar(&cfg.ASCII, "ascii", false, "disable unicode boxes and emoji")
	fs.BoolVar(&cfg.NoEmoji, "no-emoji", false, "disable emoji while keeping unicode boxes")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if cfg.Logos {
		fmt.Print(logo.RenderCatalog())
		return nil
	}

	models, err := registry.LoadBundledModels()
	if err != nil {
		return err
	}
	info := system.Detect()

	if cfg.JSON {
		out, err := json.MarshalIndent(Snapshot{System: info, Models: models}, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}

	opts := render.Options{
		Color:   !cfg.NoColor && os.Getenv("NO_COLOR") == "",
		Emoji:   !cfg.NoEmoji && !cfg.ASCII,
		Unicode: !cfg.ASCII,
	}

	if cfg.Snapshot {
		fmt.Println(render.New(opts).Dashboard(info, models))
		return nil
	}

	return tui.Run(info, models, opts)
}
