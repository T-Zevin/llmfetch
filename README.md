# LLMFetch

> 中文优先 | English below each section

[![Release](https://img.shields.io/github/v/release/T-Zevin/llmfetch?style=flat-square)](https://github.com/T-Zevin/llmfetch/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/T-Zevin/llmfetch/release.yml?style=flat-square)](https://github.com/T-Zevin/llmfetch/actions)
[![License](https://img.shields.io/github/license/T-Zevin/llmfetch?style=flat-square)](LICENSE)

LLMFetch 是一个面向本地大模型玩家的 AI 工作站仪表盘。它像 `fastfetch` 一样展示你的机器配置，又像轻量版 `htop` 一样浏览、搜索、排序本地 LLM 推荐榜。

LLMFetch is a fastfetch-style AI workstation dashboard for local LLM discovery, hardware fit checks, and model recommendations.

```text
LLMFetch  AI workstation dashboard

┌──────────────────────┬──────────────────────────────┬────────────────────────┐
│        Apple Logo     │ System                       │ AI Stack               │
│                      │ CPU       Apple M3 Max        │ Runtime   MLX / Ollama  │
│                      │ Memory    36 GB               │ Profile   MLX Powerhouse│
│                      │ Display   3 screens           │ AI Score  82            │
└──────────────────────┴──────────────────────────────┴────────────────────────┘

Rank  Model                 Provider  Score  Runtime     Out tok/s  Memory  Fit
1     DeepSeek-R1-0528      DeepSeek  97     MLX Native  90         8GB     Best
2     Qwen3-Coder-30B-A3B   Alibaba   96     MLX Native  55         24GB    Best
```

## ✨ 核心亮点 | Highlights

- 🖥️ **系统识别**：macOS、Apple Silicon、内存、磁盘、屏幕、电池、网络、终端环境。
- 🧠 **AI Stack**：检测 MLX、Ollama、LM Studio、llama.cpp、vLLM 等本地运行环境。
- 📊 **模型排行榜**：按综合分、速度、内存占用、上下文、适配度等指标排序。
- 🔎 **交互浏览**：默认进入 htop 风格 TUI，支持搜索、筛选、排序、详情展开。
- 🎞️ **长模型名滚动**：模型名过长时在选中行做水平滚动展示。
- 🎨 **彩色终端 UI**：重点列高亮，支持 `--no-color`、`--ascii`、`--no-emoji`。
- 🧩 **500 模型注册表**：内置主流本地模型样本，并可用脚本刷新。
- 🐧 **Logo Catalog**：内置 OS/发行版 logo 候选清单，方便后续确认 Linux 支持。

- 🖥️ **System detection** for macOS, Apple Silicon, memory, disk, displays, battery, network, and terminal.
- 🧠 **AI runtime detection** for MLX, Ollama, LM Studio, llama.cpp, and vLLM.
- 📊 **Model ranking** by score, speed, memory, context, fit, license, and provider.
- 🔎 **Interactive TUI** with search, filters, sorting, selection, and detail view.
- 🎞️ **Marquee model names** for long model IDs.
- 🎨 **Color terminal UI** with compatibility switches.
- 🧩 **Bundled 500-model registry** with refresh script.
- 🐧 **Reviewable OS/distro logo catalog** before Linux mapping is finalized.

## 🚀 安装 | Install

下载最新版本：

Download the latest release:

[https://github.com/T-Zevin/llmfetch/releases/latest](https://github.com/T-Zevin/llmfetch/releases/latest)

Apple Silicon Mac:

```bash
curl -L -o llmfetch.tar.gz \
  https://github.com/T-Zevin/llmfetch/releases/download/v0.2.1/llmfetch-0.2.1-aarch64-apple-darwin.tar.gz
tar -xzf llmfetch.tar.gz
cd llmfetch-0.2.1-aarch64-apple-darwin
./llmfetch
```

macOS 如果提示未验证开发者，可以先本地解除隔离属性：

If macOS blocks the binary, remove the quarantine attribute locally:

```bash
xattr -dr com.apple.quarantine ./llmfetch
```

## ⚡ 快速使用 | Quick Start

```bash
# 默认：打开交互界面
# Default: open interactive TUI
llmfetch

# 快照模式，类似 fastfetch
# Static snapshot, similar to fastfetch
llmfetch --snapshot

# JSON 输出，方便脚本或前端读取
# JSON output for scripts or frontends
llmfetch --json

# 查看候选 OS / Linux 发行版 logo
# Preview candidate OS / Linux distro logos
llmfetch --logos
```

## 🕹️ 交互快捷键 | Interactive Keys

| Key | 中文 | English |
| --- | --- | --- |
| `/` | 进入搜索 | Enter search mode |
| `Esc` / `Enter` | 退出搜索 | Leave search mode |
| `Ctrl+U` | 清空搜索输入 | Clear search input |
| `c` | 清空搜索 | Clear search |
| `s` | 切换排序：Score、Out tok/s、Memory、Context、Fit、Trend | Cycle sort modes |
| `f` | 切换适配筛选：All、Best、Good、Near | Cycle fit filters |
| `↑/↓` or `j/k` | 移动选中行 | Move selection |
| `d` / `Enter` | 展开/收起模型详情 | Toggle model detail |
| `q` | 退出 | Quit |

## 🧰 参数 | CLI Options

| 参数 | 中文说明 | English |
| --- | --- | --- |
| `llmfetch` | 默认打开交互界面 | Open interactive TUI by default |
| `-i`, `--interactive` | 显式打开交互界面 | Open interactive TUI |
| `-s`, `--snapshot` | 输出一次性快照 | Print static dashboard snapshot |
| `--json` | 输出 JSON | Print JSON snapshot |
| `--logos` | 打印候选 logo 清单 | Print logo catalog |
| `--ascii` | 禁用 Unicode 边框和 emoji | Disable Unicode boxes and emoji |
| `--no-color` | 禁用 ANSI 颜色 | Disable ANSI colors |
| `--no-emoji` | 禁用 emoji，保留 Unicode 边框 | Disable emoji only |
| `--help` | 查看帮助 | Show help |

## 📦 发布包 | Release Artifacts

当前 GitHub Release 会自动构建这些平台：

Current GitHub Release builds these targets automatically:

| Platform | Arch | Package |
| --- | --- | --- |
| macOS | Apple Silicon | `aarch64-apple-darwin.tar.gz` |
| macOS | Intel | `x86_64-apple-darwin.tar.gz` |
| Linux | ARM64 | `aarch64-unknown-linux-gnu.tar.gz` |
| Linux | x86_64 | `x86_64-unknown-linux-gnu.tar.gz` |
| Windows | ARM64 | `aarch64-pc-windows-msvc.zip` |
| Windows | x86_64 | `x86_64-pc-windows-msvc.zip` |

每个版本都会附带 `checksums.txt`。

Each release includes `checksums.txt`.

## 🧠 模型注册表 | Model Registry

内置模型数据：

Bundled model data:

```text
registry/models.json
internal/registry/models.json
```

刷新 500 个模型样本：

Refresh 500 model entries:

```bash
python3 scripts/collect_models.py --target 500
make build
```

当前采集脚本使用 Hugging Face public API 做发现和归一化。长期目标是做一个“自动采集 + 人工校准”的本地模型 registry，而不是每次运行都依赖实时 API。

The collector currently uses the Hugging Face public API for discovery and normalization. The long-term goal is a curated registry enriched by automated metadata fetchers, not a live API dependency at runtime.

## 🐧 Logo 确认 | Logo Review

预览 LLMFetch 内置候选 logo：

Preview built-in candidate logos:

```bash
./bin/llmfetch --logos
```

如果本机安装了 fastfetch / neofetch，可以对照确认 Linux 发行版风格：

If fastfetch / neofetch is installed, compare distro styles:

```bash
fastfetch --list-logos
fastfetch --print-logos | less -R
neofetch -L --ascii_distro Ubuntu
```

后续 Linux 自动识别会优先读取 `/etc/os-release`：

Linux auto-detection will later map from `/etc/os-release`:

```text
ID=ubuntu
ID_LIKE=debian
NAME="Ubuntu"
PRETTY_NAME="Ubuntu 24.04.2 LTS"
```

## 🛠️ 本地开发 | Development

```bash
git clone git@github.com:T-Zevin/llmfetch.git
cd llmfetch
go test ./...
make build
./bin/llmfetch
```

生成本地 macOS 包：

Create a local macOS package:

```bash
make build
mkdir -p dist
tar -C bin -czf dist/llmfetch-<version>-aarch64-apple-darwin.tar.gz llmfetch
shasum -a 256 dist/llmfetch-<version>-aarch64-apple-darwin.tar.gz
```

## 🚢 发布 | Release

推送 `v*` tag 会触发 GitHub Actions + GoReleaser：

Push a `v*` tag to trigger GitHub Actions + GoReleaser:

```bash
git tag v0.2.1
git push origin main --tags
```

配置文件：

Config files:

```text
.github/workflows/release.yml
.goreleaser.yaml
```

## 📍 路线图 | Roadmap

- Linux 发行版 logo 与 `/etc/os-release` 自动映射
- 更完整的模型来源、license、quant、backend 可信度校准
- `--limit`、`--sort`、`--filter`、`--provider` 等 CLI 参数
- 可选 benchmark / live-bench 模块
- Homebrew tap / install script

- Linux distro logo mapping via `/etc/os-release`
- Better provider, license, quant, and backend confidence metadata
- More CLI filters such as `--limit`, `--sort`, `--filter`, and `--provider`
- Optional benchmark / live-bench modules
- Homebrew tap / install script

## 📄 License

MIT
