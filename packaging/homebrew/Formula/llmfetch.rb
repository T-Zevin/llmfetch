class Llmfetch < Formula
  desc "AI workstation dashboard for local LLM discovery and fit checks"
  homepage "https://github.com/T-Zevin/llmfetch"
  version "0.5.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/T-Zevin/llmfetch/releases/download/v0.5.0/llmfetch-0.5.0-aarch64-apple-darwin.tar.gz"
      sha256 "1bb1d0258fb3c11cc60febe7abbde742cd883b676172a6a1006aa4097ab0dd67"
    else
      url "https://github.com/T-Zevin/llmfetch/releases/download/v0.5.0/llmfetch-0.5.0-x86_64-apple-darwin.tar.gz"
      sha256 "faeb159b5b58b1ca844d85d4a2cdb574ee444292d378ef6373d8817392e016e0"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/T-Zevin/llmfetch/releases/download/v0.5.0/llmfetch-0.5.0-aarch64-unknown-linux-gnu.tar.gz"
      sha256 "3cb462611d3141f79f9a35d772a47d9cd7318ebb565f0fdd28137cfffb5a8252"
    else
      url "https://github.com/T-Zevin/llmfetch/releases/download/v0.5.0/llmfetch-0.5.0-x86_64-unknown-linux-gnu.tar.gz"
      sha256 "ef93fa66615f00a99a947fd5978fcbec68ab462f46835245cb70a6d2360f03ff"
    end
  end

  def install
    bin.install "llmfetch"
  end

  test do
    assert_match "models", shell_output("#{bin}/llmfetch --json")
  end
end
