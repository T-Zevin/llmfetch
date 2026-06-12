# LLMFetch

Fastfetch-style AI workstation dashboard for local LLM discovery, fit checks, and model recommendations.

## Run

```bash
go run ./cmd/llmfetch
```

Or build:

```bash
make build
./bin/llmfetch
```

## Modes

```bash
llmfetch
llmfetch -i
llmfetch --snapshot
llmfetch -s
llmfetch --json
llmfetch --logos
llmfetch --ascii
llmfetch --no-color
llmfetch --no-emoji
```

`llmfetch` opens the interactive browser by default. Use `--snapshot` or `-s` for a fastfetch-style static dashboard.
Use `--logos` to review candidate OS and distro logos before they are wired into automatic detection.

## Interactive Commands

```text
/          enter search mode
Esc/Enter  leave search mode
Ctrl+U     clear search while searching
c          clear search
s          cycle sort: Score, Out tok/s, Memory, Context, Fit, Trend
f          cycle Fit filter: All, Best, Good, Near
↑/↓ or j/k move selection
d/Enter    toggle selected model detail
q          quit
```

## Current Status

This is the first Go implementation. It already includes:

- macOS system detection
- Apple Silicon chip and memory detection
- runtime detection for MLX, Ollama, LM Studio, llama.cpp, vLLM
- fastfetch-style dashboard
- model ranking table
- JSON output
- zero-dependency htop-style interactive browser
- bundled model registry snapshot
- reviewable OS/distro logo catalog

The current interactive mode is intentionally dependency-free. A BubbleTea/LipGloss TUI can be added once dependency fetching is stable.

## Model Registry

The bundled sample registry lives in:

```text
registry/models.json
internal/registry/models.json
```

The long-term design is a curated registry enriched by automated metadata fetchers, rather than relying on live APIs at runtime.

Refresh the registry:

```bash
python3 scripts/collect_models.py --target 500
make build
```

The collector currently uses the Hugging Face public API for discovery, filters to mainstream local model families, normalizes fields, writes `registry/models.json`, and syncs the embedded copy at `internal/registry/models.json`.

## Logo Review

Preview the built-in logo candidates:

```bash
./bin/llmfetch --logos
```

For Linux coverage research, compare against the host tools if they are installed:

```bash
fastfetch --list-logos
fastfetch --print-logos | less -R
neofetch -L --ascii_distro Ubuntu
```

Confirmed Linux logos can be mapped from `/etc/os-release` fields such as `ID`, `ID_LIKE`, `NAME`, and `PRETTY_NAME`.

## Release

Local package:

```bash
make build
mkdir -p dist
tar -C bin -czf dist/llmfetch-<version>-aarch64-apple-darwin.tar.gz llmfetch
shasum -a 256 dist/llmfetch-<version>-aarch64-apple-darwin.tar.gz
```

GitHub release:

```bash
git tag v0.1.0
git push origin main --tags
```

The included GitHub Actions workflow builds release archives from `.goreleaser.yaml` when a `v*` tag is pushed.
