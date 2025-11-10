# Build guide

## Purpose

- Build commands to create multiple platform-specific binaries

## Commands

- Build four different binaries under `./bin` folder for the Windows, Linux and macOS x86 and ARM architectures

```bash
mkdir -p ./bin/eureka-cli-windows-x86 \
  ./bin/eureka-cli-linux-x86 \
  ./bin/eureka-cli-darwin-x86 \
  ./bin/eureka-cli-darwin-arm

env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/eureka-cli-windows-x86 .
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/eureka-cli-linux-x86 .
env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/eureka-cli-darwin-x86 .
env GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/eureka-cli-darwin-arm .
```
