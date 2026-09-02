# Homebrew formula for pgbook. Lives in the pgrundev/homebrew-tap repo
# as Formula/pgbook.rb; this copy is the template kept next to the code.
#
# After each release, update `version` and the four sha256 values from
# the release's checksums.txt, then push to the tap.
#
#   brew install pgrundev/tap/pgbook
class Pgbook < Formula
  desc "Postgres Book in your terminal — one topic at a time"
  homepage "https://pgbook.dev"
  version "0.1.0"
  license "MIT"

  base = "https://github.com/pgrundev/pgbook/releases/download/v#{version}"

  on_macos do
    on_arm do
      url "#{base}/pgbook_#{version}_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_ARM64_SHA256"
    end
    on_intel do
      url "#{base}/pgbook_#{version}_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_AMD64_SHA256"
    end
  end

  on_linux do
    on_arm do
      url "#{base}/pgbook_#{version}_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_ARM64_SHA256"
    end
    on_intel do
      url "#{base}/pgbook_#{version}_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "pgbook"
  end

  test do
    assert_match "pgbook", shell_output("#{bin}/pgbook --version")
  end
end
