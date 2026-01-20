#!/bin/bash

set -euo pipefail

export GOOS=windows
export GOARCH=amd64

VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PACKAGE="github.com/folio-org/eureka-setup/eureka-cli/cmd"
BIN="./bin/eureka-cli.exe"

echo "Building binary"
go build -ldflags \
  "-X '$PACKAGE.Version=$VERSION' -X '$PACKAGE.Commit=$COMMIT' -X '$PACKAGE.BuildDate=$BUILD_DATE'" \
  -o $BIN .

echo "Installing binary"
go install -ldflags \
  "-X '$PACKAGE.Version=$VERSION' -X '$PACKAGE.Commit=$COMMIT' -X '$PACKAGE.BuildDate=$BUILD_DATE'" .

echo "Overwritting config files"
eureka-cli help -od