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

If Homebrew fails while cloning GitHub over HTTPS:

```text
fatal: unable to access 'https://github.com/T-Zevin/homebrew-tap/': Recv failure: Connection reset by peer
```

Use the SSH tap URL instead:

```bash
brew tap T-Zevin/tap git@github.com:T-Zevin/homebrew-tap.git
brew install T-Zevin/tap/llmfetch
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
