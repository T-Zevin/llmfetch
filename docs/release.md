# Release Checklist

## One-Time GitHub Setup

```bash
git init
git add .
git commit -m "Initial llmfetch release"
git branch -M main
git remote add origin git@github.com:<owner>/llmfetch.git
git push -u origin main
```

## Release A Version

1. Update the version in release notes and package names.
2. Run tests and build locally.
3. Create a version tag.
4. Push the tag to GitHub.

```bash
go test ./...
make build
git tag v0.2.0
git push origin main --tags
```

The GitHub Actions workflow at `.github/workflows/release.yml` uses GoReleaser and `.goreleaser.yaml` to publish Darwin, Linux, and Windows archives plus SHA256 checksums.

## Local macOS Package

```bash
make build
mkdir -p dist
tar -C bin -czf dist/llmfetch-v0.2.0-aarch64-apple-darwin.tar.gz llmfetch
shasum -a 256 dist/llmfetch-v0.2.0-aarch64-apple-darwin.tar.gz > dist/llmfetch-v0.2.0-aarch64-apple-darwin.tar.gz.sha256
```

## Logo Confirmation Flow

Before wiring Linux distro detection into the main dashboard:

```bash
./bin/llmfetch --logos
fastfetch --list-logos
fastfetch --print-logos | less -R
```

Then map confirmed names from `/etc/os-release`:

```text
ID=ubuntu
ID_LIKE=debian
NAME="Ubuntu"
PRETTY_NAME="Ubuntu 24.04.2 LTS"
```
