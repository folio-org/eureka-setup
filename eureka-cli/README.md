# Eureka CLI

## Purpose

- A CLI to deploy local Eureka development environment

## Commands

### Prerequisites

- Install prerequisites:
  - [GO](<https://go.dev/doc/install>)
  - [Rancher Desktop](<https://rancherdesktop.io/>)
- Configure hosts:
  - Add `127.0.0.1 keycloak.eureka` entry to `/etc/hosts`
  - Add `127.0.0.1 kafka.eureka` entry to `/etc/hosts`
- Git clone:
  - [folio-kong](<https://github.com/folio-org/folio-kong>) into `./misc` folder using a `master`
  - [folio-keycloak](<https://github.com/folio-org/folio-keycloak>) into `./misc` folder using a `master`
- Monitor using below system components:
  - [Keycloak](<http://keycloak.eureka:8080>): admin:admin
  - [Vault](<http://localhost:8200>): Vault token from the container logs using `docker logs vault`
  - [Kafka](<http://localhost:9080>): No auth
  - [Kong](<http://localhost:8002>): No auth  

### Build a binary
  
```shell
mkdir -p ./bin
env GOOS=windows GOARCH=amd64 go build -o ./bin .
```

> See BUILD.md to build a platform-specific binary

### (Optional) Setup config in home folder

- This config will be used by default if `-c` or `--config` flag is not specified

```shell
./bin/eureka-cli.exe setup
```

### Deploy minimal platform application

- Use a specific config: `-c` or `--config`
- Enable debug: `-d` or `--debug`

```shell
./bin/eureka-cli.exe -c ./config.minimal.yaml deployApplication
```

- Undeploy using:

> ./bin/eureka-cli.exe -c ./config.minimal.yaml undeployApplication
