# Build Binary Guide

## Purpose

- Provide standard build commands that create platform-specific binaries

## Commands

Below are commands to create platform-specific binaries. If on Windows, it is recommended to run all commands from the Windows Terminal with the Git Bash (i.e. not *Cmd* and not *PowerShell*).

### Build a binary for the Windows platform (x86 architecture)

```bash
mkdir -p ./bin/eureka-cli-windows-x86
env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/eureka-cli-windows-x86 .
```

> The prefix `env GOOS=windows GOARCH=amd64` can be omitted if the shell instance cannot interpret it, with the binary being built with the default settings

### Build a binary for the Linux platform (x86 architecture)

```bash
mkdir -p ./bin/eureka-cli-linux-x86
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/eureka-cli-linux-x86 .
```

### Build a binary for the macOS platform (x86 or arm architectures)

- For x86

```bash
mkdir -p ./bin/eureka-cli-darwin-x86
env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/eureka-cli-darwin-x86 .
```

- For arm64

```bash
mkdir -p ./bin/eureka-cli-darwin-arm
env GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/eureka-cli-darwin-arm .
```

### Install binary for global use

```bash
go install
```

> `go install` must be run from the cloned project directory, i.e. from `./eureka-cli` folder

- Run the help command from any directory to test availability of the installed CLI

```bash
eureka-cli help -od
```
