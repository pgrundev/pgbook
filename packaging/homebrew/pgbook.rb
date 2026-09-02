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
      sha256 "9a4bcc74c2281baa778cbd09939a1ea4a3b352e50a987f001182115b750a2da4"
    end
    on_intel do
      url "#{base}/pgbook_#{version}_darwin_amd64.tar.gz"
      sha256 "1a1c21b06e9cc8b867bbbd8cf3f7686a0bb68b60ec30d789cf799dd355630d77"
    end
  end

  on_linux do
    on_arm do
      url "#{base}/pgbook_#{version}_linux_arm64.tar.gz"
      sha256 "e7a0f80a7530347a37797196921fa2940dbd82c52d510e952f8ed77336835aa1"
    end
    on_intel do
      url "#{base}/pgbook_#{version}_linux_amd64.tar.gz"
      sha256 "f38b1971c7d670aa270fcdf1f9915d2662d0588876499d370799c95ed31987e5"
    end
  end

  def install
    bin.install "pgbook"
  end

  test do
    assert_match "pgbook", shell_output("#{bin}/pgbook --version")
  end
end
