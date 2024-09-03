# Build

## Purpose

- Cross-platform build commands
  
## Commands

```bash
mkdir -p ./bin/eureka-cli-windows-x86 ./bin/eureka-cli-darwin-x86 ./bin/eureka-cli-darwin-arm ./bin/eureka-cli-linux-x86 
env GOOS=windows GOARCH=amd64 go build -o ./bin/eureka-cli-windows-x86 .
env GOOS=darwin GOARCH=amd64 go build -o ./bin/eureka-cli-darwin-x86 .
env GOOS=darwin GOARCH=arm64 go build -o ./bin/eureka-cli-darwin-arm .
env GOOS=linux GOARCH=amd64 go build -o ./bin/eureka-cli-linux-x86 .
```
