# Changelog

## v0.4.1

- Updated installation docs for the public Homebrew tap.

## v0.4.0

- Expanded the bundled model registry target to 10000 entries.
- Updated the registry collector to include high-signal community models when known family rules do not cover enough candidates.

## v0.3.0

- Expanded the bundled model registry from 500 to 5000 entries.
- Refreshed the README into separate Chinese and English pages with project assets.

## v0.2.1

- Fixed the GoReleaser archive configuration for current GoReleaser v2 releases.

## v0.2.0

- Default `llmfetch` opens the interactive model browser.
- Added fastfetch-style system and AI stack header.
- Added colorized ranking table with column-level highlights.
- Added model name marquee scrolling for long model IDs.
- Added model metrics: type, quantization bit, input speed, output token speed, TPM, memory percent, context, license, and fit.
- Added a bundled 500-model registry refresh flow.
- Added `--logos` for OS and distro logo review before Linux detection is finalized.
- Added GitHub Actions release workflow for GoReleaser.

## v0.1.0

- Initial Go prototype.
- Added static snapshot output, JSON output, macOS system detection, and bundled registry support.
