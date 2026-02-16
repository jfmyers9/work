# Homebrew Formula for work
# To use this formula:
#   1. Create a tap: brew tap jfmyers9/work
#   2. Copy this file to the tap repository as Formula/work.rb
#   3. Update the version and SHA256 for each release
#
# Users can then install with:
#   brew install jfmyers9/work/work

class Work < Formula
  desc "Lightweight, filesystem-based issue tracker CLI"
  homepage "https://github.com/jfmyers9/work"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jfmyers9/work/releases/download/v#{version}/work-v#{version}-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_ARM64_SHA256"
    else
      url "https://github.com/jfmyers9/work/releases/download/v#{version}/work-v#{version}-darwin-amd64.tar.gz"
      sha256 "REPLACE_WITH_AMD64_SHA256"
    end
  end

  def install
    if Hardware::CPU.arm?
      bin.install "work-darwin-arm64" => "work"
    else
      bin.install "work-darwin-amd64" => "work"
    end
  end

  def caveats
    <<~EOS
      To get started with work:

      1. Initialize a tracker in your project:
           work init

      2. Create your first issue:
           work create "My first issue"

      3. View your issues:
           work list

      For more information, visit:
           https://github.com/jfmyers9/work
    EOS
  end

  test do
    system "#{bin}/work", "version"
  end
end
