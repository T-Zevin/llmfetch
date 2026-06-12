class Llmfetch < Formula
  desc "AI workstation dashboard for local LLM discovery and fit checks"
  homepage "https://github.com/T-Zevin/llmfetch"
  version "0.4.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/T-Zevin/llmfetch/releases/download/v0.4.0/llmfetch-0.4.0-aarch64-apple-darwin.tar.gz"
      sha256 "0cf61e505cadc16cda3f87d18d6d8417b8dc5f830ba607fffc3ea6090d58d612"
    else
      url "https://github.com/T-Zevin/llmfetch/releases/download/v0.4.0/llmfetch-0.4.0-x86_64-apple-darwin.tar.gz"
      sha256 "ae122f93e9c679e6a18db342fb4ee8eca9a0a8d12f351fbb8e7258e58fdf49e9"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/T-Zevin/llmfetch/releases/download/v0.4.0/llmfetch-0.4.0-aarch64-unknown-linux-gnu.tar.gz"
      sha256 "6b0d06434509544a660905e5890ed1f2c904bc23a4f26e50bfa3812af9ca3d14"
    else
      url "https://github.com/T-Zevin/llmfetch/releases/download/v0.4.0/llmfetch-0.4.0-x86_64-unknown-linux-gnu.tar.gz"
      sha256 "5bdd4f148fa5a7bcc855bf0b85b447036acba5618df254ae505d326c3ac36a19"
    end
  end

  def install
    bin.install "llmfetch"
  end

  test do
    assert_match "models", shell_output("#{bin}/llmfetch --json")
  end
end
