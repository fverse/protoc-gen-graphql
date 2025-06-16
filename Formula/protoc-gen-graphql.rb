# Homebrew formula for protoc-gen-graphql
# To use this formula, create a tap repository: homebrew-tap
# Then users can install with: brew install fverse/tap/protoc-gen-graphql

class ProtocGenGraphql < Formula
  desc "Protoc plugin to generate GraphQL schema from .proto files"
  homepage "https://github.com/fverse/protoc-graphql"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/fverse/protoc-graphql/releases/download/v#{version}/protoc-gen-graphql-darwin-arm64"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"

      def install
        bin.install "protoc-gen-graphql-darwin-arm64" => "protoc-gen-graphql"
      end
    end

    on_intel do
      url "https://github.com/fverse/protoc-graphql/releases/download/v#{version}/protoc-gen-graphql-darwin-amd64"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"

      def install
        bin.install "protoc-gen-graphql-darwin-amd64" => "protoc-gen-graphql"
      end
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/fverse/protoc-graphql/releases/download/v#{version}/protoc-gen-graphql-linux-arm64"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"

      def install
        bin.install "protoc-gen-graphql-linux-arm64" => "protoc-gen-graphql"
      end
    end

    on_intel do
      url "https://github.com/fverse/protoc-graphql/releases/download/v#{version}/protoc-gen-graphql-linux-amd64"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"

      def install
        bin.install "protoc-gen-graphql-linux-amd64" => "protoc-gen-graphql"
      end
    end
  end

  test do
    system "#{bin}/protoc-gen-graphql", "--version"
  end
end
