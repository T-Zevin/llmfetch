<div align="center">

<img src="docs/assets/llmfetch-logo.png" alt="LLMFetch" width="860">

# LLMFetch

**An AI workstation dashboard that helps you find the best local models for your machine.**

[中文](README.md) · [Latest Release](https://github.com/T-Zevin/llmfetch/releases/latest) · [Changelog](CHANGELOG.md)

![Release](https://img.shields.io/github/v/release/T-Zevin/llmfetch?style=for-the-badge&logo=github)
![Build](https://img.shields.io/github/actions/workflow/status/T-Zevin/llmfetch/release.yml?style=for-the-badge&logo=githubactions)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/github/license/T-Zevin/llmfetch?style=for-the-badge)

</div>

## Contents

- [Overview](#overview)
- [Screenshot](#screenshot)
- [Install](#install)
- [Usage](#usage)
- [Interactive Keys](#interactive-keys)
- [Highlights](#highlights)
- [Model Registry](#model-registry)
- [Logo Review](#logo-review)
- [Development](#development)
- [Roadmap](#roadmap)

## Overview

LLMFetch is a terminal tool for local LLM users. It shows machine details like `fastfetch`, then gives you an htop-style interactive browser for model ranking, search, filtering, and fit checks.

The goal is simple: open a terminal and quickly understand which models fit your hardware, which runtime to use, how much memory they need, and what speed to expect.

## Screenshot

<div align="center">
  <img src="docs/assets/llmfetch-screenshot.png" alt="LLMFetch terminal screenshot" width="960">
</div>

## Install

macOS / Linux one-line installer:

```bash
curl -fsSL https://raw.githubusercontent.com/T-Zevin/llmfetch/main/install.sh | sh
```

Install into a user directory:

```bash
curl -fsSL https://raw.githubusercontent.com/T-Zevin/llmfetch/main/install.sh | LLMFETCH_INSTALL_DIR="$HOME/.local/bin" sh
```

The Homebrew Tap formula is included in this repository. After creating `T-Zevin/homebrew-tap`, copy `packaging/homebrew/Formula/llmfetch.rb` into `Formula/llmfetch.rb` in the tap repository. Users can then run:

```bash
brew install T-Zevin/tap/llmfetch
```

Manual download:

[https://github.com/T-Zevin/llmfetch/releases/latest](https://github.com/T-Zevin/llmfetch/releases/latest)

Apple Silicon Mac:

```bash
curl -L -o llmfetch.tar.gz \
  https://github.com/T-Zevin/llmfetch/releases/download/v0.4.0/llmfetch-0.4.0-aarch64-apple-darwin.tar.gz
tar -xzf llmfetch.tar.gz
cd llmfetch-0.4.0-aarch64-apple-darwin
./llmfetch
```

If macOS blocks the binary:

```bash
xattr -dr com.apple.quarantine ./llmfetch
```

## Usage

```bash
# Default: open interactive TUI
llmfetch

# Static snapshot, similar to fastfetch
llmfetch --snapshot

# JSON output for scripts or frontends
llmfetch --json

# Preview candidate OS / Linux distro logos
llmfetch --logos
```

Options:

| Option | Description |
| --- | --- |
| `llmfetch` | Open the interactive model browser by default |
| `-i`, `--interactive` | Open the interactive model browser |
| `-s`, `--snapshot` | Print a static system and model snapshot |
| `--json` | Print JSON data |
| `--logos` | Print candidate OS / Linux distro logos |
| `--ascii` | Disable Unicode boxes and emoji |
| `--no-color` | Disable ANSI colors |
| `--no-emoji` | Disable emoji while keeping Unicode boxes |
| `--help` | Show help |

## Interactive Keys

| Key | Action |
| --- | --- |
| `/` | Enter search mode |
| `Esc` / `Enter` | Leave search mode |
| `Ctrl+U` | Clear search input |
| `c` | Clear search |
| `s` | Cycle sort: Score, Out tok/s, Memory, Context, Fit, Trend |
| `f` | Cycle fit filter: All, Best, Good, Near |
| `↑/↓` or `j/k` | Move selection |
| `d` / `Enter` | Toggle model detail |
| `q` | Quit |

## Highlights

- System detection for macOS, Apple Silicon, memory, disk, displays, battery, network, and terminal.
- AI runtime detection for MLX, Ollama, LM Studio, llama.cpp, and vLLM.
- Model ranking by score, speed, memory, context, fit, license, and provider.
- Interactive search, filtering, sorting, selection, and detail view.
- Marquee scrolling for long selected model names.
- Color terminal UI with no-color, ASCII, and no-emoji compatibility modes.
- Multi-platform releases for macOS, Linux, and Windows on arm64 and x86_64.

## Model Registry

Bundled model data:

```text
registry/models.json
internal/registry/models.json
```

Refresh model entries:

```bash
python3 scripts/collect_models.py --target 10000
make build
```

The collector currently uses the Hugging Face public API for discovery and normalization. The long-term goal is a curated registry enriched by automated metadata fetchers, not a live API dependency at runtime.

## Logo Review

Preview built-in candidate logos:

```bash
./bin/llmfetch --logos
```

If fastfetch / neofetch is installed, compare distro styles:

```bash
fastfetch --list-logos
fastfetch --print-logos | less -R
neofetch -L --ascii_distro Ubuntu
```

Linux auto-detection will later map from `/etc/os-release`:

```text
ID=ubuntu
ID_LIKE=debian
NAME="Ubuntu"
PRETTY_NAME="Ubuntu 24.04.2 LTS"
```

## Development

```bash
git clone git@github.com:T-Zevin/llmfetch.git
cd llmfetch
go test ./...
make build
./bin/llmfetch
```

Create a local macOS package:

```bash
make build
mkdir -p dist
tar -C bin -czf dist/llmfetch-<version>-aarch64-apple-darwin.tar.gz llmfetch
shasum -a 256 dist/llmfetch-<version>-aarch64-apple-darwin.tar.gz
```

## Roadmap

- Linux distro logo mapping via `/etc/os-release`
- Better provider, license, quant, and backend confidence metadata
- CLI filters such as `--limit`, `--sort`, `--filter`, and `--provider`
- Benchmark / live-bench modules
- Homebrew Tap and one-line install script

## License

MIT
