# Eureka CLI

## Purpose

- A CLI to deploy local Eureka development environment

## Commands

### Prerequisites

- Install a compiler and a container daemon:
  - [GO](<https://go.dev/doc/install>): last development-tested version is `go1.22.4 windows/amd64`
  - [Rancher Desktop](<https://rancherdesktop.io/>): latest supported version is `v1.16.0` (make sure to enable **dockerd (Moby)** container engine)
- Configure hosts:
  - Add `127.0.0.1 keycloak.eureka` entry to `/etc/hosts`
  - Add `127.0.0.1 kafka.eureka` entry to `/etc/hosts`
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

#### Using Public DockerHub image registries (folioci & folioorg namespaces)

- Use a specific config: `-c` or `--config`
- Enable debug: `-d` or `--debug`

```shell
./bin/eureka-cli.exe -c ./config.minimal.yaml deployApplication
```

#### Using Private AWS ECR image registry

- Use a specific config: `-c` or `--config`
- Enable debug: `-d` or `--debug`

```shell
AWS_SDK_LOAD_CONFIG=true ./bin/eureka-cli.exe -c ./config.minimal.yaml deployApplication
```

> See AWS_CLI_PREPARATION.md to prepare AWS CLI beforehand

- Undeploy using:

> ./bin/eureka-cli.exe -c ./config.minimal.yaml undeployApplication

### Use the environment

- Test Keycloak authentication on the UI using the created `diku` realm and `diku-application` public client

> Open in browser `http://keycloak.eureka:8080/realms/diku/protocol/openid-connect/auth?client_id=diku-application&response_type=code&redirect_uri=http://localhost:3000&scope=openid`

- Gateway is available at `localhost:8000` or `api-gateway.eureka:8000`
- Login and get a token:

```shell
curl --request POST \
  --url localhost:8000/authn/login-with-expiry \
  --header 'Content-Type: application/json' \
  --header 'X-Okapi-Tenant: diku' \
  --data '{"username":"diku_admin","password": "admin"}' \
  --verbose
```

### Use AWS ECR

To use AWS ECR as your container registry rather than the public folio dockerhub, set `AWS_ECR_FOLIO_REPO` in your environment. When this env variable is defined it is assumed that this repo is private and you have also defined credentials in your environment.

### Troubleshooting

- Verify that all shell scripts located under `./misc` folder are saved using the **LF** (Line Feed) line break
- If you get a **SIGKILL** when trying to build the stripes container configure Rancher Desktop or other docker env with more RAM
- If health checks are failing make sure localhost is mapped to host.docker.internal in your `/etc/hosts` file
- If using Rancher Desktop on a system that also uses Docker Desktop, you may need to do set `DOCKER_HOST` in your env to where `docker.sock` is
