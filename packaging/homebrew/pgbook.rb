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
  version "0.2.0"
  license "MIT"

  base = "https://github.com/pgrundev/pgbook/releases/download/v#{version}"

  on_macos do
    on_arm do
      url "#{base}/pgbook_#{version}_darwin_arm64.tar.gz"
      sha256 "8e3fd30f6130a6ec303c115de763e9a49d52d183a359a857f3ad9ae8c5335c2b"
    end
    on_intel do
      url "#{base}/pgbook_#{version}_darwin_amd64.tar.gz"
      sha256 "866180949b855f628db5b12cb86e3e678df1c5a9c7f05586429c81df02415c9c"
    end
  end

  on_linux do
    on_arm do
      url "#{base}/pgbook_#{version}_linux_arm64.tar.gz"
      sha256 "ae9297f8b13e427c3e03506c0314b369850a32e78585dd947311d03bdf8a234d"
    end
    on_intel do
      url "#{base}/pgbook_#{version}_linux_amd64.tar.gz"
      sha256 "1ef8f0f87b963dadd40ef0b6aee20310fa86962449802118f10127a3dd9ce570"
    end
  end

  def install
    bin.install "pgbook"
  end

  test do
    assert_match "pgbook", shell_output("#{bin}/pgbook --version")
  end
end
