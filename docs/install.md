# Install

## macOS: Homebrew Tap

Users can install with:

```bash
brew install T-Zevin/tap/llmfetch
```

Or:

```bash
brew tap T-Zevin/tap
brew install llmfetch
```

The tap repository is:

```text
https://github.com/T-Zevin/homebrew-tap
```

## macOS / Linux: Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/T-Zevin/llmfetch/main/install.sh | sh
```

Custom install directory:

```bash
curl -fsSL https://raw.githubusercontent.com/T-Zevin/llmfetch/main/install.sh | LLMFETCH_INSTALL_DIR="$HOME/.local/bin" sh
```

## Manual Download

Download the latest release:

```text
https://github.com/T-Zevin/llmfetch/releases/latest
```

Available release archives:

- `aarch64-apple-darwin.tar.gz`
- `x86_64-apple-darwin.tar.gz`
- `aarch64-unknown-linux-gnu.tar.gz`
- `x86_64-unknown-linux-gnu.tar.gz`
- `aarch64-pc-windows-msvc.zip`
- `x86_64-pc-windows-msvc.zip`
