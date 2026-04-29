#!/bin/bash

# Script to build and install the eureka-cli binary for Windows (amd64)
# Injects version, commit hash, and build date via ldflags
# Usage: ./build_cli.sh

set -euo pipefail

# Set build target platform
export GOOS=windows
export GOARCH=amd64

# Resolve version metadata from git
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PACKAGE="github.com/folio-org/eureka-setup/eureka-cli/cmd"
BIN="./bin/eureka-cli.exe"

# Build the binary
echo "Building binary"
go build -ldflags \
  "-X '$PACKAGE.Version=$VERSION' -X '$PACKAGE.Commit=$COMMIT' -X '$PACKAGE.BuildDate=$BUILD_DATE'" \
  -o $BIN .

# Install to GOPATH/bin
echo "Installing binary"
go install -ldflags \
  "-X '$PACKAGE.Version=$VERSION' -X '$PACKAGE.Commit=$COMMIT' -X '$PACKAGE.BuildDate=$BUILD_DATE'" .

# Regenerate default config files
echo "Overwritting config files"
eureka-cli help -od