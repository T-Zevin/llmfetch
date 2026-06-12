<div align="center">

<img src="docs/assets/llmfetch-logo.png" alt="LLMFetch" width="860">

# LLMFetch

**本地大模型工作站仪表盘，帮你发现最适合当前机器的模型。**

[English](README_EN.md) · [下载 Release](https://github.com/T-Zevin/llmfetch/releases/latest) · [更新日志](CHANGELOG.md)

![Release](https://img.shields.io/github/v/release/T-Zevin/llmfetch?style=for-the-badge&logo=github)
![Build](https://img.shields.io/github/actions/workflow/status/T-Zevin/llmfetch/release.yml?style=for-the-badge&logo=githubactions)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/github/license/T-Zevin/llmfetch?style=for-the-badge)

</div>

## 目录

- [项目介绍](#项目介绍)
- [截图](#截图)
- [快速安装](#快速安装)
- [使用方式](#使用方式)
- [交互快捷键](#交互快捷键)
- [指标说明](#指标说明)
- [功能亮点](#功能亮点)
- [模型注册表](#模型注册表)
- [Logo 确认](#logo-确认)
- [本地开发](#本地开发)
- [路线图](#路线图)

## 项目介绍

LLMFetch 是一个面向本地 LLM 用户的终端工具。它一部分像 `fastfetch`，展示你的系统、芯片、内存、屏幕、电池、运行环境；另一部分像轻量版 `htop`，让你在终端里搜索、排序、筛选模型排行榜。

目标很直接：打开终端，立刻知道这台机器适合跑什么模型、用什么 runtime、需要多少内存、速度大概多少。

## 截图

<div align="center">
  <img src="docs/assets/llmfetch-screenshot.png" alt="LLMFetch terminal screenshot" width="960">
</div>

## 快速安装

macOS / Linux 一行安装：

```bash
curl -fsSL https://raw.githubusercontent.com/T-Zevin/llmfetch/main/install.sh | sh
```

如果想装到用户目录：

```bash
curl -fsSL https://raw.githubusercontent.com/T-Zevin/llmfetch/main/install.sh | LLMFETCH_INSTALL_DIR="$HOME/.local/bin" sh
```

Homebrew 安装：

```bash
brew install T-Zevin/tap/llmfetch
```

或者：

```bash
brew tap T-Zevin/tap
brew install llmfetch
```

如果 GitHub HTTPS clone 被 reset，可以改用 SSH tap：

```bash
brew tap T-Zevin/tap git@github.com:T-Zevin/homebrew-tap.git
brew install T-Zevin/tap/llmfetch
```

也可以手动下载 Release 对应平台的压缩包：

[https://github.com/T-Zevin/llmfetch/releases/latest](https://github.com/T-Zevin/llmfetch/releases/latest)

Apple Silicon Mac:

```bash
curl -L -o llmfetch.tar.gz \
  https://github.com/T-Zevin/llmfetch/releases/download/v0.4.1/llmfetch-0.4.1-aarch64-apple-darwin.tar.gz
tar -xzf llmfetch.tar.gz
cd llmfetch-0.4.1-aarch64-apple-darwin
./llmfetch
```

macOS 如果提示未验证开发者，可以执行：

```bash
xattr -dr com.apple.quarantine ./llmfetch
```

## 使用方式

```bash
# 默认进入交互界面
llmfetch

# 快照模式，类似 fastfetch
llmfetch --snapshot

# 输出 JSON，方便脚本或前端读取
llmfetch --json

# 查看候选 OS / Linux 发行版 logo
llmfetch --logos
```

参数说明：

| 参数 | 说明 |
| --- | --- |
| `llmfetch` | 默认打开交互式模型浏览器 |
| `-i`, `--interactive` | 显式打开交互式模型浏览器 |
| `-s`, `--snapshot` | 输出一次性系统和模型快照 |
| `--json` | 输出 JSON 数据 |
| `--logos` | 打印候选 OS / Linux 发行版 logo |
| `--ascii` | 禁用 Unicode 边框和 emoji |
| `--no-color` | 禁用 ANSI 颜色 |
| `--no-emoji` | 禁用 emoji，保留 Unicode 边框 |
| `--help` | 查看帮助 |

## 交互快捷键

| 按键 | 功能 |
| --- | --- |
| `/` | 进入搜索 |
| `Esc` / `Enter` | 退出搜索 |
| `Ctrl+U` | 清空搜索输入 |
| `c` | 清空搜索 |
| `s` | 切换排序：Score、Out tok/s、Memory、Context、Fit、Trend |
| `f` | 切换适配筛选：All、Best、Good、Near |
| `↑/↓` 或 `j/k` | 移动选中行 |
| `d` / `Enter` | 展开或收起模型详情 |
| `q` | 退出 |

## 指标说明

| 列名 | 含义 |
| --- | --- |
| `Rank` | 当前筛选和排序条件下的排名 |
| `Model` | 模型名称，长名称会在选中行横向滚动 |
| `Provider` | 模型来源或主要发布方，例如 DeepSeek、Alibaba、Meta |
| `Best` | 推荐使用场景，例如 Reasoning、Coding、Vision、Chat |
| `Type` | 模型类型，用于区分通用、代码、推理、视觉、Embedding 等 |
| `Bit` | 量化或精度信息，例如 Q4、8-bit、FP16；无法可靠识别时显示 `Auto` |
| `Score` | 综合评分，结合热度、用途、模型家族和估算适配度 |
| `Runtime` | 推荐运行环境，例如 MLX Native、Ollama、llama.cpp、Transformers |
| `In/s` | 估算输入处理速度，单位为 tokens/s |
| `Out tok/s` | 估算输出生成速度，单位为 tokens/s |
| `TPM` | 估算每分钟输出 token 数，基于 `Out tok/s × 60` |
| `Memory` | 估算推荐内存占用 |
| `Mem%` | 估算占当前机器总内存的比例 |
| `Fit` | 当前机器适配等级：Best、Good、Near |
| `Context` | 模型上下文长度，例如 32K、128K、262K |
| `License` | 模型许可证，例如 MIT、Apache-2、Llama、unknown |
| `Trend` | 基于下载量等信号估算的热度趋势 |

## 功能亮点

- 系统识别：macOS、Apple Silicon、内存、磁盘、屏幕、电池、网络、终端环境。
- AI Stack：检测 MLX、Ollama、LM Studio、llama.cpp、vLLM 等本地运行环境。
- 模型排行榜：按综合分、速度、内存、上下文、适配度、许可证等指标排序。
- 交互浏览：支持搜索、筛选、排序、选择、详情展开。
- 长模型名滚动：选中行自动水平滚动长模型名。
- 彩色终端 UI：重点列高亮，同时支持无颜色、ASCII、无 emoji 模式。
- 多平台发布：macOS、Linux、Windows，支持 arm64 和 x86_64。

## 模型注册表

内置模型数据位于：

```text
registry/models.json
internal/registry/models.json
```

刷新模型：

```bash
python3 scripts/collect_models.py --target 10000
make build
```

当前采集脚本使用 Hugging Face public API 做发现和归一化。长期目标是做一个“自动采集 + 人工校准”的本地模型 registry，而不是每次运行都依赖实时 API。

## Logo 确认

预览 LLMFetch 内置候选 logo：

```bash
./bin/llmfetch --logos
```

如果本机安装了 fastfetch / neofetch，可以对照确认 Linux 发行版风格：

```bash
fastfetch --list-logos
fastfetch --print-logos | less -R
neofetch -L --ascii_distro Ubuntu
```

后续 Linux 自动识别会优先读取 `/etc/os-release`：

```text
ID=ubuntu
ID_LIKE=debian
NAME="Ubuntu"
PRETTY_NAME="Ubuntu 24.04.2 LTS"
```

## 本地开发

```bash
git clone git@github.com:T-Zevin/llmfetch.git
cd llmfetch
go test ./...
make build
./bin/llmfetch
```

生成本地 macOS 包：

```bash
make build
mkdir -p dist
tar -C bin -czf dist/llmfetch-<version>-aarch64-apple-darwin.tar.gz llmfetch
shasum -a 256 dist/llmfetch-<version>-aarch64-apple-darwin.tar.gz
```

## 路线图

- Linux 发行版 logo 与 `/etc/os-release` 自动映射
- 更完整的模型来源、license、quant、backend 可信度校准
- `--limit`、`--sort`、`--filter`、`--provider` 等 CLI 参数
- benchmark / live-bench 模块
- Homebrew Tap 与一行安装脚本

## License

MIT
