# Eureka CLI

## Purpose

- A CLI to deploy local Eureka development environment

## Commands

### Prerequisites

- Install a compiler and a container daemon:
  - [GO](<https://go.dev/doc/install>): last development-tested version is `go1.22.4 windows/amd64`
  - [Rancher Desktop](<https://rancherdesktop.io/>): make sure to enable **dockerd (Moby)** container engine
- Configure hosts:
  - Add `127.0.0.1 keycloak.eureka` entry to `/etc/hosts`
  - Add `127.0.0.1 kafka.eureka` entry to `/etc/hosts`
- Git clone:
  - [folio-kong](<https://github.com/folio-org/folio-kong>) into `./misc` folder using a `master` branch
  - [folio-keycloak](<https://github.com/folio-org/folio-keycloak>) into `./misc` folder using a `master` branch
- Monitor using below system components:
  - [Keycloak](<http://keycloak.eureka:8080>): admin:admin
  - [Vault](<http://localhost:8200>): Find a Vault root token in the container logs using `docker logs vault` or use `getVaultRootToken` command
  - [Kafka](<http://localhost:9080>): No auth
  - [Kong](<http://localhost:8002>): No auth  

### Build a binary
  
```shell
mkdir -p ./bin
env GOOS=windows GOARCH=amd64 go build -o ./bin .
```

> See BUILD.md to build a platform-specific binary

### (Optional) Setup a default config in the home folder

- This config will be used by default if `-c` or `--config` flag is not specified

```shell
./bin/eureka-cli.exe setup
```

### Deploy a minimal platform application

- Use a specific config: `-c` or `--config`
- Enable debug: `-d` or `--debug`

```shell
./bin/eureka-cli.exe -c ./config.minimal.yaml deployApplication
```

- Undeploy using:

> ./bin/eureka-cli.exe -c ./config.minimal.yaml undeployApplication

- Test Keycloak authentication on the UI using the created `diku` realm and `diku-login-app` public client

> Open in browser `http://keycloak.eureka:8080/realms/diku/protocol/openid-connect/auth?client_id=diku-login-app&response_type=code&redirect_uri=http://localhost:3000&scope=openid`

### Troubleshooting

- Verify that all shell scripts located under `./misc` folder are saved using the **LF** (Line Feed) line break
